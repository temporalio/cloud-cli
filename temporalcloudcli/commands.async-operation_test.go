//go:build integration
// +build integration

package temporalcloudcli_test

import (
	"encoding/json"
	"fmt"
	"io"

	"go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *SharedServerSuite) TestAsyncOperationCommands() {
	// First, let's create a namespace asynchronously to get an operation ID
	nsName := s.generateRandomNamespaceName()

	spec := map[string]interface{}{
		"name":           nsName,
		"regions":        []string{"aws-us-east-1"},
		"retention_days": 1,
	}
	specBytes, err := json.Marshal(spec)
	s.Suite.Require().NoError(err)
	specArg := string(specBytes)

	// Note: We use --async to get the operation ID right away without waiting
	createRes := s.Execute(
		"namespace", "apply",
		fmt.Sprintf("--server=%s", s.server),
		"--spec", specArg,
		"--async",
		"-o=json",
		"--auto-confirm",
	)
	s.Suite.Require().NoError(createRes.Err, "failed to initiate namespace create")

	buf, err := io.ReadAll(&createRes.Stdout)
	s.Suite.Require().NoError(err)

	defer func() {
		// Clean up the namespace after the test
		_, _ = s.Execute(
			"namespace", "delete",
			fmt.Sprintf("--server=%s", s.server),
			nsName,
			"--auto-confirm",
		)
	}()

	// Extract the async operation ID from the output
	var output map[string]interface{}
	err = json.Unmarshal(buf, &output)
	s.Suite.Require().NoError(err)

	asyncOpMap, ok := output["asyncOperation"].(map[string]interface{})
	s.Suite.Require().True(ok, "expected asyncOperation in output")

	opID, ok := asyncOpMap["id"].(string)
	s.Suite.Require().True(ok, "expected id in asyncOperation")
	s.Suite.Require().NotEmpty(opID, "async operation ID should not be empty")

	// Now try to fetch the operation using our new command
	getRes := s.Execute(
		"async-operation", "get",
		fmt.Sprintf("--server=%s", s.server),
		"-i", opID,
		"-o=json",
	)
	s.Suite.Require().NoError(getRes.Err, "failed to get async operation")

	getBuf, err := io.ReadAll(&getRes.Stdout)
	s.Suite.Require().NoError(err)

	fetchedOp := &operation.AsyncOperation{}
	err = protojson.Unmarshal(getBuf, fetchedOp)
	s.Suite.Require().NoError(err, "failed to unmarshal async operation")

	s.Suite.Require().Equal(opID, fetchedOp.Id, "fetched operation ID should match requested ID")
	s.Suite.Require().Equal("CreateNamespace", fetchedOp.OperationType, "should have the correct operation type")

	// Use await command to wait until completion
	awaitRes := s.Execute(
		"async-operation", "await",
		fmt.Sprintf("--server=%s", s.server),
		"-i", opID,
		"-o=json",
	)
	s.Suite.Require().NoError(awaitRes.Err, "failed to await async operation")

	awaitBuf, err := io.ReadAll(&awaitRes.Stdout)
	s.Suite.Require().NoError(err)

	// Since we used `-o=json`, `await` prints out the final `MutationResult` or error state json
	var asyncOp operation.AsyncOperation
	err = json.Unmarshal(awaitBuf, &asyncOp)
	s.Suite.Require().NoError(err)
	s.Suite.Require().Equal("STATE_FULFILLED", asyncOp.GetState(), "should be fulfilled")
}
