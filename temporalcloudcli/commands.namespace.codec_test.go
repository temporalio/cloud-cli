package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

var testCodecNamespace = "test-namespace.test-account"

var testNamespaceWithCodec = &namespacev1.Namespace{
	Namespace:       testCodecNamespace,
	ResourceVersion: "rv-1",
	Spec: &namespacev1.NamespaceSpec{
		CodecServer: &namespacev1.CodecServerSpec{
			Endpoint:        "https://codec.example.com",
			PassAccessToken: true,
		},
	},
}

var testNamespaceWithoutCodec = &namespacev1.Namespace{
	Namespace:       testCodecNamespace,
	ResourceVersion: "rv-1",
	Spec:            &namespacev1.NamespaceSpec{},
}

// --- GetCodecServer ---

func TestGetCodecServer_WithCodecServer(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecGetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testCodecNamespace}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithCodec}, nil)
		},
		JSONOutput:         true,
		ExpectedOutputJson: struct {
			Namespace string
			Spec      *namespacev1.CodecServerSpec
		}{
			Namespace: testCodecNamespace,
			Spec:      testNamespaceWithCodec.Spec.CodecServer,
		},
	})
}

func TestGetCodecServer_NoCodecServer(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecGetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testCodecNamespace}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithoutCodec}, nil)
		},
		JSONOutput: true,
		ExpectedOutputJson: struct {
			Namespace string
			Spec      *namespacev1.CodecServerSpec
		}{
			Namespace: testCodecNamespace,
			Spec:      nil,
		},
	})
}

func TestGetCodecServer_Error(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecGetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("API error"))
		},
		ExpectedError: "API error",
	})
}

// --- SetCodecServer ---

func TestSetCodecServer_Success(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-set-codec"}

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecSetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
		Endpoint:         "https://codec.example.com",
		PassAccessToken:  true,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testCodecNamespace}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithoutCodec}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.Namespace == testCodecNamespace &&
						req.Spec.CodecServer != nil &&
						req.Spec.CodecServer.Endpoint == "https://codec.example.com" &&
						req.Spec.CodecServer.PassAccessToken == true &&
						req.ResourceVersion == "rv-1"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: op}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: op.Id},
	})
}

func TestSetCodecServer_WithCustomErrorMessage(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-set-codec-custom"}

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecSetCommand{
		NamespaceOptions:                 temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
		Endpoint:                         "https://codec.example.com",
		CustomErrorMessageDefaultMessage: "Codec unavailable",
		CustomErrorMessageDefaultLink:    "https://docs.example.com",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithoutCodec}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					cs := req.Spec.GetCodecServer()
					if cs == nil || cs.CustomErrorMessage == nil || cs.CustomErrorMessage.Default == nil {
						return false
					}
					return cs.CustomErrorMessage.Default.Message == "Codec unavailable" &&
						cs.CustomErrorMessage.Default.Link == "https://docs.example.com"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: op}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: op.Id},
	})
}

func TestSetCodecServer_ResourceVersionOverride(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-set-codec-rv"}

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecSetCommand{
		NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
		Endpoint:               "https://codec.example.com",
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithoutCodec}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.ResourceVersion == "rv-override"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: op}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: op.Id},
	})
}

func TestSetCodecServer_GetNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecSetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
		Endpoint:         "https://codec.example.com",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("get namespace failed"))
		},
		ExpectedError: "get namespace failed",
	})
}

func TestSetCodecServer_UpdateNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecSetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
		Endpoint:         "https://codec.example.com",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithoutCodec}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("update failed"))
		},
		ExpectedError: "update failed",
	})
}

func TestSetCodecServer_UpdateNamespaceNothingToChange(t *testing.T) {
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecSetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
		Endpoint:         "https://codec.example.com",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithoutCodec}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, nothingToChangeErr)
		},
		ExpectedError: "nothing to change",
	})
}

// --- DeleteCodecServer ---

func TestDeleteCodecServer_Success(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-del-codec"}

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
	}, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testCodecNamespace}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithCodec}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.Namespace == testCodecNamespace &&
						req.Spec.CodecServer == nil &&
						req.ResourceVersion == "rv-1"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: op}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: op.Id},
	})
}

func TestDeleteCodecServer_UserDeclinesPrompt(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
	}, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    false,
		},
		ExpectedError: "Aborting delete.",
	})
}

func TestDeleteCodecServer_PromptError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
	}, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptError:     errors.New("prompt failed"),
		},
		ExpectedError: "prompt failed",
	})
}

func TestDeleteCodecServer_GetNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
	}, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("get namespace failed"))
		},
		ExpectedError: "get namespace failed",
	})
}

func TestDeleteCodecServer_UpdateNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
	}, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithCodec}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("update failed"))
		},
		ExpectedError: "update failed",
	})
}

func TestDeleteCodecServer_ResourceVersionOverride(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-del-codec-rv"}

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCodecDeleteCommand{
		NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: testCodecNamespace},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
	}, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithCodec}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.ResourceVersion == "rv-override"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: op}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: op.Id},
	})
}
