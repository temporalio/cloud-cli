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
