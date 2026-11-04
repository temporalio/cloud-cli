package temporalcloudcli

import (
	"errors"
	"fmt"
	"strings"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

const callerTypeNamespace = "namespace"

func (c *CloudNamespaceCallerRateLimitGetCommand) run(cctx *CommandContext, _ []string) error {
	filter, err := callerFilterFromFlags(c.CallerNamespaceDefault, c.CallerNamespace)
	if err != nil {
		return err
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	rl, err := findCallerRateLimit(cctx, client, c.Namespace, filter)
	if err != nil {
		return err
	}
	if rl == nil {
		if filter.Default {
			return fmt.Errorf("default caller rate limit not found on namespace %q", c.Namespace)
		}
		return fmt.Errorf("caller rate limit for caller namespace %q not found on namespace %q",
			filter.Caller, c.Namespace)
	}
	// EmitDefaultValues so a rule with rps 0 (blocks all) still shows its rps.
	return cctx.Printer.PrintResource(rl, printer.PrintResourceOptions{EmitDefaultValues: true})
}

func (c *CloudNamespaceCallerRateLimitListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	// AIDEV-NOTE: default + caller in one CallerFilter is a server-side union
	// (returns the default rule and the caller's rule). Intentionally allowed.
	var filter *cloudservice.GetCallerRateLimitsRequest_CallerFilter
	if c.Type.Value != "" || c.CallerNamespaceDefault || c.CallerNamespace != "" {
		// --type is validated by its FlagStringEnum; "namespace" is the only caller type today.
		callerType := c.Type.Value
		if callerType == "" {
			callerType = callerTypeNamespace
		}
		filter = &cloudservice.GetCallerRateLimitsRequest_CallerFilter{
			CallerType: callerType,
			Default:    c.CallerNamespaceDefault,
			Caller:     c.CallerNamespace,
		}
	}
	res, err := client.GetCallerRateLimits(cctx, &cloudservice.GetCallerRateLimitsRequest{
		Namespace:    c.Namespace,
		PageSize:     int32(c.PageSize),
		PageToken:    c.PageToken,
		CallerFilter: filter,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			CallerRateLimits []*namespacev1.CallerRateLimit
			NextPageToken    string
		}{CallerRateLimits: res.CallerRateLimits, NextPageToken: res.NextPageToken},
		// EmitDefaultValues so a rule with rps 0 (blocks all) still shows its rps.
		printer.PrintResourceOptions{Fields: []string{"Id", "State"}, EmitDefaultValues: true},
		printer.TableOptions{},
	)
}

// callerFilterFromFlags builds the server-side caller filter from the mutually
// exclusive --caller-namespace-default and --caller-namespace flags.
func callerFilterFromFlags(
	isDefault bool,
	callerNamespace string,
) (*cloudservice.GetCallerRateLimitsRequest_CallerFilter, error) {
	if !isDefault && callerNamespace == "" {
		return nil, errors.New("either --caller-namespace-default or --caller-namespace is required")
	}
	if isDefault && callerNamespace != "" {
		return nil, errors.New("--caller-namespace-default and --caller-namespace are mutually exclusive")
	}
	return &cloudservice.GetCallerRateLimitsRequest_CallerFilter{
		CallerType: callerTypeNamespace,
		Default:    isDefault,
		Caller:     callerNamespace,
	}, nil
}

// AIDEV-NOTE: The server-side caller filter returns at most one rule per
// (namespace, caller), so a single call without paging is enough. We still
// error on more than one match rather than silently picking one.
// Returns (nil, nil) when no rule matches; callers decide whether that is an error.
func findCallerRateLimit(
	cctx *CommandContext,
	client cloudservice.CloudServiceClient,
	namespace string,
	filter *cloudservice.GetCallerRateLimitsRequest_CallerFilter,
) (*namespacev1.CallerRateLimit, error) {
	res, err := client.GetCallerRateLimits(cctx, &cloudservice.GetCallerRateLimitsRequest{
		Namespace:    namespace,
		CallerFilter: filter,
	})
	if err != nil {
		return nil, err
	}
	switch len(res.CallerRateLimits) {
	case 0:
		return nil, nil
	case 1:
		return res.CallerRateLimits[0], nil
	}
	ids := make([]string, len(res.CallerRateLimits))
	for i, rl := range res.CallerRateLimits {
		ids[i] = rl.Id
	}
	return nil, fmt.Errorf("expected at most one caller rate limit on namespace %q, found %d (ids: %s)",
		namespace, len(ids), strings.Join(ids, ", "))
}
