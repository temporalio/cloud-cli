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

type ResourceModifyOptions struct {
	ResourceVersion  string
	Idempotent       bool
	AsyncOperationId string
	Async            bool
	FlagSet          *pflag.FlagSet
}

func (v *ResourceModifyOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVarP(&v.ResourceVersion, "resource-version", "v", "", "Resource version for optimistic concurrency control. If not provided, the current version is fetched automatically.")
	f.BoolVar(&v.Idempotent, "idempotent", false, "Succeed silently if the namespace already matches the specification. Without this flag, the command errors when no changes are needed.")
	f.StringVar(&v.AsyncOperationId, "async-operation-id", "", "Custom identifier for tracking this async operation. If not provided, a unique ID is generated automatically.")
	f.BoolVar(&v.Async, "async", false, "Return immediately after initiating the operation instead of waiting for completion. Use the returned operation ID to check status later.")
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
	s.Command.AddCommand(&NewCloudLoginCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudLogoutCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVar(&s.ConfigDir, "config-dir", "", "Directory path where CLI configuration files are stored, including authentication tokens and settings.")
	s.Command.PersistentFlags().BoolVar(&s.DisablePopUp, "disable-pop-up", false, "Prevent the CLI from opening a browser window during authentication. Useful for headless environments or when using alternative auth methods.")
	s.Command.PersistentFlags().BoolVar(&s.AutoConfirm, "auto-confirm", false, "Automatically confirm prompts and actions that require user confirmation. Useful for scripting and automation.")
	s.ClientOptions.BuildFlags(s.Command.PersistentFlags())
	s.CommonOptions.BuildFlags(s.Command.PersistentFlags())
	s.initCommand(cctx)
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
	s.Command.AddCommand(&NewCloudNamespaceDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceEditCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceLifecycleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceRetentionCommand(cctx, &s).Command)
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
	s.Command.AddCommand(&NewCloudNamespaceCertCaAddCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCertCaDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCertCaListCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceCertCaAddCommand struct {
	Parent  *CloudNamespaceCertCaCommand
	Command cobra.Command
	ClientOptions
	NamespaceOptions
	ResourceModifyOptions
	CaCertificateFile string
}

func NewCloudNamespaceCertCaAddCommand(cctx *CommandContext, parent *CloudNamespaceCertCaCommand) *CloudNamespaceCertCaAddCommand {
	var s CloudNamespaceCertCaAddCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "add [flags]"
	s.Command.Short = "TODO"
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.CaCertificateFile, "ca-certificate-file", "", "Path to a CA certificate PEM file.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.ResourceModifyOptions.BuildFlags(s.Command.Flags())
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
	ResourceModifyOptions
	CaCertificateFile string
}

func NewCloudNamespaceCertCaDeleteCommand(cctx *CommandContext, parent *CloudNamespaceCertCaCommand) *CloudNamespaceCertCaDeleteCommand {
	var s CloudNamespaceCertCaDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "TODO"
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.CaCertificateFile, "ca-certificate-file", "", "Path to a CA certificate PEM file.")
	s.ClientOptions.BuildFlags(s.Command.Flags())
	s.NamespaceOptions.BuildFlags(s.Command.Flags())
	s.ResourceModifyOptions.BuildFlags(s.Command.Flags())
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
	s.Command.Short = "TODO"
	s.Command.Long = "TODO"
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
	Namespace string
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

type CloudNamespaceRetentionSetCommand struct {
	Parent  *CloudNamespaceRetentionCommand
	Command cobra.Command
	ClientOptions
	DiffOptions
	Namespace        string
	AsyncOperationId string
	Async            bool
	Idempotent       bool
	RetentionDays    int
	ResourceVersion  string
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
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "The fully qualified namespace name in the format 'namespace.account' (e.g., 'my-namespace.my-account'). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace")
	s.Command.Flags().StringVar(&s.AsyncOperationId, "async-operation-id", "", "Custom identifier for tracking this async operation. If not provided, a unique ID is generated automatically.")
	s.Command.Flags().BoolVar(&s.Async, "async", false, "Return immediately after initiating the operation instead of waiting for completion. Use the returned operation ID to check status later.")
	s.Command.Flags().BoolVar(&s.Idempotent, "idempotent", false, "Succeed silently if the retention period is already set to the specified value. Without this flag, the command errors when no change is needed.")
	s.Command.Flags().IntVar(&s.RetentionDays, "retention-days", 0, "New retention period in days for closed workflow history data. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "retention-days")
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
