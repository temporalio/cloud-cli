package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"
)

var defaultCreateResponse = &cloudservice.CreateNamespaceResponse{
	AsyncOperation: &operationv1.AsyncOperation{Id: "test-operation-id"},
	Namespace:      "my-namespace.my-account",
}

// baseNamespaceSpec returns the NamespaceSpec built from the minimal test params
// (Name: "my-namespace", Regions: ["aws-us-east-1"], all other fields zero-valued).
func baseNamespaceSpec() *namespacev1.NamespaceSpec {
	return &namespacev1.NamespaceSpec{
		Name:       "my-namespace",
		Regions:    []string{"aws-us-east-1"},
		ApiKeyAuth: &namespacev1.ApiKeyAuthSpec{Enabled: false},
		Lifecycle:  &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
	}
}

func createReqMatcher(expected *namespacev1.NamespaceSpec) interface{} {
	return mock.MatchedBy(func(req *cloudservice.CreateNamespaceRequest) bool {
		return proto.Equal(req.Spec, expected)
	})
}

// TestCreateNamespace_Success verifies that a namespace is created and polled to completion.
func TestCreateNamespace_Success(t *testing.T) {
	cmd := temporalcloudcli.CloudNamespaceCreateCommand{
		Name:   "my-namespace",
		Region: []string{"aws-us-east-1"},
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateNamespace(mock.Anything, createReqMatcher(baseNamespaceSpec()), mock.Anything).
				Return(defaultCreateResponse, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
			AsyncOperationID: "test-operation-id",
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
	})
}

// TestCreateNamespace_BuildsSpec verifies that all create flags are correctly wired into the NamespaceSpec.
func TestCreateNamespace_BuildsSpec(t *testing.T) {
	expectedSpec := &namespacev1.NamespaceSpec{
		Name:          "my-namespace",
		Regions:       []string{"aws-us-east-1"},
		RetentionDays: 30,
		ApiKeyAuth:    &namespacev1.ApiKeyAuthSpec{Enabled: true},
		Lifecycle:     &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
		MtlsAuth: &namespacev1.MtlsAuthSpec{
			CertificateFilters: []*namespacev1.CertificateFilterSpec{
				{CommonName: "test.temporal.io"},
				{SubjectAlternativeName: "*.temporal.io"},
			},
		},
		SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
			"MyText":    namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
			"MyKeyword": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
		},
	}

	cmd := temporalcloudcli.CloudNamespaceCreateCommand{
		Name:              "my-namespace",
		Region:            []string{"aws-us-east-1"},
		RetentionDays:     30,
		ApiKeyAuthEnabled: true,
		CertificateFilterOptions: temporalcloudcli.CertificateFilterOptions{
			CertificateFilter: []string{
				`{"commonName":"test.temporal.io"}`,
				`{"subjectAlternativeName":"*.temporal.io"}`,
			},
		},
		SearchAttribute: []string{"MyText=Text", "MyKeyword=Keyword"},
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateNamespace(mock.Anything, createReqMatcher(expectedSpec), mock.Anything).
				Return(defaultCreateResponse, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
			AsyncOperationID: "test-operation-id",
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
	})
}

// TestCreateNamespace_CreateError verifies that an API error is propagated.
func TestCreateNamespace_CreateError(t *testing.T) {
	cmd := temporalcloudcli.CloudNamespaceCreateCommand{
		Name:   "my-namespace",
		Region: []string{"aws-us-east-1"},
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("create failed"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		ExpectedError: "create failed",
	})
}

// TestCreateNamespace_PromptDeclined verifies that the API is never called when the user declines the prompt.
func TestCreateNamespace_PromptDeclined(t *testing.T) {
	cmd := temporalcloudcli.CloudNamespaceCreateCommand{
		Name:   "my-namespace",
		Region: []string{"aws-us-east-1"},
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     false,
		},
		ExpectedError: "Aborting create.",
	})
}

// TestCreateNamespace_PollerError verifies that a polling error is propagated.
func TestCreateNamespace_PollerError(t *testing.T) {
	cmd := temporalcloudcli.CloudNamespaceCreateCommand{
		Name:   "my-namespace",
		Region: []string{"aws-us-east-1"},
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(defaultCreateResponse, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
			ErrorToReturn: errors.New("polling failed"),
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		ExpectedError: "polling failed",
	})
}

// TestCreateNamespace_InvalidInput verifies that input validation errors are returned before any API call.
func TestCreateNamespace_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		cmd           temporalcloudcli.CloudNamespaceCreateCommand
		expectedError string
	}{
		{
			name: "invalid search attribute format",
			cmd: temporalcloudcli.CloudNamespaceCreateCommand{
				Name:            "my-namespace",
				Region:          []string{"aws-us-east-1"},
				SearchAttribute: []string{"MissingEquals"},
			},
			expectedError: "invalid search attribute format",
		},
		{
			name: "invalid search attribute type",
			cmd: temporalcloudcli.CloudNamespaceCreateCommand{
				Name:            "my-namespace",
				Region:          []string{"aws-us-east-1"},
				SearchAttribute: []string{"MyAttr=NotAType"},
			},
			expectedError: "invalid search attribute type",
		},
		{
			name: "invalid certificate filter JSON",
			cmd: temporalcloudcli.CloudNamespaceCreateCommand{
				Name:   "my-namespace",
				Region: []string{"aws-us-east-1"},
				CertificateFilterOptions: temporalcloudcli.CertificateFilterOptions{
					CertificateFilter: []string{"not-valid-json"},
				},
			},
			expectedError: "failed to parse certificate filter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				ExpectedError: tt.expectedError,
			})
		})
	}
}
