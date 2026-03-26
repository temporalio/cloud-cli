# Architecture

## Command Design Rules

1. All mutation commands (create, update, delete, patch) must have an `--async-operation-id` flag.
2. All top-level spec-based commands (e.g. namespace) that include a mutation must have `apply` and `edit` subcommands. The only exception is `apikey`.
3. All mutation commands must have a `--resource-version` flag for optimistic locking.

## Command Architecture

All commands implement application logic directly in the `run` method. Dependencies are injected via `CommandContext` hooks, which makes `run` testable without exported functions or params structs.

### Pattern

**Client acquisition** — use `cctx.GetCloudClient(opts)` instead of the two-step `cctx.BuildCloudClient(opts).CloudService()`. It returns a `cloudservice.CloudServiceClient` and respects the test override hook (`cctx.getCloudClientOverride`).

**Async operations** — use `cctx.GetPoller(client, asyncOpts)` which returns an `async.Poller`. Call the appropriate method based on the operation type:
- `poller.HandleCreateAsyncOperationResponse(ctx, resp, err)` — for creates; handles `AlreadyExists` if idempotent
- `poller.HandleUpdateOperation(ctx, resp, err)` — for updates; handles "nothing to change" if idempotent
- `poller.HandleDeleteOperation(ctx, resp, err)` — for deletes; handles `NotFound` if idempotent

The poller polls `GetAsyncOperation` in a loop until the op reaches a terminal state, then prints the final response (with the terminal op state embedded). Respects the test override hook (`cctx.getAsyncPollerOverride`).

**Read command skeleton:**
```go
func (c *CloudNamespaceFooGetCommand) run(cctx *CommandContext, _ []string) error {
    client, err := cctx.GetCloudClient(c.ClientOptions)
    if err != nil {
        return err
    }
    res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
    if err != nil {
        return err
    }
    return cctx.Printer.PrintResource(res.Namespace, printer.PrintResourceOptions{})
}
```

**Mutation command skeleton:**
```go
func (c *CloudNamespaceFooSetCommand) run(cctx *CommandContext, _ []string) error {
    client, err := cctx.GetCloudClient(c.ClientOptions)
    if err != nil {
        return err
    }
    res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
    if err != nil {
        return err
    }
    ns := res.Namespace
    newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
    // mutate newSpec...

    rv := ns.ResourceVersion
    if c.ResourceVersion != "" {
        rv = c.ResourceVersion
    }
    resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
        Namespace:        c.Namespace,
        Spec:             newSpec,
        ResourceVersion:  rv,
        AsyncOperationId: c.AsyncOperationId,
    })
    return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}
```

See `temporalcloudcli/commands.apikey.go` (`CloudApikeyGetCommand` and `CloudApikeyCreateForMeCommand`) for the canonical examples.

### Key Interfaces

- `cloudservice.CloudServiceClient` (from `go.temporal.io/cloud-sdk/api/cloudservice/v1`) — the gRPC client interface; injected via `cctx.GetCloudClient`; overridden in tests via `cctx.getCloudClientOverride`
- `async.Poller` (from `temporalcloudcli/async`) — handles async op lifecycle and polling; injected via `cctx.GetPoller`; overridden in tests via `cctx.getAsyncPollerOverride`
- `cliprompter.Prompter` (from `temporalcloudcli/prompter`) — shows diffs and prompts for confirmation; use `cctx.GetPrompter()`; overridden in tests via `cctx.getPrompterOverride`
- `editor.Editor` (from `temporalcloudcli/editor`) — opens `$EDITOR` for interactive proto editing; use `cctx.GetEditor()`; overridden in tests via `cctx.getEditorOverride`

### Legacy Pattern (being migrated)

Older commands (e.g. `commands.namespace.retention.go`) use an exported `XxxParams` struct and exported application logic function, with `cctx.BuildCloudClient` and `NewOperationHandler`/`AsyncOperationHandler`. When touching these files, migrate them to the new pattern. Do not write new commands using the old pattern.

## Setting Up Mocks

The mockery-generated mocks live in:
- `internal/cloudservice/mock/` — provides `MockCloudServiceClient` for the gRPC interface
- `temporalcloudcli/async/mock/` — provides `MockPoller` for the `async.Poller` interface
- `temporalcloudcli/prompter/mock/` — provides `MockPrompter` for the `cliprompter.Prompter` interface
- `temporalcloudcli/editor/mock/` — provides `MockEditor` for the `editor.Editor` interface

To regenerate all mocks after interface changes:
```bash
make mocks
```

To add a new interface to an existing package's mock, update `.mockery.yml` and run `make mocks`.
