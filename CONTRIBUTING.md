# Contributing

## Building

With `make` and the latest `go` version installed, simply run the following:

    make build

## Testing

    make test

See other tests for how to leverage things like the command harness.

### Integration tests

In order to run the integration tests, you need a temporal cloud API key.
Create one in the [cloud dashboard](https://cloud.temporal.io/) and place it in a `.env`
file at the root directory of the repo as follows:

```
TEMPORAL_API_KEY=<api key>
TEMPORAL_CLOUD_SERVER=saas-api.tmprl.cloud:443
```

Then run with `make test-integration` to run the tests.

*NOTE* This will create and delete resources on the account.

## Adding/updating commands

First, update [commands.yml](temporalcloudcli/commands.yml) following the rules in that file. Then to regenerate the
[commands.gen.go](internal/temporalcli/commands.gen.go) file from code, ensure you have
GOBIN or GOHOME/bin on your path, and run:

    make gen

This will expect every non-parent command to have a `run` method, so for new commands developers will have to implement
`run` on the new command in a separate file before it will compile.
