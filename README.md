# cloud-cli
CLI Plugin for Temporal Cloud

## Installing

### Using `go install`

#### Dependencies

- [`temporal` CLI](https://temporal.io/setup/install-temporal-cli) 
- [go](https://go.dev/doc/install) (check [go.mod](./go.mod) for the version)

Clone the repo, then run the following command from the root of the project to build and install the cloud extension:

```sh
$ go install ./cmd/temporal-cloud
```

You should now be able to run cloud extension commands by running `temporal cloud`. 

## Testing
In order to run the tests, you need a temporal cloud api key. Create one in the [cloud dashboard](https://cloud.temporal.io/) and place it in a .env file at the root directory of the repo as follows:
```
TEMPORAL_API_KEY=<api key>
TEMPORAL_CLOUD_SERVER=<server>
```

*NOTE* This will create and delete resources on the account. 

Then run with `mise run test` to run the tests.
