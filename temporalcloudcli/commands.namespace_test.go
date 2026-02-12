//go:build integration
// +build integration

package temporalcloudcli_test

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"go.temporal.io/api/temporalproto"
	"go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespace "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	e2eNamespacePrefix = "e2e-namespace"
)

func (s *SharedServerSuite) generateRandomNamespaceName() string {
	return fmt.Sprintf("%s-%s", e2eNamespacePrefix, s.generateRandomID())
}

func (s *SharedServerSuite) TestBasicNamespaceOperations() {
	s.cleanupNamespaces()
	s.testnamespaceCRUD()
	s.cleanupNamespaces()
}

func (s *SharedServerSuite) testnamespaceCRUD() {
	cloudClient := s.getCloudClient()
	// create a new namespace
	newNamespaceName := s.generateRandomNamespaceName()

	namespaceSpec := &namespace.NamespaceSpec{
		Name:          newNamespaceName,
		Regions:       []string{"aws-us-east-1"},
		RetentionDays: 30,
		ApiKeyAuth: &namespace.ApiKeyAuthSpec{
			Enabled: true,
		},
		SearchAttributes: map[string]namespace.NamespaceSpec_SearchAttributeType{
			"test": namespace.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
		},
		Lifecycle: &namespace.LifecycleSpec{},
	}

	buf, err := temporalproto.CustomJSONMarshalOptions{}.Marshal(namespaceSpec)
	s.Suite.Require().NoError(err)

	res := s.Execute(
		"namespace",
		fmt.Sprintf("--server=%s", s.server), // TODO (gmankes): remove this when the server is defaulted back to prod
		"apply",
		"--auto-confirm=true",
		"--spec", fmt.Sprintf(`%s`, string(buf)),
		"-o=json",
	)

	s.Suite.Require().NoError(res.Err)
	buf, err = io.ReadAll(&res.Stdout)
	s.Suite.Require().NoError(err)
	result := &temporalcloudcli.MutationResult{}
	err = json.Unmarshal(buf, result)
	s.Suite.Require().NoError(err)

	namespaceID := result.ID

	// make sure ns apply is completed
	asyncOpID := result.AsyncOp.Id
	err = s.pollAsyncOperation(cloudClient, asyncOpID)
	s.Suite.Require().NoError(err)

	// get the namespace
	res = s.Execute(
		"namespace",
		fmt.Sprintf("--server=%s", s.server), // TODO (gmankes): remove this when the server is defaulted back to prod
		"get",
		"-n", namespaceID,
		"-o=json",
	)
	s.Suite.Require().NoError(res.Err)

	buf, err = io.ReadAll(&res.Stdout)
	s.Suite.Require().NoError(err)

	readNamespace := &namespace.Namespace{}
	err = protojson.Unmarshal(buf, readNamespace)
	s.Suite.Require().NoError(err)

	// compare it to the inputted spec
	s.Suite.Require().NotNil(readNamespace)
	s.Suite.Require().NotNil(readNamespace.Spec)
	s.Suite.Equal(namespaceSpec.Name, readNamespace.Spec.Name)
	s.Suite.Equal(namespaceSpec.Regions, readNamespace.Spec.Regions)
	s.Suite.Equal(namespaceSpec.SearchAttributes, readNamespace.Spec.SearchAttributes)
	s.Suite.Equal(namespaceSpec.RetentionDays, readNamespace.Spec.RetentionDays)

	// get the namespace via listing
	res = s.Execute(
		"namespace",
		fmt.Sprintf("--server=%s", s.server), // TODO (gmankes): remove this when the server is defaulted back to prod
		"list",
		fmt.Sprintf("--name=%s", newNamespaceName),
		"-o=json",
	)
	s.Suite.Require().NoError(res.Err)

	buf, err = io.ReadAll(&res.Stdout)
	s.Suite.Require().NoError(err)

	// simply assert that the list contains the namespace
	s.Suite.Contains(string(buf), newNamespaceName)

	// update the namespace
	namespaceSpec.RetentionDays--

	buf, err = temporalproto.CustomJSONMarshalOptions{}.Marshal(namespaceSpec)
	s.Suite.Require().NoError(err)

	res = s.Execute(
		"namespace",
		fmt.Sprintf("--server=%s", s.server), // TODO (gmankes): remove this when the server is defaulted back to prod
		"apply",
		"--auto-confirm=true",
		"--spec", fmt.Sprintf(`%s`, string(buf)),
		"-o=json",
	)
	s.Suite.Require().NoError(res.Err)
	buf, err = io.ReadAll(&res.Stdout)
	s.Suite.Require().NoError(err)
	result = &temporalcloudcli.MutationResult{}
	err = json.Unmarshal(buf, result)
	s.Suite.Require().NoError(err)

	// make sure ns apply is completed
	asyncOpID = result.AsyncOp.Id
	err = s.pollAsyncOperation(cloudClient, asyncOpID)
	s.Suite.Require().NoError(err)

	// get the namespace (after updating)
	res = s.Execute(
		"namespace",
		fmt.Sprintf("--server=%s", s.server), // TODO (gmankes): remove this when the server is defaulted back to prod
		"get",
		"-n", namespaceID,
		"-o=json",
	)
	s.Suite.Require().NoError(res.Err)

	buf, err = io.ReadAll(&res.Stdout)
	s.Suite.Require().NoError(err)

	readNamespace = &namespace.Namespace{}
	err = protojson.Unmarshal(buf, readNamespace)
	s.Suite.Require().NoError(err)

	// compare it to the inputted spec
	s.Suite.Require().NotNil(readNamespace)
	s.Suite.Require().NotNil(readNamespace.Spec)
	s.Suite.Equal(namespaceSpec.Name, readNamespace.Spec.Name)
	s.Suite.Equal(namespaceSpec.Regions, readNamespace.Spec.Regions)
	s.Suite.Equal(namespaceSpec.SearchAttributes, readNamespace.Spec.SearchAttributes)
	s.Suite.Equal(namespaceSpec.RetentionDays, readNamespace.Spec.RetentionDays)

	// delete the namespace
	res = s.Execute(
		"namespace",
		fmt.Sprintf("--server=%s", s.server), // TODO (gmankes): remove this when the server is defaulted back to prod
		"delete",
		"-n", namespaceID,
		"--idempotent",
		"--auto-confirm=true",
		"-o=json",
	)
	s.Suite.Require().NoError(res.Err)

	// try to get the namespace
	res = s.Execute(
		"namespace",
		fmt.Sprintf("--server=%s", s.server), // TODO (gmankes): remove this when the server is defaulted back to prod
		"get",
		"-n", namespaceID,
	)

	s.Suite.Require().Error(res.Err)
	// should say not found
	stdOut, err := io.ReadAll(&res.Stdout)
	s.Suite.Require().NoError(err)

	s.Suite.NotContains(strings.ToLower(string(stdOut)), newNamespaceName)
}

func (s *SharedServerSuite) cleanupNamespaces() {
	cloudClient := s.getCloudClient()

	pageToken := ""
	namespacesToClean := []*namespace.Namespace{}
	for {
		res, err := cloudClient.CloudService().GetNamespaces(s.Context, &cloudservice.GetNamespacesRequest{
			PageToken: pageToken,
		})
		s.Suite.Require().NoError(err)
		for _, ns := range res.Namespaces {
			if strings.HasPrefix(ns.Namespace, e2eNamespacePrefix) {
				namespacesToClean = append(namespacesToClean, ns)
			}
		}
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	for _, ns := range namespacesToClean {
		res, err := cloudClient.CloudService().DeleteNamespace(s.Context, &cloudservice.DeleteNamespaceRequest{
			ResourceVersion: ns.ResourceVersion,
			Namespace:       ns.Namespace,
		})
		s.Suite.NoError(err)
		if err == nil {
			pollErr := s.pollAsyncOperation(cloudClient, res.AsyncOperation.Id)
			s.Suite.NoError(pollErr)
		}
	}
}
