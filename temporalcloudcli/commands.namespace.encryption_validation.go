package temporalcloudcli

import (
	"errors"
	"fmt"
	"strings"

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
	mode, err := parseEncryptionValidationMode(c.Mode)
	if err != nil {
		return err
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
	newSpec.EncryptionValidation = &namespacev1.EncryptionValidationSpec{
		Mode:           mode,
		MetadataKey:    c.MetadataKey,
		MetadataValues: c.MetadataValue,
		InspectHeader:  c.InspectHeader,
		InspectFailure: c.InspectFailure,
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
