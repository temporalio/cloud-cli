package temporalcloudcli_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// testAccount builds an account with the given PEM bytes as the metrics AcceptedClientCa.
func testAccount(acceptedClientCa []byte) *accountv1.Account {
	return &accountv1.Account{
		Spec: &accountv1.AccountSpec{
			Metrics: &accountv1.MetricsSpec{
				AcceptedClientCa: acceptedClientCa,
			},
		},
		ResourceVersion: "rv-1",
	}
}

// testAccountNoCerts builds an account with no metrics cert configured.
func testAccountNoCerts() *accountv1.Account {
	return &accountv1.Account{
		Spec:            &accountv1.AccountSpec{},
		ResourceVersion: "rv-1",
	}
}

// --- List ---

func TestListAccountMetricsCertCA(t *testing.T) {
	parsedCerts, _, certData := setupTestCertFile(t)

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAccount(mock.Anything, &cloudservice.GetAccountRequest{}, mock.Anything).
					Return(&cloudservice.GetAccountResponse{Account: testAccount(certData)}, nil)
			},
			expectedJsonOutput: parsedCerts,
		},
		{
			name: "Empty",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAccount(mock.Anything, &cloudservice.GetAccountRequest{}, mock.Anything).
					Return(&cloudservice.GetAccountResponse{Account: testAccountNoCerts()}, nil)
			},
			expectedJsonOutput: []cert.CACert{},
		},
		{
			name: "GetAccountError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAccount(mock.Anything, &cloudservice.GetAccountRequest{}, mock.Anything).
					Return(nil, errors.New("get account failed"))
			},
			expectedErr: "get account failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudAccountMetricsCertCaListCommand{}, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

// --- Create ---

func TestCreateAccountMetricsCertCA_Success(t *testing.T) {
	parsedCerts, certPath, certData := setupTestCertFile(t)
	base64Cert := base64.StdEncoding.EncodeToString(certData)

	// Compute the expected cert bundle bytes (trimmed PEM, same as encodeCACertBundle output).
	expectedBundle, err := base64.StdEncoding.DecodeString(parsedCerts[0].Base64EncodedData)
	require.NoError(t, err)

	asyncOp := &operation.AsyncOperation{Id: "op-create"}

	tests := []struct {
		name string
		cmd  temporalcloudcli.CloudAccountMetricsCertCaCreateCommand
	}{
		{
			name: "WithFile",
			cmd:  temporalcloudcli.CloudAccountMetricsCertCaCreateCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath}},
		},
		{
			name: "WithBase64",
			cmd:  temporalcloudcli.CloudAccountMetricsCertCaCreateCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificate: base64Cert}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					c.EXPECT().
						GetAccount(mock.Anything, &cloudservice.GetAccountRequest{}, mock.Anything).
						Return(&cloudservice.GetAccountResponse{Account: testAccountNoCerts()}, nil)
					c.EXPECT().
						UpdateAccount(mock.Anything, mock.MatchedBy(func(r *cloudservice.UpdateAccountRequest) bool {
							return string(r.Spec.GetMetrics().GetAcceptedClientCa()) == string(expectedBundle) &&
								r.ResourceVersion == "rv-1"
						}), mock.Anything).
						Return(&cloudservice.UpdateAccountResponse{AsyncOperation: asyncOp}, nil)
				},
				AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
				JSONOutput:         true,
			})
		})
	}
}

func TestCreateAccountMetricsCertCA_CertAlreadyExists(t *testing.T) {
	parsedCerts, certPath, certData := setupTestCertFile(t)

	// Compute expected bundle: same cert already in account, so bundle stays the same.
	expectedBundle, err := base64.StdEncoding.DecodeString(parsedCerts[0].Base64EncodedData)
	require.NoError(t, err)

	asyncOp := &operation.AsyncOperation{Id: "op-noop"}

	cmd := temporalcloudcli.CloudAccountMetricsCertCaCreateCommand{
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetAccount(mock.Anything, &cloudservice.GetAccountRequest{}, mock.Anything).
				Return(&cloudservice.GetAccountResponse{Account: testAccount(certData)}, nil)
			// UpdateAccount is still called — the same bundle is sent; server decides if there's a change.
			c.EXPECT().
				UpdateAccount(mock.Anything, mock.MatchedBy(func(r *cloudservice.UpdateAccountRequest) bool {
					return string(r.Spec.GetMetrics().GetAcceptedClientCa()) == string(expectedBundle)
				}), mock.Anything).
				Return(&cloudservice.UpdateAccountResponse{AsyncOperation: asyncOp}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-noop"},
		JSONOutput:         true,
	})
}

func TestCreateAccountMetricsCertCA_Error(t *testing.T) {
	_, certPath, _ := setupTestCertFile(t)

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudAccountMetricsCertCaCreateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name:        "NoCertFlags",
			cmd:         temporalcloudcli.CloudAccountMetricsCertCaCreateCommand{},
			expectedErr: "either --ca-certificate-file or --ca-certificate must be provided",
		},
		{
			name:        "BothCertFlags",
			cmd:         temporalcloudcli.CloudAccountMetricsCertCaCreateCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath, CaCertificate: "data"}},
			expectedErr: "cannot specify both",
		},
		{
			name:        "InvalidBase64",
			cmd:         temporalcloudcli.CloudAccountMetricsCertCaCreateCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificate: "!!!invalid"}},
			expectedErr: "invalid base64 encoded certificate data",
		},
		{
			name: "GetAccountError",
			cmd:  temporalcloudcli.CloudAccountMetricsCertCaCreateCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("account not found"))
			},
			expectedErr: "account not found",
		},
		{
			name: "UpdateAccountError",
			cmd:  temporalcloudcli.CloudAccountMetricsCertCaCreateCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetAccountResponse{Account: testAccountNoCerts()}, nil)
				c.EXPECT().
					UpdateAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			expectedErr: "update failed",
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

// --- Delete ---

func TestDeleteAccountMetricsCertCA_Success(t *testing.T) {
	_, certPath, certData := setupTestCertFile(t)

	asyncOp := &operation.AsyncOperation{Id: "op-delete"}

	cmd := temporalcloudcli.CloudAccountMetricsCertCaDeleteCommand{
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetAccount(mock.Anything, &cloudservice.GetAccountRequest{}, mock.Anything).
				Return(&cloudservice.GetAccountResponse{Account: testAccount(certData)}, nil)
			c.EXPECT().
				UpdateAccount(mock.Anything, mock.MatchedBy(func(r *cloudservice.UpdateAccountRequest) bool {
					// cert removed → AcceptedClientCa should be empty/nil
					return len(r.Spec.GetMetrics().GetAcceptedClientCa()) == 0 &&
						r.ResourceVersion == "rv-1"
				}), mock.Anything).
				Return(&cloudservice.UpdateAccountResponse{AsyncOperation: asyncOp}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-delete"},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		JSONOutput:         true,
	})
}

func TestDeleteAccountMetricsCertCA_Error(t *testing.T) {
	_, certPath, certData := setupTestCertFile(t)

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudAccountMetricsCertCaDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		expectedErr             string
	}{
		{
			name:        "NoCertFlags",
			cmd:         temporalcloudcli.CloudAccountMetricsCertCaDeleteCommand{},
			expectedErr: "either --ca-certificate-file or --ca-certificate must be provided",
		},
		{
			name:          "UserDeclines",
			cmd:           temporalcloudcli.CloudAccountMetricsCertCaDeleteCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath}},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "GetAccountError",
			cmd:  temporalcloudcli.CloudAccountMetricsCertCaDeleteCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("get account failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "get account failed",
		},
		{
			name: "UpdateAccountError",
			cmd:  temporalcloudcli.CloudAccountMetricsCertCaDeleteCommand{CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetAccountResponse{Account: testAccount(certData)}, nil)
				c.EXPECT().
					UpdateAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
