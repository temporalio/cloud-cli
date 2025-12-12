.PHONY: all gen build

all: gen build

gen: 
	go run github.com/temporalio/cli/cmd/gen-commands@main -input ./temporalcloudcli/commandsgen/commands.yml -pkg temporalcloudcli > ./temporalcloudcli/commands.gen.go

build:
	go build ./cmd/temporal-cloud
