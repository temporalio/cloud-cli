package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	connectivityrulev1 "go.temporal.io/cloud-sdk/api/connectivityrule/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

// TestListConnectivityRules_Success verifies that ListConnectivityRules prints the returned rules.
func TestListConnectivityRules_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	rules := []*connectivityrulev1.ConnectivityRule{
		{Id: "rule-1", ResourceVersion: "rv-1"},
		{Id: "rule-2", ResourceVersion: "rv-2"},
	}
	mockCloud.EXPECT().
		GetConnectivityRules(context.Background(), &cloudservice.GetConnectivityRulesRequest{}).
		Return(&cloudservice.GetConnectivityRulesResponse{
			ConnectivityRules: rules,
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListConnectivityRules(context.Background(), temporalcloudcli.ListConnectivityRulesParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	type listConnectivityRulesOutput struct {
		ConnectivityRules []*connectivityrulev1.ConnectivityRule `json:"connectivityRules"`
		NextPageToken     string                                 `json:"nextPageToken"`
	}
	var out listConnectivityRulesOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, listConnectivityRulesOutput{ConnectivityRules: rules, NextPageToken: ""}, out)
}

// TestListConnectivityRules_WithNamespace verifies that the namespace filter is passed through.
func TestListConnectivityRules_WithNamespace(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetConnectivityRules(context.Background(), &cloudservice.GetConnectivityRulesRequest{
			Namespace: "my-namespace.my-account",
		}).
		Return(&cloudservice.GetConnectivityRulesResponse{
			ConnectivityRules: []*connectivityrulev1.ConnectivityRule{},
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListConnectivityRules(context.Background(), temporalcloudcli.ListConnectivityRulesParams{
		Namespace: "my-namespace.my-account",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)
}

// TestListConnectivityRules_Error verifies that an API error propagates.
func TestListConnectivityRules_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetConnectivityRules(context.Background(), &cloudservice.GetConnectivityRulesRequest{}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ListConnectivityRules(context.Background(), temporalcloudcli.ListConnectivityRulesParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}

// TestDescribeConnectivityRule_Success verifies that the rule is printed.
func TestDescribeConnectivityRule_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	rule := &connectivityrulev1.ConnectivityRule{
		Id:              "rule-1",
		ResourceVersion: "rv-1",
	}
	mockCloud.EXPECT().
		GetConnectivityRule(context.Background(), &cloudservice.GetConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
		}).
		Return(&cloudservice.GetConnectivityRuleResponse{
			ConnectivityRule: rule,
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.DescribeConnectivityRule(context.Background(), temporalcloudcli.DescribeConnectivityRuleParams{
		ID:      "rule-1",
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	var out connectivityrulev1.ConnectivityRule
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, connectivityrulev1.ConnectivityRule{Id: "rule-1"}, out)
}

// TestDescribeConnectivityRule_Error verifies that an API error propagates.
func TestDescribeConnectivityRule_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetConnectivityRule(context.Background(), &cloudservice.GetConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
		}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.DescribeConnectivityRule(context.Background(), temporalcloudcli.DescribeConnectivityRuleParams{
		ID:      "rule-1",
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}

// TestCreatePublicConnectivityRule_Success verifies that a public rule is created and the ID is printed.
func TestCreatePublicConnectivityRule_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	op := &operation.AsyncOperation{Id: "op-123"}
	mockPrompter.EXPECT().
		PromptApply(&connectivityrulev1.ConnectivityRuleSpec{}, &connectivityrulev1.ConnectivityRuleSpec{
			ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
				PublicRule: &connectivityrulev1.PublicConnectivityRule{},
			},
		}, false).
		Return(nil)
	mockCloud.EXPECT().
		CreateConnectivityRule(context.Background(), &cloudservice.CreateConnectivityRuleRequest{
			Spec: &connectivityrulev1.ConnectivityRuleSpec{
				ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
					PublicRule: &connectivityrulev1.PublicConnectivityRule{},
				},
			},
		}).
		Return(&cloudservice.CreateConnectivityRuleResponse{
			ConnectivityRuleId: "rule-abc",
			AsyncOperation:     op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "rule-abc").Return(nil)

	err := temporalcloudcli.CreatePublicConnectivityRule(context.Background(), temporalcloudcli.CreatePublicConnectivityRuleParams{
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestCreatePublicConnectivityRule_APIError verifies that an API error goes through HandleErr.
func TestCreatePublicConnectivityRule_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	apiErr := errors.New("api error")

	mockPrompter.EXPECT().
		PromptApply(&connectivityrulev1.ConnectivityRuleSpec{}, &connectivityrulev1.ConnectivityRuleSpec{
			ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
				PublicRule: &connectivityrulev1.PublicConnectivityRule{},
			},
		}, false).
		Return(nil)
	mockCloud.EXPECT().
		CreateConnectivityRule(context.Background(), &cloudservice.CreateConnectivityRuleRequest{
			Spec: &connectivityrulev1.ConnectivityRuleSpec{
				ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
					PublicRule: &connectivityrulev1.PublicConnectivityRule{},
				},
			},
		}).
		Return(nil, apiErr)

	mockHandler.EXPECT().HandleCreateErr(apiErr).Return(apiErr)

	err := temporalcloudcli.CreatePublicConnectivityRule(context.Background(), temporalcloudcli.CreatePublicConnectivityRuleParams{
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestCreatePublicConnectivityRule_PromptDeclined verifies that declining the prompt aborts the create.
func TestCreatePublicConnectivityRule_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	promptErr := errors.New("aborted")

	mockPrompter.EXPECT().
		PromptApply(&connectivityrulev1.ConnectivityRuleSpec{}, &connectivityrulev1.ConnectivityRuleSpec{
			ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
				PublicRule: &connectivityrulev1.PublicConnectivityRule{},
			},
		}, false).
		Return(promptErr)

	err := temporalcloudcli.CreatePublicConnectivityRule(context.Background(), temporalcloudcli.CreatePublicConnectivityRuleParams{
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestCreatePrivateConnectivityRule_Success verifies that a private rule is created with connection-id and region.
func TestCreatePrivateConnectivityRule_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	op := &operation.AsyncOperation{Id: "op-456"}
	mockPrompter.EXPECT().
		PromptApply(&connectivityrulev1.ConnectivityRuleSpec{}, &connectivityrulev1.ConnectivityRuleSpec{
			ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PrivateRule{
				PrivateRule: &connectivityrulev1.PrivateConnectivityRule{
					ConnectionId: "vpce-12345",
					Region:       "aws-us-west-2",
				},
			},
		}, false).
		Return(nil)
	mockCloud.EXPECT().
		CreateConnectivityRule(context.Background(), &cloudservice.CreateConnectivityRuleRequest{
			Spec: &connectivityrulev1.ConnectivityRuleSpec{
				ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PrivateRule{
					PrivateRule: &connectivityrulev1.PrivateConnectivityRule{
						ConnectionId: "vpce-12345",
						Region:       "aws-us-west-2",
					},
				},
			},
		}).
		Return(&cloudservice.CreateConnectivityRuleResponse{
			ConnectivityRuleId: "rule-def",
			AsyncOperation:     op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "rule-def").Return(nil)

	err := temporalcloudcli.CreatePrivateConnectivityRule(context.Background(), temporalcloudcli.CreatePrivateConnectivityRuleParams{
		ConnectionID:     "vpce-12345",
		Region:           "aws-us-west-2",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestCreatePrivateConnectivityRule_MissingConnectionID verifies an error when no connection-id.
func TestCreatePrivateConnectivityRule_MissingConnectionID(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.CreatePrivateConnectivityRule(context.Background(), temporalcloudcli.CreatePrivateConnectivityRuleParams{
		Region:           "aws-us-west-2",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--connection-id is required")
}

// TestCreatePrivateConnectivityRule_MissingRegion verifies an error when no region.
func TestCreatePrivateConnectivityRule_MissingRegion(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.CreatePrivateConnectivityRule(context.Background(), temporalcloudcli.CreatePrivateConnectivityRuleParams{
		ConnectionID:     "vpce-12345",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--region is required")
}

// TestDeleteConnectivityRule_Success verifies that the rule is fetched, diff shown, then deleted.
func TestDeleteConnectivityRule_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	ruleSpec := &connectivityrulev1.ConnectivityRuleSpec{
		ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
			PublicRule: &connectivityrulev1.PublicConnectivityRule{},
		},
	}
	mockCloud.EXPECT().
		GetConnectivityRule(context.Background(), &cloudservice.GetConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
		}).
		Return(&cloudservice.GetConnectivityRuleResponse{
			ConnectivityRule: &connectivityrulev1.ConnectivityRule{
				Id:              "rule-1",
				ResourceVersion: "rv-1",
				Spec:            ruleSpec,
			},
		}, nil)

	mockPrompter.EXPECT().
		PromptApply(ruleSpec, &connectivityrulev1.ConnectivityRuleSpec{}, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-789"}
	mockCloud.EXPECT().
		DeleteConnectivityRule(context.Background(), &cloudservice.DeleteConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
			ResourceVersion:    "rv-1",
		}).
		Return(&cloudservice.DeleteConnectivityRuleResponse{
			AsyncOperation: op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "rule-1").Return(nil)

	err := temporalcloudcli.DeleteConnectivityRule(context.Background(), temporalcloudcli.DeleteConnectivityRuleParams{
		ID:               "rule-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestDeleteConnectivityRule_ExplicitResourceVersion verifies that the provided resource version is used directly.
func TestDeleteConnectivityRule_ExplicitResourceVersion(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	ruleSpec := &connectivityrulev1.ConnectivityRuleSpec{
		ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
			PublicRule: &connectivityrulev1.PublicConnectivityRule{},
		},
	}
	mockCloud.EXPECT().
		GetConnectivityRule(context.Background(), &cloudservice.GetConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
		}).
		Return(&cloudservice.GetConnectivityRuleResponse{
			ConnectivityRule: &connectivityrulev1.ConnectivityRule{
				Id:              "rule-1",
				ResourceVersion: "rv-1",
				Spec:            ruleSpec,
			},
		}, nil)

	mockPrompter.EXPECT().
		PromptApply(ruleSpec, &connectivityrulev1.ConnectivityRuleSpec{}, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-101"}
	mockCloud.EXPECT().
		DeleteConnectivityRule(context.Background(), &cloudservice.DeleteConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
			ResourceVersion:    "rv-explicit",
		}).
		Return(&cloudservice.DeleteConnectivityRuleResponse{
			AsyncOperation: op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "rule-1").Return(nil)

	err := temporalcloudcli.DeleteConnectivityRule(context.Background(), temporalcloudcli.DeleteConnectivityRuleParams{
		ID:               "rule-1",
		ResourceVersion:  "rv-explicit",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestDeleteConnectivityRule_GetError verifies that an error fetching the rule propagates.
func TestDeleteConnectivityRule_GetError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetConnectivityRule(context.Background(), &cloudservice.GetConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
		}).
		Return(nil, apiErr)

	mockHandler.EXPECT().HandleDeleteErr(apiErr).Return(apiErr)

	err := temporalcloudcli.DeleteConnectivityRule(context.Background(), temporalcloudcli.DeleteConnectivityRuleParams{
		ID:               "rule-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestDeleteConnectivityRule_APIError verifies that a delete API error goes through HandleErr.
func TestDeleteConnectivityRule_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	deleteErr := errors.New("delete error")

	ruleSpec := &connectivityrulev1.ConnectivityRuleSpec{
		ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
			PublicRule: &connectivityrulev1.PublicConnectivityRule{},
		},
	}
	mockCloud.EXPECT().
		GetConnectivityRule(context.Background(), &cloudservice.GetConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
		}).
		Return(&cloudservice.GetConnectivityRuleResponse{
			ConnectivityRule: &connectivityrulev1.ConnectivityRule{
				Id:              "rule-1",
				ResourceVersion: "rv-1",
				Spec:            ruleSpec,
			},
		}, nil)

	mockPrompter.EXPECT().
		PromptApply(ruleSpec, &connectivityrulev1.ConnectivityRuleSpec{}, false).
		Return(nil)

	mockCloud.EXPECT().
		DeleteConnectivityRule(context.Background(), &cloudservice.DeleteConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
			ResourceVersion:    "rv-1",
		}).
		Return(nil, deleteErr)

	mockHandler.EXPECT().HandleDeleteErr(deleteErr).Return(deleteErr)

	err := temporalcloudcli.DeleteConnectivityRule(context.Background(), temporalcloudcli.DeleteConnectivityRuleParams{
		ID:               "rule-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, deleteErr)
}

// TestDeleteConnectivityRule_PromptDeclined verifies that declining the prompt aborts the delete.
func TestDeleteConnectivityRule_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	promptErr := errors.New("aborted")

	ruleSpec := &connectivityrulev1.ConnectivityRuleSpec{
		ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
			PublicRule: &connectivityrulev1.PublicConnectivityRule{},
		},
	}
	mockCloud.EXPECT().
		GetConnectivityRule(context.Background(), &cloudservice.GetConnectivityRuleRequest{
			ConnectivityRuleId: "rule-1",
		}).
		Return(&cloudservice.GetConnectivityRuleResponse{
			ConnectivityRule: &connectivityrulev1.ConnectivityRule{
				Id:              "rule-1",
				ResourceVersion: "rv-1",
				Spec:            ruleSpec,
			},
		}, nil)

	mockPrompter.EXPECT().
		PromptApply(ruleSpec, &connectivityrulev1.ConnectivityRuleSpec{}, false).
		Return(promptErr)

	err := temporalcloudcli.DeleteConnectivityRule(context.Background(), temporalcloudcli.DeleteConnectivityRuleParams{
		ID:               "rule-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}
