# Agents

## Overview

This repository contains the CLI Plugin for Temporal Cloud (`temporal-cloud`).

## Project Structure

- `cmd/temporal-cloud/` - Main entry point for the CLI
- `temporalcloudcli/` - Core CLI implementation
  - `cloud.go` - The cloud specific client implementation including auth and clients for cloud ops api service
  - `commands.go` - The command context thats passed around and the root command implementation
  - `commands.gen.go` - Generated command code, do not edit
  - `commands.login.go` - Login command implementation
  - `commands.namespace.go` - Namespace command implementation
  - `commands.yml` - Command configuration, used to generate commands.gen.go
  - `common.go` - Shared utilities, constants, and types used across the CLI
  - `namespace.go` - The namespace client implementation
  - `internal/printer/` - Output formatting utilities

## Building

```bash
make
```

## Usage

The plugin is meant to be an extension to the Temporal CLI. After building, the plugin binary should be copied to some place in your `PATH` and renamed to `temporal-cloud`. After that, you can use it as follows:

```bash
temporal cloud <command> [flags]
```


### Use Anchor Comments
Add specially formatted comments throughout the codebase, where appropriate, for yourself as inline knowledge that can be easily `grep`ped for.

- Use `AIDEV-NOTE:`, `AIDEV-TODO:`, or `AIDEV-QUESTION:` (all-caps prefix) for comments aimed at AI and developers.
- **Important:** Before scanning files, always first try to **grep for existing anchors** `AIDEV-*` in relevant subdirectories.
- **Update relevant anchors** when modifying associated code.
- **Do not remove `AIDEV-NOTE`s** without explicit human instruction.
- Make sure to add relevant anchor comments, whenever a file or piece of code is:
    * too complex, or
    * very important, or
    * confusing, or
    * could have a bug

### General Code Style Guide
- Keep all `const`, `type`, and `var` at the top of the file.
- Group multiple `const`, `type` and `var` declarations together. e.g.
```go
type (
    CompanyName string 
    CompanyID   string
)
```

- Limit lines to 120 characters.
- Avoid stuttering in names. When a name's context (like the enclosing type or package) already implies part of the name, do not repeat it.

**Bad:**
```go
type User struct {
    UserName string
    UserID   int
}

func (u User) GetUserName() string {
    return u.UserName
}
```

**Good:**
```go
type User struct {
    Name string
    ID   int
}

func (u User) GetName() string {
    return u.Name
}
```

- Add an EOF newline for new files you create.

## AI Communication Guidelines

### Plan Before Implementing
Before providing your final implementation, use <implementation_planning> tags to:

1. Break down the feature into smaller, manageable tasks.
2. Consider potential challenges for each task and how to address them.
3. Provide a high-level outline of the code structure, including function names and their purposes.
4. List specific test cases you plan to implement.
5. State which error handling approaches you will use for different scenarios.
6. Discuss the trade-offs inherent in your design decisions, including:
    * Performance trade-offs
    * Scalability trade-offs
    * Complexity trade-offs
    * Security trade-offs
7. Reason about the failure modes of your design. How does it handle crashes? A 10x increase in load?

### Code Explanation Format
When explaining code mechanics with verb constructions (e.g., "calls", "sends", "is processed"), follow this format:
1. **Bold** the verb in the sentence
2. Immediately after the sentence, provide a clickable code citation showing the exact line where that action occurs
3. Use the format: `[filepath:line](filepath:line)` for clickable code citations

**Example:**
When the account entity module initializes, it **registers** activities with the Temporal worker using prefixed names:

[entities/account/internal/fx.go:22](entities/account/internal/fx.go:22)
```go
w.RegisterActivityWithOptions(a, activity.RegisterOptions{Name: ActivityNamePrefix})
```

## What AI _Must_ Do

1. Use environment variables instead of committing secrets
2. Always ask instead of assuming business logic
3. Add to AIDEV comments or ask instead of removing them
4. Stay focused on the task at hand. When in doubt about whether a change is related to the task, ask.

## Development Process
1. Follow the style guide rules above.
2. When a new `.go` file is created, git add it.
3. Before saying that you're "done", ensure the following checks pass:
   - `make all`

## Command Generation Guidelines
1. All commands should have an `--async-operation-id` for mutations. For example, anything that is a create, update, or delete, or any command specific patch commands.
2. All top level spec based commands (for example, namespace) that include a mutation should have an `apply` and `edit` subcommand. The only exception to this will be apikey.
3. All mutation commands should have a `--resource-version` flag that allows the user to specify a resource version for optimistic locking.

## How to Implement a New Command

Adding a command is a four-step process. Read this section top-to-bottom before touching any file.

### Layered Architecture Pattern

**IMPORTANT:** All resource commands (namespace, apikey, etc.) follow a two-layer architecture for separation of concerns and testability:

#### Layer 1: Command Layer (`temporalcloudcli/commands.<resource>.go`)
Responsibilities:
- Parse and validate user input (flags, arguments)
- Handle user interactions (prompts, confirmations)
- Call application layer methods
- Format and print output using printer utilities
- Should NOT contain business logic or direct API calls (except for simple one-off commands)

#### Layer 2: Application Layer (`internal/<resource>/client.go`)
Responsibilities:
- Contains business logic and domain operations
- Communicates with the cloud API
- Defines mockable interfaces (e.g., `CloudService`) for all external dependencies
- Returns domain objects, not command-specific types
- Should be testable without CLI dependencies

#### Testing Strategy

**Application Layer Tests:**
- Unit tests with mocks generated from interfaces
- Tests live in `internal/<resource>/*_test.go`
- Mock the `CloudService` interface to test business logic in isolation

**Command Layer Tests:**
- Integration tests in `temporalcloudcli/commands.<resource>_test.go`
- Use real API calls (require live credentials)
- Can inject mocked application client via `CommandContext.NamespaceClient` for unit testing command logic

#### Setting Up Mocks

When creating a new resource client with mockable interfaces:

1. **Define the interface in the application layer:**
```go
// internal/myresource/client.go
type CloudService interface {
    GetResource(ctx context.Context, req *cloudservice.GetResourceRequest, opts ...grpc.CallOption) (*cloudservice.GetResourceResponse, error)
    // ... other methods
}

type Client struct {
    Cloud CloudService
}

func NewClient(cloudClient cloudservice.CloudServiceClient) *Client {
    return &Client{Cloud: cloudClient}
}
```

2. **Update `.mockery.yml` to generate mocks:**
```yaml
packages:
  github.com/temporalio/cloud-cli/internal/myresource:
    interfaces:
      CloudService:
```

3. **Regenerate mocks:**
```bash
make mocks
```

This generates `internal/myresource/mock/mock.go` with a `MockCloudService` type.

4. **Add the client interface to command context:**

If the resource needs to be accessible across multiple commands, add it to `temporalcloudcli/commands.go`:
```go
type MyResourceClient interface {
    GetResource(context.Context, string) (*resourcev1.Resource, error)
    // ... other methods from internal/myresource/client.go
}

type CommandContext struct {
    // ... existing fields
    MyResourceClient MyResourceClient  // For testing, can inject mock
}
```

Then add it to `.mockery.yml`:
```yaml
packages:
  github.com/temporalio/cloud-cli/temporalcloudcli:
    interfaces:
      MyResourceClient:
```

And run `make mocks` again.

#### Example Implementation

**Command Layer (`temporalcloudcli/commands.namespace.cert_ca.go`):**
```go
func (c *CloudNamespaceCertCaListCommand) run(cctx *CommandContext, _ []string) error {
    // 1. Get application client (real or mock for testing)
    namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
    if err != nil {
        return err
    }

    // 2. Validate input
    if c.Namespace == "" {
        return errors.New("namespace is required")
    }

    // 3. Call application layer
    certs, err := namespaceClient.ListCACerts(cctx.Context, c.Namespace)
    if err != nil {
        return err
    }

    // 4. Format and print output
    return cctx.Printer.PrintStructured(certs, printer.StructuredOptions{})
}

// Helper to get client (allows mock injection for testing)
func getNamespaceClient(cctx *CommandContext, opts ClientOptions) (NamespaceClient, error) {
    if cctx.NamespaceClient != nil {
        return cctx.NamespaceClient, nil  // Use mock if provided
    }
    cloudClient, err := cctx.BuildCloudClient(opts)
    if err != nil {
        return nil, err
    }
    return namespace.NewClient(cloudClient.CloudService()), nil
}
```

**Application Layer (`internal/namespace/client.go`):**
```go
// CloudService defines mockable interface for cloud operations
type CloudService interface {
    GetNamespace(ctx context.Context, req *cloudservice.GetNamespaceRequest, opts ...grpc.CallOption) (*cloudservice.GetNamespaceResponse, error)
    UpdateNamespace(ctx context.Context, req *cloudservice.UpdateNamespaceRequest, opts ...grpc.CallOption) (*cloudservice.UpdateNamespaceResponse, error)
}

type Client struct {
    Cloud CloudService  // Mockable interface, not concrete type
}

func NewClient(cloudClient cloudservice.CloudServiceClient) *Client {
    return &Client{Cloud: cloudClient}
}

func (c *Client) ListCACerts(ctx context.Context, namespace string) ([]cert.CACert, error) {
    // Business logic: fetch namespace, parse certs, validate, etc.
    res, err := c.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{
        Namespace: namespace,
    })
    if err != nil {
        return nil, err
    }
    // ... parse and return certs
}
```

**Application Layer Unit Test (`internal/namespace/cert_test.go`):**
```go
func TestClient_ListCACerts(t *testing.T) {
    mockCloud := mock.NewMockCloudService(t)
    client := &Client{Cloud: mockCloud}

    mockCloud.EXPECT().GetNamespace(mock.Anything, mock.Anything).Return(
        &cloudservice.GetNamespaceResponse{/* ... */}, nil,
    )

    certs, err := client.ListCACerts(context.Background(), "my-namespace")
    require.NoError(t, err)
    require.NotEmpty(t, certs)
}
```

#### When to Use This Pattern

**Use the two-layer pattern when:**
- The resource has multiple operations (namespace, apikey, identity, etc.)
- Business logic is non-trivial and needs unit testing
- The same logic might be reused across multiple commands

**Skip the application layer for:**
- Simple one-off commands (like `whoami`)
- Commands with no business logic (direct API passthrough)

For simple commands, call `cloudClient.CloudService()` directly in the `run` method.

### Best Practices and Common Patterns

Follow these patterns to avoid rework and match existing code style:

#### Server-side vs Client-side Responsibility
- **Don't implement client-side validation that the server handles.** Examples:
  - Duplicate detection (e.g., adding the same cert filter twice) — let the server return an error
  - Complex business rules and constraints — trust the server's validation
- **Do implement client-side validation for:**
  - Required fields before making API calls
  - Flag combinations that don't make sense (e.g., both `--file` and `--data` provided)
  - Input format validation (e.g., valid base64, file exists)

#### Prefer Standard Library Over Custom Helpers
- Use Go standard library functions instead of writing custom helpers
- **Good:** Use `slices.ContainsFunc()` for filtering
- **Bad:** Write a custom `certFilterKey()` function to build map keys
- **Example from cert_filter.go:**
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

#### Choosing the Right Printer Method
- **`PrintResourceList`** — for list commands returning collections (most common for list operations)
  ```go
  return cctx.Printer.PrintResourceList(
      struct {
          CertificateFilters []*namespacev1.CertificateFilterSpec
      }{CertificateFilters: filters},
      printer.PrintResourceOptions{},
      printer.TableOptions{},
  )
  ```
- **`PrintStructured`** — for single objects or simple data structures
- **`PrintResource`** — for resources with metadata and spec fields

#### Mutation Commands Must Prompt for Confirmation
- **All create and delete commands** must prompt the user before making changes
- Use `cctx.promptYes()` with `cctx.RootCommand.AutoConfirm` to allow bypass
- **Pattern:**
  ```go
  yes, err := cctx.promptYes("Create (y/yes)?", cctx.RootCommand.AutoConfirm)
  if err != nil {
      return err
  }
  if !yes {
      return errors.New("Aborting create.")
  }
  ```
- Update commands typically don't need prompts (user explicitly provides new values)

#### Testing Patterns

**When to use table-driven tests:**
- Use table-driven tests when test cases follow the **same flow** with simple variations
- Good for: different flag combinations, different input values, same assertion pattern
- **Avoid complex branching** in table-driven tests (nested ifs, different setup per case)
- If you need conditional logic to handle different test cases, use individual test functions instead

**Application Layer Tests (`internal/<resource>/<feature>_test.go`):**
- Use **external test package** with `_test` suffix: `package namespace_test`
- Write **individual test functions** (most application logic tests have unique setup/assertions)
- Use **`proto.Equal()`** for comparing protobuf messages (not field-by-field comparison)
- **Look at existing tests** like `internal/namespace/cert_test.go` for the pattern
- Example naming: `TestClient_ListCertFilters_Success`, `TestClient_AddCertFilters_GetNamespaceError`

**Command Layer Tests (`temporalcloudcli/commands.<name>_test.go`):**
- Use **external test package**: `package temporalcloudcli_test`
- **Table-driven tests** work well for success cases with simple variations:
  - Example: async vs sync mode, different flag combinations, with/without resource version
  - All cases follow same pattern: setup command → run → assert result
  - See `TestCloudNamespaceCertCaCreateCommand_Success` in `cert_ca_test.go`
- **Individual test functions** for cases requiring different logic:
  - Validation errors (different error messages to check)
  - API errors (different mock setup)
  - Prompt scenarios (different stdin/AutoConfirm setup)
  - Idempotent behavior (different assertion logic for success vs error)
- **Look at existing tests** like `temporalcloudcli/commands.namespace.cert_ca_test.go` for the pattern
- Always set `RootCommand.AutoConfirm: true` in tests to bypass prompts (unless testing prompt behavior)
- Test prompt scenarios as individual functions:
  - User declining prompt (`AutoConfirm: false`, provide `"n\n"` as stdin)
  - JSON output without auto-confirm (should error)

#### Code Simplicity
- Keep code simple and readable over clever abstractions
- **Good:** Simple loop with `continue` for filtering
- **Bad:** Complex `slices.DeleteFunc` with nested logic
- If logic is hard to understand at a glance, it's probably too complex

### Step 1 — Declare the command in `commands.yml`

Every command is declared in `temporalcloudcli/commands.yml`. The name field is the full command path (e.g. `cloud whoami`, `cloud namespace get`). Key fields:

```yaml
- name: cloud <subcommand>          # full dotted path; determines the cobra tree
  summary: One-line description
  description: |
    Longer description shown in --help.
    Include a usage example block.
  has-init: false                   # true only for the root `cloud` command
  option-sets:
    - client                        # adds --api-key / --server flags
  options:                          # command-specific flags
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
- `client` — adds `--api-key` and `--server` (hidden); include this on every command that calls the API
- `diff` — adds `--verbose-diff`; include on mutation commands that show a diff before applying
- `common` — external package options; only on the root `cloud` command

**Read-only commands** do not need `--async-operation-id` or `--resource-version`.

### Step 2 — Regenerate `commands.gen.go`

```bash
make gen
```

This clones the `temporalio/cli` repo, builds the `gen-commands` tool, and regenerates `temporalcloudcli/commands.gen.go`. **Never edit `commands.gen.go` by hand.**

The generator derives the Go struct name from the command path:
`cloud namespace get` → `CloudNamespaceGetCommand`
`cloud whoami`        → `CloudWhoamiCommand`

Each generated struct embeds the option-set structs (e.g. `ClientOptions`) and a `cobra.Command`. The generator wires `s.Command.Run` to call `s.run(cctx, args)`, which you implement in Step 3.

### Step 3 — Implement `commands.<name>.go`

Create `temporalcloudcli/commands.<name>.go` and implement the `run` method on the generated struct.

**Standard read command skeleton:**

```go
package temporalcloudcli

import (
    "go.temporal.io/cloud-sdk/api/cloudservice/v1"
    "github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudMyCommand) run(cctx *CommandContext, _ []string) error {
    cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
    if err != nil {
        return err
    }
    res, err := cloudClient.CloudService().SomeRPCMethod(cctx.Context, &cloudservice.SomeRequest{})
    if err != nil {
        return err
    }
    return cctx.Printer.PrintStructured(res, printer.StructuredOptions{})
}
```

**Printer methods:**
- `PrintStructured(v, opts)` — single object; JSON or text table
- `PrintResource(proto, opts)` — standard resource with metadata + spec fields
- `PrintResourceList(list, printOpts, tableOpts)` — slice of resources as a table

**Resource-layer client (for complex resources):** If the command belongs to a resource with multiple operations (like namespace), add a `<resource>.go` file with a thin client wrapper over `cloudclient.Client` — see `namespace.go` as the canonical example. For simple one-off calls (like `whoami`), call `cloudClient.CloudService()` directly in the `run` method.

**Common helpers in `common.go`:**
- `loadJSONSpec(spec string)` — loads JSON from inline string or `@file` path
- `runEditorForJSONEditForProtos(current, target)` — opens `$EDITOR` for interactive editing
- `promptApplyResource(cctx, old, new, verbose)` — shows diff and asks for confirmation
- `PollAsyncOperation(cctx, client, opID, resourceID)` — polls until terminal state
- `isNotFoundErr(err)` / `isNothingChangedErr(idempotent, err)` — gRPC error helpers

After creating the file, run `git add temporalcloudcli/commands.<name>.go`.

### Step 4 — Write integration tests in `commands.<name>_test.go`

All tests are integration tests (require a live API key + server). Follow this template:

```go
//go:build integration
// +build integration

package temporalcloudcli_test

import (
    "fmt"
    "io"

    "google.golang.org/protobuf/encoding/protojson"
    "go.temporal.io/cloud-sdk/api/cloudservice/v1"
)

func (s *SharedServerSuite) TestMyCommand() {
    res := s.Execute(
        "<subcommand>",
        fmt.Sprintf("--server=%s", s.server),
        "-o=json",
    )
    s.Suite.Require().NoError(res.Err)

    buf, err := io.ReadAll(&res.Stdout)
    s.Suite.Require().NoError(err)

    out := &cloudservice.MyResponse{}
    err = protojson.Unmarshal(buf, out)
    s.Suite.Require().NoError(err)

    // assert fields on out
}
```

Key test infrastructure (defined in `commands_test.go`):
- `s.Execute(args...)` — runs the CLI with the given args, captures stdout/stderr
- `s.getCloudClient()` — returns a raw `*cloudclient.Client` for direct SDK calls
- `s.pollAsyncOperation(client, opID)` — polls an async op to completion
- `s.generateRandomID()` — returns a random 10-char string for unique resource names
- `s.ContainsOnSameLine(text, pieces...)` — asserts pieces appear in order on one line

After creating the file, run `git add temporalcloudcli/commands.<name>_test.go`.

### Step 5 — Verify

```bash
make all   # gen + build + test (unit tests only; integration tests need -tags=integration)
```