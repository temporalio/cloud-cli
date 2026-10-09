package temporalcloudcli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

const (
	defaultEncryptionValidationMetadataKey   = converter.MetadataEncoding
	defaultEncryptionValidationMetadataValue = "binary/encrypted"

	encryptionValidationModeDisabled = "disabled"
	encryptionValidationModeWarn     = "warn"
	encryptionValidationModeDeny     = "deny"
)

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
		spec = res.Namespace.Spec.EncryptionValidation
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
	if flags.Changed("mode") {
		if _, err := parseEncryptionValidationMode(c.Mode); err != nil {
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
	if err := applyEncryptionValidationFlags(newSpec.EncryptionValidation, flags, ""); err != nil {
		return err
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

func (c *CloudNamespaceEncryptionValidationEnableCommand) run(cctx *CommandContext, _ []string) error {
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
	newSpec.EncryptionValidation = enableEncryptionValidation(newSpec.EncryptionValidation, c.Deny)

	yes, err := cctx.GetPrompter().PromptApply(ns.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting enable.")
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
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	if newSpec.EncryptionValidation == nil {
		newSpec.EncryptionValidation = &namespacev1.EncryptionValidationSpec{}
	} else {
		newSpec.EncryptionValidation = proto.Clone(newSpec.EncryptionValidation).(*namespacev1.EncryptionValidationSpec)
	}
	newSpec.EncryptionValidation.Mode = namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DISABLED

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

func enableEncryptionValidation(existing *namespacev1.EncryptionValidationSpec, deny bool) *namespacev1.EncryptionValidationSpec {
	mode := namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN
	if deny {
		mode = namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DENY
	}

	if existing == nil {
		return &namespacev1.EncryptionValidationSpec{
			Mode:           mode,
			MetadataKey:    defaultEncryptionValidationMetadataKey,
			MetadataValues: []string{defaultEncryptionValidationMetadataValue},
		}
	}

	spec := proto.Clone(existing).(*namespacev1.EncryptionValidationSpec)
	if spec.Mode == namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DISABLED {
		spec.Mode = mode
	}
	return spec
}

// applyEncryptionValidationFlags overwrites only the spec fields whose flags (prefix + name) were explicitly passed.
func applyEncryptionValidationFlags(
	spec *namespacev1.EncryptionValidationSpec,
	flagSet *pflag.FlagSet,
	prefix string,
) error {
	if flagSet.Changed(prefix + "mode") {
		raw, err := flagSet.GetString(prefix + "mode")
		if err != nil {
			return err
		}
		mode, err := parseEncryptionValidationMode(raw)
		if err != nil {
			return err
		}
		spec.Mode = mode
	}
	if flagSet.Changed(prefix + "metadata-key") {
		key, err := flagSet.GetString(prefix + "metadata-key")
		if err != nil {
			return err
		}
		spec.MetadataKey = key
	}
	if flagSet.Changed(prefix + "metadata-value") {
		values, err := flagSet.GetStringArray(prefix + "metadata-value")
		if err != nil {
			return err
		}
		spec.MetadataValues = values
	}
	if flagSet.Changed(prefix + "inspect-header") {
		inspectHeader, err := flagSet.GetBool(prefix + "inspect-header")
		if err != nil {
			return err
		}
		spec.InspectHeader = inspectHeader
	}
	if flagSet.Changed(prefix + "inspect-failure") {
		inspectFailure, err := flagSet.GetBool(prefix + "inspect-failure")
		if err != nil {
			return err
		}
		spec.InspectFailure = inspectFailure
	}
	return nil
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
	spec := &namespacev1.EncryptionValidationSpec{}
	if err := applyEncryptionValidationFlags(spec, c.Command.Flags(), "encryption-validation-"); err != nil {
		return nil, err
	}
	if proto.Equal(spec, &namespacev1.EncryptionValidationSpec{}) {
		return nil, nil
	}
	if spec.Mode == namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_UNSPECIFIED {
		return nil, errors.New("--encryption-validation-mode is required when any encryption-validation flag is set")
	}
	return spec, nil
}
