package temporalcloudcli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli/async"
	"github.com/temporalio/cloud-cli/temporalcloudcli/editor"
	editormock "github.com/temporalio/cloud-cli/temporalcloudcli/editor/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cliprompter "github.com/temporalio/cloud-cli/temporalcloudcli/prompter"
	promptermock "github.com/temporalio/cloud-cli/temporalcloudcli/prompter/mock"
)

type (
	CommandIfc interface {
		run(cctx *CommandContext, args []string) error
	}

	TestAsyncPollerOptions struct {
		AsyncOperationID string
		ErrorToReturn    error
	}

	TestPromptOptions struct {
		ExpectPrompApply bool
		PromptApplyError error

		ExpectPromptYes bool
		PromptYesResult bool
		PromptYesError  error
	}

	TestEditorOptions struct {
		Modified    proto.Message
		EditorError error
	}

	TestCommandOptions struct {
		Args                    []string
		CloudClientExpectations func(cloudClient *cloudmock.MockCloudServiceClient)
		AsyncPollerOptions      TestAsyncPollerOptions
		PromptOptions           TestPromptOptions
		EditorOptions           TestEditorOptions
		JSONOutput              bool
		ExpectedError           string
		ExpectedOutput          string
		ExpectedOutputJson      any
	}
)

func getMockEditor(t *testing.T, printer *printer.Printer, opts TestEditorOptions) editor.Editor {
	editorMock := editormock.NewMockEditor(t)
	if opts.Modified != nil || opts.EditorError != nil {
		editorMock.EXPECT().EditProto(mock.Anything).
			Return(opts.Modified, opts.EditorError).
			Times(1)
	}
	return editorMock
}

func getMockPrompter(t *testing.T, printer *printer.Printer, opts TestPromptOptions) cliprompter.Prompter {
	promptMock := promptermock.NewMockPrompter(t)
	if opts.ExpectPrompApply {
		promptMock.EXPECT().
			PromptApply(mock.Anything, mock.Anything, mock.Anything).
			Return(opts.PromptApplyError).
			Times(1)
	}

	if opts.ExpectPromptYes {
		promptMock.EXPECT().
			PromptYes(mock.Anything, mock.Anything).
			Return(opts.PromptYesResult, opts.PromptYesError).
			Times(1)
	}
	return promptMock
}

func getMockAsyncPoller(t *testing.T, printer *printer.Printer, opts *TestAsyncPollerOptions) async.Poller {
	asyncPollerMockCloudClient := cloudmock.NewMockCloudServiceClient(t)
	switch {
	case opts.ErrorToReturn != nil:
		asyncPollerMockCloudClient.EXPECT().
			GetAsyncOperation(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, in *cloudservice.GetAsyncOperationRequest, _ ...grpc.CallOption) {
				if opts.AsyncOperationID != "" {
					assert.Equal(t, opts.AsyncOperationID, in.AsyncOperationId)
				}
			}).
			Return(nil, opts.ErrorToReturn).
			Times(1)
	case opts.AsyncOperationID != "":
		// first call returns pending, second call returns success
		asyncPollerMockCloudClient.EXPECT().
			GetAsyncOperation(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, in *cloudservice.GetAsyncOperationRequest, _ ...grpc.CallOption) {
				assert.Equal(t, opts.AsyncOperationID, in.AsyncOperationId)
			}).
			Return(&cloudservice.GetAsyncOperationResponse{
				AsyncOperation: &operation.AsyncOperation{
					Id:    opts.AsyncOperationID,
					State: operation.AsyncOperation_STATE_PENDING,
				},
			}, nil).
			Times(1)
		asyncPollerMockCloudClient.EXPECT().
			GetAsyncOperation(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, in *cloudservice.GetAsyncOperationRequest, _ ...grpc.CallOption) {
				assert.Equal(t, opts.AsyncOperationID, in.AsyncOperationId)
			}).
			Return(&cloudservice.GetAsyncOperationResponse{
				AsyncOperation: &operation.AsyncOperation{
					Id:    opts.AsyncOperationID,
					State: operation.AsyncOperation_STATE_FULFILLED,
				},
			}, nil).
			Times(1)
		// AsyncOperationID == "" and ErrorToReturn == nil: no poller expectations (command errors before polling)
	}
	testPoller := async.NewPoller(
		asyncPollerMockCloudClient,
		printer,
		false,               // idempotent
		false,               // async
		time.Millisecond*10, // pollInterval
	)
	return testPoller
}

func TestCommand(t *testing.T, ctx context.Context, command CommandIfc, opts TestCommandOptions) {
	mockCloudClient := cloudmock.NewMockCloudServiceClient(t)
	printerBuf := &bytes.Buffer{}
	printer := &printer.Printer{Output: printerBuf, JSON: opts.JSONOutput}
	if opts.CloudClientExpectations != nil {
		opts.CloudClientExpectations(mockCloudClient)
	}
	asyncPoller := getMockAsyncPoller(t, printer, &opts.AsyncPollerOptions)
	prompter := getMockPrompter(t, printer, opts.PromptOptions)
	cctx := &CommandContext{
		Context:     context.Background(),
		Printer:     printer,
		RootCommand: &CloudCommand{},
		getCloudClientOverride: func() cloudservice.CloudServiceClient {
			return mockCloudClient
		},
		getAsyncPollerOverride: func() async.Poller {
			return asyncPoller
		},
		getPrompterOverride: func() cliprompter.Prompter {
			return prompter
		},
		getEditorOverride: func() editor.Editor {
			return getMockEditor(t, printer, opts.EditorOptions)
		},
	}
	err := command.run(cctx, opts.Args)
	if opts.ExpectedError != "" {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), opts.ExpectedError)
	} else {
		assert.NoError(t, err)
		t.Logf("Command output: %s", printerBuf.String())
		if opts.ExpectedOutput != "" {
			assert.Equal(t, opts.ExpectedOutput, printerBuf.String())
		}
		if opts.ExpectedOutputJson != nil {
			var js []byte
			var err error
			if protoMessage, ok := opts.ExpectedOutputJson.(proto.Message); ok {
				js, err = protojson.Marshal(protoMessage)
			} else {
				js, err = json.Marshal(opts.ExpectedOutputJson)
			}
			assert.NoError(t, err)
			assert.JSONEq(t, string(js), printerBuf.String())
		}
	}
}
