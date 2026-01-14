.PHONY: all gen build test

include .env
export $(shell sed 's/=.*//' .env)

all: gen build test

gen: 
	go tool gen-commands -input ./temporalcloudcli/commands.yml -pkg temporalcloudcli > ./temporalcloudcli/commands.gen.go

build:
	go build ./cmd/temporal-cloud

test-integration:
	go test -tags=integration ./...

test:
	go test ./...
