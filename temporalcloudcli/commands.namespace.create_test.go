package temporalcloudcli_test

import (
	"bytes"
	"context"
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
	"google.golang.org/protobuf/proto"
)

var defaultCreateResponse = &cloudservice.CreateNamespaceResponse{
	AsyncOperation: &operationv1.AsyncOperation{Id: "test-operation-id"},
	Namespace:      "my-namespace.my-account",
}

func noopUnmarshalProtoJSON(b []byte, m proto.Message) error {
	return temporalcloudcli.UnmarshalProtoJSONWithOptions(b, m, false)
}

// baseNamespaceSpec returns the NamespaceSpec built from the minimal test params
// (Name: "my-namespace", Regions: ["aws-us-east-1"], all other fields zero-valued).
func baseNamespaceSpec() *namespacev1.NamespaceSpec {
	return &namespacev1.NamespaceSpec{
		Name:       "my-namespace",
		Regions:    []string{"aws-us-east-1"},
		ApiKeyAuth: &namespacev1.ApiKeyAuthSpec{Enabled: false},
		Lifecycle:  &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
		Fairness:   &namespacev1.FairnessSpec{TaskQueueFairnessEnabled: false},
	}
}

func specMatcher(expected *namespacev1.NamespaceSpec) interface{} {
	return mock.MatchedBy(func(s proto.Message) bool { return proto.Equal(s, expected) })
}

func createReqMatcher(expected *namespacev1.NamespaceSpec) interface{} {
	return mock.MatchedBy(func(req *cloudservice.CreateNamespaceRequest) bool {
		return proto.Equal(req.Spec, expected)
	})
}

// TestCreateNamespace_Success verifies that CreateNamespace calls HandleOperation with the API response.
func TestCreateNamespace_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	spec := baseNamespaceSpec()

	mockPrompter.EXPECT().
		PromptApply(&namespacev1.NamespaceSpec{}, specMatcher(spec), false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), createReqMatcher(spec)).
		Return(defaultCreateResponse, nil)

	mockHandler.EXPECT().
		HandleOperation(defaultCreateResponse.AsyncOperation, "my-namespace.my-account").
		Return(nil)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		OperationHandler:   mockHandler,
	})
	require.NoError(t, err)
}

// TestCreateNamespace_BuildsSpec verifies that CreateNamespace correctly wires params into the NamespaceSpec.
func TestCreateNamespace_BuildsSpec(t *testing.T) {
	expectedSpec := &namespacev1.NamespaceSpec{
		Name:          "my-namespace",
		Regions:       []string{"aws-us-east-1"},
		RetentionDays: 30,
		ApiKeyAuth:    &namespacev1.ApiKeyAuthSpec{Enabled: true},
		Lifecycle:     &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
		Fairness:      &namespacev1.FairnessSpec{TaskQueueFairnessEnabled: true},
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
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	mockPrompter.EXPECT().
		PromptApply(&namespacev1.NamespaceSpec{}, specMatcher(expectedSpec), false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), createReqMatcher(expectedSpec)).
		Return(defaultCreateResponse, nil)

	mockHandler.EXPECT().
		HandleOperation(defaultCreateResponse.AsyncOperation, "my-namespace.my-account").
		Return(nil)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:                    "my-namespace",
		Regions:                 []string{"aws-us-east-1"},
		RetentionDays:           30,
		ApiKeyAuthEnabled:       true,
		EnableTaskQueueFairness: true,
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
		OperationHandler:   mockHandler,
	})
	require.NoError(t, err)
}

// TestCreateNamespace_Error verifies that an API error is forwarded to HandleCreateErr.
func TestCreateNamespace_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("create failed")

	spec := baseNamespaceSpec()

	mockPrompter.EXPECT().
		PromptApply(&namespacev1.NamespaceSpec{}, specMatcher(spec), false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), createReqMatcher(spec)).
		Return(nil, apiErr)

	mockHandler.EXPECT().
		HandleCreateErr(apiErr).
		Return(apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		OperationHandler:   mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestCreateNamespace_PromptDeclined verifies that the API is never called when the prompt is declined.
func TestCreateNamespace_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := fmt.Errorf("Aborting apply.")

	spec := baseNamespaceSpec()

	mockPrompter.EXPECT().
		PromptApply(&namespacev1.NamespaceSpec{}, specMatcher(spec), false).
		Return(promptErr)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		OperationHandler:   mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestCreateNamespace_HandleOperationError verifies that an error from HandleOperation propagates.
func TestCreateNamespace_HandleOperationError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	handleErr := errors.New("polling failed")

	spec := baseNamespaceSpec()

	mockPrompter.EXPECT().
		PromptApply(&namespacev1.NamespaceSpec{}, specMatcher(spec), false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespace(context.Background(), createReqMatcher(spec)).
		Return(defaultCreateResponse, nil)

	mockHandler.EXPECT().
		HandleOperation(defaultCreateResponse.AsyncOperation, "my-namespace.my-account").
		Return(handleErr)

	var buf bytes.Buffer
	err := temporalcloudcli.CreateNamespace(context.Background(), temporalcloudcli.CreateNamespaceParams{
		Name:               "my-namespace",
		Regions:            []string{"aws-us-east-1"},
		Cloud:              mockCloud,
		Printer:            &printer.Printer{Output: &buf, JSON: true},
		Prompter:           mockPrompter,
		UnmarshalProtoJSON: noopUnmarshalProtoJSON,
		OperationHandler:   mockHandler,
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
			tt.params.OperationHandler = cmdmock.NewMockAsyncOperationHandler(t)

			err := temporalcloudcli.CreateNamespace(context.Background(), tt.params)
			require.Error(t, err)
			tt.assertError(t, err)
		})
	}
}
