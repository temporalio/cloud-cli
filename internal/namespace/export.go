package namespace

import (
	"context"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
)

// S3ExportSinkParams contains S3-specific parameters for export sink operations.
type S3ExportSinkParams struct {
	RoleName     string
	BucketName   string
	Region       string
	AwsAccountID string
	KmsArn       string
}

// GCSExportSinkParams contains GCS-specific parameters for export sink operations.
type GCSExportSinkParams struct {
	SaID         string
	BucketName   string
	GcpProjectID string
	Region       string
}

// CreateExportSinkParams contains parameters for creating an export sink.
// Exactly one of S3 or GCS must be set.
type CreateExportSinkParams struct {
	Namespace        string
	SinkName         string
	S3               *S3ExportSinkParams
	GCS              *GCSExportSinkParams
	AsyncOperationID string
}

// UpdateExportSinkParams contains parameters for updating an export sink.
// Exactly one of S3 or GCS must be set.
type UpdateExportSinkParams struct {
	Namespace        string
	SinkName         string
	S3               *S3ExportSinkParams
	GCS              *GCSExportSinkParams
	ResourceVersion  string
	AsyncOperationID string
}

// ValidateExportSinkParams contains parameters for validating an export sink configuration.
// Exactly one of S3 or GCS must be set.
type ValidateExportSinkParams struct {
	Namespace string
	SinkName  string
	S3        *S3ExportSinkParams
	GCS       *GCSExportSinkParams
}

// EnableExportSinkParams contains parameters for enabling an export sink.
type EnableExportSinkParams struct {
	Namespace        string
	SinkName         string
	ResourceVersion  string
	AsyncOperationID string
}

// DisableExportSinkParams contains parameters for disabling an export sink.
type DisableExportSinkParams struct {
	Namespace        string
	SinkName         string
	ResourceVersion  string
	AsyncOperationID string
}

// DeleteExportSinkParams contains parameters for deleting an export sink.
type DeleteExportSinkParams struct {
	Namespace        string
	SinkName         string
	ResourceVersion  string
	AsyncOperationID string
}

// GetExportSink retrieves a single export sink by name for the specified namespace.
func (c *Client) GetExportSink(ctx context.Context, namespaceName, sinkName string) (*namespacev1.ExportSink, error) {
	res, err := c.Cloud.GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: namespaceName,
		Name:      sinkName,
	})
	if err != nil {
		return nil, err
	}
	return res.Sink, nil
}

// ListExportSinks returns all export sinks for the specified namespace, handling pagination.
func (c *Client) ListExportSinks(ctx context.Context, namespaceName string) ([]*namespacev1.ExportSink, error) {
	var sinks []*namespacev1.ExportSink
	var pageToken string
	for {
		res, err := c.Cloud.GetNamespaceExportSinks(ctx, &cloudservice.GetNamespaceExportSinksRequest{
			Namespace: namespaceName,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, res.Sinks...)
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	return sinks, nil
}

// CreateExportSink creates a new export sink for the specified namespace.
// The sink is created in the enabled state. Exactly one of params.S3 or params.GCS must be set.
func (c *Client) CreateExportSink(ctx context.Context, params CreateExportSinkParams) (*operation.AsyncOperation, error) {
	spec := &namespacev1.ExportSinkSpec{
		Name:    params.SinkName,
		Enabled: true,
	}
	switch {
	case params.S3 != nil:
		spec.S3 = &sinkv1.S3Spec{
			RoleName:     params.S3.RoleName,
			BucketName:   params.S3.BucketName,
			Region:       params.S3.Region,
			AwsAccountId: params.S3.AwsAccountID,
			KmsArn:       params.S3.KmsArn,
		}
	case params.GCS != nil:
		spec.Gcs = &sinkv1.GCSSpec{
			SaId:         params.GCS.SaID,
			BucketName:   params.GCS.BucketName,
			GcpProjectId: params.GCS.GcpProjectID,
			Region:       params.GCS.Region,
		}
	}
	res, err := c.Cloud.CreateNamespaceExportSink(ctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// UpdateExportSink updates an existing export sink, preserving the current enabled state.
// Exactly one of params.S3 or params.GCS must be set.
func (c *Client) UpdateExportSink(ctx context.Context, params UpdateExportSinkParams) (*operation.AsyncOperation, error) {
	sink, err := c.GetExportSink(ctx, params.Namespace, params.SinkName)
	if err != nil {
		return nil, err
	}

	resourceVersion := sink.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	spec := &namespacev1.ExportSinkSpec{
		Name:    params.SinkName,
		Enabled: sink.GetSpec().GetEnabled(),
	}
	switch {
	case params.S3 != nil:
		spec.S3 = &sinkv1.S3Spec{
			RoleName:     params.S3.RoleName,
			BucketName:   params.S3.BucketName,
			Region:       params.S3.Region,
			AwsAccountId: params.S3.AwsAccountID,
			KmsArn:       params.S3.KmsArn,
		}
	case params.GCS != nil:
		spec.Gcs = &sinkv1.GCSSpec{
			SaId:         params.GCS.SaID,
			BucketName:   params.GCS.BucketName,
			GcpProjectId: params.GCS.GcpProjectID,
			Region:       params.GCS.Region,
		}
	}

	res, err := c.Cloud.UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Spec:             spec,
		ResourceVersion:  resourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// ValidateExportSink validates an export sink configuration without creating or updating it.
// Exactly one of params.S3 or params.GCS must be set.
func (c *Client) ValidateExportSink(ctx context.Context, params ValidateExportSinkParams) error {
	spec := &namespacev1.ExportSinkSpec{
		Name:    params.SinkName,
		Enabled: true,
	}
	switch {
	case params.S3 != nil:
		spec.S3 = &sinkv1.S3Spec{
			RoleName:     params.S3.RoleName,
			BucketName:   params.S3.BucketName,
			Region:       params.S3.Region,
			AwsAccountId: params.S3.AwsAccountID,
			KmsArn:       params.S3.KmsArn,
		}
	case params.GCS != nil:
		spec.Gcs = &sinkv1.GCSSpec{
			SaId:         params.GCS.SaID,
			BucketName:   params.GCS.BucketName,
			GcpProjectId: params.GCS.GcpProjectID,
			Region:       params.GCS.Region,
		}
	}
	_, err := c.Cloud.ValidateNamespaceExportSink(ctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec:      spec,
	})
	return err
}

// EnableExportSink enables a previously disabled export sink.
func (c *Client) EnableExportSink(ctx context.Context, params EnableExportSinkParams) (*operation.AsyncOperation, error) {
	sink, err := c.GetExportSink(ctx, params.Namespace, params.SinkName)
	if err != nil {
		return nil, err
	}

	resourceVersion := sink.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	newSpec := sink.Spec
	newSpec.Enabled = true

	res, err := c.Cloud.UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Spec:             newSpec,
		ResourceVersion:  resourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// DisableExportSink disables an export sink while preserving its configuration.
func (c *Client) DisableExportSink(ctx context.Context, params DisableExportSinkParams) (*operation.AsyncOperation, error) {
	sink, err := c.GetExportSink(ctx, params.Namespace, params.SinkName)
	if err != nil {
		return nil, err
	}

	resourceVersion := sink.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	newSpec := sink.Spec
	newSpec.Enabled = false

	res, err := c.Cloud.UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Spec:             newSpec,
		ResourceVersion:  resourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// DeleteExportSink deletes an export sink from the specified namespace.
func (c *Client) DeleteExportSink(ctx context.Context, params DeleteExportSinkParams) (*operation.AsyncOperation, error) {
	res, err := c.Cloud.DeleteNamespaceExportSink(ctx, &cloudservice.DeleteNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Name:             params.SinkName,
		ResourceVersion:  params.ResourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}
