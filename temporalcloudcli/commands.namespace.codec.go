package temporalcloudcli

import (
	"errors"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCodecGetCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	codecServer, err := namespaceClient.GetCodecServer(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	result := struct {
		Namespace string
		Spec      *namespacev1.CodecServerSpec
	}{
		Namespace: c.Namespace,
		Spec:      codecServer,
	}
	return cctx.Printer.PrintResource(result, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceCodecSetCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	setCodec := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.SetCodec)
	return setCodec(namespace.SetCodecParams{
		Namespace:                        c.Namespace,
		Endpoint:                         c.Endpoint,
		PassAccessToken:                  c.PassAccessToken,
		IncludeCrossOriginCredentials:    c.IncludeCrossOriginCredentials,
		CustomErrorMessageDefaultMessage: c.CustomErrorMessageDefaultMessage,
		CustomErrorMessageDefaultLink:    c.CustomErrorMessageDefaultLink,
		ResourceVersion:                  c.ResourceVersion,
		AsyncOperationID:                 c.AsyncOperationId,
	})
}

func (c *CloudNamespaceCodecDeleteCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	deleteCodec := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.DeleteCodec)
	return deleteCodec(namespace.DeleteCodecParams{
		Namespace:        c.Namespace,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}
