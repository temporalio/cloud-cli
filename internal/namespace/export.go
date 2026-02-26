package namespace

import (
	"context"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
)

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

// CreateS3ExportSinkParams contains parameters for creating an S3 export sink.
type CreateS3ExportSinkParams struct {
	Namespace        string
	SinkName         string
	RoleName         string
	BucketName       string
	Region           string
	AwsAccountID     string
	KmsArn           string
	AsyncOperationID string
}

// CreateS3ExportSink creates a new S3 export sink for the specified namespace.
// The sink is created in the enabled state.
func (c *Client) CreateS3ExportSink(ctx context.Context, params CreateS3ExportSinkParams) (*operation.AsyncOperation, error) {
	res, err := c.Cloud.CreateNamespaceExportSink(ctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec: &namespacev1.ExportSinkSpec{
			Name:    params.SinkName,
			Enabled: true,
			S3: &sinkv1.S3Spec{
				RoleName:     params.RoleName,
				BucketName:   params.BucketName,
				Region:       params.Region,
				AwsAccountId: params.AwsAccountID,
				KmsArn:       params.KmsArn,
			},
		},
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// UpdateS3ExportSinkParams contains parameters for updating an S3 export sink.
type UpdateS3ExportSinkParams struct {
	Namespace        string
	SinkName         string
	RoleName         string
	BucketName       string
	Region           string
	AwsAccountID     string
	KmsArn           string
	ResourceVersion  string
	AsyncOperationID string
}

// UpdateS3ExportSink updates an existing S3 export sink, preserving the current enabled state.
func (c *Client) UpdateS3ExportSink(ctx context.Context, params UpdateS3ExportSinkParams) (*operation.AsyncOperation, error) {
	sink, err := c.GetExportSink(ctx, params.Namespace, params.SinkName)
	if err != nil {
		return nil, err
	}

	resourceVersion := sink.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	res, err := c.Cloud.UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec: &namespacev1.ExportSinkSpec{
			Name:    params.SinkName,
			Enabled: sink.GetSpec().GetEnabled(),
			S3: &sinkv1.S3Spec{
				RoleName:     params.RoleName,
				BucketName:   params.BucketName,
				Region:       params.Region,
				AwsAccountId: params.AwsAccountID,
				KmsArn:       params.KmsArn,
			},
		},
		ResourceVersion:  resourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// ValidateS3ExportSinkParams contains parameters for validating an S3 export sink configuration.
type ValidateS3ExportSinkParams struct {
	Namespace    string
	SinkName     string
	RoleName     string
	BucketName   string
	Region       string
	AwsAccountID string
	KmsArn       string
}

// ValidateS3ExportSink validates an S3 export sink configuration without creating or updating it.
func (c *Client) ValidateS3ExportSink(ctx context.Context, params ValidateS3ExportSinkParams) error {
	_, err := c.Cloud.ValidateNamespaceExportSink(ctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec: &namespacev1.ExportSinkSpec{
			Name:    params.SinkName,
			Enabled: true,
			S3: &sinkv1.S3Spec{
				RoleName:     params.RoleName,
				BucketName:   params.BucketName,
				Region:       params.Region,
				AwsAccountId: params.AwsAccountID,
				KmsArn:       params.KmsArn,
			},
		},
	})
	return err
}

// CreateGCSExportSinkParams contains parameters for creating a GCS export sink.
type CreateGCSExportSinkParams struct {
	Namespace        string
	SinkName         string
	SaID             string
	BucketName       string
	GcpProjectID     string
	Region           string
	AsyncOperationID string
}

// CreateGCSExportSink creates a new GCS export sink for the specified namespace.
// The sink is created in the enabled state.
func (c *Client) CreateGCSExportSink(ctx context.Context, params CreateGCSExportSinkParams) (*operation.AsyncOperation, error) {
	res, err := c.Cloud.CreateNamespaceExportSink(ctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec: &namespacev1.ExportSinkSpec{
			Name:    params.SinkName,
			Enabled: true,
			Gcs: &sinkv1.GCSSpec{
				SaId:         params.SaID,
				BucketName:   params.BucketName,
				GcpProjectId: params.GcpProjectID,
				Region:       params.Region,
			},
		},
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// UpdateGCSExportSinkParams contains parameters for updating a GCS export sink.
type UpdateGCSExportSinkParams struct {
	Namespace        string
	SinkName         string
	SaID             string
	BucketName       string
	GcpProjectID     string
	Region           string
	ResourceVersion  string
	AsyncOperationID string
}

// UpdateGCSExportSink updates an existing GCS export sink, preserving the current enabled state.
func (c *Client) UpdateGCSExportSink(ctx context.Context, params UpdateGCSExportSinkParams) (*operation.AsyncOperation, error) {
	sink, err := c.GetExportSink(ctx, params.Namespace, params.SinkName)
	if err != nil {
		return nil, err
	}

	resourceVersion := sink.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	res, err := c.Cloud.UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec: &namespacev1.ExportSinkSpec{
			Name:    params.SinkName,
			Enabled: sink.GetSpec().GetEnabled(),
			Gcs: &sinkv1.GCSSpec{
				SaId:         params.SaID,
				BucketName:   params.BucketName,
				GcpProjectId: params.GcpProjectID,
				Region:       params.Region,
			},
		},
		ResourceVersion:  resourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// ValidateGCSExportSinkParams contains parameters for validating a GCS export sink configuration.
type ValidateGCSExportSinkParams struct {
	Namespace    string
	SinkName     string
	SaID         string
	BucketName   string
	GcpProjectID string
	Region       string
}

// ValidateGCSExportSink validates a GCS export sink configuration without creating or updating it.
func (c *Client) ValidateGCSExportSink(ctx context.Context, params ValidateGCSExportSinkParams) error {
	_, err := c.Cloud.ValidateNamespaceExportSink(ctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec: &namespacev1.ExportSinkSpec{
			Name:    params.SinkName,
			Enabled: true,
			Gcs: &sinkv1.GCSSpec{
				SaId:         params.SaID,
				BucketName:   params.BucketName,
				GcpProjectId: params.GcpProjectID,
				Region:       params.Region,
			},
		},
	})
	return err
}

// EnableExportSinkParams contains parameters for enabling an export sink.
type EnableExportSinkParams struct {
	Namespace        string
	SinkName         string
	ResourceVersion  string
	AsyncOperationID string
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

// DisableExportSinkParams contains parameters for disabling an export sink.
type DisableExportSinkParams struct {
	Namespace        string
	SinkName         string
	ResourceVersion  string
	AsyncOperationID string
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

// DeleteExportSinkParams contains parameters for deleting an export sink.
type DeleteExportSinkParams struct {
	Namespace        string
	SinkName         string
	ResourceVersion  string
	AsyncOperationID string
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
