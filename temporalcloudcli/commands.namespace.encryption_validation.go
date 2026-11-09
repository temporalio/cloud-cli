package temporalcloudcli

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

const (
	encryptionValidationModeDisabled = "disabled"
	encryptionValidationModeWarn     = "warn"
	encryptionValidationModeDeny     = "deny"

	// Server-side defaults applied when the metadata key or values are left empty.
	defaultEncryptionValidationMetadataKey   = "encoding"
	defaultEncryptionValidationMetadataValue = "binary/encrypted"
)

var encryptionValidationSetFlags = []string{"mode", "metadata-key", "metadata-value", "inspect-header", "inspect-failure"}

func (c *CloudNamespaceEncryptionValidationGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	var spec *namespacev1.EncryptionValidationSpec
	if res.Namespace.Spec != nil {
		spec = withEncryptionValidationDefaults(res.Namespace.Spec.EncryptionValidation)
	}
	return cctx.Printer.PrintResource(struct {
		Namespace string                                `json:"namespace"`
		Spec      *namespacev1.EncryptionValidationSpec `json:"spec"`
	}{
		Namespace: res.Namespace.Namespace,
		Spec:      spec,
	}, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceEncryptionValidationSetCommand) run(cctx *CommandContext, _ []string) error {
	flags := c.Command.Flags()
	if !slices.ContainsFunc(encryptionValidationSetFlags, flags.Changed) {
		return errors.New("at least one of --mode, --metadata-key, --metadata-value, --inspect-header, " +
			"or --inspect-failure must be set")
	}
	if flags.Changed("metadata-value") && slices.ContainsFunc(c.MetadataValue, isBlank) {
		return errors.New("--metadata-value cannot be blank")
	}
	var mode namespacev1.EncryptionValidationSpec_EncryptionValidationMode
	if flags.Changed("mode") {
		var err error
		mode, err = parseEncryptionValidationMode(c.Mode)
		if err != nil {
			return err
		}
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	ns := res.Namespace
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	if newSpec.EncryptionValidation == nil {
		if !flags.Changed("mode") {
			return errors.New("--mode is required when encryption validation is not configured")
		}
		newSpec.EncryptionValidation = &namespacev1.EncryptionValidationSpec{}
	}
	if flags.Changed("mode") {
		newSpec.EncryptionValidation.Mode = mode
	}
	if flags.Changed("metadata-key") {
		newSpec.EncryptionValidation.MetadataKey = c.MetadataKey
	}
	if flags.Changed("metadata-value") {
		newSpec.EncryptionValidation.MetadataValues = c.MetadataValue
	}
	if flags.Changed("inspect-header") {
		newSpec.EncryptionValidation.InspectHeader = c.InspectHeader
	}
	if flags.Changed("inspect-failure") {
		newSpec.EncryptionValidation.InspectFailure = c.InspectFailure
	}
	if proto.Equal(ns.Spec, newSpec) {
		return encryptionValidationNothingChanged(cctx, c.Idempotent)
	}

	yes, err := cctx.GetPrompter().PromptApply(ns.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set.")
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceEncryptionValidationDisableCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	ns := res.Namespace
	if ns.Spec.EncryptionValidation == nil {
		return encryptionValidationNothingChanged(cctx, c.Idempotent)
	}
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	newSpec.EncryptionValidation.Mode = namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DISABLED
	if proto.Equal(ns.Spec, newSpec) {
		return encryptionValidationNothingChanged(cctx, c.Idempotent)
	}

	yes, err := cctx.GetPrompter().PromptApply(ns.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting disable.")
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

// encryptionValidationNothingChanged handles a request that would not change the namespace.
// It matches the server's "nothing to change" handling used by other commands: an error by
// default, or an "unchanged" result when --idempotent is set.
func encryptionValidationNothingChanged(cctx *CommandContext, idempotent bool) error {
	if idempotent {
		return cctx.Printer.PrintStructured(newUnchangedResult(), printer.StructuredOptions{})
	}
	return errors.New("nothing to change: encryption validation is already in the requested state " +
		"(use --idempotent to succeed without changes)")
}

// withEncryptionValidationDefaults returns a copy of spec with the server defaults filled in
// for an empty metadata key or empty metadata values, so `get` shows what is actually enforced.
func withEncryptionValidationDefaults(spec *namespacev1.EncryptionValidationSpec) *namespacev1.EncryptionValidationSpec {
	if spec == nil {
		return nil
	}
	out := proto.Clone(spec).(*namespacev1.EncryptionValidationSpec)
	if out.MetadataKey == "" {
		out.MetadataKey = defaultEncryptionValidationMetadataKey
	}
	if len(out.MetadataValues) == 0 {
		out.MetadataValues = []string{defaultEncryptionValidationMetadataValue}
	}
	return out
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func parseEncryptionValidationMode(value string) (namespacev1.EncryptionValidationSpec_EncryptionValidationMode, error) {
	switch strings.ToLower(value) {
	case encryptionValidationModeDisabled:
		return namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DISABLED, nil
	case encryptionValidationModeWarn:
		return namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN, nil
	case encryptionValidationModeDeny:
		return namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DENY, nil
	default:
		return namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_UNSPECIFIED, fmt.Errorf(
			"invalid encryption validation mode %q: must be disabled, warn, or deny",
			value,
		)
	}
}

func encryptionValidationFromCreateFlags(c *CloudNamespaceCreateCommand) (*namespacev1.EncryptionValidationSpec, error) {
	flags := c.Command.Flags()
	spec := &namespacev1.EncryptionValidationSpec{}
	if flags.Changed("encryption-validation-mode") {
		mode, err := parseEncryptionValidationMode(c.EncryptionValidationMode)
		if err != nil {
			return nil, err
		}
		spec.Mode = mode
	}
	if flags.Changed("encryption-validation-metadata-key") {
		spec.MetadataKey = c.EncryptionValidationMetadataKey
	}
	if flags.Changed("encryption-validation-metadata-value") {
		spec.MetadataValues = c.EncryptionValidationMetadataValue
	}
	if flags.Changed("encryption-validation-inspect-header") {
		spec.InspectHeader = c.EncryptionValidationInspectHeader
	}
	if flags.Changed("encryption-validation-inspect-failure") {
		spec.InspectFailure = c.EncryptionValidationInspectFailure
	}
	if proto.Equal(spec, &namespacev1.EncryptionValidationSpec{}) {
		return nil, nil
	}
	if spec.Mode == namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_UNSPECIFIED {
		return nil, errors.New("--encryption-validation-mode is required when any encryption-validation flag is set")
	}
	return spec, nil
}
