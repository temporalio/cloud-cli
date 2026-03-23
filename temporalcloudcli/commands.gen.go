// Code generated. DO NOT EDIT.

package temporalcloudcli

import (
	"github.com/mattn/go-isatty"

	"github.com/spf13/cobra"

	"github.com/spf13/pflag"

	"github.com/temporalio/cli/cliext"

	"os"
)

var hasHighlighting = isatty.IsTerminal(os.Stdout.Fd())

type ClientOptions struct {
	ApiKey  string
	Server  string
	FlagSet *pflag.FlagSet
}

func (v *ClientOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.ApiKey, "api-key", "", "API key for authenticating with Temporal Cloud. Can be used instead of interactive login for automation and CI/CD pipelines.")
	f.StringVar(&v.Server, "server", "saas-api.tmprl-test.cloud:443", "Override the Temporal Cloud API server address. Used for connecting to non-production environments.")
	_ = f.MarkHidden("server")
}

type DiffOptions struct {
	VerboseDiff bool
	FlagSet     *pflag.FlagSet
}

func (v *DiffOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.BoolVar(&v.VerboseDiff, "verbose-diff", false, "Show detailed differences between the current and desired namespace configurations when changes are detected.")
}

type NamespaceOptions struct {
	Namespace string
	FlagSet   *pflag.FlagSet
}

func (v *NamespaceOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVarP(&v.Namespace, "namespace", "n", "", "The fully qualified namespace name in the format 'namespace.account' (e.g., 'my-namespace.my-account'). Required.")
	_ = cobra.MarkFlagRequired(f, "namespace")
}

type ResourceVersionOptions struct {
	ResourceVersion string
	FlagSet         *pflag.FlagSet
}

func (v *ResourceVersionOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVarP(&v.ResourceVersion, "resource-version", "v", "", "Resource version for optimistic concurrency control. If not provided, the current version is fetched automatically.")
}

type AsyncOperationOptions struct {
	Idempotent       bool
	AsyncOperationId string
	Async            bool
	FlagSet          *pflag.FlagSet
}

func (v *AsyncOperationOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.BoolVar(&v.Idempotent, "idempotent", false, "Succeed silently if the resource already exists or matches the specification. Without this flag, the command errors when no changes are needed.")
	f.StringVar(&v.AsyncOperationId, "async-operation-id", "", "Custom identifier for tracking this async operation. If not provided, a unique ID is generated automatically.")
	f.BoolVar(&v.Async, "async", false, "Return immediately after initiating the operation instead of waiting for completion. Use the returned operation ID to check status later.")
}

type UserIdentificationOptions struct {
	UserId    string
	UserEmail string
	FlagSet   *pflag.FlagSet
}

func (v *UserIdentificationOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.UserId, "user-id", "", "The ID of the user. Mutually exclusive with --user-email.")
	f.StringVar(&v.UserEmail, "user-email", "", "The email address of the user. Mutually exclusive with --user-id.")
}

type CodecServerOptions struct {
	CodecEndpoint                      string
	CodecPassAccessToken               bool
	CodecIncludeCrossOriginCredentials bool
	FlagSet                            *pflag.FlagSet
}

func (v *CodecServerOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.CodecEndpoint, "codec-endpoint", "", "HTTPS codec server endpoint URL.")
	f.BoolVar(&v.CodecPassAccessToken, "codec-pass-access-token", false, "Pass the user access token to the codec server endpoint.")
	f.BoolVar(&v.CodecIncludeCrossOriginCredentials, "codec-include-cross-origin-credentials", false, "Include cross-origin credentials in codec server requests.")
}

type CaCertificateOptions struct {
	CaCertificate     string
	CaCertificateFile string
	FlagSet           *pflag.FlagSet
}

func (v *CaCertificateOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.CaCertificate, "ca-certificate", "", "Base64-encoded CA certificate for mTLS authentication. Mutually exclusive with --ca-certificate-file.")
	f.StringVar(&v.CaCertificateFile, "ca-certificate-file", "", "Path to a CA certificate PEM file for mTLS authentication. Mutually exclusive with --ca-certificate.")
}

type CertificateFilterOptions struct {
	CertificateFilter     []string
	CertificateFilterFile string
	FlagSet               *pflag.FlagSet
}

func (v *CertificateFilterOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringArrayVar(&v.CertificateFilter, "certificate-filter", nil, "Certificate filter as a JSON object (e.g. '{\"commonName\":\"foo\"}'). Repeat to add multiple.")
	f.StringVar(&v.CertificateFilterFile, "certificate-filter-file", "", "Path to a JSON file containing a certificate filter object.")
}

type GroupIdOptions struct {
	GroupId string
	FlagSet *pflag.FlagSet
}

func (v *GroupIdOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.GroupId, "group-id", "", "The ID of the user group. Required.")
	_ = cobra.MarkFlagRequired(f, "group-id")
}

type ExportSinkOptions struct {
	SinkName string
	FlagSet  *pflag.FlagSet
}

func (v *ExportSinkOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.SinkName, "sink-name", "", "The name of the export sink. Required.")
	_ = cobra.MarkFlagRequired(f, "sink-name")
}

type ExportS3Options struct {
	RoleName     string
	BucketName   string
	Region       string
	AwsAccountId string
	KmsArn       string
	FlagSet      *pflag.FlagSet
}

func (v *ExportS3Options) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.RoleName, "role-name", "", "The IAM role ARN that Temporal Cloud assumes for writing to S3. Required.")
	_ = cobra.MarkFlagRequired(f, "role-name")
	f.StringVar(&v.BucketName, "bucket-name", "", "The name of the destination S3 bucket. Required.")
	_ = cobra.MarkFlagRequired(f, "bucket-name")
	f.StringVar(&v.Region, "region", "", "The AWS region where the S3 bucket is located. Required.")
	_ = cobra.MarkFlagRequired(f, "region")
	f.StringVar(&v.AwsAccountId, "aws-account-id", "", "The AWS account ID associated with the bucket and role. Required.")
	_ = cobra.MarkFlagRequired(f, "aws-account-id")
	f.StringVar(&v.KmsArn, "kms-arn", "", "The AWS KMS key ARN for server-side encryption of exported data. Optional.")
}

type ExportGcsOptions struct {
	SaId         string
	BucketName   string
	GcpProjectId string
	Region       string
	FlagSet      *pflag.FlagSet
}

func (v *ExportGcsOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.SaId, "sa-id", "", "The customer service account ID that Temporal Cloud impersonates for writing to GCS. Required.")
	_ = cobra.MarkFlagRequired(f, "sa-id")
	f.StringVar(&v.BucketName, "bucket-name", "", "The name of the destination GCS bucket. Required.")
	_ = cobra.MarkFlagRequired(f, "bucket-name")
	f.StringVar(&v.GcpProjectId, "gcp-project-id", "", "The GCP project ID associated with the bucket and service account. Required.")
	_ = cobra.MarkFlagRequired(f, "gcp-project-id")
	f.StringVar(&v.Region, "region", "", "The GCS bucket region. Required.")
	_ = cobra.MarkFlagRequired(f, "region")
}

type CloudCommand struct {
	Command cobra.Command
	ClientOptions
	cliext.CommonOptions
	ConfigDir    string
	DisablePopUp bool
	AutoConfirm  bool
}

func NewCloudCommand(cctx *CommandContext) *CloudCommand {
	var s CloudCommand
	s.Command.Use = "cloud"
	s.Command.Short = "Temporal Cloud command-line interface"
	if hasHighlighting {
		s.Command.Long = "The Temporal Cloud CLI provides commands for managing and operating Temporal Cloud resources,\nincluding namespaces, users, and account settings.\n\nExample:\n\n\x1b[1mcloud namespace get --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "The Temporal Cloud CLI provides commands for managing and operating Temporal Cloud resources,\nincluding namespaces, users, and account settings.\n\nExample:\n\n```\ncloud namespace get --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudAccountCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudConnectivityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudLoginCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudLogoutCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudWhoamiCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVar(&s.ConfigDir, "config-dir", "", "Directory path where CLI configuration files are stored, including authentication tokens and settings.")
	s.Command.PersistentFlags().BoolVar(&s.DisablePopUp, "disable-pop-up", false, "Prevent the CLI from opening a browser window during authentication. Useful for headless environments or when using alternative auth methods.")
	s.Command.PersistentFlags().BoolVar(&s.AutoConfirm, "auto-confirm", false, "Automatically confirm prompts and actions that require user confirmation. Useful for scripting and automation.")
	s.ClientOptions.BuildFlags(s.Command.PersistentFlags())
	s.CommonOptions.BuildFlags(s.Command.PersistentFlags())
	s.initCommand(cctx)
	return &s
}

type CloudAccountCommand struct {
	Parent  *CloudCommand
	Command cobra.Command
}

func NewCloudAccountCommand(cctx *CommandContext, parent *CloudCommand) *CloudAccountCommand {
	var s CloudAccountCommand
	s.Parent = parent
	s.Command.Use = "account"
	s.Command.Short = "Manage Temporal Cloud account"
	s.Command.Long = "Manage the Temporal Cloud account."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudAccountAuditLogCommand(cctx, &s).Command)
	return &s
}

type CloudAccountAuditLogCommand struct {
	Parent  *CloudAccountCommand
	Command cobra.Command
}

func NewCloudAccountAuditLogCommand(cctx *CommandContext, parent *CloudAccountCommand) *CloudAccountAuditLogCommand {
	var s CloudAccountAuditLogCommand
	s.Parent = parent
	s.Command.Use = "audit-log"
	s.Command.Short = "Manage audit logs"
	s.Command.Long = "Commands for working with account audit logs."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudAccountAuditLogGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkCommand(cctx, &s).Command)
	return &s
}

type CloudAccountAuditLogGetCommand struct {
	Parent  *CloudAccountAuditLogCommand
	Command cobra.Command
	ClientOptions
	PageSize  int
	PageToken string
	StartTime cliext.FlagTimestamp
	EndTime   cliext.FlagTimestamp
}

func NewCloudAccountAuditLogGetCommand(cctx *CommandContext, parent *CloudAccountAuditLogCommand) *CloudAccountAuditLogGetCommand {
	var s CloudAccountAuditLogGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get audit logs"
	s.Command.Long = "Returns a paginated list of audit logs for the account, optionally filtered by time range.\n\nExample:\n  temporal cloud account audit-log get --page-size 50\n  temporal cloud account audit-log get --start-time 2024-01-01T00:00:00Z --end-time 2024-02-01T00:00:00Z"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Number of logs to retrieve per page. Cannot exceed 1000. Defaults to 100.")
	s.Command.Flags().StringVar(&s.PageToken, "page-token", "", "Page token from a previous response to retrieve the next page.")
	s.Command.Flags().Var(&s.StartTime, "start-time", "Filter for logs at or after this UTC time (RFC3339 format, e.g. 2024-01-01T00:00:00Z). Defaults to 30 days ago.")
	s.Command.Flags().Var(&s.EndTime, "end-time", "Filter for logs before this UTC time (RFC3339 format, e.g. 2024-02-01T00:00:00Z). Defaults to current time.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkCommand struct {
	Parent  *CloudAccountAuditLogCommand
	Command cobra.Command
}

func NewCloudAccountAuditLogSinkCommand(cctx *CommandContext, parent *CloudAccountAuditLogCommand) *CloudAccountAuditLogSinkCommand {
	var s CloudAccountAuditLogSinkCommand
	s.Parent = parent
	s.Command.Use = "sink"
	s.Command.Short = "Manage audit log sinks"
	s.Command.Long = "Commands for working with account audit log sinks."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkDisableCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkEnableCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkKinesisCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkPubsubCommand(cctx, &s).Command)
	return &s
}

type CloudAccountAuditLogSinkDeleteCommand struct {
	Parent  *CloudAccountAuditLogSinkCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Name string
}

func NewCloudAccountAuditLogSinkDeleteCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkCommand) *CloudAccountAuditLogSinkDeleteCommand {
	var s CloudAccountAuditLogSinkDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete an audit log sink"
	s.Command.Long = "Delete an audit log sink for the account. This action is irreversible.\n\nExample:\n  temporal cloud account audit-log sink delete --name my-sink"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "The name of the audit log sink to delete. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkDisableCommand struct {
	Parent  *CloudAccountAuditLogSinkCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Name string
}

func NewCloudAccountAuditLogSinkDisableCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkCommand) *CloudAccountAuditLogSinkDisableCommand {
	var s CloudAccountAuditLogSinkDisableCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "disable [flags]"
	s.Command.Short = "Disable an audit log sink"
	s.Command.Long = "Disable an audit log sink for the account.\n\nExample:\n  temporal cloud account audit-log sink disable --name my-sink"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "The name of the audit log sink to disable. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkEnableCommand struct {
	Parent  *CloudAccountAuditLogSinkCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Name string
}

func NewCloudAccountAuditLogSinkEnableCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkCommand) *CloudAccountAuditLogSinkEnableCommand {
	var s CloudAccountAuditLogSinkEnableCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "enable [flags]"
	s.Command.Short = "Enable an audit log sink"
	s.Command.Long = "Enable an audit log sink for the account.\n\nExample:\n  temporal cloud account audit-log sink enable --name my-sink"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "The name of the audit log sink to enable. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkGetCommand struct {
	Parent  *CloudAccountAuditLogSinkCommand
	Command cobra.Command
	ClientOptions
	Name string
}

func NewCloudAccountAuditLogSinkGetCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkCommand) *CloudAccountAuditLogSinkGetCommand {
	var s CloudAccountAuditLogSinkGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get an audit log sink"
	s.Command.Long = "Returns the details of an audit log sink for the account.\n\nExample:\n  temporal cloud account audit-log sink get --name my-sink"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "The name of the audit log sink to get. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkKinesisCommand struct {
	Parent  *CloudAccountAuditLogSinkCommand
	Command cobra.Command
}

func NewCloudAccountAuditLogSinkKinesisCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkCommand) *CloudAccountAuditLogSinkKinesisCommand {
	var s CloudAccountAuditLogSinkKinesisCommand
	s.Parent = parent
	s.Command.Use = "kinesis"
	s.Command.Short = "Manage Kinesis audit log sinks"
	s.Command.Long = "Commands for managing Kinesis-based audit log sinks."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkKinesisCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkKinesisUpdateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkKinesisValidateCommand(cctx, &s).Command)
	return &s
}

type CloudAccountAuditLogSinkKinesisCreateCommand struct {
	Parent  *CloudAccountAuditLogSinkKinesisCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	Name           string
	RoleName       string
	DestinationUri string
	Region         string
}

func NewCloudAccountAuditLogSinkKinesisCreateCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkKinesisCommand) *CloudAccountAuditLogSinkKinesisCreateCommand {
	var s CloudAccountAuditLogSinkKinesisCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a Kinesis audit log sink"
	s.Command.Long = "Create an account audit log sink that streams audit events to Amazon Kinesis.\n\nTemporal Cloud assumes the specified IAM role to write events to the Kinesis\nstream identified by the destination URI.\n\nExample:\n  temporal cloud account audit-log sink kinesis create \\\n    --name my-sink \\\n    --role-name arn:aws:iam::123456789012:role/MyRole \\\n    --destination-uri arn:aws:kinesis:us-east-1:123456789012:stream/MyStream \\\n    --region us-east-1"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Name of the audit log sink. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVar(&s.RoleName, "role-name", "", "ARN of the IAM role that Temporal Cloud assumes to write to the Kinesis stream. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "role-name")
	s.Command.Flags().StringVar(&s.DestinationUri, "destination-uri", "", "ARN of the Kinesis stream to deliver audit log events to. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "destination-uri")
	s.Command.Flags().StringVar(&s.Region, "region", "", "AWS region where the Kinesis stream is located (e.g. us-east-1). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "region")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkKinesisUpdateCommand struct {
	Parent  *CloudAccountAuditLogSinkKinesisCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Name           string
	RoleName       string
	DestinationUri string
	Region         string
}

func NewCloudAccountAuditLogSinkKinesisUpdateCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkKinesisCommand) *CloudAccountAuditLogSinkKinesisUpdateCommand {
	var s CloudAccountAuditLogSinkKinesisUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update a Kinesis audit log sink"
	s.Command.Long = "Update an existing Kinesis audit log sink. Only the flags you provide are changed;\nomitted string flags retain their current values. \n\nExample:\n  temporal cloud account audit-log sink kinesis update \\\n    --name my-sink \\\n    --role-name arn:aws:iam::123456789012:role/NewRole"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Name of the audit log sink to update. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVar(&s.RoleName, "role-name", "", "ARN of the IAM role that Temporal Cloud assumes to write to the Kinesis stream. If omitted, the current value is kept.")
	s.Command.Flags().StringVar(&s.DestinationUri, "destination-uri", "", "ARN of the Kinesis stream to deliver audit log events to. If omitted, the current value is kept.")
	s.Command.Flags().StringVar(&s.Region, "region", "", "AWS region where the Kinesis stream is located (e.g. us-east-1). If omitted, the current value is kept.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkKinesisValidateCommand struct {
	Parent  *CloudAccountAuditLogSinkKinesisCommand
	Command cobra.Command
	ClientOptions
	RoleName       string
	DestinationUri string
	Region         string
}

func NewCloudAccountAuditLogSinkKinesisValidateCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkKinesisCommand) *CloudAccountAuditLogSinkKinesisValidateCommand {
	var s CloudAccountAuditLogSinkKinesisValidateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "validate [flags]"
	s.Command.Short = "Validate a Kinesis audit log sink configuration"
	s.Command.Long = "Validate an audit log sink configuration against Amazon Kinesis without creating it.\nUse this to verify that the IAM role and Kinesis stream are correctly configured\nbefore creating or updating the sink.\n\nExample:\n  temporal cloud account audit-log sink kinesis validate \\\n    --name my-sink \\\n    --role-name arn:aws:iam::123456789012:role/MyRole \\\n    --destination-uri arn:aws:kinesis:us-east-1:123456789012:stream/MyStream \\\n    --region us-east-1"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.RoleName, "role-name", "", "ARN of the IAM role that Temporal Cloud assumes to write to the Kinesis stream. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "role-name")
	s.Command.Flags().StringVar(&s.DestinationUri, "destination-uri", "", "ARN of the Kinesis stream to deliver audit log events to. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "destination-uri")
	s.Command.Flags().StringVar(&s.Region, "region", "", "AWS region where the Kinesis stream is located (e.g. us-east-1). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "region")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkListCommand struct {
	Parent  *CloudAccountAuditLogSinkCommand
	Command cobra.Command
	ClientOptions
	PageSize  int
	PageToken string
}

func NewCloudAccountAuditLogSinkListCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkCommand) *CloudAccountAuditLogSinkListCommand {
	var s CloudAccountAuditLogSinkListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List audit log sinks"
	s.Command.Long = "Returns a paginated list of audit log sinks for the account.\n\nExample:\n  temporal cloud account audit-log sink list\n  temporal cloud account audit-log sink list --page-size 50"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Number of sinks to retrieve per page. Cannot exceed 1000. Defaults to 100.")
	s.Command.Flags().StringVar(&s.PageToken, "page-token", "", "Page token from a previous response to retrieve the next page.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkPubsubCommand struct {
	Parent  *CloudAccountAuditLogSinkCommand
	Command cobra.Command
}

func NewCloudAccountAuditLogSinkPubsubCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkCommand) *CloudAccountAuditLogSinkPubsubCommand {
	var s CloudAccountAuditLogSinkPubsubCommand
	s.Parent = parent
	s.Command.Use = "pubsub"
	s.Command.Short = "Manage PubSub audit log sinks"
	s.Command.Long = "Commands for managing PubSub audit log sinks."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkPubsubCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkPubsubUpdateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudAccountAuditLogSinkPubsubValidateCommand(cctx, &s).Command)
	return &s
}

type CloudAccountAuditLogSinkPubsubCreateCommand struct {
	Parent  *CloudAccountAuditLogSinkPubsubCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	Name             string
	ServiceAccountId string
	TopicName        string
	GcpProjectId     string
}

func NewCloudAccountAuditLogSinkPubsubCreateCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkPubsubCommand) *CloudAccountAuditLogSinkPubsubCreateCommand {
	var s CloudAccountAuditLogSinkPubsubCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a PubSub audit log sink"
	if hasHighlighting {
		s.Command.Long = "Creates a new PubSub audit log sink for the account using Google Cloud Pub/Sub.\n\nExample:\n\n\x1b[1mtemporal cloud account audit-log sink pubsub create \\\n  --name my-sink \\\n  --service-account-id my-sa@project.iam.gserviceaccount.com \\\n  --topic-name my-topic \\\n  --gcp-project-id my-project\x1b[0m"
	} else {
		s.Command.Long = "Creates a new PubSub audit log sink for the account using Google Cloud Pub/Sub.\n\nExample:\n\n```\ntemporal cloud account audit-log sink pubsub create \\\n  --name my-sink \\\n  --service-account-id my-sa@project.iam.gserviceaccount.com \\\n  --topic-name my-topic \\\n  --gcp-project-id my-project\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "The name of the audit log sink. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVar(&s.ServiceAccountId, "service-account-id", "", "The GCP service account ID that Temporal Cloud impersonates for writing records to the customer's PubSub topic. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "service-account-id")
	s.Command.Flags().StringVar(&s.TopicName, "topic-name", "", "The destination PubSub topic name where audit logs will be sent. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "topic-name")
	s.Command.Flags().StringVar(&s.GcpProjectId, "gcp-project-id", "", "The GCP project ID of the PubSub topic and service account. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "gcp-project-id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkPubsubUpdateCommand struct {
	Parent  *CloudAccountAuditLogSinkPubsubCommand
	Command cobra.Command
	ClientOptions
	ResourceVersionOptions
	AsyncOperationOptions
	Name             string
	ServiceAccountId string
	TopicName        string
	GcpProjectId     string
}

func NewCloudAccountAuditLogSinkPubsubUpdateCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkPubsubCommand) *CloudAccountAuditLogSinkPubsubUpdateCommand {
	var s CloudAccountAuditLogSinkPubsubUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update a PubSub audit log sink"
	if hasHighlighting {
		s.Command.Long = "Updates an existing PubSub audit log sink for the account.\n\nExample:\n\n\x1b[1mtemporal cloud account audit-log sink pubsub update \\\n  --name my-sink \\\n  --service-account-id new-sa@project.iam.gserviceaccount.com \\\n  --topic-name new-topic \\\n  --gcp-project-id new-project\x1b[0m"
	} else {
		s.Command.Long = "Updates an existing PubSub audit log sink for the account.\n\nExample:\n\n```\ntemporal cloud account audit-log sink pubsub update \\\n  --name my-sink \\\n  --service-account-id new-sa@project.iam.gserviceaccount.com \\\n  --topic-name new-topic \\\n  --gcp-project-id new-project\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "The name of the audit log sink to update. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVar(&s.ServiceAccountId, "service-account-id", "", "The GCP service account ID that Temporal Cloud impersonates for writing records to the customer's PubSub topic.")
	s.Command.Flags().StringVar(&s.TopicName, "topic-name", "", "The destination PubSub topic name where audit logs will be sent.")
	s.Command.Flags().StringVar(&s.GcpProjectId, "gcp-project-id", "", "The GCP project ID of the PubSub topic and service account.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudAccountAuditLogSinkPubsubValidateCommand struct {
	Parent  *CloudAccountAuditLogSinkPubsubCommand
	Command cobra.Command
	ClientOptions
	ServiceAccountId string
	TopicName        string
	GcpProjectId     string
}

func NewCloudAccountAuditLogSinkPubsubValidateCommand(cctx *CommandContext, parent *CloudAccountAuditLogSinkPubsubCommand) *CloudAccountAuditLogSinkPubsubValidateCommand {
	var s CloudAccountAuditLogSinkPubsubValidateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "validate [flags]"
	s.Command.Short = "Validate a PubSub audit log sink"
	if hasHighlighting {
		s.Command.Long = "Validates a PubSub audit log sink specification without creating or modifying any resources.\n\nExample:\n\n\x1b[1mtemporal cloud account audit-log sink pubsub validate \\\n  --name my-sink \\\n  --service-account-id my-sa@project.iam.gserviceaccount.com \\\n  --topic-name my-topic \\\n  --gcp-project-id my-project\x1b[0m"
	} else {
		s.Command.Long = "Validates a PubSub audit log sink specification without creating or modifying any resources.\n\nExample:\n\n```\ntemporal cloud account audit-log sink pubsub validate \\\n  --name my-sink \\\n  --service-account-id my-sa@project.iam.gserviceaccount.com \\\n  --topic-name my-topic \\\n  --gcp-project-id my-project\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ServiceAccountId, "service-account-id", "", "The GCP service account ID that Temporal Cloud impersonates for writing records to the customer's PubSub topic. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "service-account-id")
	s.Command.Flags().StringVar(&s.TopicName, "topic-name", "", "The destination PubSub topic name where audit logs will be sent. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "topic-name")
	s.Command.Flags().StringVar(&s.GcpProjectId, "gcp-project-id", "", "The GCP project ID of the PubSub topic and service account. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "gcp-project-id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyCommand struct {
	Parent  *CloudCommand
	Command cobra.Command
}

func NewCloudApikeyCommand(cctx *CommandContext, parent *CloudCommand) *CloudApikeyCommand {
	var s CloudApikeyCommand
	s.Parent = parent
	s.Command.Use = "apikey"
	s.Command.Short = "Manage Temporal Cloud API keys"
	s.Command.Long = "Commands for creating, listing, and managing Temporal Cloud API keys.\n\nAPI keys authenticate requests to the Temporal Cloud API."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudApikeyCreateForMeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyCreateForServiceAccountCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyDisableCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyEditCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyEnableCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudApikeyUpdateCommand(cctx, &s).Command)
	return &s
}

type CloudApikeyCreateForMeCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	DisplayName    string
	Description    string
	ExpiryTime     cliext.FlagTimestamp
	ExpiryDuration cliext.FlagDuration
}

func NewCloudApikeyCreateForMeCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyCreateForMeCommand {
	var s CloudApikeyCreateForMeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create-for-me [flags]"
	s.Command.Short = "Create an API key for the current user"
	if hasHighlighting {
		s.Command.Long = "Create a new API key owned by the currently authenticated user.\nThe token is printed once on creation and cannot be retrieved again.\n\nExample:\n\n\x1b[1mcloud apikey create-for-me --display-name \"My Key\"\x1b[0m"
	} else {
		s.Command.Long = "Create a new API key owned by the currently authenticated user.\nThe token is printed once on creation and cannot be retrieved again.\n\nExample:\n\n```\ncloud apikey create-for-me --display-name \"My Key\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.DisplayName, "display-name", "", "A human-readable display name for the API key. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "display-name")
	s.Command.Flags().StringVar(&s.Description, "description", "", "An optional description for the API key.")
	s.Command.Flags().Var(&s.ExpiryTime, "expiry-time", "Expiry time for the API key in RFC3339 format (e.g. 2025-12-31T00:00:00Z). Mutually exclusive with --expiry-duration.")
	s.ExpiryDuration = 0
	s.Command.Flags().Var(&s.ExpiryDuration, "expiry-duration", "Expiry duration relative to now (e.g. 30d, 24h, 90m). Supports days (d), hours (h), minutes (m), and seconds (s). Mutually exclusive with --expiry-time.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyCreateForServiceAccountCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ServiceAccountId string
	DisplayName      string
	Description      string
	ExpiryTime       cliext.FlagTimestamp
	ExpiryDuration   cliext.FlagDuration
}

func NewCloudApikeyCreateForServiceAccountCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyCreateForServiceAccountCommand {
	var s CloudApikeyCreateForServiceAccountCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create-for-service-account [flags]"
	s.Command.Short = "Create an API key for a service account"
	if hasHighlighting {
		s.Command.Long = "Create a new API key owned by the specified service account.\nThe token is printed once on creation and cannot be retrieved again.\n\nExample:\n\n\x1b[1mcloud apikey create-for-service-account --service-account-id my-sa-id --display-name \"My Key\"\x1b[0m"
	} else {
		s.Command.Long = "Create a new API key owned by the specified service account.\nThe token is printed once on creation and cannot be retrieved again.\n\nExample:\n\n```\ncloud apikey create-for-service-account --service-account-id my-sa-id --display-name \"My Key\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ServiceAccountId, "service-account-id", "", "The ID of the service account to create the API key for. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "service-account-id")
	s.Command.Flags().StringVar(&s.DisplayName, "display-name", "", "A human-readable display name for the API key. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "display-name")
	s.Command.Flags().StringVar(&s.Description, "description", "", "An optional description for the API key.")
	s.Command.Flags().Var(&s.ExpiryTime, "expiry-time", "Expiry time for the API key in RFC3339 format (e.g. 2025-12-31T00:00:00Z). Mutually exclusive with --expiry-duration.")
	s.ExpiryDuration = 0
	s.Command.Flags().Var(&s.ExpiryDuration, "expiry-duration", "Expiry duration relative to now (e.g. 30d, 24h, 90m). Supports days (d), hours (h), minutes (m), and seconds (s). Mutually exclusive with --expiry-time.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyDeleteCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	KeyId string
}

func NewCloudApikeyDeleteCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyDeleteCommand {
	var s CloudApikeyDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete an API key"
	if hasHighlighting {
		s.Command.Long = "Delete a Temporal Cloud API key. This action is irreversible.\n\nExample:\n\n\x1b[1mcloud apikey delete --key-id my-key-id\x1b[0m"
	} else {
		s.Command.Long = "Delete a Temporal Cloud API key. This action is irreversible.\n\nExample:\n\n```\ncloud apikey delete --key-id my-key-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.KeyId, "key-id", "", "The ID of the API key to delete. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key-id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyDisableCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	KeyId string
}

func NewCloudApikeyDisableCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyDisableCommand {
	var s CloudApikeyDisableCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "disable [flags]"
	s.Command.Short = "Disable an API key"
	if hasHighlighting {
		s.Command.Long = "Disable a Temporal Cloud API key. Disabled keys cannot be used for authentication.\n\nExample:\n\n\x1b[1mcloud apikey disable --key-id my-key-id\x1b[0m"
	} else {
		s.Command.Long = "Disable a Temporal Cloud API key. Disabled keys cannot be used for authentication.\n\nExample:\n\n```\ncloud apikey disable --key-id my-key-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.KeyId, "key-id", "", "The ID of the API key to disable. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key-id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyEditCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	AsyncOperationOptions
	ResourceVersionOptions
	KeyId string
}

func NewCloudApikeyEditCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyEditCommand {
	var s CloudApikeyEditCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "edit [flags]"
	s.Command.Short = "Interactively edit an API key"
	if hasHighlighting {
		s.Command.Long = "Open an API key configuration in your default editor for interactive\nmodification. After saving and closing the editor, the changes are\napplied to Temporal Cloud.\n\nThe editor is determined by the EDITOR environment variable, falling\nback to 'vi' if not set.\n\nExample:\n\n\x1b[1mcloud apikey edit --key-id my-key-id\x1b[0m"
	} else {
		s.Command.Long = "Open an API key configuration in your default editor for interactive\nmodification. After saving and closing the editor, the changes are\napplied to Temporal Cloud.\n\nThe editor is determined by the EDITOR environment variable, falling\nback to 'vi' if not set.\n\nExample:\n\n```\ncloud apikey edit --key-id my-key-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.KeyId, "key-id", "", "The ID of the API key to edit. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key-id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.DiffOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyEnableCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	KeyId string
}

func NewCloudApikeyEnableCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyEnableCommand {
	var s CloudApikeyEnableCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "enable [flags]"
	s.Command.Short = "Enable an API key"
	if hasHighlighting {
		s.Command.Long = "Enable a previously disabled Temporal Cloud API key.\n\nExample:\n\n\x1b[1mcloud apikey enable --key-id my-key-id\x1b[0m"
	} else {
		s.Command.Long = "Enable a previously disabled Temporal Cloud API key.\n\nExample:\n\n```\ncloud apikey enable --key-id my-key-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.KeyId, "key-id", "", "The ID of the API key to enable. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key-id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyGetCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	KeyId string
}

func NewCloudApikeyGetCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyGetCommand {
	var s CloudApikeyGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get API key details"
	if hasHighlighting {
		s.Command.Long = "Retrieve the configuration and status of a Temporal Cloud API key.\n\nExample:\n\n\x1b[1mcloud apikey get --key-id my-key-id\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the configuration and status of a Temporal Cloud API key.\n\nExample:\n\n```\ncloud apikey get --key-id my-key-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.KeyId, "key-id", "", "The ID of the API key to retrieve. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key-id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyListCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	UserId           string
	UserEmail        string
	ServiceAccountId string
	PageSize         int
	PageToken        string
}

func NewCloudApikeyListCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyListCommand {
	var s CloudApikeyListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List API keys"
	if hasHighlighting {
		s.Command.Long = "List API keys. Optionally filter by user ID, user email, or service account ID.\nAt most one filter may be specified.\n\nExample:\n\n\x1b[1mcloud apikey list\ncloud apikey list --user-id my-user-id\ncloud apikey list --service-account-id my-sa-id\x1b[0m"
	} else {
		s.Command.Long = "List API keys. Optionally filter by user ID, user email, or service account ID.\nAt most one filter may be specified.\n\nExample:\n\n```\ncloud apikey list\ncloud apikey list --user-id my-user-id\ncloud apikey list --service-account-id my-sa-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.UserId, "user-id", "", "Filter API keys by user ID. Mutually exclusive with --user-email and --service-account-id.")
	s.Command.Flags().StringVar(&s.UserEmail, "user-email", "", "Filter API keys by user email. Mutually exclusive with --user-id and --service-account-id.")
	s.Command.Flags().StringVar(&s.ServiceAccountId, "service-account-id", "", "Filter API keys by service account ID. Mutually exclusive with --user-id and --user-email.")
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Number of API keys to return per page.")
	s.Command.Flags().StringVar(&s.PageToken, "page-token", "", "Token for retrieving the next page of results.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudApikeyUpdateCommand struct {
	Parent  *CloudApikeyCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	KeyId       string
	DisplayName string
	Description string
	Disabled    bool
}

func NewCloudApikeyUpdateCommand(cctx *CommandContext, parent *CloudApikeyCommand) *CloudApikeyUpdateCommand {
	var s CloudApikeyUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update an API key"
	if hasHighlighting {
		s.Command.Long = "Update an API key's display name, description, or disabled status.\nOnly flags that are explicitly provided are changed.\n\nExample:\n\n\x1b[1mcloud apikey update --key-id my-key-id --display-name \"New Name\"\ncloud apikey update --key-id my-key-id --disabled=true\x1b[0m"
	} else {
		s.Command.Long = "Update an API key's display name, description, or disabled status.\nOnly flags that are explicitly provided are changed.\n\nExample:\n\n```\ncloud apikey update --key-id my-key-id --display-name \"New Name\"\ncloud apikey update --key-id my-key-id --disabled=true\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.KeyId, "key-id", "", "The ID of the API key to update. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key-id")
	s.Command.Flags().StringVar(&s.DisplayName, "display-name", "", "New display name for the API key.")
	s.Command.Flags().StringVar(&s.Description, "description", "", "New description for the API key.")
	s.Command.Flags().BoolVar(&s.Disabled, "disabled", false, "Set to true to disable the API key, or false to enable it.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudConnectivityCommand struct {
	Parent  *CloudCommand
	Command cobra.Command
}

func NewCloudConnectivityCommand(cctx *CommandContext, parent *CloudCommand) *CloudConnectivityCommand {
	var s CloudConnectivityCommand
	s.Parent = parent
	s.Command.Use = "connectivity"
	s.Command.Short = "Manage Temporal Cloud connectivity rules"
	s.Command.Long = "Commands for managing connectivity rules for Temporal Cloud."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudConnectivityDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudConnectivityGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudConnectivityListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudConnectivityPrivateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudConnectivityPublicCommand(cctx, &s).Command)
	return &s
}

type CloudConnectivityDeleteCommand struct {
	Parent  *CloudConnectivityCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Id string
}

func NewCloudConnectivityDeleteCommand(cctx *CommandContext, parent *CloudConnectivityCommand) *CloudConnectivityDeleteCommand {
	var s CloudConnectivityDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a connectivity rule"
	if hasHighlighting {
		s.Command.Long = "Delete a connectivity rule by its ID.\n\nExample:\n\n\x1b[1mcloud connectivity delete --id <connectivity-rule-id>\x1b[0m"
	} else {
		s.Command.Long = "Delete a connectivity rule by its ID.\n\nExample:\n\n```\ncloud connectivity delete --id <connectivity-rule-id>\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Id, "id", "", "The ID of the connectivity rule. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudConnectivityGetCommand struct {
	Parent  *CloudConnectivityCommand
	Command cobra.Command
	ClientOptions
	Id string
}

func NewCloudConnectivityGetCommand(cctx *CommandContext, parent *CloudConnectivityCommand) *CloudConnectivityGetCommand {
	var s CloudConnectivityGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get details of a connectivity rule"
	if hasHighlighting {
		s.Command.Long = "Get details of a specific connectivity rule by its ID.\n\nExample:\n\n\x1b[1mcloud connectivity get --id <connectivity-rule-id>\x1b[0m"
	} else {
		s.Command.Long = "Get details of a specific connectivity rule by its ID.\n\nExample:\n\n```\ncloud connectivity get --id <connectivity-rule-id>\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Id, "id", "", "The ID of the connectivity rule. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "id")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudConnectivityListCommand struct {
	Parent  *CloudConnectivityCommand
	Command cobra.Command
	ClientOptions
	Namespace string
	PageSize  int
	PageToken string
}

func NewCloudConnectivityListCommand(cctx *CommandContext, parent *CloudConnectivityCommand) *CloudConnectivityListCommand {
	var s CloudConnectivityListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List connectivity rules"
	if hasHighlighting {
		s.Command.Long = "List connectivity rules, optionally filtered by namespace.\n\nExample:\n\n\x1b[1mcloud connectivity list --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "List connectivity rules, optionally filtered by namespace.\n\nExample:\n\n```\ncloud connectivity list --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "Filter connectivity rules by namespace (e.g., 'my-namespace.my-account').")
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Number of connectivity rules to return per page.")
	s.Command.Flags().StringVar(&s.PageToken, "page-token", "", "Page token for pagination.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudConnectivityPrivateCommand struct {
	Parent  *CloudConnectivityCommand
	Command cobra.Command
}

func NewCloudConnectivityPrivateCommand(cctx *CommandContext, parent *CloudConnectivityCommand) *CloudConnectivityPrivateCommand {
	var s CloudConnectivityPrivateCommand
	s.Parent = parent
	s.Command.Use = "private"
	s.Command.Short = "Manage private connectivity rules"
	s.Command.Long = "Commands for managing private connectivity rules."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudConnectivityPrivateCreateCommand(cctx, &s).Command)
	return &s
}

type CloudConnectivityPrivateCreateCommand struct {
	Parent  *CloudConnectivityPrivateCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ConnectionId string
	Region       string
	GcpProjectId string
}

func NewCloudConnectivityPrivateCreateCommand(cctx *CommandContext, parent *CloudConnectivityPrivateCommand) *CloudConnectivityPrivateCreateCommand {
	var s CloudConnectivityPrivateCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a private connectivity rule"
	if hasHighlighting {
		s.Command.Long = "Create a new private VPC connectivity rule. Requires --connection-id and --region.\n\nExample:\n\n\x1b[1mcloud connectivity private create --connection-id vpce-12345 --region aws-us-west-2\x1b[0m"
	} else {
		s.Command.Long = "Create a new private VPC connectivity rule. Requires --connection-id and --region.\n\nExample:\n\n```\ncloud connectivity private create --connection-id vpce-12345 --region aws-us-west-2\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ConnectionId, "connection-id", "", "The connection ID for private connectivity. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "connection-id")
	s.Command.Flags().StringVar(&s.Region, "region", "", "The region for private connectivity. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "region")
	s.Command.Flags().StringVar(&s.GcpProjectId, "gcp-project-id", "", "The GCP project ID (only for GCP private connectivity).")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudConnectivityPublicCommand struct {
	Parent  *CloudConnectivityCommand
	Command cobra.Command
}

func NewCloudConnectivityPublicCommand(cctx *CommandContext, parent *CloudConnectivityCommand) *CloudConnectivityPublicCommand {
	var s CloudConnectivityPublicCommand
	s.Parent = parent
	s.Command.Use = "public"
	s.Command.Short = "Manage public connectivity rules"
	s.Command.Long = "Commands for managing public connectivity rules."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudConnectivityPublicCreateCommand(cctx, &s).Command)
	return &s
}

type CloudConnectivityPublicCreateCommand struct {
	Parent  *CloudConnectivityPublicCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
}

func NewCloudConnectivityPublicCreateCommand(cctx *CommandContext, parent *CloudConnectivityPublicCommand) *CloudConnectivityPublicCreateCommand {
	var s CloudConnectivityPublicCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a public connectivity rule"
	if hasHighlighting {
		s.Command.Long = "Create a new public internet connectivity rule.\n\nExample:\n\n\x1b[1mcloud connectivity public create\x1b[0m"
	} else {
		s.Command.Long = "Create a new public internet connectivity rule.\n\nExample:\n\n```\ncloud connectivity public create\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudLoginCommand struct {
	Parent      *CloudCommand
	Command     cobra.Command
	Domain      string
	Audience    string
	ClientId    string
	RedirectUrl string
	Reset       bool
}

func NewCloudLoginCommand(cctx *CommandContext, parent *CloudCommand) *CloudLoginCommand {
	var s CloudLoginCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "login [flags]"
	s.Command.Short = "Authenticate with Temporal Cloud"
	if hasHighlighting {
		s.Command.Long = "Authenticate with Temporal Cloud using browser-based OAuth login.\n\nThis command opens your default browser to complete authentication. Once\nlogged in, your credentials are stored locally for subsequent commands.\n\nExample:\n\n\x1b[1mcloud login\x1b[0m\n\nFor headless environments, use --disable-pop-up and follow the printed URL."
	} else {
		s.Command.Long = "Authenticate with Temporal Cloud using browser-based OAuth login.\n\nThis command opens your default browser to complete authentication. Once\nlogged in, your credentials are stored locally for subsequent commands.\n\nExample:\n\n```\ncloud login\n```\n\nFor headless environments, use --disable-pop-up and follow the printed URL."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Domain, "domain", "login.tmprl-test.cloud", "Authentication domain for the OAuth provider.")
	_ = s.Command.Flags().MarkHidden("domain")
	s.Command.Flags().StringVar(&s.Audience, "audience", "https://saas-api.tmprl-test.cloud", "OAuth audience parameter for token generation.")
	_ = s.Command.Flags().MarkHidden("audience")
	s.Command.Flags().StringVar(&s.ClientId, "client-id", "XBimMwn90eAnjsiGVbAJ3Hgd9z06jjJB", "OAuth client identifier for authentication.")
	_ = s.Command.Flags().MarkHidden("client-id")
	s.Command.Flags().StringVar(&s.RedirectUrl, "redirect-url", "http://127.0.0.1:56628/callback", "Redirect URL for OAuth authentication flow.")
	_ = s.Command.Flags().MarkHidden("redirect-url")
	s.Command.Flags().BoolVar(&s.Reset, "reset", false, "Clear stored login credentials and configuration, then re-authenticate. Use this if you need to switch accounts or fix authentication issues.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudLogoutCommand struct {
	Parent  *CloudCommand
	Command cobra.Command
	Domain  string
}

func NewCloudLogoutCommand(cctx *CommandContext, parent *CloudCommand) *CloudLogoutCommand {
	var s CloudLogoutCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "logout [flags]"
	s.Command.Short = "Clear Temporal Cloud authentication credentials"
	if hasHighlighting {
		s.Command.Long = "Log out from Temporal Cloud by clearing stored authentication tokens\nand credentials from the local configuration.\n\nExample:\n\n\x1b[1mcloud logout\x1b[0m"
	} else {
		s.Command.Long = "Log out from Temporal Cloud by clearing stored authentication tokens\nand credentials from the local configuration.\n\nExample:\n\n```\ncloud logout\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Domain, "domain", "login.tmprl-test.cloud", "Authentication domain for the OAuth provider.")
	_ = s.Command.Flags().MarkHidden("domain")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCommand struct {
	Parent  *CloudCommand
	Command cobra.Command
}

func NewCloudNamespaceCommand(cctx *CommandContext, parent *CloudCommand) *CloudNamespaceCommand {
	var s CloudNamespaceCommand
	s.Parent = parent
	s.Command.Use = "namespace"
	s.Command.Short = "Manage Temporal Cloud namespaces"
	s.Command.Long = "Commands for creating, updating, and managing Temporal Cloud namespaces.\n\nNamespaces provide isolation for workflows and activities. Each namespace\nhas its own configuration including retention period, region, and access\ncontrols."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceApplyCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCertCaCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCertFilterCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCodecCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceEditCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceHaCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceLifecycleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceRetentionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceSearchAttributeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceTagCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceApplyCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	Spec             string
	AsyncOperationId string
	Idempotent       bool
	Async            bool
	ResourceVersion  string
}

func NewCloudNamespaceApplyCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceApplyCommand {
	var s CloudNamespaceApplyCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "apply [flags]"
	s.Command.Short = "Create or update a namespace from a specification"
	if hasHighlighting {
		s.Command.Long = "Apply a namespace configuration to Temporal Cloud. Creates a new namespace\nif it doesn't exist, or updates an existing one to match the specification.\n\nThe specification can be provided as inline JSON or loaded from a file\nby prefixing the path with '@'.\n\nExample with inline JSON:\n\n\x1b[1mcloud namespace apply --spec '{\"name\": \"namespace-name\", \"region\": \"us-west-2\", \"retention_days\": 7}'\x1b[0m\n\nExample with file path:\n\n\x1b[1mcloud namespace apply --spec @namespace-spec.json\x1b[0m"
	} else {
		s.Command.Long = "Apply a namespace configuration to Temporal Cloud. Creates a new namespace\nif it doesn't exist, or updates an existing one to match the specification.\n\nThe specification can be provided as inline JSON or loaded from a file\nby prefixing the path with '@'.\n\nExample with inline JSON:\n\n```\ncloud namespace apply --spec '{\"name\": \"namespace-name\", \"region\": \"us-west-2\", \"retention_days\": 7}'\n```\n\nExample with file path:\n\n```\ncloud namespace apply --spec @namespace-spec.json\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Spec, "spec", "", "Namespace configuration in JSON format. Provide inline JSON directly, or use '@path/to/file.json' to load from a file. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "spec")
	s.Command.Flags().StringVar(&s.AsyncOperationId, "async-operation-id", "", "Custom identifier for tracking this async operation. If not provided, a unique ID is generated automatically.")
	s.Command.Flags().BoolVar(&s.Idempotent, "idempotent", false, "Succeed silently if the namespace already matches the specification. Without this flag, the command errors when no changes are needed.")
	s.Command.Flags().BoolVar(&s.Async, "async", false, "Return immediately after initiating the operation instead of waiting for completion. Use the returned operation ID to check status later.")
	s.Command.Flags().StringVarP(&s.ResourceVersion, "resource-version", "v", "", "Resource version for optimistic concurrency control. If not provided, the current version is fetched automatically.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.DiffOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCertCaCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceCertCaCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceCertCaCommand {
	var s CloudNamespaceCertCaCommand
	s.Parent = parent
	s.Command.Use = "cert-ca"
	s.Command.Short = "Manage client CA certificates for namespaces"
	s.Command.Long = "Commands for managing the client CA certificates of Temporal Cloud namespaces."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceCertCaCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCertCaDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCertCaListCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceCertCaCreateCommand struct {
	Parent  *CloudNamespaceCertCaCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	CaCertificateOptions
}

func NewCloudNamespaceCertCaCreateCommand(cctx *CommandContext, parent *CloudNamespaceCertCaCommand) *CloudNamespaceCertCaCreateCommand {
	var s CloudNamespaceCertCaCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Add CA certificates to a namespace"
	if hasHighlighting {
		s.Command.Long = "Add client CA certificates to a Temporal Cloud namespace from a PEM file\nor base64 encoded string. These certificates are used to verify client\nconnections and enable mTLS authentication.\n\nSpecify either --ca-certificate-file or --ca-certificate, but not both.\n\nExample with file:\n\n\x1b[1mcloud namespace cert-ca create --namespace my-namespace.my-account --ca-certificate-file ca-cert.pem\x1b[0m\n\nExample with base64 encoded data:\n\n\x1b[1mcloud namespace cert-ca create --namespace my-namespace.my-account --ca-certificate <base64-encoded-cert>\x1b[0m"
	} else {
		s.Command.Long = "Add client CA certificates to a Temporal Cloud namespace from a PEM file\nor base64 encoded string. These certificates are used to verify client\nconnections and enable mTLS authentication.\n\nSpecify either --ca-certificate-file or --ca-certificate, but not both.\n\nExample with file:\n\n```\ncloud namespace cert-ca create --namespace my-namespace.my-account --ca-certificate-file ca-cert.pem\n```\n\nExample with base64 encoded data:\n\n```\ncloud namespace cert-ca create --namespace my-namespace.my-account --ca-certificate <base64-encoded-cert>\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.CaCertificateOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCertCaDeleteCommand struct {
	Parent  *CloudNamespaceCertCaCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	CaCertificateOptions
}

func NewCloudNamespaceCertCaDeleteCommand(cctx *CommandContext, parent *CloudNamespaceCertCaCommand) *CloudNamespaceCertCaDeleteCommand {
	var s CloudNamespaceCertCaDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete CA certificates from a namespace"
	if hasHighlighting {
		s.Command.Long = "Delete client CA certificates from a Temporal Cloud namespace. This operation\nrequires confirmation and will remove the specified certificates from the\nnamespace configuration.\n\nSpecify either --ca-certificate-file or --ca-certificate, but not both.\n\nExample with file:\n\n\x1b[1mcloud namespace cert-ca delete --namespace my-namespace.my-account --ca-certificate-file ca-cert.pem\x1b[0m\n\nExample with base64 encoded data:\n\n\x1b[1mcloud namespace cert-ca delete --namespace my-namespace.my-account --ca-certificate <base64-encoded-cert>\x1b[0m"
	} else {
		s.Command.Long = "Delete client CA certificates from a Temporal Cloud namespace. This operation\nrequires confirmation and will remove the specified certificates from the\nnamespace configuration.\n\nSpecify either --ca-certificate-file or --ca-certificate, but not both.\n\nExample with file:\n\n```\ncloud namespace cert-ca delete --namespace my-namespace.my-account --ca-certificate-file ca-cert.pem\n```\n\nExample with base64 encoded data:\n\n```\ncloud namespace cert-ca delete --namespace my-namespace.my-account --ca-certificate <base64-encoded-cert>\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.CaCertificateOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCertCaListCommand struct {
	Parent  *CloudNamespaceCertCaCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceCertCaListCommand(cctx *CommandContext, parent *CloudNamespaceCertCaCommand) *CloudNamespaceCertCaListCommand {
	var s CloudNamespaceCertCaListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List CA certificates for a namespace"
	if hasHighlighting {
		s.Command.Long = "Retrieve the list of client CA certificates configured for a Temporal Cloud\nnamespace. These certificates are used for client authentication.\n\nExample:\n\n\x1b[1mcloud namespace cert-ca list --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the list of client CA certificates configured for a Temporal Cloud\nnamespace. These certificates are used for client authentication.\n\nExample:\n\n```\ncloud namespace cert-ca list --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCertFilterCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceCertFilterCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceCertFilterCommand {
	var s CloudNamespaceCertFilterCommand
	s.Parent = parent
	s.Command.Use = "cert-filter"
	s.Command.Short = "Manage certificate filters for namespaces"
	s.Command.Long = "Commands for managing certificate filters for Temporal Cloud namespaces.\nCertificate filters restrict mTLS connections to client certificates with\nspecific distinguished name properties."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceCertFilterCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCertFilterDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCertFilterListCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceCertFilterCreateCommand struct {
	Parent  *CloudNamespaceCertFilterCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	CommonName             string
	Organization           string
	OrganizationalUnit     string
	SubjectAlternativeName string
}

func NewCloudNamespaceCertFilterCreateCommand(cctx *CommandContext, parent *CloudNamespaceCertFilterCommand) *CloudNamespaceCertFilterCreateCommand {
	var s CloudNamespaceCertFilterCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Add certificate filters to a namespace"
	s.Command.Long = "Add new certificate filters to a Temporal Cloud namespace. Certificate\nfilters restrict mTLS connections to client certificates whose distinguished\nname properties match at least one of the filters."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.CommonName, "common-name", "", "The common name (CN) field from the certificate's distinguished name.")
	s.Command.Flags().StringVar(&s.Organization, "organization", "", "The organization (O) field from the certificate's distinguished name.")
	s.Command.Flags().StringVar(&s.OrganizationalUnit, "organizational-unit", "", "The organizational unit (OU) field from the certificate's distinguished name.")
	s.Command.Flags().StringVar(&s.SubjectAlternativeName, "subject-alternative-name", "", "The subject alternative name (SAN) from the certificate.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCertFilterDeleteCommand struct {
	Parent  *CloudNamespaceCertFilterCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	CommonName             string
	Organization           string
	OrganizationalUnit     string
	SubjectAlternativeName string
}

func NewCloudNamespaceCertFilterDeleteCommand(cctx *CommandContext, parent *CloudNamespaceCertFilterCommand) *CloudNamespaceCertFilterDeleteCommand {
	var s CloudNamespaceCertFilterDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete certificate filters from a namespace"
	s.Command.Long = "Delete certificate filters from a Temporal Cloud namespace. Filters are\nmatched by exact field equality."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.CommonName, "common-name", "", "The common name (CN) field from the certificate's distinguished name.")
	s.Command.Flags().StringVar(&s.Organization, "organization", "", "The organization (O) field from the certificate's distinguished name.")
	s.Command.Flags().StringVar(&s.OrganizationalUnit, "organizational-unit", "", "The organizational unit (OU) field from the certificate's distinguished name.")
	s.Command.Flags().StringVar(&s.SubjectAlternativeName, "subject-alternative-name", "", "The subject alternative name (SAN) from the certificate.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCertFilterListCommand struct {
	Parent  *CloudNamespaceCertFilterCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceCertFilterListCommand(cctx *CommandContext, parent *CloudNamespaceCertFilterCommand) *CloudNamespaceCertFilterListCommand {
	var s CloudNamespaceCertFilterListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List certificate filters for a namespace"
	s.Command.Long = "List all certificate filters configured for a Temporal Cloud namespace."
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCodecCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceCodecCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceCodecCommand {
	var s CloudNamespaceCodecCommand
	s.Parent = parent
	s.Command.Use = "codec"
	s.Command.Short = "Manage codec server settings for namespaces"
	s.Command.Long = "Commands for managing the codec server configuration of Temporal Cloud namespaces.\n\nThe codec server is used to encode and decode payloads for workflows and activities."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceCodecDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCodecGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCodecSetCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceCodecDeleteCommand struct {
	Parent  *CloudNamespaceCodecCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
}

func NewCloudNamespaceCodecDeleteCommand(cctx *CommandContext, parent *CloudNamespaceCodecCommand) *CloudNamespaceCodecDeleteCommand {
	var s CloudNamespaceCodecDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete codec server configuration from a namespace"
	if hasHighlighting {
		s.Command.Long = "Delete the codec server configuration from a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace codec delete --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Delete the codec server configuration from a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace codec delete --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCodecGetCommand struct {
	Parent  *CloudNamespaceCodecCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceCodecGetCommand(cctx *CommandContext, parent *CloudNamespaceCodecCommand) *CloudNamespaceCodecGetCommand {
	var s CloudNamespaceCodecGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get codec server configuration for a namespace"
	if hasHighlighting {
		s.Command.Long = "Retrieve the current codec server configuration for a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace codec get --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the current codec server configuration for a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace codec get --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCodecSetCommand struct {
	Parent  *CloudNamespaceCodecCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Endpoint                         string
	PassAccessToken                  bool
	IncludeCrossOriginCredentials    bool
	CustomErrorMessageDefaultMessage string
	CustomErrorMessageDefaultLink    string
}

func NewCloudNamespaceCodecSetCommand(cctx *CommandContext, parent *CloudNamespaceCodecCommand) *CloudNamespaceCodecSetCommand {
	var s CloudNamespaceCodecSetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set [flags]"
	s.Command.Short = "Set codec server configuration for a namespace"
	if hasHighlighting {
		s.Command.Long = "Set the codec server configuration for a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace codec set --namespace my-namespace.my-account --endpoint https://my-codec.example.com\x1b[0m"
	} else {
		s.Command.Long = "Set the codec server configuration for a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace codec set --namespace my-namespace.my-account --endpoint https://my-codec.example.com\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Endpoint, "endpoint", "", "The codec server endpoint URL. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "endpoint")
	s.Command.Flags().BoolVar(&s.PassAccessToken, "pass-access-token", false, "Whether to pass the user access token to the codec server endpoint.")
	s.Command.Flags().BoolVar(&s.IncludeCrossOriginCredentials, "include-cross-origin-credentials", false, "Whether to include cross-origin credentials in requests to the codec server.")
	s.Command.Flags().StringVar(&s.CustomErrorMessageDefaultMessage, "custom-error-message-default-message", "", "A custom message to display for remote codec server errors.")
	s.Command.Flags().StringVar(&s.CustomErrorMessageDefaultLink, "custom-error-message-default-link", "", "A link to display alongside the custom error message for remote codec server errors.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceCreateCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	CodecServerOptions
	CaCertificateOptions
	CertificateFilterOptions
	Name                   string
	Region                 []string
	RetentionDays          int
	ApiKeyAuthEnabled      bool
	EnableDeleteProtection bool
	SearchAttribute        []string
	ConnectionRuleId       []string
}

func NewCloudNamespaceCreateCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceCreateCommand {
	var s CloudNamespaceCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a new Temporal Cloud namespace"
	if hasHighlighting {
		s.Command.Long = "Create a new Temporal Cloud namespace with the specified configuration.\n\nOptions are passed as individual flags. To create or update a namespace\nusing a full JSON specification, use 'namespace apply' instead.\n\nExample:\n\n\x1b[1mcloud namespace create --name my-namespace --region aws-us-east-1 --retention-days 30\x1b[0m"
	} else {
		s.Command.Long = "Create a new Temporal Cloud namespace with the specified configuration.\n\nOptions are passed as individual flags. To create or update a namespace\nusing a full JSON specification, use 'namespace apply' instead.\n\nExample:\n\n```\ncloud namespace create --name my-namespace --region aws-us-east-1 --retention-days 30\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Name, "name", "n", "", "The name for the new namespace (becomes part of the namespace ID). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringArrayVar(&s.Region, "region", nil, "Cloud region where the namespace will be hosted. Repeat to specify multiple regions for High Availability (e.g. --region aws-us-east-1 --region aws-us-west-2). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "region")
	s.Command.Flags().IntVar(&s.RetentionDays, "retention-days", 0, "Number of days to retain closed workflow history. If not specified, the server default applies.")
	s.Command.Flags().BoolVar(&s.ApiKeyAuthEnabled, "api-key-auth-enabled", false, "Enable API key authentication for the namespace.")
	s.Command.Flags().BoolVar(&s.EnableDeleteProtection, "enable-delete-protection", false, "Prevent accidental deletion of this namespace.")
	s.Command.Flags().StringArrayVar(&s.SearchAttribute, "search-attribute", nil, "Custom search attribute as 'name=Type' (e.g. --search-attribute myAttr=Keyword). Valid types: Text, Keyword, Int, Double, Bool, Datetime, KeywordList. Repeat to add multiple.")
	s.Command.Flags().StringArrayVar(&s.ConnectionRuleId, "connection-rule-id", nil, "Private connectivity rule ID. Repeat to specify multiple.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.CodecServerOptions.BuildFlags(s.Command.Flags())
	s.CaCertificateOptions.BuildFlags(s.Command.Flags())
	s.CertificateFilterOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceDeleteCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
	ClientOptions
	Namespace        string
	AsyncOperationId string
	Async            bool
	Idempotent       bool
	ResourceVersion  string
}

func NewCloudNamespaceDeleteCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceDeleteCommand {
	var s CloudNamespaceDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a Temporal Cloud namespace"
	if hasHighlighting {
		s.Command.Long = "Delete a Temporal Cloud namespace and all associated data. This action is\nirreversible and will permanently remove all workflows, activities, and\nhistory within the namespace.\n\nExample:\n\n\x1b[1mcloud namespace delete --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Delete a Temporal Cloud namespace and all associated data. This action is\nirreversible and will permanently remove all workflows, activities, and\nhistory within the namespace.\n\nExample:\n\n```\ncloud namespace delete --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "The fully qualified namespace name in the format 'namespace.account' (e.g., 'my-namespace.my-account'). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace")
	s.Command.Flags().StringVar(&s.AsyncOperationId, "async-operation-id", "", "Custom identifier for tracking this async operation. If not provided, a unique ID is generated automatically.")
	s.Command.Flags().BoolVar(&s.Async, "async", false, "Return immediately after initiating the operation instead of waiting for completion. Use the returned operation ID to check status later.")
	s.Command.Flags().BoolVar(&s.Idempotent, "idempotent", false, "Succeed silently if the namespace does not exist. Without this flag, the command errors if the namespace is not found.")
	s.Command.Flags().StringVarP(&s.ResourceVersion, "resource-version", "v", "", "Resource version for optimistic concurrency control. If not provided, the current version is fetched automatically.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceEditCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	Namespace        string
	AsyncOperationId string
	Idempotent       bool
	Async            bool
	ResourceVersion  string
}

func NewCloudNamespaceEditCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceEditCommand {
	var s CloudNamespaceEditCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "edit [flags]"
	s.Command.Short = "Interactively edit a namespace configuration"
	if hasHighlighting {
		s.Command.Long = "Open a namespace configuration in your default editor for interactive\nmodification. After saving and closing the editor, the changes are\napplied to Temporal Cloud.\n\nThe editor is determined by the EDITOR environment variable, falling\nback to 'vi' if not set.\n\nExample:\n\n\x1b[1mcloud namespace edit --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Open a namespace configuration in your default editor for interactive\nmodification. After saving and closing the editor, the changes are\napplied to Temporal Cloud.\n\nThe editor is determined by the EDITOR environment variable, falling\nback to 'vi' if not set.\n\nExample:\n\n```\ncloud namespace edit --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "The fully qualified namespace name in the format 'namespace.account' (e.g., 'my-namespace.my-account'). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace")
	s.Command.Flags().StringVar(&s.AsyncOperationId, "async-operation-id", "", "Custom identifier for tracking this async operation. If not provided, a unique ID is generated automatically.")
	s.Command.Flags().BoolVar(&s.Idempotent, "idempotent", false, "Succeed silently if no changes were made in the editor. Without this flag, the command errors when the configuration is unchanged.")
	s.Command.Flags().BoolVar(&s.Async, "async", false, "Return immediately after initiating the operation instead of waiting for completion. Use the returned operation ID to check status later.")
	s.Command.Flags().StringVarP(&s.ResourceVersion, "resource-version", "v", "", "Resource version for optimistic concurrency control. If not provided, the current version is fetched automatically.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.DiffOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceExportCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceExportCommand {
	var s CloudNamespaceExportCommand
	s.Parent = parent
	s.Command.Use = "export"
	s.Command.Short = "Manage workflow history export sinks for namespaces"
	s.Command.Long = "Commands for managing workflow history export sinks for Temporal Cloud namespaces.\n\nExport sinks define destinations (S3 or GCS) to which workflow history is exported."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceExportDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportDisableCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportEnableCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportGcsCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportS3Command(cctx, &s).Command)
	return &s
}

type CloudNamespaceExportDeleteCommand struct {
	Parent  *CloudNamespaceExportCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	ExportSinkOptions
}

func NewCloudNamespaceExportDeleteCommand(cctx *CommandContext, parent *CloudNamespaceExportCommand) *CloudNamespaceExportDeleteCommand {
	var s CloudNamespaceExportDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a workflow history export sink"
	if hasHighlighting {
		s.Command.Long = "Delete a workflow history export sink from a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace export delete --namespace my-namespace.my-account --sink-name my-sink\x1b[0m"
	} else {
		s.Command.Long = "Delete a workflow history export sink from a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace export delete --namespace my-namespace.my-account --sink-name my-sink\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportDisableCommand struct {
	Parent  *CloudNamespaceExportCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	ExportSinkOptions
}

func NewCloudNamespaceExportDisableCommand(cctx *CommandContext, parent *CloudNamespaceExportCommand) *CloudNamespaceExportDisableCommand {
	var s CloudNamespaceExportDisableCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "disable [flags]"
	s.Command.Short = "Disable a workflow history export sink"
	if hasHighlighting {
		s.Command.Long = "Disable a workflow history export sink for a Temporal Cloud namespace.\nThe sink configuration is preserved and can be re-enabled later.\n\nExample:\n\n\x1b[1mcloud namespace export disable --namespace my-namespace.my-account --sink-name my-sink\x1b[0m"
	} else {
		s.Command.Long = "Disable a workflow history export sink for a Temporal Cloud namespace.\nThe sink configuration is preserved and can be re-enabled later.\n\nExample:\n\n```\ncloud namespace export disable --namespace my-namespace.my-account --sink-name my-sink\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportEnableCommand struct {
	Parent  *CloudNamespaceExportCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	ExportSinkOptions
}

func NewCloudNamespaceExportEnableCommand(cctx *CommandContext, parent *CloudNamespaceExportCommand) *CloudNamespaceExportEnableCommand {
	var s CloudNamespaceExportEnableCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "enable [flags]"
	s.Command.Short = "Enable a workflow history export sink"
	if hasHighlighting {
		s.Command.Long = "Enable a previously disabled workflow history export sink for a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace export enable --namespace my-namespace.my-account --sink-name my-sink\x1b[0m"
	} else {
		s.Command.Long = "Enable a previously disabled workflow history export sink for a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace export enable --namespace my-namespace.my-account --sink-name my-sink\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportGcsCommand struct {
	Parent  *CloudNamespaceExportCommand
	Command cobra.Command
}

func NewCloudNamespaceExportGcsCommand(cctx *CommandContext, parent *CloudNamespaceExportCommand) *CloudNamespaceExportGcsCommand {
	var s CloudNamespaceExportGcsCommand
	s.Parent = parent
	s.Command.Use = "gcs"
	s.Command.Short = "Manage GCS workflow history export sinks"
	s.Command.Long = "Commands for managing GCS workflow history export sinks for Temporal Cloud namespaces."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceExportGcsCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportGcsUpdateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportGcsValidateCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceExportGcsCreateCommand struct {
	Parent  *CloudNamespaceExportGcsCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ExportSinkOptions
	ExportGcsOptions
}

func NewCloudNamespaceExportGcsCreateCommand(cctx *CommandContext, parent *CloudNamespaceExportGcsCommand) *CloudNamespaceExportGcsCreateCommand {
	var s CloudNamespaceExportGcsCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a GCS workflow history export sink"
	if hasHighlighting {
		s.Command.Long = "Create a new GCS workflow history export sink for a Temporal Cloud namespace.\nThe sink is created in the enabled state.\n\nExample:\n\n\x1b[1mcloud namespace export gcs create --namespace my-namespace.my-account --sink-name my-sink \\\n  --sa-id my-service-account@my-project.iam.gserviceaccount.com \\\n  --bucket-name my-bucket --gcp-project-id my-project --region us-central1\x1b[0m"
	} else {
		s.Command.Long = "Create a new GCS workflow history export sink for a Temporal Cloud namespace.\nThe sink is created in the enabled state.\n\nExample:\n\n```\ncloud namespace export gcs create --namespace my-namespace.my-account --sink-name my-sink \\\n  --sa-id my-service-account@my-project.iam.gserviceaccount.com \\\n  --bucket-name my-bucket --gcp-project-id my-project --region us-central1\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.ExportGcsOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportGcsUpdateCommand struct {
	Parent  *CloudNamespaceExportGcsCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	ExportSinkOptions
	ExportGcsOptions
}

func NewCloudNamespaceExportGcsUpdateCommand(cctx *CommandContext, parent *CloudNamespaceExportGcsCommand) *CloudNamespaceExportGcsUpdateCommand {
	var s CloudNamespaceExportGcsUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update a GCS workflow history export sink"
	if hasHighlighting {
		s.Command.Long = "Update the configuration of an existing GCS workflow history export sink.\nThe enabled/disabled state is preserved.\n\nExample:\n\n\x1b[1mcloud namespace export gcs update --namespace my-namespace.my-account --sink-name my-sink \\\n  --sa-id my-service-account@my-project.iam.gserviceaccount.com \\\n  --bucket-name my-bucket --gcp-project-id my-project --region us-central1\x1b[0m"
	} else {
		s.Command.Long = "Update the configuration of an existing GCS workflow history export sink.\nThe enabled/disabled state is preserved.\n\nExample:\n\n```\ncloud namespace export gcs update --namespace my-namespace.my-account --sink-name my-sink \\\n  --sa-id my-service-account@my-project.iam.gserviceaccount.com \\\n  --bucket-name my-bucket --gcp-project-id my-project --region us-central1\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.ExportGcsOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportGcsValidateCommand struct {
	Parent  *CloudNamespaceExportGcsCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	ExportSinkOptions
	ExportGcsOptions
}

func NewCloudNamespaceExportGcsValidateCommand(cctx *CommandContext, parent *CloudNamespaceExportGcsCommand) *CloudNamespaceExportGcsValidateCommand {
	var s CloudNamespaceExportGcsValidateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "validate [flags]"
	s.Command.Short = "Validate a GCS workflow history export sink configuration"
	if hasHighlighting {
		s.Command.Long = "Validate a GCS workflow history export sink configuration without creating or updating it.\nA successful response means the configuration is valid.\n\nExample:\n\n\x1b[1mcloud namespace export gcs validate --namespace my-namespace.my-account --sink-name my-sink \\\n  --sa-id my-service-account@my-project.iam.gserviceaccount.com \\\n  --bucket-name my-bucket --gcp-project-id my-project --region us-central1\x1b[0m"
	} else {
		s.Command.Long = "Validate a GCS workflow history export sink configuration without creating or updating it.\nA successful response means the configuration is valid.\n\nExample:\n\n```\ncloud namespace export gcs validate --namespace my-namespace.my-account --sink-name my-sink \\\n  --sa-id my-service-account@my-project.iam.gserviceaccount.com \\\n  --bucket-name my-bucket --gcp-project-id my-project --region us-central1\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.ExportGcsOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportGetCommand struct {
	Parent  *CloudNamespaceExportCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	ExportSinkOptions
}

func NewCloudNamespaceExportGetCommand(cctx *CommandContext, parent *CloudNamespaceExportCommand) *CloudNamespaceExportGetCommand {
	var s CloudNamespaceExportGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get a workflow history export sink"
	if hasHighlighting {
		s.Command.Long = "Retrieve the configuration and status of a workflow history export sink for a\nTemporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace export get --namespace my-namespace.my-account --sink-name my-sink\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the configuration and status of a workflow history export sink for a\nTemporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace export get --namespace my-namespace.my-account --sink-name my-sink\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportListCommand struct {
	Parent  *CloudNamespaceExportCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceExportListCommand(cctx *CommandContext, parent *CloudNamespaceExportCommand) *CloudNamespaceExportListCommand {
	var s CloudNamespaceExportListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List workflow history export sinks for a namespace"
	if hasHighlighting {
		s.Command.Long = "List all workflow history export sinks configured for a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace export list --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "List all workflow history export sinks configured for a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace export list --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportS3Command struct {
	Parent  *CloudNamespaceExportCommand
	Command cobra.Command
}

func NewCloudNamespaceExportS3Command(cctx *CommandContext, parent *CloudNamespaceExportCommand) *CloudNamespaceExportS3Command {
	var s CloudNamespaceExportS3Command
	s.Parent = parent
	s.Command.Use = "s3"
	s.Command.Short = "Manage S3 workflow history export sinks"
	s.Command.Long = "Commands for managing S3 workflow history export sinks for Temporal Cloud namespaces."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceExportS3CreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportS3UpdateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceExportS3ValidateCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceExportS3CreateCommand struct {
	Parent  *CloudNamespaceExportS3Command
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ExportSinkOptions
	ExportS3Options
}

func NewCloudNamespaceExportS3CreateCommand(cctx *CommandContext, parent *CloudNamespaceExportS3Command) *CloudNamespaceExportS3CreateCommand {
	var s CloudNamespaceExportS3CreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create an S3 workflow history export sink"
	if hasHighlighting {
		s.Command.Long = "Create a new S3 workflow history export sink for a Temporal Cloud namespace.\nThe sink is created in the enabled state.\n\nExample:\n\n\x1b[1mcloud namespace export s3 create --namespace my-namespace.my-account --sink-name my-sink \\\n  --role-name arn:aws:iam::123456789012:role/my-role --bucket-name my-bucket \\\n  --region us-east-1 --aws-account-id 123456789012\x1b[0m"
	} else {
		s.Command.Long = "Create a new S3 workflow history export sink for a Temporal Cloud namespace.\nThe sink is created in the enabled state.\n\nExample:\n\n```\ncloud namespace export s3 create --namespace my-namespace.my-account --sink-name my-sink \\\n  --role-name arn:aws:iam::123456789012:role/my-role --bucket-name my-bucket \\\n  --region us-east-1 --aws-account-id 123456789012\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.ExportS3Options.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportS3UpdateCommand struct {
	Parent  *CloudNamespaceExportS3Command
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	ExportSinkOptions
	ExportS3Options
}

func NewCloudNamespaceExportS3UpdateCommand(cctx *CommandContext, parent *CloudNamespaceExportS3Command) *CloudNamespaceExportS3UpdateCommand {
	var s CloudNamespaceExportS3UpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update an S3 workflow history export sink"
	if hasHighlighting {
		s.Command.Long = "Update the configuration of an existing S3 workflow history export sink.\nThe enabled/disabled state is preserved.\n\nExample:\n\n\x1b[1mcloud namespace export s3 update --namespace my-namespace.my-account --sink-name my-sink \\\n  --role-name arn:aws:iam::123456789012:role/my-new-role --bucket-name my-bucket \\\n  --region us-east-1 --aws-account-id 123456789012\x1b[0m"
	} else {
		s.Command.Long = "Update the configuration of an existing S3 workflow history export sink.\nThe enabled/disabled state is preserved.\n\nExample:\n\n```\ncloud namespace export s3 update --namespace my-namespace.my-account --sink-name my-sink \\\n  --role-name arn:aws:iam::123456789012:role/my-new-role --bucket-name my-bucket \\\n  --region us-east-1 --aws-account-id 123456789012\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.ExportS3Options.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceExportS3ValidateCommand struct {
	Parent  *CloudNamespaceExportS3Command
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	ExportSinkOptions
	ExportS3Options
}

func NewCloudNamespaceExportS3ValidateCommand(cctx *CommandContext, parent *CloudNamespaceExportS3Command) *CloudNamespaceExportS3ValidateCommand {
	var s CloudNamespaceExportS3ValidateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "validate [flags]"
	s.Command.Short = "Validate an S3 workflow history export sink configuration"
	if hasHighlighting {
		s.Command.Long = "Validate an S3 workflow history export sink configuration without creating or updating it.\nA successful response means the configuration is valid.\n\nExample:\n\n\x1b[1mcloud namespace export s3 validate --namespace my-namespace.my-account --sink-name my-sink \\\n  --role-name arn:aws:iam::123456789012:role/my-role --bucket-name my-bucket \\\n  --region us-east-1 --aws-account-id 123456789012\x1b[0m"
	} else {
		s.Command.Long = "Validate an S3 workflow history export sink configuration without creating or updating it.\nA successful response means the configuration is valid.\n\nExample:\n\n```\ncloud namespace export s3 validate --namespace my-namespace.my-account --sink-name my-sink \\\n  --role-name arn:aws:iam::123456789012:role/my-role --bucket-name my-bucket \\\n  --region us-east-1 --aws-account-id 123456789012\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.ExportSinkOptions.BuildFlags(s.Command.Flags())
	s.ExportS3Options.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceGetCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
	ClientOptions
	Namespace string
	Spec      bool
}

func NewCloudNamespaceGetCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceGetCommand {
	var s CloudNamespaceGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Retrieve namespace details"
	if hasHighlighting {
		s.Command.Long = "Retrieve the configuration and status of a Temporal Cloud namespace.\n\nReturns details including region, retention period, endpoints, and\ncertificate information.\n\nExample:\n\n\x1b[1mcloud namespace get --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the configuration and status of a Temporal Cloud namespace.\n\nReturns details including region, retention period, endpoints, and\ncertificate information.\n\nExample:\n\n```\ncloud namespace get --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "The fully qualified namespace name in the format 'namespace.account' (e.g., 'my-namespace.my-account'). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace")
	s.Command.Flags().BoolVar(&s.Spec, "spec", false, "Output only the namespace specification in JSON format, omitting metadata and status information.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceHaCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceHaCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceHaCommand {
	var s CloudNamespaceHaCommand
	s.Parent = parent
	s.Command.Use = "ha"
	s.Command.Short = "Manage High Availability settings for namespaces"
	s.Command.Long = "Commands for managing High Availability (HA) settings of Temporal Cloud namespaces.\n\nHA settings control active region, managed failover, and replica regions."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceHaFailoverCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceHaGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceHaRegionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceHaUpdateCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceHaFailoverCommand struct {
	Parent  *CloudNamespaceHaCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	Region string
}

func NewCloudNamespaceHaFailoverCommand(cctx *CommandContext, parent *CloudNamespaceHaCommand) *CloudNamespaceHaFailoverCommand {
	var s CloudNamespaceHaFailoverCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "failover [flags]"
	s.Command.Short = "Trigger a failover to a different region"
	if hasHighlighting {
		s.Command.Long = "Trigger a failover for a Temporal Cloud namespace to a different region.\nThe target region must already be a replica region of the namespace.\n\nExample:\n\n\x1b[1mcloud namespace ha failover --namespace my-namespace.my-account --region aws-us-west-2\x1b[0m"
	} else {
		s.Command.Long = "Trigger a failover for a Temporal Cloud namespace to a different region.\nThe target region must already be a replica region of the namespace.\n\nExample:\n\n```\ncloud namespace ha failover --namespace my-namespace.my-account --region aws-us-west-2\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Region, "region", "", "The target region to failover to (e.g., aws-us-west-2). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "region")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceHaGetCommand struct {
	Parent  *CloudNamespaceHaCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceHaGetCommand(cctx *CommandContext, parent *CloudNamespaceHaCommand) *CloudNamespaceHaGetCommand {
	var s CloudNamespaceHaGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get High Availability configuration for a namespace"
	if hasHighlighting {
		s.Command.Long = "Retrieve the current High Availability configuration for a Temporal Cloud namespace.\nShows the active region and whether managed failover is enabled.\n\nExample:\n\n\x1b[1mcloud namespace ha get --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the current High Availability configuration for a Temporal Cloud namespace.\nShows the active region and whether managed failover is enabled.\n\nExample:\n\n```\ncloud namespace ha get --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceHaRegionCommand struct {
	Parent  *CloudNamespaceHaCommand
	Command cobra.Command
}

func NewCloudNamespaceHaRegionCommand(cctx *CommandContext, parent *CloudNamespaceHaCommand) *CloudNamespaceHaRegionCommand {
	var s CloudNamespaceHaRegionCommand
	s.Parent = parent
	s.Command.Use = "region"
	s.Command.Short = "Manage replica regions for a namespace"
	s.Command.Long = "Commands for managing replica regions of Temporal Cloud namespaces."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceHaRegionAddCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceHaRegionDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceHaRegionListCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceHaRegionAddCommand struct {
	Parent  *CloudNamespaceHaRegionCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Region string
}

func NewCloudNamespaceHaRegionAddCommand(cctx *CommandContext, parent *CloudNamespaceHaRegionCommand) *CloudNamespaceHaRegionAddCommand {
	var s CloudNamespaceHaRegionAddCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "add [flags]"
	s.Command.Short = "Add a replica region to a namespace"
	if hasHighlighting {
		s.Command.Long = "Add a replica region to a Temporal Cloud namespace. The region will be added\nas a passive replica and can later be used for failover.\n\nExample:\n\n\x1b[1mcloud namespace ha region add --namespace my-namespace.my-account --region aws-us-west-2\x1b[0m"
	} else {
		s.Command.Long = "Add a replica region to a Temporal Cloud namespace. The region will be added\nas a passive replica and can later be used for failover.\n\nExample:\n\n```\ncloud namespace ha region add --namespace my-namespace.my-account --region aws-us-west-2\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Region, "region", "", "The region ID to add as a replica (e.g., aws-us-west-2). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "region")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceHaRegionDeleteCommand struct {
	Parent  *CloudNamespaceHaRegionCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Region string
}

func NewCloudNamespaceHaRegionDeleteCommand(cctx *CommandContext, parent *CloudNamespaceHaRegionCommand) *CloudNamespaceHaRegionDeleteCommand {
	var s CloudNamespaceHaRegionDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Remove a replica region from a namespace"
	if hasHighlighting {
		s.Command.Long = "Remove a replica region from a Temporal Cloud namespace. Note that a 7-day\ncooldown period applies before the same region can be re-added.\n\nExample:\n\n\x1b[1mcloud namespace ha region delete --namespace my-namespace.my-account --region aws-us-west-2\x1b[0m"
	} else {
		s.Command.Long = "Remove a replica region from a Temporal Cloud namespace. Note that a 7-day\ncooldown period applies before the same region can be re-added.\n\nExample:\n\n```\ncloud namespace ha region delete --namespace my-namespace.my-account --region aws-us-west-2\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Region, "region", "", "The region ID to remove (e.g., aws-us-west-2). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "region")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceHaRegionListCommand struct {
	Parent  *CloudNamespaceHaRegionCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceHaRegionListCommand(cctx *CommandContext, parent *CloudNamespaceHaRegionCommand) *CloudNamespaceHaRegionListCommand {
	var s CloudNamespaceHaRegionListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List regions for a namespace"
	if hasHighlighting {
		s.Command.Long = "List all regions and their states for a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace ha region list --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "List all regions and their states for a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace ha region list --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceHaUpdateCommand struct {
	Parent  *CloudNamespaceHaCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	DisableAutoFailover bool
}

func NewCloudNamespaceHaUpdateCommand(cctx *CommandContext, parent *CloudNamespaceHaCommand) *CloudNamespaceHaUpdateCommand {
	var s CloudNamespaceHaUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update High Availability configuration for a namespace"
	if hasHighlighting {
		s.Command.Long = "Update the High Availability configuration for a Temporal Cloud namespace.\nUse --disable-auto-failover to toggle Temporal-managed automatic failover.\n\nExample:\n\n\x1b[1mcloud namespace ha update --namespace my-namespace.my-account --disable-auto-failover true\x1b[0m"
	} else {
		s.Command.Long = "Update the High Availability configuration for a Temporal Cloud namespace.\nUse --disable-auto-failover to toggle Temporal-managed automatic failover.\n\nExample:\n\n```\ncloud namespace ha update --namespace my-namespace.my-account --disable-auto-failover true\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.DisableAutoFailover, "disable-auto-failover", false, "Set to true to disable Temporal-managed automatic failover for the namespace. Set to false to re-enable automatic failover. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "disable-auto-failover")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceLifecycleCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceLifecycleCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceLifecycleCommand {
	var s CloudNamespaceLifecycleCommand
	s.Parent = parent
	s.Command.Use = "lifecycle"
	s.Command.Short = "Manage namespace lifecycle settings"
	s.Command.Long = "Commands for managing lifecycle settings of Temporal Cloud namespaces.\n\nLifecycle settings control the behavior and protection of namespaces,\nincluding delete protection to prevent accidental deletion."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceLifecycleGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceLifecycleSetCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceLifecycleGetCommand struct {
	Parent  *CloudNamespaceLifecycleCommand
	Command cobra.Command
	ClientOptions
	Namespace string
}

func NewCloudNamespaceLifecycleGetCommand(cctx *CommandContext, parent *CloudNamespaceLifecycleCommand) *CloudNamespaceLifecycleGetCommand {
	var s CloudNamespaceLifecycleGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get namespace lifecycle configuration"
	if hasHighlighting {
		s.Command.Long = "Retrieve the current lifecycle configuration for a Temporal Cloud namespace.\nLifecycle settings include delete protection status.\n\nExample:\n\n\x1b[1mcloud namespace lifecycle get --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the current lifecycle configuration for a Temporal Cloud namespace.\nLifecycle settings include delete protection status.\n\nExample:\n\n```\ncloud namespace lifecycle get --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "The fully qualified namespace name in the format 'namespace.account' (e.g., 'my-namespace.my-account'). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceLifecycleSetCommand struct {
	Parent  *CloudNamespaceLifecycleCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	Namespace              string
	EnableDeleteProtection bool
	AsyncOperationId       string
	Async                  bool
	Idempotent             bool
	ResourceVersion        string
}

func NewCloudNamespaceLifecycleSetCommand(cctx *CommandContext, parent *CloudNamespaceLifecycleCommand) *CloudNamespaceLifecycleSetCommand {
	var s CloudNamespaceLifecycleSetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set [flags]"
	s.Command.Short = "Set namespace lifecycle configuration"
	if hasHighlighting {
		s.Command.Long = "Set the lifecycle configuration for a Temporal Cloud namespace.\nLifecycle settings include delete protection to prevent accidental deletion.\n\nExample:\n\n\x1b[1mcloud namespace lifecycle set --namespace my-namespace.my-account --enable-delete-protection true\x1b[0m"
	} else {
		s.Command.Long = "Set the lifecycle configuration for a Temporal Cloud namespace.\nLifecycle settings include delete protection to prevent accidental deletion.\n\nExample:\n\n```\ncloud namespace lifecycle set --namespace my-namespace.my-account --enable-delete-protection true\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "The fully qualified namespace name in the format 'namespace.account' (e.g., 'my-namespace.my-account'). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace")
	s.Command.Flags().BoolVar(&s.EnableDeleteProtection, "enable-delete-protection", false, "Enable or disable delete protection for the namespace. When enabled, the namespace cannot be deleted until this flag is set to false. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "enable-delete-protection")
	s.Command.Flags().StringVar(&s.AsyncOperationId, "async-operation-id", "", "Custom identifier for tracking this async operation. If not provided, a unique ID is generated automatically.")
	s.Command.Flags().BoolVar(&s.Async, "async", false, "Return immediately after initiating the operation instead of waiting for completion. Use the returned operation ID to check status later.")
	s.Command.Flags().BoolVar(&s.Idempotent, "idempotent", false, "Succeed silently if the lifecycle configuration is already set to the specified value. Without this flag, the command errors when no change is needed.")
	s.Command.Flags().StringVar(&s.ResourceVersion, "resource-version", "", "Resource version for optimistic concurrency control. If not provided, the current version is fetched automatically.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.DiffOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceListCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
	ClientOptions
	PageSize  int
	PageToken string
	Name      string
}

func NewCloudNamespaceListCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceListCommand {
	var s CloudNamespaceListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List Temporal Cloud namespaces"
	if hasHighlighting {
		s.Command.Long = "List all Temporal Cloud namespaces accessible with the current\nauthentication credentials.\n\nExample:\n\n\x1b[1mcloud namespace list\x1b[0m"
	} else {
		s.Command.Long = "List all Temporal Cloud namespaces accessible with the current\nauthentication credentials.\n\nExample:\n\n```\ncloud namespace list\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Number of namespaces to return per page. Use for paginated results.")
	s.Command.Flags().StringVar(&s.PageToken, "page-token", "", "Token for retrieving the next page of results in a paginated list.")
	s.Command.Flags().StringVar(&s.Name, "name", "", "Filter namespaces by the name as defined in the specification of the namespace.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceRetentionCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceRetentionCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceRetentionCommand {
	var s CloudNamespaceRetentionCommand
	s.Parent = parent
	s.Command.Use = "retention"
	s.Command.Short = "Manage namespace retention settings"
	s.Command.Long = "Commands for managing data retention settings of Temporal Cloud namespaces.\n\nRetention determines how long closed workflow history data are stored before\nbeing automatically deleted."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceRetentionGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceRetentionSetCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceRetentionGetCommand struct {
	Parent  *CloudNamespaceRetentionCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceRetentionGetCommand(cctx *CommandContext, parent *CloudNamespaceRetentionCommand) *CloudNamespaceRetentionGetCommand {
	var s CloudNamespaceRetentionGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get namespace retention period"
	if hasHighlighting {
		s.Command.Long = "Retrieve the current data retention period for a Temporal Cloud namespace.\nThe retention period defines how long closed workflow history data are stored.\n\nExample:\n\n\x1b[1mcloud namespace retention get --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the current data retention period for a Temporal Cloud namespace.\nThe retention period defines how long closed workflow history data are stored.\n\nExample:\n\n```\ncloud namespace retention get --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceRetentionSetCommand struct {
	Parent  *CloudNamespaceRetentionCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	RetentionDays int
}

func NewCloudNamespaceRetentionSetCommand(cctx *CommandContext, parent *CloudNamespaceRetentionCommand) *CloudNamespaceRetentionSetCommand {
	var s CloudNamespaceRetentionSetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set [flags]"
	s.Command.Short = "Set namespace retention period"
	if hasHighlighting {
		s.Command.Long = "Set the data retention period for a Temporal Cloud namespace. The\nretention period defines how long closed workflow history data are stored.\n\nExample:\n\n\x1b[1mcloud namespace retention set --namespace my-namespace.my-account --retention-days 14\x1b[0m"
	} else {
		s.Command.Long = "Set the data retention period for a Temporal Cloud namespace. The\nretention period defines how long closed workflow history data are stored.\n\nExample:\n\n```\ncloud namespace retention set --namespace my-namespace.my-account --retention-days 14\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.RetentionDays, "retention-days", 0, "New retention period in days for closed workflow history data. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "retention-days")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceSearchAttributeCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceSearchAttributeCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceSearchAttributeCommand {
	var s CloudNamespaceSearchAttributeCommand
	s.Parent = parent
	s.Command.Use = "search-attribute"
	s.Command.Short = "Manage custom search attributes for namespaces"
	s.Command.Long = "Commands for managing custom search attributes for Temporal Cloud namespaces.\nSearch attributes enable filtering and searching workflows by custom fields."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceSearchAttributeCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceSearchAttributeListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceSearchAttributeRenameCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceSearchAttributeCreateCommand struct {
	Parent  *CloudNamespaceSearchAttributeCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Name string
	Type string
}

func NewCloudNamespaceSearchAttributeCreateCommand(cctx *CommandContext, parent *CloudNamespaceSearchAttributeCommand) *CloudNamespaceSearchAttributeCreateCommand {
	var s CloudNamespaceSearchAttributeCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a custom search attribute for a namespace"
	if hasHighlighting {
		s.Command.Long = "Create a new custom search attribute for a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace search-attribute create --namespace my-namespace.my-account --name MyField --type Keyword\x1b[0m"
	} else {
		s.Command.Long = "Create a new custom search attribute for a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace search-attribute create --namespace my-namespace.my-account --name MyField --type Keyword\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "The name of the search attribute to create. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVar(&s.Type, "type", "", "The type of the search attribute. Valid values: Text, Keyword, Int, Double, Bool, Datetime, KeywordList. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "type")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceSearchAttributeListCommand struct {
	Parent  *CloudNamespaceSearchAttributeCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceSearchAttributeListCommand(cctx *CommandContext, parent *CloudNamespaceSearchAttributeCommand) *CloudNamespaceSearchAttributeListCommand {
	var s CloudNamespaceSearchAttributeListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List custom search attributes for a namespace"
	if hasHighlighting {
		s.Command.Long = "List all custom search attributes configured for a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace search-attribute list --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "List all custom search attributes configured for a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace search-attribute list --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceSearchAttributeRenameCommand struct {
	Parent  *CloudNamespaceSearchAttributeCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	ResourceVersionOptions
	ExistingName string
	NewName      string
}

func NewCloudNamespaceSearchAttributeRenameCommand(cctx *CommandContext, parent *CloudNamespaceSearchAttributeCommand) *CloudNamespaceSearchAttributeRenameCommand {
	var s CloudNamespaceSearchAttributeRenameCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "rename [flags]"
	s.Command.Short = "Rename a custom search attribute"
	if hasHighlighting {
		s.Command.Long = "Rename an existing custom search attribute for a Temporal Cloud namespace.\nThis operation preserves all existing data associated with the search attribute.\n\nExample:\n\n\x1b[1mcloud namespace search-attribute rename --namespace my-namespace.my-account --existing-name OldField --new-name NewField\x1b[0m"
	} else {
		s.Command.Long = "Rename an existing custom search attribute for a Temporal Cloud namespace.\nThis operation preserves all existing data associated with the search attribute.\n\nExample:\n\n```\ncloud namespace search-attribute rename --namespace my-namespace.my-account --existing-name OldField --new-name NewField\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ExistingName, "existing-name", "", "The current name of the search attribute to rename. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "existing-name")
	s.Command.Flags().StringVar(&s.NewName, "new-name", "", "The new name for the search attribute. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "new-name")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceTagCommand struct {
	Parent  *CloudNamespaceCommand
	Command cobra.Command
}

func NewCloudNamespaceTagCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceTagCommand {
	var s CloudNamespaceTagCommand
	s.Parent = parent
	s.Command.Use = "tag"
	s.Command.Short = "Manage namespace tags"
	s.Command.Long = "Commands for managing tags of Temporal Cloud namespaces.\n\nTags are key-value pairs used for organization and categorization of namespaces."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceTagCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceTagDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceTagListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceTagUpdateCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceTagCreateCommand struct {
	Parent  *CloudNamespaceTagCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	Key   string
	Value string
}

func NewCloudNamespaceTagCreateCommand(cctx *CommandContext, parent *CloudNamespaceTagCommand) *CloudNamespaceTagCreateCommand {
	var s CloudNamespaceTagCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a tag for a namespace"
	if hasHighlighting {
		s.Command.Long = "Create a new tag for a Temporal Cloud namespace. Fails if a tag with\nthe specified key already exists.\n\nExample:\n\n\x1b[1mcloud namespace tag create --namespace my-namespace.my-account --key environment --value production\x1b[0m"
	} else {
		s.Command.Long = "Create a new tag for a Temporal Cloud namespace. Fails if a tag with\nthe specified key already exists.\n\nExample:\n\n```\ncloud namespace tag create --namespace my-namespace.my-account --key environment --value production\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Key, "key", "", "The tag key. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key")
	s.Command.Flags().StringVar(&s.Value, "value", "", "The tag value. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "value")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceTagDeleteCommand struct {
	Parent  *CloudNamespaceTagCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	Key string
}

func NewCloudNamespaceTagDeleteCommand(cctx *CommandContext, parent *CloudNamespaceTagCommand) *CloudNamespaceTagDeleteCommand {
	var s CloudNamespaceTagDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a tag from a namespace"
	if hasHighlighting {
		s.Command.Long = "Delete a tag from a Temporal Cloud namespace by its key.\n\nExample:\n\n\x1b[1mcloud namespace tag delete --namespace my-namespace.my-account --key environment\x1b[0m"
	} else {
		s.Command.Long = "Delete a tag from a Temporal Cloud namespace by its key.\n\nExample:\n\n```\ncloud namespace tag delete --namespace my-namespace.my-account --key environment\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Key, "key", "", "The tag key to delete. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceTagListCommand struct {
	Parent  *CloudNamespaceTagCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
}

func NewCloudNamespaceTagListCommand(cctx *CommandContext, parent *CloudNamespaceTagCommand) *CloudNamespaceTagListCommand {
	var s CloudNamespaceTagListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List tags for a namespace"
	if hasHighlighting {
		s.Command.Long = "List all tags configured for a Temporal Cloud namespace.\n\nExample:\n\n\x1b[1mcloud namespace tag list --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "List all tags configured for a Temporal Cloud namespace.\n\nExample:\n\n```\ncloud namespace tag list --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceTagUpdateCommand struct {
	Parent  *CloudNamespaceTagCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	AsyncOperationOptions
	Key   string
	Value string
}

func NewCloudNamespaceTagUpdateCommand(cctx *CommandContext, parent *CloudNamespaceTagCommand) *CloudNamespaceTagUpdateCommand {
	var s CloudNamespaceTagUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update a tag for a namespace"
	if hasHighlighting {
		s.Command.Long = "Update the value of an existing tag for a Temporal Cloud namespace.\nFails if the specified tag key does not exist.\n\nExample:\n\n\x1b[1mcloud namespace tag update --namespace my-namespace.my-account --key environment --value staging\x1b[0m"
	} else {
		s.Command.Long = "Update the value of an existing tag for a Temporal Cloud namespace.\nFails if the specified tag key does not exist.\n\nExample:\n\n```\ncloud namespace tag update --namespace my-namespace.my-account --key environment --value staging\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Key, "key", "", "The tag key to update. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key")
	s.Command.Flags().StringVar(&s.Value, "value", "", "The new value for the tag. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "value")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserCommand struct {
	Parent  *CloudCommand
	Command cobra.Command
}

func NewCloudUserCommand(cctx *CommandContext, parent *CloudCommand) *CloudUserCommand {
	var s CloudUserCommand
	s.Parent = parent
	s.Command.Use = "user"
	s.Command.Short = "Manage Temporal Cloud users"
	s.Command.Long = "Commands for managing Temporal Cloud users."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudUserApplyCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserEditCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserInviteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserSetAccountRoleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserSetNamespacePermissionsCommand(cctx, &s).Command)
	return &s
}

type CloudUserApplyCommand struct {
	Parent  *CloudUserCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Spec string
}

func NewCloudUserApplyCommand(cctx *CommandContext, parent *CloudUserCommand) *CloudUserApplyCommand {
	var s CloudUserApplyCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "apply [flags]"
	s.Command.Short = "Create or update a user from a specification"
	if hasHighlighting {
		s.Command.Long = "Apply a user configuration to Temporal Cloud. Creates a new user invitation\nif the email does not exist, or updates the existing user to match the specification.\n\nThe specification can be provided as inline JSON or loaded from a file\nby prefixing the path with '@'.\n\nExample with inline JSON:\n\n\x1b[1mcloud user apply --spec '{\"email\": \"alice@example.com\", \"access\": {\"account_access\": {\"role\": \"developer\"}}}'\x1b[0m\n\nExample with file path:\n\n\x1b[1mcloud user apply --spec @user-spec.json\x1b[0m"
	} else {
		s.Command.Long = "Apply a user configuration to Temporal Cloud. Creates a new user invitation\nif the email does not exist, or updates the existing user to match the specification.\n\nThe specification can be provided as inline JSON or loaded from a file\nby prefixing the path with '@'.\n\nExample with inline JSON:\n\n```\ncloud user apply --spec '{\"email\": \"alice@example.com\", \"access\": {\"account_access\": {\"role\": \"developer\"}}}'\n```\n\nExample with file path:\n\n```\ncloud user apply --spec @user-spec.json\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Spec, "spec", "", "User configuration in JSON format. Provide inline JSON directly, or use '@path/to/file.json' to load from a file. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "spec")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.DiffOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserDeleteCommand struct {
	Parent  *CloudUserCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	UserIdentificationOptions
}

func NewCloudUserDeleteCommand(cctx *CommandContext, parent *CloudUserCommand) *CloudUserDeleteCommand {
	var s CloudUserDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a Temporal Cloud user"
	if hasHighlighting {
		s.Command.Long = "Delete a Temporal Cloud user. This action is irreversible.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n\x1b[1mcloud user delete --user-id my-user-id\ncloud user delete --user-email alice@example.com\x1b[0m"
	} else {
		s.Command.Long = "Delete a Temporal Cloud user. This action is irreversible.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n```\ncloud user delete --user-id my-user-id\ncloud user delete --user-email alice@example.com\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.UserIdentificationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserEditCommand struct {
	Parent  *CloudUserCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	AsyncOperationOptions
	ResourceVersionOptions
	UserIdentificationOptions
}

func NewCloudUserEditCommand(cctx *CommandContext, parent *CloudUserCommand) *CloudUserEditCommand {
	var s CloudUserEditCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "edit [flags]"
	s.Command.Short = "Interactively edit a user configuration"
	if hasHighlighting {
		s.Command.Long = "Open a user configuration in your default editor for interactive\nmodification. After saving and closing the editor, the changes are\napplied to Temporal Cloud.\n\nThe editor is determined by the EDITOR environment variable, falling\nback to 'vi' if not set.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n\x1b[1mcloud user edit --user-id my-user-id\ncloud user edit --user-email alice@example.com\x1b[0m"
	} else {
		s.Command.Long = "Open a user configuration in your default editor for interactive\nmodification. After saving and closing the editor, the changes are\napplied to Temporal Cloud.\n\nThe editor is determined by the EDITOR environment variable, falling\nback to 'vi' if not set.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n```\ncloud user edit --user-id my-user-id\ncloud user edit --user-email alice@example.com\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.DiffOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.UserIdentificationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGetCommand struct {
	Parent  *CloudUserCommand
	Command cobra.Command
	ClientOptions
	UserIdentificationOptions
}

func NewCloudUserGetCommand(cctx *CommandContext, parent *CloudUserCommand) *CloudUserGetCommand {
	var s CloudUserGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Retrieve user details"
	if hasHighlighting {
		s.Command.Long = "Retrieve the configuration and status of a Temporal Cloud user.\n\nExample:\n\n\x1b[1mcloud user get --user-id my-user-id\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the configuration and status of a Temporal Cloud user.\n\nExample:\n\n```\ncloud user get --user-id my-user-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.UserIdentificationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserInviteCommand struct {
	Parent  *CloudUserCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	Email           string
	AccountRole     string
	NamespaceAccess []string
}

func NewCloudUserInviteCommand(cctx *CommandContext, parent *CloudUserCommand) *CloudUserInviteCommand {
	var s CloudUserInviteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "invite [flags]"
	s.Command.Short = "Invite a user to Temporal Cloud"
	if hasHighlighting {
		s.Command.Long = "Invite a user to Temporal Cloud by email. Optionally assign an account-level\nrole and namespace-level access permissions.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\n\nExample:\n\n\x1b[1mcloud user invite --email alice@example.com --account-role developer \\\n  --namespace-access my-namespace.my-account=write\x1b[0m"
	} else {
		s.Command.Long = "Invite a user to Temporal Cloud by email. Optionally assign an account-level\nrole and namespace-level access permissions.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\n\nExample:\n\n```\ncloud user invite --email alice@example.com --account-role developer \\\n  --namespace-access my-namespace.my-account=write\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Email, "email", "", "The email address of the user to invite. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "email")
	s.Command.Flags().StringVar(&s.AccountRole, "account-role", "", "The account-level role to assign. Valid values: owner, admin, developer, finance-admin, read, metrics-read.")
	s.Command.Flags().StringArrayVar(&s.NamespaceAccess, "namespace-access", nil, "Namespace access to grant, in the format 'namespace=permission'. Permission must be one of: admin, write, read. Can be repeated.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserListCommand struct {
	Parent  *CloudUserCommand
	Command cobra.Command
	ClientOptions
	PageSize  int
	PageToken string
	Email     string
	Namespace string
}

func NewCloudUserListCommand(cctx *CommandContext, parent *CloudUserCommand) *CloudUserListCommand {
	var s CloudUserListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List Temporal Cloud users"
	if hasHighlighting {
		s.Command.Long = "List all Temporal Cloud users accessible with the current\nauthentication credentials.\n\nExample:\n\n\x1b[1mcloud user list\x1b[0m"
	} else {
		s.Command.Long = "List all Temporal Cloud users accessible with the current\nauthentication credentials.\n\nExample:\n\n```\ncloud user list\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Number of users to return per page. Use for paginated results.")
	s.Command.Flags().StringVar(&s.PageToken, "page-token", "", "Token for retrieving the next page of results in a paginated list.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Filter users by email address.")
	s.Command.Flags().StringVar(&s.Namespace, "namespace", "", "Filter users by the namespace they have access to.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserSetAccountRoleCommand struct {
	Parent  *CloudUserCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	UserIdentificationOptions
	AccountRole string
}

func NewCloudUserSetAccountRoleCommand(cctx *CommandContext, parent *CloudUserCommand) *CloudUserSetAccountRoleCommand {
	var s CloudUserSetAccountRoleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set-account-role [flags]"
	s.Command.Short = "Set the account role for a user"
	if hasHighlighting {
		s.Command.Long = "Set the account-level role for a Temporal Cloud user.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n\x1b[1mcloud user set-account-role --user-id my-user-id --account-role developer\ncloud user set-account-role --user-email alice@example.com --account-role admin\x1b[0m"
	} else {
		s.Command.Long = "Set the account-level role for a Temporal Cloud user.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n```\ncloud user set-account-role --user-id my-user-id --account-role developer\ncloud user set-account-role --user-email alice@example.com --account-role admin\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.AccountRole, "account-role", "", "The account-level role to assign. Valid values: owner, admin, developer, finance-admin, read, metrics-read. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "account-role")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.UserIdentificationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserSetNamespacePermissionsCommand struct {
	Parent  *CloudUserCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	ResourceVersionOptions
	UserIdentificationOptions
	NamespaceAccess []string
}

func NewCloudUserSetNamespacePermissionsCommand(cctx *CommandContext, parent *CloudUserCommand) *CloudUserSetNamespacePermissionsCommand {
	var s CloudUserSetNamespacePermissionsCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set-namespace-permissions [flags]"
	s.Command.Short = "Set namespace permissions for a user"
	if hasHighlighting {
		s.Command.Long = "Add, update, or remove namespace-level permissions for a Temporal Cloud user.\nChanges are applied additively: namespaces not listed are left unchanged.\n\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\nTo remove access to a namespace, pass an empty permission: 'namespace='.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n\x1b[1m# Grant write access to my-namespace and read access to other-namespace:\ncloud user set-namespace-permissions --user-id my-user-id \\\n  --namespace-access my-namespace.my-account=write \\\n  --namespace-access other-namespace.my-account=read\n\n# Remove access to a namespace:\ncloud user set-namespace-permissions --user-id my-user-id \\\n  --namespace-access my-namespace.my-account=\x1b[0m"
	} else {
		s.Command.Long = "Add, update, or remove namespace-level permissions for a Temporal Cloud user.\nChanges are applied additively: namespaces not listed are left unchanged.\n\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\nTo remove access to a namespace, pass an empty permission: 'namespace='.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n```\n# Grant write access to my-namespace and read access to other-namespace:\ncloud user set-namespace-permissions --user-id my-user-id \\\n  --namespace-access my-namespace.my-account=write \\\n  --namespace-access other-namespace.my-account=read\n\n# Remove access to a namespace:\ncloud user set-namespace-permissions --user-id my-user-id \\\n  --namespace-access my-namespace.my-account=\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.NamespaceAccess, "namespace-access", nil, "Namespace access change in the format 'namespace=permission'. Permission must be one of: admin, write, read. Can be repeated. Use an empty permission (e.g. 'testns=') to remove access to a namespace. Changes are additive: namespaces not listed are left unchanged. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace-access")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.UserIdentificationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupCommand struct {
	Parent  *CloudCommand
	Command cobra.Command
}

func NewCloudUserGroupCommand(cctx *CommandContext, parent *CloudCommand) *CloudUserGroupCommand {
	var s CloudUserGroupCommand
	s.Parent = parent
	s.Command.Use = "user-group"
	s.Command.Short = "Manage Temporal Cloud user groups"
	s.Command.Long = "Commands for managing Temporal Cloud user groups."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudUserGroupApplyCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupCreateCloudGroupCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupCreateGoogleGroupCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupCreateScimGroupCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupEditCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupMembersCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupSetAccountRoleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupSetNamespacePermissionsCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupUpdateCommand(cctx, &s).Command)
	return &s
}

type CloudUserGroupApplyCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	AsyncOperationOptions
	ResourceVersionOptions
	Spec string
}

func NewCloudUserGroupApplyCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupApplyCommand {
	var s CloudUserGroupApplyCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "apply [flags]"
	s.Command.Short = "Create or update a user group from a specification"
	if hasHighlighting {
		s.Command.Long = "Apply a user group configuration to Temporal Cloud. Creates a new user group\nif no group with the given display name exists, or updates the existing one\nto match the specification.\n\nThe specification can be provided as inline JSON or loaded from a file\nby prefixing the path with '@'.\n\nExample with inline JSON:\n\n\x1b[1mcloud user-group apply --spec '{\"display_name\": \"Engineering\", \"cloud_group\": {}, \"access\": {\"account_access\": {\"role\": \"developer\"}}}'\x1b[0m\n\nExample with file path:\n\n\x1b[1mcloud user-group apply --spec @user-group-spec.json\x1b[0m"
	} else {
		s.Command.Long = "Apply a user group configuration to Temporal Cloud. Creates a new user group\nif no group with the given display name exists, or updates the existing one\nto match the specification.\n\nThe specification can be provided as inline JSON or loaded from a file\nby prefixing the path with '@'.\n\nExample with inline JSON:\n\n```\ncloud user-group apply --spec '{\"display_name\": \"Engineering\", \"cloud_group\": {}, \"access\": {\"account_access\": {\"role\": \"developer\"}}}'\n```\n\nExample with file path:\n\n```\ncloud user-group apply --spec @user-group-spec.json\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Spec, "spec", "", "User group configuration in JSON format. Provide inline JSON directly, or use '@path/to/file.json' to load from a file. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "spec")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.DiffOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupCreateCloudGroupCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	DisplayName     string
	AccountRole     string
	NamespaceAccess []string
}

func NewCloudUserGroupCreateCloudGroupCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupCreateCloudGroupCommand {
	var s CloudUserGroupCreateCloudGroupCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create-cloud-group [flags]"
	s.Command.Short = "Create a Temporal Cloud-managed user group"
	if hasHighlighting {
		s.Command.Long = "Create a new Temporal Cloud-managed user group. Members can be managed\nusing the add-member and remove-member commands.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\n\nExample:\n\n\x1b[1mcloud user-group create-cloud-group --display-name \"Engineering\" \\\n  --account-role developer \\\n  --namespace-access my-namespace.my-account=write\x1b[0m"
	} else {
		s.Command.Long = "Create a new Temporal Cloud-managed user group. Members can be managed\nusing the add-member and remove-member commands.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\n\nExample:\n\n```\ncloud user-group create-cloud-group --display-name \"Engineering\" \\\n  --account-role developer \\\n  --namespace-access my-namespace.my-account=write\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.DisplayName, "display-name", "", "The display name of the user group. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "display-name")
	s.Command.Flags().StringVar(&s.AccountRole, "account-role", "", "The account-level role to assign. Valid values: owner, admin, developer, finance-admin, read, metrics-read.")
	s.Command.Flags().StringArrayVar(&s.NamespaceAccess, "namespace-access", nil, "Namespace access to grant, in the format 'namespace=permission'. Permission must be one of: admin, write, read. Can be repeated.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupCreateGoogleGroupCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	DisplayName      string
	GoogleGroupEmail string
	AccountRole      string
	NamespaceAccess  []string
}

func NewCloudUserGroupCreateGoogleGroupCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupCreateGoogleGroupCommand {
	var s CloudUserGroupCreateGoogleGroupCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create-google-group [flags]"
	s.Command.Short = "Create a Google-group-backed user group"
	if hasHighlighting {
		s.Command.Long = "Create a new user group backed by a Google Group. Members are managed\nvia the Google Group itself.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\n\nExample:\n\n\x1b[1mcloud user-group create-google-group --display-name \"Platform\" \\\n  --google-group-email platform@example.com \\\n  --account-role developer\x1b[0m"
	} else {
		s.Command.Long = "Create a new user group backed by a Google Group. Members are managed\nvia the Google Group itself.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\n\nExample:\n\n```\ncloud user-group create-google-group --display-name \"Platform\" \\\n  --google-group-email platform@example.com \\\n  --account-role developer\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.DisplayName, "display-name", "", "The display name of the user group. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "display-name")
	s.Command.Flags().StringVar(&s.GoogleGroupEmail, "google-group-email", "", "The email address of the Google Group. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "google-group-email")
	s.Command.Flags().StringVar(&s.AccountRole, "account-role", "", "The account-level role to assign. Valid values: owner, admin, developer, finance-admin, read, metrics-read.")
	s.Command.Flags().StringArrayVar(&s.NamespaceAccess, "namespace-access", nil, "Namespace access to grant, in the format 'namespace=permission'. Permission must be one of: admin, write, read. Can be repeated.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupCreateScimGroupCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	AsyncOperationOptions
	DisplayName     string
	ScimIdpId       string
	AccountRole     string
	NamespaceAccess []string
}

func NewCloudUserGroupCreateScimGroupCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupCreateScimGroupCommand {
	var s CloudUserGroupCreateScimGroupCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create-scim-group [flags]"
	s.Command.Short = "Create a SCIM-backed user group"
	if hasHighlighting {
		s.Command.Long = "Create a new user group backed by a SCIM identity provider group.\nMembers are managed via the upstream identity provider.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\n\nExample:\n\n\x1b[1mcloud user-group create-scim-group --display-name \"Security\" \\\n  --scim-idp-id idp-group-id-123 \\\n  --account-role read\x1b[0m"
	} else {
		s.Command.Long = "Create a new user group backed by a SCIM identity provider group.\nMembers are managed via the upstream identity provider.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\n\nExample:\n\n```\ncloud user-group create-scim-group --display-name \"Security\" \\\n  --scim-idp-id idp-group-id-123 \\\n  --account-role read\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.DisplayName, "display-name", "", "The display name of the user group. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "display-name")
	s.Command.Flags().StringVar(&s.ScimIdpId, "scim-idp-id", "", "The identity provider ID for the SCIM group. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "scim-idp-id")
	s.Command.Flags().StringVar(&s.AccountRole, "account-role", "", "The account-level role to assign. Valid values: owner, admin, developer, finance-admin, read, metrics-read.")
	s.Command.Flags().StringArrayVar(&s.NamespaceAccess, "namespace-access", nil, "Namespace access to grant, in the format 'namespace=permission'. Permission must be one of: admin, write, read. Can be repeated.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupDeleteCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	GroupIdOptions
	AsyncOperationOptions
	ResourceVersionOptions
}

func NewCloudUserGroupDeleteCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupDeleteCommand {
	var s CloudUserGroupDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a Temporal Cloud user group"
	if hasHighlighting {
		s.Command.Long = "Delete a Temporal Cloud user group. This action is irreversible.\n\nExample:\n\n\x1b[1mcloud user-group delete --group-id my-group-id\x1b[0m"
	} else {
		s.Command.Long = "Delete a Temporal Cloud user group. This action is irreversible.\n\nExample:\n\n```\ncloud user-group delete --group-id my-group-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupEditCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	GroupIdOptions
	AsyncOperationOptions
	ResourceVersionOptions
}

func NewCloudUserGroupEditCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupEditCommand {
	var s CloudUserGroupEditCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "edit [flags]"
	s.Command.Short = "Interactively edit a user group configuration"
	if hasHighlighting {
		s.Command.Long = "Open a user group configuration in your default editor for interactive\nmodification. After saving and closing the editor, the changes are\napplied to Temporal Cloud.\n\nThe editor is determined by the EDITOR environment variable, falling\nback to 'vi' if not set.\n\nExample:\n\n\x1b[1mcloud user-group edit --group-id my-group-id\x1b[0m"
	} else {
		s.Command.Long = "Open a user group configuration in your default editor for interactive\nmodification. After saving and closing the editor, the changes are\napplied to Temporal Cloud.\n\nThe editor is determined by the EDITOR environment variable, falling\nback to 'vi' if not set.\n\nExample:\n\n```\ncloud user-group edit --group-id my-group-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.DiffOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupGetCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	GroupIdOptions
}

func NewCloudUserGroupGetCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupGetCommand {
	var s CloudUserGroupGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Retrieve user group details"
	if hasHighlighting {
		s.Command.Long = "Retrieve the configuration and status of a Temporal Cloud user group.\n\nExample:\n\n\x1b[1mcloud user-group get --group-id my-group-id\x1b[0m"
	} else {
		s.Command.Long = "Retrieve the configuration and status of a Temporal Cloud user group.\n\nExample:\n\n```\ncloud user-group get --group-id my-group-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupListCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	PageSize                int
	PageToken               string
	Namespace               string
	DisplayName             string
	GoogleGroupEmailAddress string
	ScimGroupIdpId          string
}

func NewCloudUserGroupListCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupListCommand {
	var s CloudUserGroupListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List Temporal Cloud user groups"
	if hasHighlighting {
		s.Command.Long = "List all Temporal Cloud user groups accessible with the current\nauthentication credentials.\n\nExample:\n\n\x1b[1mcloud user-group list\x1b[0m"
	} else {
		s.Command.Long = "List all Temporal Cloud user groups accessible with the current\nauthentication credentials.\n\nExample:\n\n```\ncloud user-group list\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Number of user groups to return per page. Use for paginated results.")
	s.Command.Flags().StringVar(&s.PageToken, "page-token", "", "Token for retrieving the next page of results in a paginated list.")
	s.Command.Flags().StringVar(&s.Namespace, "namespace", "", "Filter user groups by the namespace they have access to.")
	s.Command.Flags().StringVar(&s.DisplayName, "display-name", "", "Filter user groups by display name.")
	s.Command.Flags().StringVar(&s.GoogleGroupEmailAddress, "google-group-email-address", "", "Filter user groups by Google group email address.")
	s.Command.Flags().StringVar(&s.ScimGroupIdpId, "scim-group-idp-id", "", "Filter user groups by SCIM group IDP ID.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupMembersCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
}

func NewCloudUserGroupMembersCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupMembersCommand {
	var s CloudUserGroupMembersCommand
	s.Parent = parent
	s.Command.Use = "members"
	s.Command.Short = "Manage Temporal Cloud user group members"
	s.Command.Long = "Commands for managing members of Temporal Cloud user groups."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudUserGroupMembersAddCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupMembersListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserGroupMembersRemoveCommand(cctx, &s).Command)
	return &s
}

type CloudUserGroupMembersAddCommand struct {
	Parent  *CloudUserGroupMembersCommand
	Command cobra.Command
	ClientOptions
	GroupIdOptions
	AsyncOperationOptions
	UserIdentificationOptions
}

func NewCloudUserGroupMembersAddCommand(cctx *CommandContext, parent *CloudUserGroupMembersCommand) *CloudUserGroupMembersAddCommand {
	var s CloudUserGroupMembersAddCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "add [flags]"
	s.Command.Short = "Add a member to a Temporal Cloud user group"
	if hasHighlighting {
		s.Command.Long = "Add a user to a Temporal Cloud user group.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n\x1b[1mcloud user-group members add --group-id my-group-id --user-id my-user-id\ncloud user-group members add --group-id my-group-id --user-email alice@example.com\x1b[0m"
	} else {
		s.Command.Long = "Add a user to a Temporal Cloud user group.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n```\ncloud user-group members add --group-id my-group-id --user-id my-user-id\ncloud user-group members add --group-id my-group-id --user-email alice@example.com\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.UserIdentificationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupMembersListCommand struct {
	Parent  *CloudUserGroupMembersCommand
	Command cobra.Command
	ClientOptions
	GroupIdOptions
	PageSize  int
	PageToken string
}

func NewCloudUserGroupMembersListCommand(cctx *CommandContext, parent *CloudUserGroupMembersCommand) *CloudUserGroupMembersListCommand {
	var s CloudUserGroupMembersListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List members of a Temporal Cloud user group"
	if hasHighlighting {
		s.Command.Long = "List all members of a Temporal Cloud user group.\n\nExample:\n\n\x1b[1mcloud user-group members list --group-id my-group-id\x1b[0m"
	} else {
		s.Command.Long = "List all members of a Temporal Cloud user group.\n\nExample:\n\n```\ncloud user-group members list --group-id my-group-id\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Number of members to return per page. Use for paginated results.")
	s.Command.Flags().StringVar(&s.PageToken, "page-token", "", "Token for retrieving the next page of results in a paginated list.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupMembersRemoveCommand struct {
	Parent  *CloudUserGroupMembersCommand
	Command cobra.Command
	ClientOptions
	GroupIdOptions
	AsyncOperationOptions
	UserIdentificationOptions
}

func NewCloudUserGroupMembersRemoveCommand(cctx *CommandContext, parent *CloudUserGroupMembersCommand) *CloudUserGroupMembersRemoveCommand {
	var s CloudUserGroupMembersRemoveCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "remove [flags]"
	s.Command.Short = "Remove a member from a Temporal Cloud user group"
	if hasHighlighting {
		s.Command.Long = "Remove a user from a Temporal Cloud user group.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n\x1b[1mcloud user-group members remove --group-id my-group-id --user-id my-user-id\ncloud user-group members remove --group-id my-group-id --user-email alice@example.com\x1b[0m"
	} else {
		s.Command.Long = "Remove a user from a Temporal Cloud user group.\n\nSpecify the user with either --user-id or --user-email (not both).\n\nExample:\n\n```\ncloud user-group members remove --group-id my-group-id --user-id my-user-id\ncloud user-group members remove --group-id my-group-id --user-email alice@example.com\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.UserIdentificationOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupSetAccountRoleCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	GroupIdOptions
	AsyncOperationOptions
	ResourceVersionOptions
	AccountRole string
}

func NewCloudUserGroupSetAccountRoleCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupSetAccountRoleCommand {
	var s CloudUserGroupSetAccountRoleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set-account-role [flags]"
	s.Command.Short = "Set the account role for a user group"
	if hasHighlighting {
		s.Command.Long = "Set the account-level role for a Temporal Cloud user group.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\n\nExample:\n\n\x1b[1mcloud user-group set-account-role --group-id my-group-id --account-role developer\x1b[0m"
	} else {
		s.Command.Long = "Set the account-level role for a Temporal Cloud user group.\n\nAccount roles: owner, admin, developer, finance-admin, read, metrics-read.\n\nExample:\n\n```\ncloud user-group set-account-role --group-id my-group-id --account-role developer\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.AccountRole, "account-role", "", "The account-level role to assign. Valid values: owner, admin, developer, finance-admin, read, metrics-read. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "account-role")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupSetNamespacePermissionsCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	GroupIdOptions
	AsyncOperationOptions
	ResourceVersionOptions
	NamespaceAccess []string
}

func NewCloudUserGroupSetNamespacePermissionsCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupSetNamespacePermissionsCommand {
	var s CloudUserGroupSetNamespacePermissionsCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set-namespace-permissions [flags]"
	s.Command.Short = "Set namespace permissions for a user group"
	if hasHighlighting {
		s.Command.Long = "Add, update, or remove namespace-level permissions for a Temporal Cloud user group.\nChanges are applied additively: namespaces not listed are left unchanged.\n\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\nTo remove access to a namespace, pass an empty permission: 'namespace='.\n\nExample:\n\n\x1b[1mcloud user-group set-namespace-permissions --group-id my-group-id \\\n  --namespace-access my-namespace.my-account=write \\\n  --namespace-access other-namespace.my-account=read\x1b[0m"
	} else {
		s.Command.Long = "Add, update, or remove namespace-level permissions for a Temporal Cloud user group.\nChanges are applied additively: namespaces not listed are left unchanged.\n\nNamespace access format: 'namespace=permission' where permission is one of: admin, write, read.\nTo remove access to a namespace, pass an empty permission: 'namespace='.\n\nExample:\n\n```\ncloud user-group set-namespace-permissions --group-id my-group-id \\\n  --namespace-access my-namespace.my-account=write \\\n  --namespace-access other-namespace.my-account=read\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.NamespaceAccess, "namespace-access", nil, "Namespace access change in the format 'namespace=permission'. Permission must be one of: admin, write, read. Can be repeated. Use an empty permission (e.g. 'testns=') to remove access to a namespace. Changes are additive: namespaces not listed are left unchanged. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace-access")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudUserGroupUpdateCommand struct {
	Parent  *CloudUserGroupCommand
	Command cobra.Command
	ClientOptions
	GroupIdOptions
	AsyncOperationOptions
	ResourceVersionOptions
	AccountRole     string
	NamespaceAccess []string
}

func NewCloudUserGroupUpdateCommand(cctx *CommandContext, parent *CloudUserGroupCommand) *CloudUserGroupUpdateCommand {
	var s CloudUserGroupUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update a Temporal Cloud user group"
	if hasHighlighting {
		s.Command.Long = "Update an existing Temporal Cloud user group's access settings.\n\nProvide at least one of --account-role or --namespace-access.\n\nExample:\n\n\x1b[1mcloud user-group update --group-id my-group-id --account-role developer\ncloud user-group update --group-id my-group-id \\\n  --namespace-access my-namespace.my-account=write\ncloud user-group update --group-id my-group-id --account-role admin \\\n  --namespace-access my-namespace.my-account=write \\\n  --namespace-access other-namespace.my-account=read\x1b[0m"
	} else {
		s.Command.Long = "Update an existing Temporal Cloud user group's access settings.\n\nProvide at least one of --account-role or --namespace-access.\n\nExample:\n\n```\ncloud user-group update --group-id my-group-id --account-role developer\ncloud user-group update --group-id my-group-id \\\n  --namespace-access my-namespace.my-account=write\ncloud user-group update --group-id my-group-id --account-role admin \\\n  --namespace-access my-namespace.my-account=write \\\n  --namespace-access other-namespace.my-account=read\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.AccountRole, "account-role", "", "The account role to assign to the group. Role must be one of: admin, developer, finance-admin, read.")
	s.Command.Flags().StringArrayVar(&s.NamespaceAccess, "namespace-access", nil, "Namespace access change in the format 'namespace=permission'. Permission must be one of: admin, write, read. Can be repeated. Use an empty permission (e.g. 'testns=') to remove access to a namespace. Changes are additive: namespaces not listed are left unchanged.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.GroupIdOptions.BuildFlags(s.Command.Flags())
	s.AsyncOperationOptions.BuildFlags(s.Command.Flags())
	s.ResourceVersionOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudWhoamiCommand struct {
	Parent  *CloudCommand
	Command cobra.Command
	ClientOptions
}

func NewCloudWhoamiCommand(cctx *CommandContext, parent *CloudCommand) *CloudWhoamiCommand {
	var s CloudWhoamiCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "whoami [flags]"
	s.Command.Short = "Display the current authenticated identity"
	if hasHighlighting {
		s.Command.Long = "Display information about the currently authenticated identity.\n\nShows whether you are authenticated as a user or service account, along\nwith the associated API key if one is in use.\n\nExample:\n\n\x1b[1mcloud whoami\x1b[0m"
	} else {
		s.Command.Long = "Display information about the currently authenticated identity.\n\nShows whether you are authenticated as a user or service account, along\nwith the associated API key if one is in use.\n\nExample:\n\n```\ncloud whoami\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}
