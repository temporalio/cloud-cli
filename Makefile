.PHONY: all install buf-install gen proto build test mocks

# Load .env file if it exists (for local development)
# In CI/CD, environment variables are provided by the environment
-include .env
export

# Derive the version from git so `make build` stamps the binary. Falls back to
# the commit hash if no tag is reachable, and appends "-dirty" for uncommitted changes.
# When git is unavailable (e.g. a source tarball), VERSION is empty and we omit the
# ldflag entirely so the in-code default (0.0.0-DEV) is preserved.
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null)
VERSION_PKG := github.com/temporalio/cloud-cli/temporalcloudcli
LDFLAGS := $(if $(VERSION),-X $(VERSION_PKG).Version=$(VERSION))

all: gen proto build mocks test

# we need to install gen-commands directly because gen-commands does not have its
# own go.mod
install:
	rm -rf ./cli
	git clone https://github.com/temporalio/cli && cd ./cli && go install ./cmd/gen-commands && cd ..
	rm -rf ./cli

gen: install
	gen-commands -input ./temporalcloudcli/commands.yml -pkg temporalcloudcli > ./temporalcloudcli/commands.gen.go

buf-install:
	go install github.com/bufbuild/buf/cmd/buf@v1.50.0

# Regenerates the vendored onboarding service client (proto/ -> internal/).
# The protoc plugins are pinned as go tools in go.mod.
proto: buf-install
	buf lint
	buf generate

build:
	go build -ldflags "$(LDFLAGS)" ./cmd/temporal-cloud

test-integration:
	go test -tags=integration ./...

test:
	go test ./...

mocks:
	go tool mockery
