package temporalcloudcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	resourcev1 "go.temporal.io/cloud-sdk/api/resource/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

const (
	callerTypeNamespace = "namespace"
	// callerRateLimitDefaultKey sorts before the caller namespace keys (callerTypeNamespace + ":" + id).
	callerRateLimitDefaultKey = "default"
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

// AIDEV-NOTE: apply is an additive upsert: rules in --spec are created or updated,
// rules not in --spec are left untouched. Changed rules are applied one at a time in
// key order (the same order as the diff), failing fast. A rule is skipped only when its
// spec matches and it is active, so a rerun skips applied rules and retries failed ones.
func (c *CloudNamespaceCallerRateLimitApplyCommand) run(cctx *CommandContext, _ []string) error {
	specs, err := parseCallerRateLimitSpecs(cctx, c.Spec)
	if err != nil {
		return err
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	byKey, err := listAllCallerRateLimits(cctx, client, c.Namespace)
	if err != nil {
		return err
	}
	changes := planCallerRateLimitChanges(byKey, specs)
	if len(changes) == 0 {
		cctx.Printer.Println("No changes to apply: all rules in --spec already match.")
		return nil
	}
	if len(changes) > 1 && (c.AsyncOperationId != "" || c.ResourceVersion != "") {
		return errors.New("--async-operation-id and --resource-version can only be used when apply " +
			"changes a single rule")
	}
	if err := c.confirmChanges(cctx, changes); err != nil {
		return err
	}
	var applied []string
	for _, ch := range changes {
		err := upsertCallerRateLimit(cctx, client, c.Namespace, ch, c.AsyncOperationOptions, c.ResourceVersion)
		if err != nil {
			done := "none"
			if len(applied) > 0 {
				done = strings.Join(applied, ", ")
			}
			return fmt.Errorf("failed to apply %s: %w (already applied: %s)",
				describeCallerRateLimit(ch.spec), err, done)
		}
		applied = append(applied, describeCallerRateLimit(ch.spec))
	}
	return nil
}

// confirmChanges prints one plan line per changed rule, then shows one diff
// of the changed rules and asks for confirmation.
// AIDEV-NOTE: PromptApply diffs a single proto.Message, so the rules are wrapped in a
// GetCallerRateLimitsResponse. Only Spec is set on each rule so the diff shows spec content
// only. Creates appear only on the modified side.
func (c *CloudNamespaceCallerRateLimitApplyCommand) confirmChanges(
	cctx *CommandContext,
	changes []callerRateLimitChange,
) error {
	before := &cloudservice.GetCallerRateLimitsResponse{}
	after := &cloudservice.GetCallerRateLimitsResponse{}
	for _, ch := range changes {
		action := "Create"
		if ch.existing != nil {
			action = "Update"
			before.CallerRateLimits = append(before.CallerRateLimits,
				&namespacev1.CallerRateLimit{Spec: ch.existing.Spec})
		}
		after.CallerRateLimits = append(after.CallerRateLimits, &namespacev1.CallerRateLimit{Spec: ch.spec})
		// The diff doesn't name the rule, so the plan line does.
		line := action + " " + describeCallerRateLimit(ch.spec)
		if callerRateLimitBlocksAll(ch.spec) {
			line += " (warning: " + callerRateLimitBlocksAllWarning + ")"
		}
		cctx.Printer.Println(line)
	}
	yes, err := cctx.GetPrompter().PromptApply(before, after, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting apply.")
	}
	return nil
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

// parseCallerRateLimitSpecs parses --spec as a JSON array of specs and rejects duplicate rules.
// Entry numbers in errors are 1-based.
func parseCallerRateLimitSpecs(cctx *CommandContext, raw string) ([]*namespacev1.CallerRateLimitSpec, error) {
	data, err := loadJSONSpec(raw)
	if err != nil {
		return nil, err
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse JSON spec: expected an array of caller rate limits: %w", err)
	}
	specs := make([]*namespacev1.CallerRateLimitSpec, len(entries))
	// AIDEV-NOTE: Duplicates are rejected client-side, unlike command-implementation.md's
	// "let the server detect duplicates": each rule is a separate request, so the server
	// can't see a duplicate within one file, and it would cause a partial apply.
	firstEntry := map[string]int{}
	for i, entry := range entries {
		spec := &namespacev1.CallerRateLimitSpec{}
		if err := cctx.UnmarshalProtoJSON(entry, spec); err != nil {
			return nil, fmt.Errorf("failed to parse JSON spec entry %d: %w", i+1, err)
		}
		key := callerRateLimitKey(spec)
		if first, ok := firstEntry[key]; ok {
			return nil, fmt.Errorf("duplicate %s in spec (entries %d and %d)",
				describeCallerRateLimit(spec), first, i+1)
		}
		firstEntry[key] = i + 1
		specs[i] = spec
	}
	return specs, nil
}

// planCallerRateLimitChanges returns the specs that need a create or update, sorted by key.
// AIDEV-NOTE: A matching rule is skipped only when it is active. In any other state (e.g. a
// failed update) it is updated again: the server keeps the desired spec on a failed rule and
// only treats an update as a no-op when the rule is active.
func planCallerRateLimitChanges(
	byKey map[string]*namespacev1.CallerRateLimit,
	specs []*namespacev1.CallerRateLimitSpec,
) []callerRateLimitChange {
	var changes []callerRateLimitChange
	for _, spec := range specs {
		key := callerRateLimitKey(spec)
		rl := byKey[key]
		if rl != nil && proto.Equal(rl.Spec, spec) && rl.GetState() == resourcev1.ResourceState_RESOURCE_STATE_ACTIVE {
			continue
		}
		changes = append(changes, callerRateLimitChange{key: key, existing: rl, spec: spec})
	}
	slices.SortFunc(changes, func(a, b callerRateLimitChange) int { return strings.Compare(a.key, b.key) })
	return changes
}

// listAllCallerRateLimits pages through every rule on the namespace, indexed by key.
// AIDEV-NOTE: Existing rules whose caller type this CLI doesn't recognize (Rule decodes as
// nil, keyed "") are skipped: apply can't target them, and skipping them means a spec entry
// with no rule ({}) never matches an existing rule. It is planned as a create and the server
// rejects it.
func listAllCallerRateLimits(
	cctx *CommandContext,
	client cloudservice.CloudServiceClient,
	namespace string,
) (map[string]*namespacev1.CallerRateLimit, error) {
	byKey := map[string]*namespacev1.CallerRateLimit{}
	var pageToken string
	for {
		res, err := client.GetCallerRateLimits(cctx, &cloudservice.GetCallerRateLimitsRequest{
			Namespace: namespace,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}
		for _, rl := range res.CallerRateLimits {
			key := callerRateLimitKey(rl.Spec)
			if key == "" {
				continue
			}
			if prev, ok := byKey[key]; ok {
				return nil, fmt.Errorf("found more than one %s on namespace %q (ids: %s, %s)",
					describeCallerRateLimit(rl.Spec), namespace, prev.Id, rl.Id)
			}
			byKey[key] = rl
		}
		if res.NextPageToken == "" {
			return byKey, nil
		}
		pageToken = res.NextPageToken
	}
}

// callerRateLimitKey identifies the caller a rule applies to; "" when the spec has no rule.
func callerRateLimitKey(spec *namespacev1.CallerRateLimitSpec) string {
	switch rule := spec.GetRule().(type) {
	case *namespacev1.CallerRateLimitSpec_NamespaceDefault:
		return callerRateLimitDefaultKey
	case *namespacev1.CallerRateLimitSpec_NamespaceLimit:
		return callerTypeNamespace + ":" + rule.NamespaceLimit.GetNamespaceId()
	}
	return ""
}

// describeCallerRateLimit names a rule for messages.
func describeCallerRateLimit(spec *namespacev1.CallerRateLimitSpec) string {
	switch rule := spec.GetRule().(type) {
	case *namespacev1.CallerRateLimitSpec_NamespaceDefault:
		return "default caller rate limit"
	case *namespacev1.CallerRateLimitSpec_NamespaceLimit:
		return fmt.Sprintf("caller rate limit for caller namespace %q", rule.NamespaceLimit.GetNamespaceId())
	}
	return "caller rate limit with no rule"
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
	return nil, fmt.Errorf("found more than one %s on namespace %q (ids: %s)",
		describeCallerRateLimit(callerRateLimitSpec(filter, 0)), namespace, strings.Join(ids, ", "))
}

// callerRateLimitNotFoundErr describes a missing rule in terms of the flag the user passed.
func callerRateLimitNotFoundErr(
	namespace string,
	filter *cloudservice.GetCallerRateLimitsRequest_CallerFilter,
) error {
	return fmt.Errorf("%s not found on namespace %q",
		describeCallerRateLimit(callerRateLimitSpec(filter, 0)), namespace)
}
