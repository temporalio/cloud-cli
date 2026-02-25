# Architecture

## Command Design Rules

1. All mutation commands (create, update, delete, patch) must have an `--async-operation-id` flag.
2. All top-level spec-based commands (e.g. namespace) that include a mutation must have `apply` and `edit` subcommands. The only exception is `apikey`.
3. All mutation commands must have a `--resource-version` flag for optimistic locking.

## Two-Layer Architecture

All resource commands follow a two-layer architecture for separation of concerns and testability.

### Layer 1 — Command Layer (`temporalcloudcli/commands.<resource>.go`)

- Parse and validate user input (flags, arguments)
- Handle user interactions (prompts, confirmations)
- Call application layer methods
- Format and print output using printer utilities
- Must NOT contain business logic or direct API calls (except for simple one-off commands)

### Layer 2 — Application Layer (`internal/<resource>/client.go`)

- Contains business logic and domain operations
- Communicates with the cloud API
- Defines mockable interfaces (e.g. `CloudService`) for all external dependencies
- Returns domain objects, not command-specific types
- Must be testable without CLI dependencies

### When to Use This Pattern

**Use the two-layer pattern when:**
- The resource has multiple operations (namespace, apikey, identity, etc.)
- Business logic is non-trivial and needs unit testing
- The same logic might be reused across multiple commands

**Skip the application layer for:**
- Simple one-off commands (like `whoami`)
- Commands with no business logic (direct API passthrough)

For simple commands, call `cloudClient.CloudService()` directly in the `run` method.

## Setting Up Mocks

When creating a new resource client with mockable interfaces:

**1. Define the interface in the application layer:**
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

**2. Update `.mockery.yml`:**
```yaml
packages:
  github.com/temporalio/cloud-cli/internal/myresource:
    interfaces:
      CloudService:
```

**3. Regenerate mocks:**
```bash
make mocks
```

This generates `internal/myresource/mock/mock.go` with a `MockCloudService` type.

**4. Add the client interface to `CommandContext` if needed across multiple commands:**
```go
// temporalcloudcli/commands.go
type MyResourceClient interface {
    GetResource(context.Context, string) (*resourcev1.Resource, error)
}

type CommandContext struct {
    // ...
    MyResourceClient MyResourceClient
}
```

Then add to `.mockery.yml` and run `make mocks` again:
```yaml
packages:
  github.com/temporalio/cloud-cli/temporalcloudcli:
    interfaces:
      MyResourceClient:
```
