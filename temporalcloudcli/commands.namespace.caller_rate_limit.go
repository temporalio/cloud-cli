package temporalcloudcli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

const (
	callerTypeNamespace = "namespace"
	// callerRateLimitBlocksAllWarning is the rps 0 warning shared by set and apply. Both print it
	// with Println, a no-op in JSON mode, matching PromptApply skipping the diff there.
	callerRateLimitBlocksAllWarning = "rps 0 blocks all traffic from the matched callers"
)

// callerRateLimitChange is one rule to create (existing == nil) or update. key orders apply's changes.
type callerRateLimitChange struct {
	key      string
	existing *namespacev1.CallerRateLimit
	spec     *namespacev1.CallerRateLimitSpec
}

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
		return callerRateLimitNotFoundErr(c.Namespace, filter)
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

// AIDEV-NOTE: A missing rule is an error even with --idempotent, matching the
// custom-role and nexus endpoint deletes. --idempotent only covers the Delete RPC
// itself returning NotFound (e.g. a concurrent delete), via the poller.
func (c *CloudNamespaceCallerRateLimitDeleteCommand) run(cctx *CommandContext, _ []string) error {
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
		return callerRateLimitNotFoundErr(c.Namespace, filter)
	}
	yes, err := cctx.GetPrompter().PromptYes("Delete")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}
	rv := rl.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteCallerRateLimit(cctx, &cloudservice.DeleteCallerRateLimitRequest{
		Namespace:         c.Namespace,
		CallerRateLimitId: rl.Id,
		ResourceVersion:   rv,
		AsyncOperationId:  c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

// AIDEV-NOTE: set is an upsert. PromptApply is always called; existing is a typed nil
// when the rule doesn't exist (an untyped nil makes PrintDiff fail silently).
func (c *CloudNamespaceCallerRateLimitSetCommand) run(cctx *CommandContext, _ []string) error {
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
	rps := float32FlagToFloat64(c.Rps)
	spec := callerRateLimitSpec(filter, rps)
	var existingSpec *namespacev1.CallerRateLimitSpec
	if rl != nil {
		existingSpec = rl.Spec
	}
	if callerRateLimitBlocksAll(spec) {
		cctx.Printer.Println("Warning: " + callerRateLimitBlocksAllWarning + ".")
	}
	yes, err := cctx.GetPrompter().PromptApply(existingSpec, spec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set.")
	}
	ch := callerRateLimitChange{existing: rl, spec: spec}
	return upsertCallerRateLimit(cctx, client, c.Namespace, ch, c.AsyncOperationOptions, c.ResourceVersion)
}

// upsertCallerRateLimit creates ch.spec when ch.existing is nil, otherwise updates ch.existing,
// and, unless --async, waits for the async operation. resourceVersion overrides the existing
// rule's resource version when set.
// AIDEV-NOTE: --resource-version is ignored on create, as custom-role apply does.
func upsertCallerRateLimit(
	cctx *CommandContext,
	client cloudservice.CloudServiceClient,
	namespace string,
	ch callerRateLimitChange,
	asyncOpts AsyncOperationOptions,
	resourceVersion string,
) error {
	poller := cctx.GetPoller(client, asyncOpts)
	if ch.existing == nil {
		resp, err := client.CreateCallerRateLimit(cctx, &cloudservice.CreateCallerRateLimitRequest{
			Namespace:        namespace,
			Spec:             ch.spec,
			AsyncOperationId: asyncOpts.AsyncOperationId,
		})
		return poller.HandleCreateAsyncOperationResponse(cctx, resp, err)
	}
	rv := ch.existing.ResourceVersion
	if resourceVersion != "" {
		rv = resourceVersion
	}
	resp, err := client.UpdateCallerRateLimit(cctx, &cloudservice.UpdateCallerRateLimitRequest{
		Namespace:         namespace,
		CallerRateLimitId: ch.existing.Id,
		Spec:              ch.spec,
		ResourceVersion:   rv,
		AsyncOperationId:  asyncOpts.AsyncOperationId,
	})
	return poller.HandleUpdateOperation(cctx, resp, err)
}

// callerRateLimitBlocksAll reports whether spec has a rule whose rps is 0, which blocks all traffic
// from the matched callers.
// AIDEV-TODO: PrintDiff omits proto default values, so rps 0 is invisible in set's and apply's
// diffs; add EmitDefaultValues to printer.DiffOptions (and thread through the prompter) as a
// follow-up, then drop the warning.
// AIDEV-NOTE: The warning is not covered by unit tests; text output includes poller timestamps.
func callerRateLimitBlocksAll(spec *namespacev1.CallerRateLimitSpec) bool {
	switch rule := spec.GetRule().(type) {
	case *namespacev1.CallerRateLimitSpec_NamespaceDefault:
		return rule.NamespaceDefault.GetRps() == 0
	case *namespacev1.CallerRateLimitSpec_NamespaceLimit:
		return rule.NamespaceLimit.GetRps() == 0
	}
	return false
}

// callerRateLimitSpec builds the rule for the caller selected by filter.
func callerRateLimitSpec(
	filter *cloudservice.GetCallerRateLimitsRequest_CallerFilter,
	rps float64,
) *namespacev1.CallerRateLimitSpec {
	if filter.Default {
		return &namespacev1.CallerRateLimitSpec{
			Rule: &namespacev1.CallerRateLimitSpec_NamespaceDefault{
				NamespaceDefault: &namespacev1.NamespaceDefaultCallerRateLimit{Rps: rps},
			},
		}
	}
	return &namespacev1.CallerRateLimitSpec{
		Rule: &namespacev1.CallerRateLimitSpec_NamespaceLimit{
			NamespaceLimit: &namespacev1.NamespaceCallerRateLimit{NamespaceId: filter.Caller, Rps: rps},
		},
	}
}

// AIDEV-NOTE: gen-commands only emits float32 flags; round-trip through the
// shortest float32 decimal so --rps 0.1 is sent as 0.1, not 0.10000000149.
func float32FlagToFloat64(f float32) float64 {
	v, _ := strconv.ParseFloat(strconv.FormatFloat(float64(f), 'g', -1, 32), 64)
	return v
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

// callerRateLimitNotFoundErr describes a missing rule in terms of the flag the user passed.
func callerRateLimitNotFoundErr(
	namespace string,
	filter *cloudservice.GetCallerRateLimitsRequest_CallerFilter,
) error {
	if filter.Default {
		return fmt.Errorf("default caller rate limit not found on namespace %q", namespace)
	}
	return fmt.Errorf("caller rate limit for caller namespace %q not found on namespace %q",
		filter.Caller, namespace)
}
