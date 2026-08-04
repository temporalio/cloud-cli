package temporalcloudcli

import (
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	projectv1 "go.temporal.io/cloud-sdk/api/project/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudProjectListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetProjects(cctx, &cloudservice.GetProjectsRequest{
		PageSize:   int32(c.PageSize),
		PageToken:  c.PageToken,
		ProjectIds: c.ProjectId,
	})
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
		struct {
			Projects      []*projectv1.Project
			NextPageToken string
		}{
			Projects:      res.Projects,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State", "CreatedTime"},
			SpecFields: []string{"DisplayName", "Description"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudProjectGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetProject(cctx, &cloudservice.GetProjectRequest{ProjectId: c.ProjectId})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(res.Project, printer.PrintResourceOptions{})
}

func (c *CloudProjectCreateCommand) run(cctx *CommandContext, _ []string) error {
	spec := projectSpecFromFlags(c.DisplayName, c.Description, c.EnableDeleteProtection)

	yes, err := cctx.GetPrompter().PromptYes("Create")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	resp, err := client.CreateProject(cctx, &cloudservice.CreateProjectRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudProjectUpdateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetProject(cctx, &cloudservice.GetProjectRequest{ProjectId: c.ProjectId})
	if err != nil {
		return err
	}
	project := res.Project
	newSpec := proto.Clone(project.Spec).(*projectv1.ProjectSpec)

	if c.Command.Flags().Changed("display-name") {
		newSpec.DisplayName = c.DisplayName
	}
	if c.Command.Flags().Changed("description") {
		newSpec.Description = c.Description
	}
	if c.Command.Flags().Changed("enable-delete-protection") {
		if newSpec.Lifecycle == nil {
			newSpec.Lifecycle = &projectv1.LifecycleSpec{}
		}
		newSpec.Lifecycle.EnableDeleteProtection = c.EnableDeleteProtection
	}

	yes, err := cctx.GetPrompter().PromptApply(project.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting update.")
	}

	rv := project.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateProject(cctx, &cloudservice.UpdateProjectRequest{
		ProjectId:        c.ProjectId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudProjectApplyCommand) run(cctx *CommandContext, _ []string) error {
	specData, err := loadJSONSpec(c.Spec)
	if err != nil {
		return err
	}
	spec := &projectv1.ProjectSpec{}
	if err := cctx.UnmarshalProtoJSON(specData, spec); err != nil {
		return fmt.Errorf("failed to parse JSON spec: %w", err)
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	if c.ProjectId == "" {
		yes, err := cctx.GetPrompter().PromptApply((*projectv1.ProjectSpec)(nil), spec, c.VerboseDiff)
		if err != nil {
			return err
		}
		if !yes {
			return errors.New("Aborting apply.")
		}
		resp, err := client.CreateProject(cctx, &cloudservice.CreateProjectRequest{
			Spec:             spec,
			AsyncOperationId: c.AsyncOperationId,
		})
		return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
	}

	res, err := client.GetProject(cctx, &cloudservice.GetProjectRequest{ProjectId: c.ProjectId})
	if err != nil {
		return err
	}
	project := res.Project

	yes, err := cctx.GetPrompter().PromptApply(project.Spec, spec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting apply.")
	}

	rv := project.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateProject(cctx, &cloudservice.UpdateProjectRequest{
		ProjectId:        c.ProjectId,
		Spec:             spec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudProjectEditCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetProject(cctx, &cloudservice.GetProjectRequest{ProjectId: c.ProjectId})
	if err != nil {
		return err
	}
	project := res.Project

	edited, err := cctx.GetEditor().EditProto(project.Spec)
	if err != nil {
		return err
	}
	newSpec := edited.(*projectv1.ProjectSpec)

	yes, err := cctx.GetPrompter().PromptApply(project.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting edit.")
	}

	rv := project.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateProject(cctx, &cloudservice.UpdateProjectRequest{
		ProjectId:        c.ProjectId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudProjectDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetProject(cctx, &cloudservice.GetProjectRequest{ProjectId: c.ProjectId})
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptYes("Delete")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	rv := res.Project.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteProject(cctx, &cloudservice.DeleteProjectRequest{
		ProjectId:        c.ProjectId,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

func projectSpecFromFlags(displayName, description string, enableDeleteProtection bool) *projectv1.ProjectSpec {
	return &projectv1.ProjectSpec{
		DisplayName: displayName,
		Description: description,
		Lifecycle: &projectv1.LifecycleSpec{
			EnableDeleteProtection: enableDeleteProtection,
		},
	}
}
