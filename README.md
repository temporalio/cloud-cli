# Temporal Cloud CLI

Plugin for the Temporal command-line interface to work with [Temporal Cloud](https://docs.temporal.io/cloud)

## Quick install

### Install via Homebrew

    brew install temporalio/prerelease/cloud-cli

### Install via download

1. Install the [`temporal` CLI](https://temporal.io/setup/install-temporal-cli)
2. Download the [latest version](https://github.com/temporalio/cloud-cli/releases/latest) for your OS and architecture:
3. Extract the downloaded archive.
4. Add the `temporal-cloud` binary to your `PATH` (`temoporal-cloud.exe` for Windows).

### Build

1. Install the [`temporal` CLI](https://temporal.io/setup/install-temporal-cli)
2. Install [Go](https://go.dev/doc/install) (check [go.mod](./go.mod) for the version)
3. Clone repository
4. Switch to cloned directory, and run `go build ./cmd/temporal-cloud`

The executable will be at `temporal-cloud` (`temporal-cloud.exe` for Windows).

## Usage

Once installed, invoke the plugin via `temporal cloud`
