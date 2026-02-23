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
