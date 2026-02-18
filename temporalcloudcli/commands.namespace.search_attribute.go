package temporalcloudcli

import (
	"errors"
	"strings"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceSearchAttributeListCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	searchAttrs, err := namespaceClient.ListSearchAttributes(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(searchAttrs, printer.StructuredOptions{})
}

func (c *CloudNamespaceSearchAttributeCreateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	if c.Name == "" {
		return errors.New("name is required")
	}

	if c.Type == "" {
		return errors.New("type is required")
	}

	// Parse the type string to the enum value
	attrType, err := parseSearchAttributeType(c.Type)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Create search attribute (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting create.")
	}

	params := namespace.CreateSearchAttributeParams{
		Namespace:        c.Namespace,
		Name:             c.Name,
		Type:             attrType,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.CreateSearchAttribute(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}

func (c *CloudNamespaceSearchAttributeRenameCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	if c.ExistingName == "" {
		return errors.New("existing-name is required")
	}

	if c.NewName == "" {
		return errors.New("new-name is required")
	}

	yes, err := cctx.promptYes("Rename search attribute (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting rename.")
	}

	params := namespace.RenameSearchAttributeParams{
		Namespace:                         c.Namespace,
		ExistingCustomSearchAttributeName: c.ExistingName,
		NewCustomSearchAttributeName:      c.NewName,
		ResourceVersion:                   c.ResourceVersion,
		AsyncOperationID:                  c.AsyncOperationId,
	}
	op, err := namespaceClient.RenameSearchAttribute(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}

func (c *CloudNamespaceSearchAttributeDeleteCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	if c.Name == "" {
		return errors.New("name is required")
	}

	yes, err := cctx.promptYes("Delete search attribute (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting delete.")
	}

	params := namespace.DeleteSearchAttributeParams{
		Namespace:        c.Namespace,
		Name:             c.Name,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.DeleteSearchAttribute(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}

// parseSearchAttributeType converts a string type to the corresponding enum value.
// AIDEV-NOTE: This function provides case-insensitive matching for search attribute types
// to improve user experience when specifying types via CLI.
func parseSearchAttributeType(typeStr string) (namespacev1.NamespaceSpec_SearchAttributeType, error) {
	typeStr = strings.ToLower(strings.TrimSpace(typeStr))

	switch typeStr {
	case "text":
		return namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT, nil
	case "keyword":
		return namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD, nil
	case "int":
		return namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_INT, nil
	case "double":
		return namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_DOUBLE, nil
	case "bool":
		return namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_BOOL, nil
	case "datetime":
		return namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_DATETIME, nil
	case "keywordlist":
		return namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD_LIST, nil
	default:
		return namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_UNSPECIFIED,
			errors.New("invalid search attribute type: must be one of Text, Keyword, Int, Double, Bool, Datetime, KeywordList")
	}
}
