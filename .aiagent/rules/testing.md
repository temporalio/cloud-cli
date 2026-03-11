# Testing

## Unit Tests (`temporalcloudcli/commands.<name>_test.go`)

Unit tests call the exported application logic functions directly with mocked dependencies. No build tag is required — these run as part of the normal `make all` / `go test` flow.

**Package:** `package temporalcloudcli_test`

- Write individual test functions (most have unique setup/assertions)
- Use `assert.Equal` for protobuf message assertions. Use `proto.Equal` only inside `mock.MatchedBy` closures. Never compare proto messages field-by-field.
- Mock all three injectable dependency interfaces:
  - `nsmock.NewMockCloudService(t)` (from `internal/namespace/mock`)
  - `cmdmock.NewMockPrompter(t)` (from `temporalcloudcli/mock`)
  - `cmdmock.NewMockAsyncOperationHandler(t)` (from `temporalcloudcli/mock`)
- Example naming: `TestGetRetention_Success`, `TestSetRetention_GetNamespaceError`
- See `temporalcloudcli/commands.namespace.retention_test.go` for the canonical pattern

**Example:**
```go
func TestGetFoo_Success(t *testing.T) {
    mockCloud := nsmock.NewMockCloudService(t)

    mockCloud.EXPECT().
        GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
        Return(&cloudservice.GetNamespaceResponse{
            Namespace: &namespacev1.Namespace{
                Namespace: "my-namespace",
                Spec:      &namespacev1.NamespaceSpec{/* ... */},
            },
        }, nil)

    var buf bytes.Buffer
    err := temporalcloudcli.GetFoo(context.Background(), temporalcloudcli.GetFooParams{
        Namespace: "my-namespace",
        Cloud:     mockCloud,
        Printer:   &printer.Printer{Output: &buf, JSON: true},
    })
    require.NoError(t, err)
    // assert buf contents
}
```

## Integration Tests (`temporalcloudcli/commands.<name>_test.go`)

Integration tests require a live API key + server and use the `SharedServerSuite`.

**Build tag required:**
```go
//go:build integration
// +build integration
```

**Package:** `package temporalcloudcli_test`

**Test infrastructure** (defined in `commands_test.go`):
- `s.Execute(args...)` — runs the CLI with the given args, captures stdout/stderr
- `s.getCloudClient()` — returns a raw `*cloudclient.Client` for direct SDK calls
- `s.pollAsyncOperation(client, opID)` — polls an async op to completion
- `s.generateRandomID()` — returns a random 10-char string for unique resource names
- `s.ContainsOnSameLine(text, pieces...)` — asserts pieces appear in order on one line

**Integration test template:**
```go
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

## When to Use Table-Driven Tests

**Use table-driven tests when** test cases follow the same flow with simple variations:
- Different flag combinations
- Different input values with the same assertion pattern
- async vs sync mode

**Use individual test functions when** cases require different logic:
- Validation errors (different error messages to check)
- API errors (different mock setup)
- Prompt scenarios (different stdin/AutoConfirm setup)
- Idempotent behavior (different assertion logic for success vs error)

**Avoid complex branching** in table-driven tests (nested ifs, different setup per case). If you need conditional logic to handle different cases, use individual test functions.

See `temporalcloudcli/commands.namespace.retention_test.go` for the canonical unit test pattern.

## Common Test Patterns

**Always set `AutoConfirm: true`** to bypass prompts unless testing prompt behavior:
```go
s.RootCommand.AutoConfirm = true
```

**Test prompt scenarios as individual functions:**
- User declining: `AutoConfirm: false`, provide `"n\n"` as stdin
- JSON output without auto-confirm: should error
