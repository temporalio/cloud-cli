# Architecture

## Command Design Rules

1. All mutation commands (create, update, delete, patch) must have an `--async-operation-id` flag.
2. All top-level spec-based commands (e.g. namespace) that include a mutation must have `apply` and `edit` subcommands. The only exception is `apikey`.
3. All mutation commands must have a `--resource-version` flag for optimistic locking.

## Function-Based Architecture

All resource commands follow a function-based architecture for separation of concerns and testability.

### Pattern

Each command's application logic lives in an **exported function** in the `temporalcloudcli` package. The function takes a `context.Context` and an exported `XxxParams` struct containing both data fields and injectable dependencies. The `run` method only builds the cloud client and wires everything together.

**Application logic function** (`temporalcloudcli/commands.<resource>.go`):
- Exported function (e.g., `GetRetention`, `SetRetention`) owns the full operation
- Takes a `XxxParams` struct with data fields (e.g. `Namespace`, `RetentionDays`) and dependency fields (`Cloud`, `Prompter`, `OperationHandler`)
- Calls the gRPC API via the `Cloud` field (a `cloudservice.CloudServiceClient` interface)
- Testable by calling the function directly with mock dependencies

**`run` method wiring** (on the generated command struct):
- Builds the cloud client via `cctx.BuildCloudClient`
- Constructs the `XxxParams` struct from command flags and wired dependencies
- Calls the exported application logic function

**Example:**
```go
// XxxParams structs — exported for testability
type (
    GetFooParams struct {
        Namespace string

        Cloud   cloudservice.CloudServiceClient
        Printer *printer.Printer
    }

    SetFooParams struct {
        Namespace        string
        Value            string
        ResourceVersion  string
        AsyncOperationID string

        Cloud            cloudservice.CloudServiceClient
        Prompter         Prompter
        OperationHandler AsyncOperationHandler
    }
)

// Exported application logic functions
func GetFoo(ctx context.Context, params GetFooParams) error {
    res, err := params.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: params.Namespace})
    if err != nil {
        return err
    }
    return params.Printer.PrintStructured(res.Namespace.Spec, printer.StructuredOptions{})
}

func SetFoo(ctx context.Context, params SetFooParams) error {
    res, err := params.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: params.Namespace})
    if err != nil {
        return err
    }
    ns := res.Namespace
    newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
    // mutate newSpec...

    if err := params.Prompter.PromptApply(ns.Spec, newSpec, false); err != nil {
        return err
    }

    rv := ns.ResourceVersion
    if params.ResourceVersion != "" {
        rv = params.ResourceVersion
    }
    updateNamespace := runAsyncOperation(params.Cloud.UpdateNamespace, params.OperationHandler)
    return updateNamespace(ctx, &cloudservice.UpdateNamespaceRequest{
        Namespace:        params.Namespace,
        Spec:             newSpec,
        ResourceVersion:  rv,
        AsyncOperationId: params.AsyncOperationID,
    })
}

// run methods — wiring only
func (c *CloudNamespaceFooGetCommand) run(cctx *CommandContext, _ []string) error {
    cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
    if err != nil {
        return err
    }
    return GetFoo(cctx.Context, GetFooParams{
        Namespace: c.Namespace,
        Cloud:     cloudClient.CloudService(),
        Printer:   cctx.Printer,
    })
}

func (c *CloudNamespaceFooSetCommand) run(cctx *CommandContext, _ []string) error {
    cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
    if err != nil {
        return err
    }
    return SetFoo(cctx.Context, SetFooParams{
        Namespace:        c.Namespace,
        Value:            c.Value,
        ResourceVersion:  c.ResourceVersion,
        AsyncOperationID: c.AsyncOperationId,
        Cloud:            cloudClient.CloudService(),
        Prompter:         newPrompter(cctx),
        OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions),
    })
}
```

### Key Interfaces

- `cloudservice.CloudServiceClient` (from `go.temporal.io/cloud-sdk/api/cloudservice/v1`) — the gRPC client interface; mock with `csmock.NewMockCloudServiceClient(t)`
- `Prompter` (from `temporalcloudcli/common.go`) — shows diffs and prompts for confirmation; mock with `cmdmock.NewMockPrompter(t)`
- `AsyncOperationHandler` (from `temporalcloudcli/common.go`) — handles async op lifecycle; mock with `cmdmock.NewMockAsyncOperationHandler(t)`

### When to Skip This Pattern

For simple commands with no business logic (e.g. `whoami`), call `cloudClient.CloudService()` directly in the `run` method — no exported function or params struct needed.

## Setting Up Mocks

The mockery-generated mocks live in:
- `internal/cloudservice/mock/` — provides `MockCloudServiceClient` for the gRPC interface
- `temporalcloudcli/mock/` — provides `MockPrompter` and `MockAsyncOperationHandler`

To regenerate all mocks after interface changes:
```bash
make mocks
```

To add a new interface to an existing package's mock, update `.mockery.yml` and run `make mocks`.
