.PHONY: all gen build

all: gen build

gen: 
	go tool gen-commands -input ./temporalcloudcli/commandsgen/commands.yml -pkg temporalcloudcli > ./temporalcloudcli/commands.gen.go

build:
	go build ./cmd/temporal-cloud
