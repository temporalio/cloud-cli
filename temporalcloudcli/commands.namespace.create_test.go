package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var defaultCreateResponse = &cloudservice.CreateNamespaceResponse{
	AsyncOperation: &operationv1.AsyncOperation{Id: "test-operation-id"},
	Namespace:      "my-namespace.my-account",
}

func noopUnmarshalProtoJSON(b []byte, m proto.Message) error {
	return temporalcloudcli.UnmarshalProtoJSONWithOptions(b, m, false)
}

// TestCreateNamespace_Success verifies that CreateNamespace calls HandleResult with the API response.
func TestCreateNamespace_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	mockPrompter.EXPECT().
		PromptApply(mock.Anything, mock.Anything, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), mock.Anything).
		Return(defaultCreateResponse, nil)

	var gotOp *operationv1.AsyncOperation
	var gotNamespaceID string
	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		HandleResult: func(op *operationv1.AsyncOperation, namespaceID string) error {
			gotOp = op
			gotNamespaceID = namespaceID
			return nil
		},
	})
	require.NoError(t, err)
	assert.Equal(t, defaultCreateResponse.AsyncOperation, gotOp)
	assert.Equal(t, defaultCreateResponse.Namespace, gotNamespaceID)
}

// TestCreateNamespace_BuildsSpec verifies that CreateNamespace correctly wires params into the NamespaceSpec.
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

	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	mockPrompter.EXPECT().
		PromptApply(mock.Anything, mock.Anything, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), mock.MatchedBy(func(req *cloudservice.CreateNamespaceRequest) bool {
			return proto.Equal(req.Spec, expectedSpec)
		})).
		Return(defaultCreateResponse, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:              "my-namespace",
		Regions:           []string{"aws-us-east-1"},
		RetentionDays:     30,
		ApiKeyAuthEnabled: true,
		CertificateFilterOptions: temporalcloudcli.CertificateFilterOptions{
			CertificateFilter: []string{
				`{"commonName":"test.temporal.io"}`,
				`{"subjectAlternativeName":"*.temporal.io"}`,
			},
		},
		SearchAttribute:    []string{"MyText=Text", "MyKeyword=Keyword"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		HandleResult:       func(*operationv1.AsyncOperation, string) error { return nil },
	})
	require.NoError(t, err)
}

// TestCreateNamespace_Error verifies that an API error propagates.
func TestCreateNamespace_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	apiErr := errors.New("create failed")

	mockPrompter.EXPECT().
		PromptApply(mock.Anything, mock.Anything, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), mock.Anything).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		HandleResult:       func(*operationv1.AsyncOperation, string) error { return nil },
	})
	require.ErrorIs(t, err, apiErr)
}

// TestCreateNamespace_IdempotentAlreadyExists verifies that an AlreadyExists error prints
// an unchanged result when Idempotent is true.
func TestCreateNamespace_IdempotentAlreadyExists(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	alreadyExistsErr := status.Error(codes.AlreadyExists, "namespace already exists")

	mockPrompter.EXPECT().
		PromptApply(mock.Anything, mock.Anything, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), mock.Anything).
		Return(nil, alreadyExistsErr)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Idempotent:         true,
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		HandleResult:       func(*operationv1.AsyncOperation, string) error { return nil },
	})
	require.NoError(t, err)

	var result temporalcloudcli.Result
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "unchanged", result.Status)
	assert.Equal(t, "my-namespace", result.ID)
}

// TestCreateNamespace_PromptDeclined verifies that the API is never called when the prompt is declined.
func TestCreateNamespace_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	promptErr := fmt.Errorf("Aborting apply.")

	mockPrompter.EXPECT().
		PromptApply(mock.Anything, mock.Anything, false).
		Return(promptErr)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		HandleResult:       func(*operationv1.AsyncOperation, string) error { return nil },
	})
	require.ErrorIs(t, err, promptErr)
}

// TestCreateNamespace_HandleResultError verifies that an error from HandleResult propagates.
func TestCreateNamespace_HandleResultError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	handleErr := errors.New("polling failed")

	mockPrompter.EXPECT().
		PromptApply(mock.Anything, mock.Anything, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), mock.Anything).
		Return(defaultCreateResponse, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		HandleResult:       func(*operationv1.AsyncOperation, string) error { return handleErr },
	})
	require.ErrorIs(t, err, handleErr)
}

// TestCreateNamespace_InvalidInput verifies that input parsing errors are returned before the API is called.
func TestCreateNamespace_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		params      temporalcloudcli.CreateNamespaceParams
		assertError func(*testing.T, error)
	}{
		{
			name: "invalid search attribute format",
			params: temporalcloudcli.CreateNamespaceParams{
				Name:            "my-namespace",
				Regions:         []string{"aws-us-east-1"},
				SearchAttribute: []string{"MissingEquals"},
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid search attribute format")
			},
		},
		{
			name: "invalid search attribute type",
			params: temporalcloudcli.CreateNamespaceParams{
				Name:            "my-namespace",
				Regions:         []string{"aws-us-east-1"},
				SearchAttribute: []string{"MyAttr=NotAType"},
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid search attribute type")
			},
		},
		{
			name: "invalid certificate filter JSON",
			params: temporalcloudcli.CreateNamespaceParams{
				Name:    "my-namespace",
				Regions: []string{"aws-us-east-1"},
				CertificateFilterOptions: temporalcloudcli.CertificateFilterOptions{
					CertificateFilter: []string{"not-valid-json"},
				},
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "failed to parse certificate filter")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			mockPrompter := cmdmock.NewMockPrompter(t)

			var buf bytes.Buffer
			tt.params.Cloud = mockCloud
			tt.params.Printer = &printer.Printer{Output: &buf, JSON: true}
			tt.params.Prompter = mockPrompter
			tt.params.UnmarshalProtoJSON = noopUnmarshalProtoJSON
			tt.params.HandleResult = func(*operationv1.AsyncOperation, string) error { return nil }

			err := temporalcloudcli.CreateNamespace(context.Background(), tt.params)
			require.Error(t, err)
			tt.assertError(t, err)
		})
	}
}
