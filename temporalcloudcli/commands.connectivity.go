package temporalcloudcli

import (
	"context"
	"errors"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	connectivityrulev1 "go.temporal.io/cloud-sdk/api/connectivityrule/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	ListConnectivityRulesParams struct {
		Namespace string
		PageSize  int32
		PageToken string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	GetConnectivityRuleParams struct {
		ID string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	CreatePublicConnectivityRuleParams struct {
		EnableStableIPs  bool
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	CreatePrivateConnectivityRuleParams struct {
		ConnectionID     string
		Region           string
		GCPProjectID     string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	DeleteConnectivityRuleParams struct {
		ID               string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}
)

// ListConnectivityRules retrieves connectivity rules, optionally filtered by namespace.
func ListConnectivityRules(ctx context.Context, params ListConnectivityRulesParams) error {
	res, err := params.Cloud.GetConnectivityRules(ctx, &cloudservice.GetConnectivityRulesRequest{
		Namespace: params.Namespace,
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
	})
	if err != nil {
		return err
	}
	return params.Printer.PrintResourceList(
		struct {
			ConnectivityRules []*connectivityrulev1.ConnectivityRule
			NextPageToken     string
		}{
			ConnectivityRules: res.ConnectivityRules,
			NextPageToken:     res.GetNextPageToken(),
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State"},
			SpecFields: []string{},
		},
		printer.TableOptions{},
	)
}

// GetConnectivityRule retrieves details of a specific connectivity rule by ID.
func GetConnectivityRule(ctx context.Context, params GetConnectivityRuleParams) error {
	res, err := params.Cloud.GetConnectivityRule(ctx, &cloudservice.GetConnectivityRuleRequest{
		ConnectivityRuleId: params.ID,
	})
	if err != nil {
		return err
	}

	return params.Printer.PrintResource(res.ConnectivityRule, printer.PrintResourceOptions{})
}

// CreatePublicConnectivityRule creates a new public internet connectivity rule.
func CreatePublicConnectivityRule(ctx context.Context, params CreatePublicConnectivityRuleParams) error {
	spec := &connectivityrulev1.ConnectivityRuleSpec{
		ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
			PublicRule: &connectivityrulev1.PublicConnectivityRule{
				EnableStableIps: params.EnableStableIPs,
			},
		},
	}

	if err := params.Prompter.PromptApply(&connectivityrulev1.ConnectivityRuleSpec{}, spec, false); err != nil {
		return err
	}

	createRule := wrapCreateOperation(
		params.Cloud.CreateConnectivityRule,
		params.OperationHandler,
		func(res *cloudservice.CreateConnectivityRuleResponse) string { return res.GetConnectivityRuleId() },
	)
	return createRule(ctx, &cloudservice.CreateConnectivityRuleRequest{
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

// CreatePrivateConnectivityRule creates a new private VPC connectivity rule.
func CreatePrivateConnectivityRule(ctx context.Context, params CreatePrivateConnectivityRuleParams) error {
	if params.ConnectionID == "" {
		return errors.New("--connection-id is required for private connectivity")
	}
	if params.Region == "" {
		return errors.New("--region is required for private connectivity")
	}

	spec := &connectivityrulev1.ConnectivityRuleSpec{
		ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PrivateRule{
			PrivateRule: &connectivityrulev1.PrivateConnectivityRule{
				ConnectionId: params.ConnectionID,
				GcpProjectId: params.GCPProjectID,
				Region:       params.Region,
			},
		},
	}

	if err := params.Prompter.PromptApply(&connectivityrulev1.ConnectivityRuleSpec{}, spec, false); err != nil {
		return err
	}

	createRule := wrapCreateOperation(
		params.Cloud.CreateConnectivityRule,
		params.OperationHandler,
		func(res *cloudservice.CreateConnectivityRuleResponse) string { return res.GetConnectivityRuleId() },
	)
	return createRule(ctx, &cloudservice.CreateConnectivityRuleRequest{
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

// DeleteConnectivityRule deletes a connectivity rule by ID.
// Always fetches the rule first to show a diff before deletion.
func DeleteConnectivityRule(ctx context.Context, params DeleteConnectivityRuleParams) error {
	res, err := params.Cloud.GetConnectivityRule(ctx, &cloudservice.GetConnectivityRuleRequest{
		ConnectivityRuleId: params.ID,
	})
	if err != nil {
		return params.OperationHandler.HandleDeleteErr(err)
	}

	rule := res.ConnectivityRule

	if err := params.Prompter.PromptApply(rule.Spec, &connectivityrulev1.ConnectivityRuleSpec{}, false); err != nil {
		return err
	}

	rv := params.ResourceVersion
	if rv == "" {
		rv = rule.ResourceVersion
	}

	deleteConnectivityRule := wrapDeleteOperation(params.Cloud.DeleteConnectivityRule, params.OperationHandler, params.ID)
	return deleteConnectivityRule(ctx, &cloudservice.DeleteConnectivityRuleRequest{
		ConnectivityRuleId: params.ID,
		ResourceVersion:    rv,
		AsyncOperationId:   params.AsyncOperationID,
	})
}

func (c *CloudConnectivityListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ListConnectivityRules(cctx.Context, ListConnectivityRulesParams{
		Namespace: c.Namespace,
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}

func (c *CloudConnectivityGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return GetConnectivityRule(cctx.Context, GetConnectivityRuleParams{
		ID:      c.Id,
		Cloud:   cloudClient.CloudService(),
		Printer: cctx.Printer,
	})
}

func (c *CloudConnectivityPublicCreateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreatePublicConnectivityRule(cctx.Context, CreatePublicConnectivityRuleParams{
		EnableStableIPs:  c.EnableStableIps,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudConnectivityPrivateCreateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreatePrivateConnectivityRule(cctx.Context, CreatePrivateConnectivityRuleParams{
		ConnectionID:     c.ConnectionId,
		Region:           c.Region,
		GCPProjectID:     c.GcpProjectId,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudConnectivityDeleteCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return DeleteConnectivityRule(cctx.Context, DeleteConnectivityRuleParams{
		ID:               c.Id,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}
