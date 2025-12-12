.PHONY: all gen build

all: gen build

gen: 
	go run ./temporalcloudcli/internal/cmd/gen-commands

build:
	go build ./cmd/cloud-cli
