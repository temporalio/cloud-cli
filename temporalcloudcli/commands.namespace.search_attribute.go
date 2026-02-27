package temporalcloudcli

import (
	"fmt"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// AIDEV-NOTE: Search attribute type names map user-facing strings to proto enum values.
// The proto enum names use UPPER_SNAKE_CASE (e.g. SEARCH_ATTRIBUTE_TYPE_TEXT) but
// users provide PascalCase names (e.g. "Text") matching the Temporal UI convention.
var (
	searchAttributeTypeFromString = map[string]namespacev1.NamespaceSpec_SearchAttributeType{
		"Text":        namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
		"Keyword":     namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
		"Int":         namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_INT,
		"Double":      namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_DOUBLE,
		"Bool":        namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_BOOL,
		"Datetime":    namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_DATETIME,
		"KeywordList": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD_LIST,
	}

	searchAttributeTypeToString = map[namespacev1.NamespaceSpec_SearchAttributeType]string{
		namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT:         "Text",
		namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD:      "Keyword",
		namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_INT:          "Int",
		namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_DOUBLE:       "Double",
		namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_BOOL:         "Bool",
		namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_DATETIME:     "Datetime",
		namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD_LIST: "KeywordList",
	}
)

// SearchAttributeOutput is the display representation of a custom search attribute.
type SearchAttributeOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func parseSearchAttributeType(s string) (namespacev1.NamespaceSpec_SearchAttributeType, error) {
	t, ok := searchAttributeTypeFromString[s]
	if !ok {
		return 0, fmt.Errorf(
			"invalid search attribute type %q: valid values are Text, Keyword, Int, Double, Bool, Datetime, KeywordList",
			s,
		)
	}
	return t, nil
}

func formatSearchAttributeType(t namespacev1.NamespaceSpec_SearchAttributeType) string {
	if s, ok := searchAttributeTypeToString[t]; ok {
		return s
	}
	return t.String()
}

func (c *CloudNamespaceSearchAttributeListCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	attrs, err := namespaceClient.ListSearchAttributes(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	output := make([]SearchAttributeOutput, len(attrs))
	for i, a := range attrs {
		output[i] = SearchAttributeOutput{
			Name: a.Name,
			Type: formatSearchAttributeType(a.Type),
		}
	}

	return cctx.Printer.PrintResourceList(
		struct{ SearchAttributes []SearchAttributeOutput }{SearchAttributes: output},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceSearchAttributeCreateCommand) run(cctx *CommandContext, _ []string) error {
	attrType, err := parseSearchAttributeType(c.Type)
	if err != nil {
		return err
	}

	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	createSearchAttribute := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.CreateSearchAttribute)
	return createSearchAttribute(namespace.CreateSearchAttributeParams{
		Namespace:        c.Namespace,
		Name:             c.Name,
		Type:             attrType,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceSearchAttributeRenameCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	renameSearchAttribute := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.RenameSearchAttribute)
	return renameSearchAttribute(namespace.RenameSearchAttributeParams{
		Namespace:                         c.Namespace,
		ExistingCustomSearchAttributeName: c.ExistingName,
		NewCustomSearchAttributeName:      c.NewName,
		ResourceVersion:                   c.ResourceVersion,
		AsyncOperationID:                  c.AsyncOperationId,
	})
}
