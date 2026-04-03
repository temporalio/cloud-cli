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

// testNamespaceWithFilters returns a namespace with the given cert filters.
func testNamespaceWithFilters(filters []*namespacev1.CertificateFilterSpec) *namespacev1.Namespace {
	return &namespacev1.Namespace{
		Namespace:       "test-namespace.test-account",
		ResourceVersion: "rv-1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: filters,
			},
		},
	}
}

// --- ListCertFilters ---

func TestListCertFilters(t *testing.T) {
	existingFilters := []*namespacev1.CertificateFilterSpec{
		{CommonName: "test.temporal.io", Organization: "Temporal"},
		{SubjectAlternativeName: "*.temporal.io"},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceCertFilterListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNamespaceCertFilterListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithFilters(existingFilters)}, nil)
			},
		},
		{
			name: "EmptyFilters",
			cmd:  temporalcloudcli.CloudNamespaceCertFilterListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: testNamespaceWithFilters(nil)}, nil)
			},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceCertFilterListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- CreateCertFilter ---

func TestCreateCertFilter(t *testing.T) {
	existingNs := testNamespaceWithFilters([]*namespacev1.CertificateFilterSpec{
		{CommonName: "existing.temporal.io"},
	})
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceCertFilterCreateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceCertFilterCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
				Organization:     "Temporal",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNs}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						if req.Namespace != "test-namespace.test-account" || req.ResourceVersion != "rv-1" {
							return false
						}
						filters := req.Spec.GetMtlsAuth().GetCertificateFilters()
						return len(filters) == 2 &&
							filters[1].CommonName == "test.temporal.io" &&
							filters[1].Organization == "Temporal"
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNamespaceCertFilterCreateCommand{
				NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:             "test.temporal.io",
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNs}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.ResourceVersion == "rv-override"
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-2"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-2"},
		},
		{
			name: "NoFieldsProvided",
			cmd: temporalcloudcli.CloudNamespaceCertFilterCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
			},
			expectedErr: "at least one certificate filter field must be specified",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudNamespaceCertFilterCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceCertFilterCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "namespace not found",
		},
		{
			name: "UpdateNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceCertFilterCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNs}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "update operation failed",
		},
		{
			name: "PollingError",
			cmd: temporalcloudcli.CloudNamespaceCertFilterCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNs}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-poll-err"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-poll-err", ErrorToReturn: errors.New("polling failed")},
			expectedErr:        "failed to wait for async operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- DeleteCertFilter ---

func TestDeleteCertFilter(t *testing.T) {
	existingNs := testNamespaceWithFilters([]*namespacev1.CertificateFilterSpec{
		{CommonName: "test.temporal.io", Organization: "Temporal"},
		{CommonName: "other.temporal.io"},
	})

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceCertFilterDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceCertFilterDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
				Organization:     "Temporal",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNs}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						if req.Namespace != "test-namespace.test-account" || req.ResourceVersion != "rv-1" {
							return false
						}
						filters := req.Spec.GetMtlsAuth().GetCertificateFilters()
						// Should only have the "other.temporal.io" filter remaining
						return len(filters) == 1 && filters[0].CommonName == "other.temporal.io"
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNamespaceCertFilterDeleteCommand{
				NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:             "test.temporal.io",
				Organization:           "Temporal",
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNs}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.ResourceVersion == "rv-override"
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del-2"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del-2"},
		},
		{
			name: "NoFieldsProvided",
			cmd: temporalcloudcli.CloudNamespaceCertFilterDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
			},
			expectedErr: "at least one certificate filter field must be specified",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudNamespaceCertFilterDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceCertFilterDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "namespace not found",
		},
		{
			name: "UpdateNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceCertFilterDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
				Organization:     "Temporal",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNs}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "update operation failed",
		},
		{
			name: "PollingError",
			cmd: temporalcloudcli.CloudNamespaceCertFilterDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				CommonName:       "test.temporal.io",
				Organization:     "Temporal",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNs}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-poll-err"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-poll-err", ErrorToReturn: errors.New("polling failed")},
			expectedErr:        "failed to wait for async operation",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
