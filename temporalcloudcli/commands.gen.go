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
	s.Command.AddCommand(&NewCloudConnectivityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudLoginCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudLogoutCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudUserCommand(cctx, &s).Command)
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
