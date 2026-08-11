package temporalcloudcli

import (
	"context"
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	projectv1 "go.temporal.io/cloud-sdk/api/project/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

const projectLookupPageSize = 1000

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

	project, err := resolveProject(cctx, client, c.ProjectId)
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(project, printer.PrintResourceOptions{})
}

func (c *CloudProjectUserListCommand) run(cctx *CommandContext, _ []string) error {
	return printProjectUserAssignments(cctx, c.ClientOptions, c.ProjectId, c.PageSize, c.PageToken)
}

func (c *CloudProjectUserGroupListCommand) run(cctx *CommandContext, _ []string) error {
	return printProjectUserGroupAssignments(cctx, c.ClientOptions, c.ProjectId, c.PageSize, c.PageToken)
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

	project, err := resolveProject(cctx, client, c.ProjectId)
	if err != nil {
		return err
	}
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
		ProjectId:        project.Id,
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

	project, err := projectForApply(cctx, client, spec)
	if err != nil {
		return err
	}

	if project == nil {
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
		ProjectId:        project.Id,
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

	project, err := resolveProject(cctx, client, c.ProjectId)
	if err != nil {
		return err
	}

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
		ProjectId:        project.Id,
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

	project, err := resolveProject(cctx, client, c.ProjectId)
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

	rv := project.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteProject(cctx, &cloudservice.DeleteProjectRequest{
		ProjectId:        project.Id,
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

func projectForApply(
	ctx context.Context,
	client cloudservice.CloudServiceClient,
	spec *projectv1.ProjectSpec,
) (*projectv1.Project, error) {
	if spec.GetDisplayName() == "" {
		return nil, nil
	}
	project, err := getProjectByName(ctx, client, spec.GetDisplayName())
	if err != nil {
		return nil, err
	}
	return project, nil
}

func resolveProject(
	ctx context.Context,
	client cloudservice.CloudServiceClient,
	projectID string,
) (*projectv1.Project, error) {
	res, err := client.GetProject(ctx, &cloudservice.GetProjectRequest{ProjectId: projectID})
	if err != nil {
		return nil, err
	}
	return res.Project, nil
}

func getProjectByName(ctx context.Context, client cloudservice.CloudServiceClient, projectName string) (*projectv1.Project, error) {
	var match *projectv1.Project
	var pageToken string
	for {
		res, err := client.GetProjects(ctx, &cloudservice.GetProjectsRequest{
			PageSize:  projectLookupPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, project := range res.Projects {
			if project.GetSpec().GetDisplayName() != projectName {
				continue
			}
			if match != nil {
				return nil, fmt.Errorf("multiple projects found with display name %q", projectName)
			}
			match = project
		}

		pageToken = res.GetNextPageToken()
		if pageToken == "" {
			return match, nil
		}
	}
}
