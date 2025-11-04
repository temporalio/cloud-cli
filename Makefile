.PHONY: all gen build

all: gen build

gen: temporalcloudcli/commands.gen.go

temporalcloudcli/commands.gen.go: temporalcloudcli/commandsgen/commands.yml
	go run ./temporalcloudcli/internal/cmd/gen-commands

build:
	go build ./cmd/cloud-cli
