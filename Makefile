.PHONY: all install gen build test mocks

# Load .env file if it exists (for local development)
# In CI/CD, environment variables are provided by the environment
-include .env
export

all: gen build mocks test

# we need to install gen-commands directly because gen-commands does not have its
# own go.mod
install:
	rm -rf ./cli
	git clone https://github.com/temporalio/cli && cd ./cli && go install ./cmd/gen-commands && cd ..
	rm -rf ./cli

gen: install
	gen-commands -input ./temporalcloudcli/commands.yml -pkg temporalcloudcli > ./temporalcloudcli/commands.gen.go

build:
	go build ./cmd/temporal-cloud

test-integration:
	go test -tags=integration ./...

test:
	go test ./...

mocks:
	mockery
