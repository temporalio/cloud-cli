package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// --- HaGet ---

func TestHaGet(t *testing.T) {
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
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:    "my-ns",
							ActiveRegion: "us-west-2",
							Spec: &namespacev1.NamespaceSpec{
								HighAvailability: &namespacev1.HighAvailabilitySpec{
									DisableManagedFailover:         false,
									DisablePassivePollerForwarding: true,
								},
							},
						},
					}, nil)
			},
			expectedJsonOutput: struct {
				Namespace                      string
				ActiveRegion                   string
				ManagedFailoverEnabled         bool
				PassivePollerForwardingEnabled bool
			}{
				Namespace:                      "my-ns",
				ActiveRegion:                   "us-west-2",
				ManagedFailoverEnabled:         true,
				PassivePollerForwardingEnabled: false,
			},
		},
		{
			name: "GetNamespaceError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaGetCommand{
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

// --- HaUpdate ---

// TestHaUpdate uses table-driven tests for the ha update command.
// AIDEV-NOTE: We initialize the cobra FlagSet manually in setupCmd so Changed() returns true for
// explicitly set flags, since TestCommand calls run() directly without going through cobra's flag parsing.
func TestHaUpdate(t *testing.T) {
	baseNS := func(disableManagedFailover, disablePassivePollerForwarding bool) *namespacev1.Namespace {
		return &namespacev1.Namespace{
			Namespace:       "my-ns",
			ResourceVersion: "rv-1",
			Spec: &namespacev1.NamespaceSpec{
				HighAvailability: &namespacev1.HighAvailabilitySpec{
					DisableManagedFailover:         disableManagedFailover,
					DisablePassivePollerForwarding: disablePassivePollerForwarding,
				},
			},
		}
	}

	tests := []struct {
		name                    string
		setupCmd                func(*temporalcloudcli.CloudNamespaceHaUpdateCommand)
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "DisableAutoFailover",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceHaUpdateCommand) {
				cmd.Command.Flags().BoolVar(&cmd.DisableAutoFailover, "disable-auto-failover", false, "")
				require.NoError(t, cmd.Command.Flags().Set("disable-auto-failover", "true"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(false, false)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
						Namespace:       "my-ns",
						ResourceVersion: "rv-1",
						Spec: &namespacev1.NamespaceSpec{
							HighAvailability: &namespacev1.HighAvailabilitySpec{
								DisableManagedFailover:         true,
								DisablePassivePollerForwarding: false,
							},
						},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name: "EnableAutoFailover",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceHaUpdateCommand) {
				cmd.Command.Flags().BoolVar(&cmd.DisableAutoFailover, "disable-auto-failover", false, "")
				require.NoError(t, cmd.Command.Flags().Set("disable-auto-failover", "false"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(true, false)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
						Namespace:       "my-ns",
						ResourceVersion: "rv-1",
						Spec: &namespacev1.NamespaceSpec{
							HighAvailability: &namespacev1.HighAvailabilitySpec{
								DisableManagedFailover:         false,
								DisablePassivePollerForwarding: false,
							},
						},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name: "DisablePassivePollerForwarding",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceHaUpdateCommand) {
				cmd.Command.Flags().BoolVar(&cmd.DisablePassivePollerForwarding, "disable-passive-poller-forwarding", false, "")
				require.NoError(t, cmd.Command.Flags().Set("disable-passive-poller-forwarding", "true"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(false, false)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
						Namespace:       "my-ns",
						ResourceVersion: "rv-1",
						Spec: &namespacev1.NamespaceSpec{
							HighAvailability: &namespacev1.HighAvailabilitySpec{
								DisableManagedFailover:         false,
								DisablePassivePollerForwarding: true,
							},
						},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name: "EnablePassivePollerForwarding",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceHaUpdateCommand) {
				cmd.Command.Flags().BoolVar(&cmd.DisablePassivePollerForwarding, "disable-passive-poller-forwarding", false, "")
				require.NoError(t, cmd.Command.Flags().Set("disable-passive-poller-forwarding", "false"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(false, true)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
						Namespace:       "my-ns",
						ResourceVersion: "rv-1",
						Spec: &namespacev1.NamespaceSpec{
							HighAvailability: &namespacev1.HighAvailabilitySpec{
								DisableManagedFailover:         false,
								DisablePassivePollerForwarding: false,
							},
						},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name: "BothFlags",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceHaUpdateCommand) {
				cmd.Command.Flags().BoolVar(&cmd.DisableAutoFailover, "disable-auto-failover", false, "")
				cmd.Command.Flags().BoolVar(&cmd.DisablePassivePollerForwarding, "disable-passive-poller-forwarding", false, "")
				require.NoError(t, cmd.Command.Flags().Set("disable-auto-failover", "true"))
				require.NoError(t, cmd.Command.Flags().Set("disable-passive-poller-forwarding", "true"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(false, false)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
						Namespace:       "my-ns",
						ResourceVersion: "rv-1",
						Spec: &namespacev1.NamespaceSpec{
							HighAvailability: &namespacev1.HighAvailabilitySpec{
								DisableManagedFailover:         true,
								DisablePassivePollerForwarding: true,
							},
						},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name:        "NoChanges",
			expectedErr: "no changes specified",
		},
		{
			name: "GetNamespaceError",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceHaUpdateCommand) {
				cmd.Command.Flags().BoolVar(&cmd.DisableAutoFailover, "disable-auto-failover", false, "")
				require.NoError(t, cmd.Command.Flags().Set("disable-auto-failover", "true"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("get error"))
			},
			expectedErr: "get error",
		},
		{
			name: "UpdateNamespaceError",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceHaUpdateCommand) {
				cmd.Command.Flags().BoolVar(&cmd.DisableAutoFailover, "disable-auto-failover", false, "")
				require.NoError(t, cmd.Command.Flags().Set("disable-auto-failover", "true"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: baseNS(false, false)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update error"))
			},
			expectedErr: "update error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &temporalcloudcli.CloudNamespaceHaUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns"},
			}
			if tt.setupCmd != nil {
				tt.setupCmd(cmd)
			}
			temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
