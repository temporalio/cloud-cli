# Temporal Cloud CLI (Pre-release)

Plugin for the Temporal command-line interface to work with [Temporal Cloud](https://docs.temporal.io/cloud).

> **Pre-release:** This plugin is offered as a pre-release and is subject to change. Please reach out to [Temporal Support](https://support.temporal.io/) if you have questions.

## Prerequisites

- A [Temporal Cloud](https://docs.temporal.io/cloud) account.
- The [`temporal` CLI](https://temporal.io/setup/install-temporal-cli) installed and on your `PATH`. The Cloud CLI runs as a plugin to it.

## Quick install

### Install via Homebrew

    brew install temporalio/prerelease/cloud-cli

### Install via download

1. Download the [latest version](https://github.com/temporalio/cloud-cli/releases/latest) for your OS and architecture.
2. Extract the downloaded archive.
3. Add the `temporal-cloud` binary to your `PATH` (`temporal-cloud.exe` for Windows).

### Build

1. Install [Go](https://go.dev/doc/install) (check [go.mod](./go.mod) for the version).
2. Clone this repository.
3. From the cloned directory, run `make build` (or `go build ./cmd/temporal-cloud`).

The executable will be at `temporal-cloud` (`temporal-cloud.exe` for Windows). Add it to your `PATH` so the `temporal` CLI can discover it.

## Usage

Once installed, invoke the plugin via `temporal cloud`.

### Authenticate

```sh
temporal cloud login     # browser-based OAuth login
temporal cloud whoami    # confirm the authenticated identity
```

Alternatively, pass an API key directly to any command with `--api-key`.

### Examples

```sh
temporal cloud namespace list
temporal cloud namespace get --namespace <namespace>
temporal cloud namespace retention get --namespace <namespace>
```

Run `temporal cloud --help` (or `temporal cloud <command> --help`) to see all available commands and flags.
