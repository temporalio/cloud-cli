# Cloud CLI Developer Guide

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

The plugin is meant to be an extension to the Temporal CLI. After building, copy the binary somewhere in your `PATH` and rename it to `temporal-cloud`:

```bash
temporal cloud <command> [flags]
```

---

## What AI Must Do

1. Use environment variables instead of committing secrets.
2. Always ask instead of assuming business logic.
3. Add to AIDEV comments or ask instead of removing them.
4. Stay focused on the task at hand. When in doubt about whether a change is related to the task, ask.

## AI Communication Guidelines

### Plan Before Implementing

Before providing your final implementation, use `<implementation_planning>` tags to:

1. Break down the feature into smaller, manageable tasks.
2. Consider potential challenges for each task and how to address them.
3. Provide a high-level outline of the code structure, including function names and their purposes.
4. List specific test cases you plan to implement.
5. State which error handling approaches you will use for different scenarios.
6. Discuss the trade-offs inherent in your design decisions (performance, scalability, complexity, security).
7. Reason about failure modes. How does it handle crashes? A 10x increase in load?

### Code Explanation Format

When explaining code mechanics with verb constructions (e.g., "calls", "sends", "is processed"):

1. **Bold** the verb in the sentence.
2. Immediately after the sentence, provide a clickable code citation.
3. Use the format: `[filepath:line](filepath:line)`

**Example:**
When the account entity module initializes, it **registers** activities with the Temporal worker using prefixed names:

[entities/account/internal/fx.go:22](entities/account/internal/fx.go:22)
```go
w.RegisterActivityWithOptions(a, activity.RegisterOptions{Name: ActivityNamePrefix})
```

---

## Development Process

1. Follow the style rules in `.claude/rules/code-style.md`.
2. Follow the architecture decisions in `.claude/rules/architecture.md`.
3. When adding a new command, follow `.claude/rules/command-implementation.md`.
4. When writing tests, follow `.claude/rules/testing.md`.
5. When a new `.go` file is created, `git add` it.
6. Before saying "done", ensure `make all` passes.
