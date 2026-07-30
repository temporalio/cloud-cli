package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"
	projectv1 "go.temporal.io/cloud-sdk/api/project/v1"
	resourcev1 "go.temporal.io/cloud-sdk/api/resource/v1"
	"google.golang.org/protobuf/proto"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func testProject(id string) *projectv1.Project {
	return &projectv1.Project{
		Id:              id,
		ResourceVersion: "rv-" + id,
		State:           resourcev1.ResourceState_RESOURCE_STATE_ACTIVE,
		Spec: &projectv1.ProjectSpec{
			DisplayName: "Engineering",
			Description: "Engineering workloads",
			Lifecycle:   &projectv1.LifecycleSpec{EnableDeleteProtection: true},
		},
	}
}

func testAsyncOperation(id string) *operationv1.AsyncOperation {
	return &operationv1.AsyncOperation{Id: id, State: operationv1.AsyncOperation_STATE_PENDING}
}

func TestProjectList(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectListCommand{
		ProjectId: []string{"project-a", "project-b"},
		PageSize:  50,
		PageToken: "next",
	}
	res := &cloudservice.GetProjectsResponse{
		Projects:      []*projectv1.Project{testProject("project-a")},
		NextPageToken: "next-2",
	}

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetProjects(mock.Anything, &cloudservice.GetProjectsRequest{
					PageSize:   50,
					PageToken:  "next",
					ProjectIds: []string{"project-a", "project-b"},
				}, mock.Anything).
				Return(res, nil)
		},
		JSONOutput: true,
		ExpectedOutputJson: struct {
			Projects      []*projectv1.Project
			NextPageToken string
		}{Projects: res.Projects, NextPageToken: "next-2"},
	})
}

func TestProjectGet(t *testing.T) {
	project := testProject("project-a")

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudProjectGetCommand{ProjectId: "project-a"}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetProject(mock.Anything, &cloudservice.GetProjectRequest{ProjectId: "project-a"}, mock.Anything).
				Return(&cloudservice.GetProjectResponse{Project: project}, nil)
		},
		JSONOutput:         true,
		ExpectedOutputJson: project,
	})
}

func TestProjectCreate(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectCreateCommand{
		DisplayName:            "Engineering",
		Description:            "Engineering workloads",
		EnableDeleteProtection: true,
	}
	wantSpec := &projectv1.ProjectSpec{
		DisplayName: "Engineering",
		Description: "Engineering workloads",
		Lifecycle:   &projectv1.LifecycleSpec{EnableDeleteProtection: true},
	}

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateProject(mock.Anything, mock.MatchedBy(func(req *cloudservice.CreateProjectRequest) bool {
					return proto.Equal(req.Spec, wantSpec)
				}), mock.Anything).
				Return(&cloudservice.CreateProjectResponse{
					ProjectId:      "project-a",
					AsyncOperation: testAsyncOperation("op-create"),
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: "Create",
			PromptResult:           true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		JSONOutput:         true,
	})
}

func TestProjectCreate_PromptDecline(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectCreateCommand{DisplayName: "Engineering"}

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: "Create",
			PromptResult:           false,
		},
		ExpectedError: "Aborting create.",
	})
}

func TestProjectUpdate(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectUpdateCommand{
		ProjectId:              "project-a",
		DisplayName:            "Platform",
		EnableDeleteProtection: false,
	}
	cmd.Command.Flags().StringVar(&cmd.DisplayName, "display-name", "", "")
	require.NoError(t, cmd.Command.Flags().Set("display-name", "Platform"))
	cmd.Command.Flags().BoolVar(&cmd.EnableDeleteProtection, "enable-delete-protection", false, "")
	require.NoError(t, cmd.Command.Flags().Set("enable-delete-protection", "false"))

	existing := testProject("project-a")
	wantSpec := &projectv1.ProjectSpec{
		DisplayName: "Platform",
		Description: "Engineering workloads",
		Lifecycle:   &projectv1.LifecycleSpec{EnableDeleteProtection: false},
	}

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetProject(mock.Anything, &cloudservice.GetProjectRequest{ProjectId: "project-a"}, mock.Anything).
				Return(&cloudservice.GetProjectResponse{Project: existing}, nil)
			c.EXPECT().
				UpdateProject(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateProjectRequest) bool {
					return req.ProjectId == "project-a" &&
						req.ResourceVersion == "rv-project-a" &&
						proto.Equal(req.Spec, wantSpec)
				}), mock.Anything).
				Return(&cloudservice.UpdateProjectResponse{AsyncOperation: testAsyncOperation("op-update")}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply:          true,
			ExpectPromptApplyExisting: existing.Spec,
			ExpectPromptApplyModified: wantSpec,
			PromptResult:              true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update"},
		JSONOutput:         true,
	})
}

func TestProjectUpdate_ResourceVersionOverride(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectUpdateCommand{
		ProjectId:              "project-a",
		Description:            "New description",
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-user"},
	}
	cmd.Command.Flags().StringVar(&cmd.Description, "description", "", "")
	require.NoError(t, cmd.Command.Flags().Set("description", "New description"))

	existing := testProject("project-a")

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().GetProject(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetProjectResponse{Project: existing}, nil)
			c.EXPECT().
				UpdateProject(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateProjectRequest) bool {
					return req.ResourceVersion == "rv-user" && req.Spec.Description == "New description"
				}), mock.Anything).
				Return(&cloudservice.UpdateProjectResponse{AsyncOperation: testAsyncOperation("op-update")}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update"},
		JSONOutput:         true,
	})
}

func TestProjectApply_Create(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectApplyCommand{
		Spec: `{"display_name":"Engineering","description":"Engineering workloads"}`,
	}
	wantSpec := &projectv1.ProjectSpec{
		DisplayName: "Engineering",
		Description: "Engineering workloads",
	}

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateProject(mock.Anything, mock.MatchedBy(func(req *cloudservice.CreateProjectRequest) bool {
					return proto.Equal(req.Spec, wantSpec)
				}), mock.Anything).
				Return(&cloudservice.CreateProjectResponse{
					ProjectId:      "project-a",
					AsyncOperation: testAsyncOperation("op-create"),
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		JSONOutput:         true,
	})
}

func TestProjectApply_Update(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectApplyCommand{
		ProjectId:              "project-a",
		Spec:                   `{"display_name":"Platform","description":"Platform workloads"}`,
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-user"},
	}
	existing := testProject("project-a")
	wantSpec := &projectv1.ProjectSpec{
		DisplayName: "Platform",
		Description: "Platform workloads",
	}

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetProject(mock.Anything, &cloudservice.GetProjectRequest{ProjectId: "project-a"}, mock.Anything).
				Return(&cloudservice.GetProjectResponse{Project: existing}, nil)
			c.EXPECT().
				UpdateProject(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateProjectRequest) bool {
					return req.ProjectId == "project-a" &&
						req.ResourceVersion == "rv-user" &&
						proto.Equal(req.Spec, wantSpec)
				}), mock.Anything).
				Return(&cloudservice.UpdateProjectResponse{AsyncOperation: testAsyncOperation("op-update")}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply:          true,
			ExpectPromptApplyExisting: existing.Spec,
			ExpectPromptApplyModified: wantSpec,
			PromptResult:              true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update"},
		JSONOutput:         true,
	})
}

func TestProjectApply_InvalidSpec(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectApplyCommand{Spec: `{`}

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "failed to parse JSON spec",
	})
}

func TestProjectEdit(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectEditCommand{
		ProjectId:              "project-a",
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-user"},
	}
	existing := testProject("project-a")
	editedSpec := &projectv1.ProjectSpec{DisplayName: "Edited"}

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetProject(mock.Anything, &cloudservice.GetProjectRequest{ProjectId: "project-a"}, mock.Anything).
				Return(&cloudservice.GetProjectResponse{Project: existing}, nil)
			c.EXPECT().
				UpdateProject(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateProjectRequest) bool {
					return req.ProjectId == "project-a" &&
						req.ResourceVersion == "rv-user" &&
						proto.Equal(req.Spec, editedSpec)
				}), mock.Anything).
				Return(&cloudservice.UpdateProjectResponse{AsyncOperation: testAsyncOperation("op-update")}, nil)
		},
		EditorOptions: temporalcloudcli.TestEditorOptions{Modified: editedSpec},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply:          true,
			ExpectPromptApplyExisting: existing.Spec,
			ExpectPromptApplyModified: editedSpec,
			PromptResult:              true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update"},
		JSONOutput:         true,
	})
}

func TestProjectEdit_EditorError(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectEditCommand{ProjectId: "project-a"}
	existing := testProject("project-a")

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetProject(mock.Anything, &cloudservice.GetProjectRequest{ProjectId: "project-a"}, mock.Anything).
				Return(&cloudservice.GetProjectResponse{Project: existing}, nil)
		},
		EditorOptions: temporalcloudcli.TestEditorOptions{EditorError: errors.New("editor failed")},
		ExpectedError: "editor failed",
	})
}

func TestProjectDelete(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectDeleteCommand{
		ProjectId:              "project-a",
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-user"},
	}
	existing := testProject("project-a")

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetProject(mock.Anything, &cloudservice.GetProjectRequest{ProjectId: "project-a"}, mock.Anything).
				Return(&cloudservice.GetProjectResponse{Project: existing}, nil)
			c.EXPECT().
				DeleteProject(mock.Anything, &cloudservice.DeleteProjectRequest{
					ProjectId:        "project-a",
					ResourceVersion:  "rv-user",
					AsyncOperationId: "",
				}, mock.Anything).
				Return(&cloudservice.DeleteProjectResponse{AsyncOperation: testAsyncOperation("op-delete")}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: "Delete",
			PromptResult:           true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-delete"},
		JSONOutput:         true,
	})
}

func TestProjectDelete_PromptDecline(t *testing.T) {
	cmd := &temporalcloudcli.CloudProjectDeleteCommand{ProjectId: "project-a"}
	existing := testProject("project-a")

	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetProject(mock.Anything, &cloudservice.GetProjectRequest{ProjectId: "project-a"}, mock.Anything).
				Return(&cloudservice.GetProjectResponse{Project: existing}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: "Delete",
			PromptResult:           false,
		},
		ExpectedError: "Aborting delete.",
	})
}
