# cloud-cli
CLI Plugin for Temporal Cloud

## Testing
In order to run the tests, you need a temporal cloud api key. Create one in the [cloud dashboard](https://cloud.temporal.io/) and place it in a .env file at the root directory of the repo as follows:
```
TEMPORAL_API_KEY=<api key>
TEMPORAL_CLOUD_SERVER=<server>
TEMPORAL_ACCOUNT=<account>
```

*NOTE* This will create and delete resources on the account. 

Then run with `mise run test` to run the tests.