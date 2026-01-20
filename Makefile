.PHONY: all gen build test

# Load .env file if it exists (for local development)
# In CI/CD, environment variables are provided by the environment
-include .env
ifneq (,$(wildcard .env))
export $(shell sed 's/=.*//' .env)
endif

all: install gen build test

install:
	go install github.com/temporalio/cli/cmd/gen-commands@latest

gen: 
	gen-commands -input ./temporalcloudcli/commands.yml -pkg temporalcloudcli > ./temporalcloudcli/commands.gen.go

build:
	go build ./cmd/temporal-cloud

test-integration:
	go test -tags=integration ./...

test:
	go test ./...
