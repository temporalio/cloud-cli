package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	nexusv1 "go.temporal.io/cloud-sdk/api/nexus/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

var testEndpoint = &nexusv1.Endpoint{
	Id:              "ep-123",
	ResourceVersion: "v1",
	Spec: &nexusv1.EndpointSpec{
		Name: "my-endpoint",
		TargetSpec: &nexusv1.EndpointTargetSpec{
			Variant: &nexusv1.EndpointTargetSpec_WorkerTargetSpec{
				WorkerTargetSpec: &nexusv1.WorkerTargetSpec{
					NamespaceId: "ns-123",
					TaskQueue:   "my-task-queue",
				},
			},
		},
	},
}

// --- GetNexusEndpoint ---

func TestGetNexusEndpoint(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "SuccessEndpointFound",
			cmd:  temporalcloudcli.CloudNexusEndpointGetCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "my-endpoint",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
			},
			expectedJsonOutput: testEndpoint,
		},
		{
			name: "NotFound",
			cmd:  temporalcloudcli.CloudNexusEndpointGetCommand{Name: "missing"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "missing",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{}, nil)
			},
			expectedErr: `endpoint "missing" not found`,
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudNexusEndpointGetCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: "connection refused",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

// --- ListNexusEndpoints ---

func TestListNexusEndpoints(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "SuccessSingleEndpoint",
			cmd:  temporalcloudcli.CloudNexusEndpointListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
			},
			expectedJsonOutput: struct {
				Endpoints     []*nexusv1.Endpoint
				NextPageToken string
			}{
				Endpoints: []*nexusv1.Endpoint{testEndpoint},
			},
		},
		{
			name: "WithPageSizeAndToken",
			cmd: temporalcloudcli.CloudNexusEndpointListCommand{
				PageSize:  10,
				PageToken: "tok-abc",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						PageSize:  10,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints:     []*nexusv1.Endpoint{testEndpoint},
						NextPageToken: "tok-def",
					}, nil)
			},
			expectedJsonOutput: struct {
				Endpoints     []*nexusv1.Endpoint
				NextPageToken string
			}{
				Endpoints:     []*nexusv1.Endpoint{testEndpoint},
				NextPageToken: "tok-def",
			},
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudNexusEndpointListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: "connection refused",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

// --- CreateNexusEndpoint ---

func TestCreateNexusEndpoint(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointCreateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		promptOptions           temporalcloudcli.TestPromptOptions
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNexusEndpointCreateCommand{
				Name:            "my-endpoint",
				TargetNamespace: "ns-123",
				TargetTaskQueue: "my-tq",
				AllowNamespace:  []string{"caller-ns"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.CreateNexusEndpointRequest) bool {
						return req.Spec.Name == "my-endpoint" &&
							req.Spec.TargetSpec.GetWorkerTargetSpec().NamespaceId == "ns-123" &&
							req.Spec.TargetSpec.GetWorkerTargetSpec().TaskQueue == "my-tq" &&
							len(req.Spec.PolicySpecs) == 1 &&
							req.Spec.PolicySpecs[0].GetAllowedCloudNamespacePolicySpec().NamespaceId == "caller-ns"
					}), mock.Anything).
					Return(&cloudservice.CreateNexusEndpointResponse{
						EndpointId:     "ep-new",
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedJsonOutput: &cloudservice.CreateNexusEndpointResponse{
				EndpointId: "ep-new",
				AsyncOperation: &operation.AsyncOperation{
					Id:    "op-1",
					State: operation.AsyncOperation_STATE_FULFILLED,
				},
			},
		},
		{
			name: "UserDeclines",
			cmd: temporalcloudcli.CloudNexusEndpointCreateCommand{
				Name:            "my-endpoint",
				TargetNamespace: "ns-123",
				TargetTaskQueue: "my-tq",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name: "MutuallyExclusiveDescription",
			cmd: temporalcloudcli.CloudNexusEndpointCreateCommand{
				Name:            "my-endpoint",
				TargetNamespace: "ns-123",
				TargetTaskQueue: "my-tq",
				Description:     "inline desc",
				DescriptionFile: "/some/file",
			},
			expectedErr: "mutually exclusive",
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudNexusEndpointCreateCommand{
				Name:            "my-endpoint",
				TargetNamespace: "ns-123",
				TargetTaskQueue: "my-tq",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateNexusEndpoint(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("permission denied"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "permission denied",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				PromptOptions:           tt.promptOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

// --- DeleteNexusEndpoint ---

func TestDeleteNexusEndpoint(t *testing.T) {
	tt := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		pollerOptions           temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "SuccessDeleteEndpoint",
			cmd:  temporalcloudcli.CloudNexusEndpointDeleteCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "my-endpoint",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
				c.EXPECT().
					DeleteNexusEndpoint(mock.Anything, &cloudservice.DeleteNexusEndpointRequest{
						EndpointId:      testEndpoint.Id,
						ResourceVersion: "v1",
					}, mock.Anything).
					Return(&cloudservice.DeleteNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			pollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "NotFound",
			cmd:  temporalcloudcli.CloudNexusEndpointDeleteCommand{Name: "missing"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "missing",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{}, nil)
			},
			expectedErr: `endpoint "missing" not found`,
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudNexusEndpointDeleteCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: "connection refused",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudNexusEndpointDeleteCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "my-endpoint",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNexusEndpointDeleteCommand{
				Name:                   "my-endpoint",
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "my-endpoint",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
				c.EXPECT().
					DeleteNexusEndpoint(mock.Anything, &cloudservice.DeleteNexusEndpointRequest{
						EndpointId:      testEndpoint.Id,
						ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.DeleteNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			pollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "DeleteAPIError",
			cmd:  temporalcloudcli.CloudNexusEndpointDeleteCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "my-endpoint",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
				c.EXPECT().
					DeleteNexusEndpoint(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("delete failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPromptYes:        true,
				ExpectPromptYesMessage: "Delete",
				PromptResult:           true,
			},
			expectedErr: "delete operation failed",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tc.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tc.cloudClientExpectations,
				PromptOptions:           tc.promptOptions,
				AsyncPollerOptions:      tc.pollerOptions,
				ExpectedError:           tc.expectedErr,
			})
		})
	}
}

// --- UpdateNexusEndpoint ---

// TestUpdateNexusEndpoint uses table-driven tests for the update nexus endpoint command.
// AIDEV-NOTE: We initialize the cobra FlagSet manually in setupCmd so Changed() returns true for
// explicitly set flags, since TestCommand calls run() directly without going through cobra's flag parsing.
func TestUpdateNexusEndpoint(t *testing.T) {
	existingSpec := &nexusv1.EndpointSpec{
		Name: "my-endpoint",
		TargetSpec: &nexusv1.EndpointTargetSpec{
			Variant: &nexusv1.EndpointTargetSpec_WorkerTargetSpec{
				WorkerTargetSpec: &nexusv1.WorkerTargetSpec{
					NamespaceId: "ns-123",
					TaskQueue:   "my-task-queue",
				},
			},
		},
	}
	existingEndpoint := &nexusv1.Endpoint{
		Id:              "ep-123",
		ResourceVersion: "v1",
		Spec:            existingSpec,
	}

	getEndpointExpectation := func(c *cloudmock.MockCloudServiceClient) {
		c.EXPECT().
			GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
				Name: "my-endpoint",
			}, mock.Anything).
			Return(&cloudservice.GetNexusEndpointsResponse{
				Endpoints: []*nexusv1.Endpoint{existingEndpoint},
			}, nil)
	}

	tests := []struct {
		name                    string
		setupCmd                func(*temporalcloudcli.CloudNexusEndpointUpdateCommand)
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "TargetNamespace",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.TargetNamespace, "target-namespace", "", "")
				require.NoError(t, cmd.Command.Flags().Set("target-namespace", "new-ns"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				getEndpointExpectation(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.EndpointId == "ep-123" &&
							req.Spec.TargetSpec.GetWorkerTargetSpec().NamespaceId == "new-ns" &&
							req.Spec.TargetSpec.GetWorkerTargetSpec().TaskQueue == "my-task-queue" &&
							req.ResourceVersion == "v1"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "TargetTaskQueue",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.TargetTaskQueue, "target-task-queue", "", "")
				require.NoError(t, cmd.Command.Flags().Set("target-task-queue", "new-tq"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				getEndpointExpectation(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.EndpointId == "ep-123" &&
							req.Spec.TargetSpec.GetWorkerTargetSpec().NamespaceId == "ns-123" &&
							req.Spec.TargetSpec.GetWorkerTargetSpec().TaskQueue == "new-tq" &&
							req.ResourceVersion == "v1"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "Description",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.Description = "new description"
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				getEndpointExpectation(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.EndpointId == "ep-123" &&
							req.Spec.Description != nil &&
							req.ResourceVersion == "v1"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "UnsetDescription",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.UnsetDescription = true
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				getEndpointExpectation(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.EndpointId == "ep-123" &&
							req.Spec.Description == nil &&
							req.ResourceVersion == "v1"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "UnsetDescriptionWithDescription",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.UnsetDescription = true
				cmd.Description = "some desc"
			},
			expectedErr: "--unset-description cannot be used with --description or --description-file",
		},
		{
			name: "UnsetDescriptionWithDescriptionFile",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.UnsetDescription = true
				cmd.DescriptionFile = "/some/file"
			},
			expectedErr: "--unset-description cannot be used with --description or --description-file",
		},
		{
			name: "MutuallyExclusiveDescription",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.Description = "inline"
				cmd.DescriptionFile = "/some/file"
			},
			expectedErr: "mutually exclusive",
		},
		{
			name: "NoChanges",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				getEndpointExpectation(c)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPrompApply: true,
				PromptError:      errors.New("Aborting apply."),
			},
			expectedErr: "Aborting apply.",
		},
		{
			name: "GetError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("get error"))
			},
			expectedErr: "get error",
		},
		{
			name: "NotFound",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "my-endpoint",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{}, nil)
			},
			expectedErr: `endpoint "my-endpoint" not found`,
		},
		{
			name: "ResourceVersionOverride",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.ResourceVersion = "rv-override"
				cmd.Command.Flags().StringVar(&cmd.TargetNamespace, "target-namespace", "", "")
				require.NoError(t, cmd.Command.Flags().Set("target-namespace", "new-ns"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				getEndpointExpectation(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.ResourceVersion == "rv-override"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "UpdateError",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.TargetNamespace, "target-namespace", "", "")
				require.NoError(t, cmd.Command.Flags().Set("target-namespace", "new-ns"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				getEndpointExpectation(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update operation failed",
		},
		{
			name: "PromptDeclined",
			setupCmd: func(cmd *temporalcloudcli.CloudNexusEndpointUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.TargetNamespace, "target-namespace", "", "")
				require.NoError(t, cmd.Command.Flags().Set("target-namespace", "new-ns"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				getEndpointExpectation(c)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting update.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNexusEndpointUpdateCommand{Name: "my-endpoint"}
			if tt.setupCmd != nil {
				tt.setupCmd(&cmd)
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- Allowed Namespace Commands ---

// newTestEndpointWithPolicies returns a fresh endpoint with existing allowed namespace policy specs.
// A fresh instance is needed per test because add/set/remove commands modify the endpoint in-place.
func newTestEndpointWithPolicies() *nexusv1.Endpoint {
	return &nexusv1.Endpoint{
		Id:              "ep-123",
		ResourceVersion: "v1",
		Spec: &nexusv1.EndpointSpec{
			Name: "my-endpoint",
			TargetSpec: &nexusv1.EndpointTargetSpec{
				Variant: &nexusv1.EndpointTargetSpec_WorkerTargetSpec{
					WorkerTargetSpec: &nexusv1.WorkerTargetSpec{
						NamespaceId: "ns-123",
						TaskQueue:   "my-task-queue",
					},
				},
			},
			PolicySpecs: []*nexusv1.EndpointPolicySpec{
				{
					Variant: &nexusv1.EndpointPolicySpec_AllowedCloudNamespacePolicySpec{
						AllowedCloudNamespacePolicySpec: &nexusv1.AllowedCloudNamespacePolicySpec{
							NamespaceId: "caller-ns-1",
						},
					},
				},
				{
					Variant: &nexusv1.EndpointPolicySpec_AllowedCloudNamespacePolicySpec{
						AllowedCloudNamespacePolicySpec: &nexusv1.AllowedCloudNamespacePolicySpec{
							NamespaceId: "caller-ns-2",
						},
					},
				},
			},
		},
	}
}

func expectGetEndpointWithPolicies(c *cloudmock.MockCloudServiceClient) {
	c.EXPECT().
		GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
			Name: "my-endpoint",
		}, mock.Anything).
		Return(&cloudservice.GetNexusEndpointsResponse{
			Endpoints: []*nexusv1.Endpoint{newTestEndpointWithPolicies()},
		}, nil)
}

func TestListNexusEndpointAllowedNamespaces(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointAllowedNamespaceListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNexusEndpointAllowedNamespaceListCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
			},
			expectedJsonOutput: struct {
				Namespaces []string
			}{
				Namespaces: []string{"caller-ns-1", "caller-ns-2"},
			},
		},
		{
			name: "EmptyList",
			cmd:  temporalcloudcli.CloudNexusEndpointAllowedNamespaceListCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "my-endpoint",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
			},
			expectedJsonOutput: struct {
				Namespaces []string
			}{
				Namespaces: []string{},
			},
		},
		{
			name: "NotFound",
			cmd:  temporalcloudcli.CloudNexusEndpointAllowedNamespaceListCommand{Name: "missing"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "missing",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{}, nil)
			},
			expectedErr: `endpoint "missing" not found`,
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudNexusEndpointAllowedNamespaceListCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: "connection refused",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

func TestAddNexusEndpointAllowedNamespace(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointAllowedNamespaceAddCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceAddCommand{
				Name:      "my-endpoint",
				Namespace: []string{"new-ns"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.EndpointId == "ep-123" &&
							len(req.Spec.PolicySpecs) == 3 &&
							req.Spec.PolicySpecs[2].GetAllowedCloudNamespacePolicySpec().NamespaceId == "new-ns" &&
							req.ResourceVersion == "v1"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-add"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-add"},
		},
		{
			name: "UserDeclines",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceAddCommand{
				Name:      "my-endpoint",
				Namespace: []string{"new-ns"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting add.",
		},
		{
			name: "NotFound",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceAddCommand{
				Name:      "missing",
				Namespace: []string{"new-ns"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "missing",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{}, nil)
			},
			expectedErr: `endpoint "missing" not found`,
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceAddCommand{
				Name:      "my-endpoint",
				Namespace: []string{"new-ns"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: "connection refused",
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceAddCommand{
				Name:                   "my-endpoint",
				Namespace:              []string{"new-ns"},
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.ResourceVersion == "rv-override"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-add"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-add"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

func TestSetNexusEndpointAllowedNamespace(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointAllowedNamespaceSetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceSetCommand{
				Name:      "my-endpoint",
				Namespace: []string{"new-ns-1", "new-ns-2"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.EndpointId == "ep-123" &&
							len(req.Spec.PolicySpecs) == 2 &&
							req.Spec.PolicySpecs[0].GetAllowedCloudNamespacePolicySpec().NamespaceId == "new-ns-1" &&
							req.Spec.PolicySpecs[1].GetAllowedCloudNamespacePolicySpec().NamespaceId == "new-ns-2" &&
							req.ResourceVersion == "v1"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-set"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-set"},
		},
		{
			name: "UserDeclines",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceSetCommand{
				Name:      "my-endpoint",
				Namespace: []string{"new-ns-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting set.",
		},
		{
			name: "NotFound",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceSetCommand{
				Name:      "missing",
				Namespace: []string{"new-ns"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "missing",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{}, nil)
			},
			expectedErr: `endpoint "missing" not found`,
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceSetCommand{
				Name:      "my-endpoint",
				Namespace: []string{"new-ns"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: "connection refused",
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceSetCommand{
				Name:                   "my-endpoint",
				Namespace:              []string{"new-ns"},
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.ResourceVersion == "rv-override"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-set"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-set"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

func TestRemoveNexusEndpointAllowedNamespace(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointAllowedNamespaceRemoveCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceRemoveCommand{
				Name:      "my-endpoint",
				Namespace: []string{"caller-ns-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.EndpointId == "ep-123" &&
							len(req.Spec.PolicySpecs) == 1 &&
							req.Spec.PolicySpecs[0].GetAllowedCloudNamespacePolicySpec().NamespaceId == "caller-ns-2" &&
							req.ResourceVersion == "v1"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-rm"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-rm"},
		},
		{
			name: "UserDeclines",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceRemoveCommand{
				Name:      "my-endpoint",
				Namespace: []string{"caller-ns-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting remove.",
		},
		{
			name: "NotFound",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceRemoveCommand{
				Name:      "missing",
				Namespace: []string{"caller-ns-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "missing",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{}, nil)
			},
			expectedErr: `endpoint "missing" not found`,
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceRemoveCommand{
				Name:      "my-endpoint",
				Namespace: []string{"caller-ns-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: "connection refused",
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNexusEndpointAllowedNamespaceRemoveCommand{
				Name:                   "my-endpoint",
				Namespace:              []string{"caller-ns-1"},
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectGetEndpointWithPolicies(c)
				c.EXPECT().
					UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNexusEndpointRequest) bool {
						return req.ResourceVersion == "rv-override"
					}), mock.Anything).
					Return(&cloudservice.UpdateNexusEndpointResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-rm"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-rm"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
