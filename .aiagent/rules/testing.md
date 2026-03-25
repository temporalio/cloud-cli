# Testing

## Unit Tests (`temporalcloudcli/commands.<name>_test.go`)

Unit tests use the `TestCommand` harness (`commands.testing.go`) to call `run` directly on a command struct with injected mock dependencies. No build tag is required — these run as part of the normal `make all` / `go test` flow.

**Package:** `package temporalcloudcli_test`

- Use `assert.Equal` for protobuf message assertions. Use `proto.Equal` only inside `mock.MatchedBy` closures. Never compare proto messages field-by-field.
- Example naming: `TestGetFoo_Success`, `TestSetFoo_GetNamespaceError`
- See `temporalcloudcli/commands.apikey_test.go` for the canonical pattern

### `TestCommand` harness

`temporalcloudcli.TestCommand(t, ctx, command, opts)` builds a `CommandContext` with two override hooks and calls `command.run`:

```go
type TestCommandOptions struct {
    Args                    []string
    CloudClientExpectations func(cloudClient *cloudmock.MockCloudServiceClient)
    AsyncPollerOptions      TestAsyncPollerOptions
    JSONOutput              bool
    ExpectedError           string
    ExpectedOutput          string
    ExpectedOutputJson      any  // proto.Message or plain value; compared with JSONEq
}

type TestAsyncPollerOptions struct {
    AsyncOperationID string  // set when polling should be exercised; drives mock GetAsyncOperation expectations
    ErrorToReturn    error   // set to simulate a poll failure
}
```

- `CloudClientExpectations` — sets `EXPECT()` calls on `*cloudmock.MockCloudServiceClient`; injected via `cctx.getCloudClientOverride`
- `AsyncPollerOptions` — drives a real `async.NewPoller` backed by a mock cloud client; injected via `cctx.getAsyncPollerOverride`
  - `AsyncOperationID` set (no error): two `GetAsyncOperation` calls — first `STATE_PENDING`, second `STATE_FULFILLED`
  - `ErrorToReturn` set: one `GetAsyncOperation` call → returns the error
  - Neither set: no poller expectations (command errors before reaching the poller)

**Read command example:**
```go
func TestGetFoo(t *testing.T) {
    tests := []struct {
        name                  string
        cmd                   temporalcloudcli.CloudNamespaceFooGetCommand
        setClientExpectations func(*cloudmock.MockCloudServiceClient)
        expectedErr           string
        expectedJsonOutput    any
    }{
        {
            name: "Success",
            cmd:  temporalcloudcli.CloudNamespaceFooGetCommand{Namespace: "my-ns"},
            setClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
                c.EXPECT().
                    GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
                    Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespace}, nil)
            },
            expectedJsonOutput: testNamespace,
        },
        {
            name: "GetNamespaceError",
            cmd:  temporalcloudcli.CloudNamespaceFooGetCommand{Namespace: "my-ns"},
            setClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
                c.EXPECT().
                    GetNamespace(mock.Anything, mock.Anything, mock.Anything).
                    Return(nil, errors.New("not found"))
            },
            expectedErr: "not found",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            temporalcloudcli.TestCommand(t, context.Background(), &tt.cmd, temporalcloudcli.TestCommandOptions{
                CloudClientExpectations: tt.setClientExpectations,
                JSONOutput:              true,
                ExpectedError:           tt.expectedErr,
                ExpectedOutputJson:      tt.expectedJsonOutput,
            })
        })
    }
}
```

**Mutation command example (with async polling):**
```go
temporalcloudcli.TestCommand(t, context.Background(), &tt.cmd, temporalcloudcli.TestCommandOptions{
    CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
        c.EXPECT().GetNamespace(...).Return(...)
        c.EXPECT().UpdateNamespace(...).Return(&cloudservice.UpdateNamespaceResponse{
            AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
        }, nil)
    },
    AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
        AsyncOperationID: "op-123",
    },
    JSONOutput:         true,
    ExpectedOutputJson: expectedResponse,
})
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

See `temporalcloudcli/commands.apikey_test.go` for the canonical unit test pattern.

## Testing Editor-Based Commands

`runEditorForJSONEditForProtos` launches a real `$EDITOR` process — it cannot be called in unit tests. For `edit` commands, store the editor as a field on the command struct and inject it via `TestCommandOptions.CloudClientExpectations` or by setting it directly on the cmd before passing to `TestCommand`:

```go
// In the command struct (unexported field, set before run):
type CloudNamespaceFooEditCommand struct {
    // ...generated fields...
    runEditor func(existing, target proto.Message) error
}

func (c *CloudNamespaceFooEditCommand) run(cctx *CommandContext, _ []string) error {
    runEditor := c.runEditor
    if runEditor == nil {
        runEditor = runEditorForJSONEditForProtos
    }
    // use runEditor(existing, newSpec)
}
```

Tests set the field directly on the cmd struct before passing to `TestCommand`:
```go
cmd := &temporalcloudcli.CloudNamespaceFooEditCommand{...}
cmd.SetRunEditor(func(existing, target proto.Message) error {
    proto.Merge(target, editedSpec)
    return nil
})
temporalcloudcli.TestCommand(t, ctx, cmd, opts)
```

## Common Test Patterns

**Always set `AutoConfirm: true`** to bypass prompts unless testing prompt behavior:
```go
s.RootCommand.AutoConfirm = true
```

**Test prompt scenarios as individual functions:**
- User declining: `AutoConfirm: false`, provide `"n\n"` as stdin
- JSON output without auto-confirm: should error
