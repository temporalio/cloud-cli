package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

var testHANamespace = &namespacev1.Namespace{
	Namespace:       "my-ns.my-account",
	ResourceVersion: "rv-1",
	ActiveRegion:    "aws-us-east-1",
	Spec: &namespacev1.NamespaceSpec{
		HighAvailability: &namespacev1.HighAvailabilitySpec{
			DisableManagedFailover: false,
		},
	},
	RegionStatus: map[string]*namespacev1.NamespaceRegionStatus{
		"aws-us-east-1": {State: namespacev1.NamespaceRegionStatus_STATE_ACTIVE},
		"aws-us-west-2": {State: namespacev1.NamespaceRegionStatus_STATE_PASSIVE},
	},
}

// --- HaGet ---

func TestHaGet(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceHaGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNamespaceHaGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceHaGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- HaUpdate ---

func TestHaUpdate_Confirmed(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaUpdateCommand{
		NamespaceOptions:    temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		DisableAutoFailover: true,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.Namespace == "my-ns.my-account" &&
						req.ResourceVersion == "rv-1" &&
						req.Spec.HighAvailability.DisableManagedFailover == true
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-update"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update"},
	})
}

func TestHaUpdate_ResourceVersionOverride(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaUpdateCommand{
		NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-custom"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.ResourceVersion == "rv-custom"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-rv"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-rv"},
	})
}

func TestHaUpdate_Declined(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaUpdateCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
		ExpectedError: "Aborting update.",
	})
}

func TestHaUpdate_GetNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaUpdateCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("not found"))
		},
		ExpectedError: "not found",
	})
}

func TestHaUpdate_UpdateError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaUpdateCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("update failed"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		ExpectedError: "update failed",
	})
}

// --- HaFailover ---

func TestHaFailover_Confirmed(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaFailoverCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				FailoverNamespaceRegion(mock.Anything, &cloudservice.FailoverNamespaceRegionRequest{
					Namespace: "my-ns.my-account",
					Region:    "aws-us-west-2",
				}, mock.Anything).
				Return(&cloudservice.FailoverNamespaceRegionResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-failover"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-failover"},
	})
}

func TestHaFailover_Declined(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaFailoverCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
		ExpectedError: "Aborting failover.",
	})
}

func TestHaFailover_GetNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaFailoverCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("namespace not found"))
		},
		ExpectedError: "namespace not found",
	})
}

func TestHaFailover_FailoverError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaFailoverCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				FailoverNamespaceRegion(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("failover unavailable"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		ExpectedError: "failover unavailable",
	})
}

// --- HaRegionList ---

func TestHaRegionList(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceHaRegionListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNamespaceHaRegionListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceHaRegionListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- HaRegionAdd ---

func TestHaRegionAdd_Confirmed(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionAddCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-eu-west-1",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				AddNamespaceRegion(mock.Anything, &cloudservice.AddNamespaceRegionRequest{
					Namespace:       "my-ns.my-account",
					Region:          "aws-eu-west-1",
					ResourceVersion: "rv-1",
				}, mock.Anything).
				Return(&cloudservice.AddNamespaceRegionResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-add"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-add"},
	})
}

func TestHaRegionAdd_ResourceVersionOverride(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionAddCommand{
		NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-custom"},
		Region:                 "aws-eu-west-1",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				AddNamespaceRegion(mock.Anything, mock.MatchedBy(func(req *cloudservice.AddNamespaceRegionRequest) bool {
					return req.ResourceVersion == "rv-custom"
				}), mock.Anything).
				Return(&cloudservice.AddNamespaceRegionResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-add-rv"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-add-rv"},
	})
}

func TestHaRegionAdd_Declined(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionAddCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-eu-west-1",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
		ExpectedError: "Aborting create.",
	})
}

func TestHaRegionAdd_GetNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionAddCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-eu-west-1",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("namespace not found"))
		},
		ExpectedError: "namespace not found",
	})
}

func TestHaRegionAdd_AddError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionAddCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-eu-west-1",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				AddNamespaceRegion(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("add failed"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		ExpectedError: "add failed",
	})
}

// --- HaRegionDelete ---

func TestHaRegionDelete_Confirmed(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				DeleteNamespaceRegion(mock.Anything, &cloudservice.DeleteNamespaceRegionRequest{
					Namespace:       "my-ns.my-account",
					Region:          "aws-us-west-2",
					ResourceVersion: "rv-1",
				}, mock.Anything).
				Return(&cloudservice.DeleteNamespaceRegionResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
	})
}

func TestHaRegionDelete_ResourceVersionOverride(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionDeleteCommand{
		NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-custom"},
		Region:                 "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				DeleteNamespaceRegion(mock.Anything, mock.MatchedBy(func(req *cloudservice.DeleteNamespaceRegionRequest) bool {
					return req.ResourceVersion == "rv-custom"
				}), mock.Anything).
				Return(&cloudservice.DeleteNamespaceRegionResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-del-rv"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del-rv"},
	})
}

func TestHaRegionDelete_Declined(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
		ExpectedError: "Aborting delete.",
	})
}

func TestHaRegionDelete_GetNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("namespace not found"))
		},
		ExpectedError: "namespace not found",
	})
}

func TestHaRegionDelete_DeleteError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceHaRegionDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-account"},
		Region:           "aws-us-west-2",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testHANamespace}, nil)
			c.EXPECT().
				DeleteNamespaceRegion(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("delete failed"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		ExpectedError: "delete failed",
	})
}
