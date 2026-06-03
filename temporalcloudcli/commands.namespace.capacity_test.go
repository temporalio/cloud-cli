package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cliext "github.com/temporalio/cli/cliext"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// --- CapacityGet ---

func TestCapacityGet(t *testing.T) {
	capacityInfo := &namespacev1.NamespaceCapacityInfo{
		Namespace: "my-ns",
		CurrentCapacity: &namespacev1.Capacity{
			CurrentMode: &namespacev1.Capacity_Provisioned_{
				Provisioned: &namespacev1.Capacity_Provisioned{CurrentValue: 100},
			},
		},
	}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedJsonOutput      any
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceCapacityInfo(mock.Anything, &cloudservice.GetNamespaceCapacityInfoRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceCapacityInfoResponse{CapacityInfo: capacityInfo}, nil)
			},
			expectedJsonOutput: capacityInfo,
		},
		{
			name: "GetError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceCapacityInfo(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCapacityGetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns"},
			}, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

// --- CapacityUpdate ---

func TestCapacityUpdate(t *testing.T) {
	baseNS := func(spec *namespacev1.CapacitySpec) *namespacev1.Namespace {
		return &namespacev1.Namespace{
			Namespace:       "my-ns",
			ResourceVersion: "rv-1",
			Spec: &namespacev1.NamespaceSpec{
				RetentionDays: 14,
				CapacitySpec:  spec,
			},
		}
	}
	onDemandSpec := &namespacev1.CapacitySpec{
		Spec: &namespacev1.CapacitySpec_OnDemand_{OnDemand: &namespacev1.CapacitySpec_OnDemand{}},
	}
	provisioned100 := &namespacev1.CapacitySpec{
		Spec: &namespacev1.CapacitySpec_Provisioned_{
			Provisioned: &namespacev1.CapacitySpec_Provisioned{Value: 100},
		},
	}

	tests := []struct {
		name                    string
		mode                    string
		value                   float32
		resourceVersionOverride string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "OnDemand",
			mode: "on_demand",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(provisioned100)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
						Namespace:       "my-ns",
						ResourceVersion: "rv-1",
						Spec: &namespacev1.NamespaceSpec{
							RetentionDays: 14,
							CapacitySpec:  onDemandSpec,
						},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name:  "Provisioned",
			mode:  "provisioned",
			value: 100,
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(nil)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
						Namespace:       "my-ns",
						ResourceVersion: "rv-1",
						Spec: &namespacev1.NamespaceSpec{
							RetentionDays: 14,
							CapacitySpec:  provisioned100,
						},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name:        "ProvisionedRequiresPositiveValue",
			mode:        "provisioned",
			value:       0,
			expectedErr: "--capacity-value must be greater than 0 when --capacity-mode is 'provisioned'",
		},
		{
			name: "PromptDeclined",
			mode: "on_demand",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(provisioned100)}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting update.",
		},
		{
			name: "GetNamespaceError",
			mode: "on_demand",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
		{
			name: "UpdateNamespaceError",
			mode: "on_demand",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(provisioned100)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update operation failed",
		},
		{
			name:                    "ResourceVersionOverride",
			mode:                    "on_demand",
			resourceVersionOverride: "rv-override",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(provisioned100)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
						Namespace:       "my-ns",
						ResourceVersion: "rv-override",
						Spec: &namespacev1.NamespaceSpec{
							RetentionDays: 14,
							CapacitySpec:  onDemandSpec,
						},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceCapacityUpdateCommand{
				NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "my-ns"},
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: tt.resourceVersionOverride},
				CapacityMode:           cliext.NewFlagStringEnum([]string{"on_demand", "provisioned"}, tt.mode),
				CapacityValue:          tt.value,
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
