# Testing

## Unit Tests (`temporalcloudcli/commands.<name>_test.go`)

Unit tests use the `TestCommand` harness (`commands.testing.go`) to call `run` directly on a command struct with injected mock dependencies. No build tag is required — these run as part of the normal `make all` / `go test` flow.

**Package:** `package temporalcloudcli_test`

- Use `assert.Equal` for protobuf message assertions. Use `proto.Equal` only inside `mock.MatchedBy` closures. Never compare proto messages field-by-field.
- Example naming: `TestGetFoo_Success`, `TestSetFoo_GetNamespaceError`
- See `temporalcloudcli/commands.apikey_test.go` for the canonical pattern

### `TestCommand` harness

`temporalcloudcli.TestCommand(t, command, opts)` builds a `CommandContext` with override hooks and calls `command.run`:

```go
type TestCommandOptions struct {
    Args                    []string
    CloudClientExpectations func(cloudClient *cloudmock.MockCloudServiceClient)
    AsyncPollerOptions      TestAsyncPollerOptions
    PromptOptions           TestPromptOptions
    EditorOptions           TestEditorOptions
    JSONOutput              bool
    ExpectedError           string
    ExpectedOutput          string
    ExpectedOutputJson      any  // proto.Message or plain value; compared with JSONEq
}

type TestAsyncPollerOptions struct {
    AsyncOperationID string  // set when polling should be exercised; drives mock GetAsyncOperation expectations
    ErrorToReturn    error   // set to simulate a poll failure
}

type TestPromptOptions struct {
    ExpectPromptYes        bool
    ExpectPromptYesMessage string
    ExpectPrompApply       bool
    // optional: assert the exact messages passed to PromptApply
    ExpectPromptApplyExisting proto.Message
    ExpectPromptApplyModified proto.Message
    ExpectPromptApplyVerbose  bool
    PromptResult bool   // returned by the mock (true = confirmed)
    PromptError  error  // returned by the mock
}

type TestEditorOptions struct {
    Modified    proto.Message  // proto returned by EditProto; nil = no editor call expected
    EditorError error
}
```

- `CloudClientExpectations` — sets `EXPECT()` calls on `*cloudmock.MockCloudServiceClient`; injected via `cctx.getCloudClientOverride`
- `AsyncPollerOptions` — drives a real `async.NewPoller` backed by a mock cloud client; injected via `cctx.getAsyncPollerOverride`
  - `AsyncOperationID` set (no error): two `GetAsyncOperation` calls — first `STATE_PENDING`, second `STATE_FULFILLED`
  - `ErrorToReturn` set: one `GetAsyncOperation` call → returns the error
  - Neither set: no poller expectations (command errors before reaching the poller)
- `PromptOptions` — drives a `MockPrompter`; set `ExpectPromptYes` or `ExpectPrompApply` to assert the mock is called once; `PromptResult` controls what it returns
- `EditorOptions` — drives a `MockEditor`; set `Modified` to the proto the editor should return; set `EditorError` to simulate an error

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
            temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
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
temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
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

`editor.Editor.EditProto` launches a real `$EDITOR` process — it cannot run in unit tests. The `TestCommand` harness automatically injects a `MockEditor` via `cctx.getEditorOverride`. Use `TestEditorOptions` to control its behavior:

```go
temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
    CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) { ... },
    EditorOptions: temporalcloudcli.TestEditorOptions{
        Modified: editedSpec,  // proto returned by EditProto
    },
    JSONOutput:         true,
    ExpectedOutputJson: expectedResponse,
})
```

- `Modified` set: mock expects one `EditProto` call and returns `Modified`
- `EditorError` set: mock expects one `EditProto` call and returns the error
- Neither set: no editor call expected (command errors before reaching the editor)

In the `run` method, obtain the editor via `cctx.GetEditor()`:
```go
func (c *CloudNamespaceFooEditCommand) run(cctx *CommandContext, _ []string) error {
    // ...fetch existing...
    modified, err := cctx.GetEditor().EditProto(existing)
    if err != nil {
        return err
    }
    // use modified...
}
```

## Common Test Patterns

**Bypassing prompts in unit tests** — the `MockPrompter` only fires expectations you explicitly set in `PromptOptions`. If a command calls `GetPrompter().PromptYes` but you don't set `ExpectPromptYes: true`, the test will fail with an unexpected call. To simulate auto-confirm, set `PromptOptions: TestPromptOptions{ExpectPromptYes: true, PromptResult: true}`.

**Test prompt scenarios as individual functions:**
- User confirming: `PromptOptions: TestPromptOptions{ExpectPromptYes: true, PromptResult: true}`
- User declining: `PromptOptions: TestPromptOptions{ExpectPromptYes: true, PromptResult: false}`
- Prompt error: `PromptOptions: TestPromptOptions{ExpectPromptYes: true, PromptError: errors.New("...")}`

**Integration tests** use `s.RootCommand.AutoConfirm = true` to bypass real prompts.
