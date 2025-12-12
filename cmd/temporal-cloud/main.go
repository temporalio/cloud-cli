package main

import (
	"context"

	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	temporalcloudcli.Execute(ctx, temporalcloudcli.CommandOptions{})
}
