# Implementing a New Command

Adding a command is a five-step process. Read this top-to-bottom before touching any file.

## Step 1 — Declare the command in `commands.yml`

Every command is declared in `temporalcloudcli/commands.yml`. The `name` field is the full command path. Key fields:

```yaml
- name: cloud <subcommand>          # full dotted path; determines the cobra tree
  summary: One-line description
  description: |
    Longer description shown in --help.
    Include a usage example block.
  has-init: false                   # true only for the root `cloud` command
  option-sets:
    - client                        # adds --api-key / --server flags
  options:
    - name: my-flag
      type: string                  # string | bool | int
      required: true
      short: f                      # optional single-char shorthand
      description: |
        Flag description.
  docs:                             # optional; used for generated docs site
    keywords: [...]
    description-header: ...
    tags: [...]
```

**Reusable option-sets** (defined at the bottom of `commands.yml`):
- `client` — adds `--api-key` and `--server` (hidden); include on every command that calls the API
- `diff` — adds `--verbose-diff`; include **only** on `apply` and `edit` commands that show a full spec diff before applying. Do NOT include on targeted mutation commands like `set` or `update` — those don't show a diff.
- `common` — external package options; only on the root `cloud` command

**Read-only commands** do not need `--async-operation-id` or `--resource-version`.

## Step 2 — Regenerate `commands.gen.go`

```bash
make gen
```

This clones the `temporalio/cli` repo, builds the `gen-commands` tool, and regenerates `temporalcloudcli/commands.gen.go`. **Never edit `commands.gen.go` by hand.**

The generator derives the Go struct name from the command path:
- `cloud namespace get` → `CloudNamespaceGetCommand`
- `cloud whoami` → `CloudWhoamiCommand`

Each generated struct embeds the option-set structs (e.g. `ClientOptions`) and a `cobra.Command`. The generator wires `s.Command.Run` to call `s.run(cctx, args)`, which you implement in Step 3.

## Step 3 — Implement `commands.<name>.go`

Create `temporalcloudcli/commands.<name>.go` and implement the application logic function(s) plus the `run` wiring method(s).

Each command gets an exported `XxxParams` struct with data fields and injectable dependency fields, and an exported function that contains all application logic. The `run` method only builds the cloud client and calls the exported function.

**Read command skeleton:**
```go
package temporalcloudcli

import (
    "context"

    cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
    "github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type GetFooParams struct {
    Namespace string

    Cloud   cloudservice.CloudServiceClient
    Printer *printer.Printer
}

func GetFoo(ctx context.Context, params GetFooParams) error {
    res, err := params.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: params.Namespace})
    if err != nil {
        return err
    }
    return params.Printer.PrintStructured(res.Namespace.Spec, printer.StructuredOptions{})
}

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
```

**Mutation command skeleton:**
```go
type SetFooParams struct {
    Namespace        string
    Value            string
    ResourceVersion  string
    AsyncOperationID string

    Cloud            cloudservice.CloudServiceClient
    Prompter         Prompter
    OperationHandler AsyncOperationHandler
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

See `temporalcloudcli/commands.namespace.retention.go` for the canonical example.

**Printer methods:**
- `PrintStructured(v, opts)` — flat data or raw responses with no clear spec boundary
- `PrintResource(v, opts)` — use for any get command that returns a resource or sub-resource. Pass a struct with top-level identity fields (e.g. `Namespace string`) and a `Spec` field for the content. In text mode this renders the identity at the top and spec fields indented under `Spec:`; in JSON mode it falls back to `PrintStructured`.
- `PrintResourceList(list, printOpts, tableOpts)` — slice of resources as a table

**Common helpers in `common.go`:**
- `loadJSONSpec(spec string)` — loads JSON from inline string or `@file` path
- `runEditorForJSONEditForProtos(current, target)` — opens `$EDITOR` for interactive editing
- `promptApplyResource(cctx, old, new, verbose)` — shows diff and asks for confirmation
- `PollAsyncOperation(cctx, client, opID, resourceID)` — polls until terminal state
- `isNotFoundErr(err)` / `isNothingChangedErr(idempotent, err)` — gRPC error helpers

After creating the file, run `git add temporalcloudcli/commands.<name>.go`.

## Step 4 — Write tests in `commands.<name>_test.go`

See `.claude/rules/testing.md` for testing patterns and the integration test template.

After creating the file, run `git add temporalcloudcli/commands.<name>_test.go`.

## Step 5 — Verify

```bash
make all   # gen + build + test (unit tests only; integration tests need -tags=integration)
```

---

## Best Practices

### Targeted `set` Commands on Map Fields Are Additive

When a `set` command targets a map field (e.g. namespace permissions), changes must be **merged into the existing map** rather than replacing it wholesale. Provide empty-value syntax to remove an entry (e.g. `namespace=`).

```go
// Good: merge changes into existing map; empty permission = remove
func applyNamespaceAccessChanges(existing map[string]*identityv1.NamespaceAccess, changes []string) (map[string]*identityv1.NamespaceAccess, error) {
    result := make(map[string]*identityv1.NamespaceAccess, len(existing))
    for k, v := range existing {
        result[k] = v
    }
    for _, a := range changes {
        ns, perm, _ := strings.Cut(a, "=")
        if perm == "" {
            delete(result, ns)
        } else {
            result[ns] = &identityv1.NamespaceAccess{Permission: permissionNames[perm]}
        }
    }
    ...
}
```

**Validate inputs before any API call** using a dry-run (nil existing map) so format errors are returned immediately without network round-trips:

```go
// Validate before API calls — applyNamespaceAccessChanges(nil, ...) is safe (range over nil map is a no-op)
if _, err := applyNamespaceAccessChanges(nil, params.NamespaceAccesses); err != nil {
    return err
}
user, err := resolveUser(...)  // only called after inputs are known-good
...
accesses, _ := applyNamespaceAccessChanges(user.Spec.Access.NamespaceAccesses, params.NamespaceAccesses)
```

### Server-side vs Client-side Responsibility

**Don't implement client-side validation that the server handles:**
- Duplicate detection (e.g. adding the same cert filter twice) — let the server return an error
- Complex business rules and constraints — trust the server's validation

**Do implement client-side validation for:**
- Required fields before making API calls
- Flag combinations that don't make sense (e.g. both `--file` and `--data` provided)
- Input format validation (e.g. valid base64, file exists)

### Prefer Standard Library Over Custom Helpers

Use Go standard library functions instead of writing custom helpers.

```go
// Good: simple loop with slices.ContainsFunc
var newFilters []*namespacev1.CertificateFilterSpec
for _, existing := range existingFilters {
    shouldRemove := slices.ContainsFunc(params.Filters, func(toRemove *namespacev1.CertificateFilterSpec) bool {
        return certFiltersEqual(existing, toRemove)
    })
    if shouldRemove {
        continue
    }
    newFilters = append(newFilters, existing)
}
```

### Mutation Commands Must Prompt for Confirmation

All create and delete commands must prompt the user before making changes:

```go
yes, err := cctx.promptYes("Create (y/yes)?", cctx.RootCommand.AutoConfirm)
if err != nil {
    return err
}
if !yes {
    return errors.New("Aborting create.")
}
```

Update commands typically don't need prompts (user explicitly provides new values).

### Code Simplicity

Keep code simple and readable over clever abstractions.
- **Good:** Simple loop with `continue` for filtering
- **Bad:** Complex `slices.DeleteFunc` with nested logic

If logic is hard to understand at a glance, it's probably too complex.
