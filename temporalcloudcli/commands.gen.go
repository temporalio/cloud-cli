// Code generated. DO NOT EDIT.

package temporalcloudcli

import (
	"github.com/mattn/go-isatty"

	"github.com/spf13/cobra"

	"os"
)

var hasHighlighting = isatty.IsTerminal(os.Stdout.Fd())

type CloudCommand struct {
	Command                 cobra.Command
	ConfigFile              string
	Profile                 string
	DisableConfigFile       bool
	DisableConfigEnv        bool
	LogLevel                StringEnum
	LogFormat               StringEnum
	Output                  StringEnum
	TimeFormat              StringEnum
	Color                   StringEnum
	NoJsonShorthandPayloads bool
	CommandTimeout          Duration
	ClientConnectTimeout    Duration
	ConfigDir               string
	DisablePopUp            bool
	ApiKey                  string
}

func NewCloudCommand(cctx *CommandContext) *CloudCommand {
	var s CloudCommand
	s.Command.Use = "cloud"
	s.Command.Short = "Temporal Cloud command-line interface"
	if hasHighlighting {
		s.Command.Long = "The Temporal Cloud CLI provides management and operations for Temporal Cloud.\n\nExample:\n\n\x1b[1mcloud namespace\x1b[0m"
	} else {
		s.Command.Long = "The Temporal Cloud CLI provides management and operations for Temporal Cloud.\n\nExample:\n\n```\ncloud namespace\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudLoginCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVar(&s.ConfigFile, "config-file", "", "File path to read TOML config from, defaults to `$CONFIG_PATH/temporal/temporal.toml` where `$CONFIG_PATH` is defined as `$HOME/.config` on Unix, \"$HOME/Library/Application Support\" on macOS, and %AppData% on Windows. EXPERIMENTAL.")
	s.Command.PersistentFlags().StringVar(&s.Profile, "profile", "", "Profile to use for config file. EXPERIMENTAL.")
	s.Command.PersistentFlags().BoolVar(&s.DisableConfigFile, "disable-config-file", false, "If set, disables loading environment config from config file. EXPERIMENTAL.")
	s.Command.PersistentFlags().BoolVar(&s.DisableConfigEnv, "disable-config-env", false, "If set, disables loading environment config from environment variables. EXPERIMENTAL.")
	s.LogLevel = NewStringEnum([]string{"debug", "info", "warn", "error", "never"}, "info")
	s.Command.PersistentFlags().Var(&s.LogLevel, "log-level", "Log level. Default is \"info\" for most commands and \"warn\" for `server start-dev`. Accepted values: debug, info, warn, error, never.")
	s.LogFormat = NewStringEnum([]string{"text", "json", "pretty"}, "text")
	s.Command.PersistentFlags().Var(&s.LogFormat, "log-format", "Log format. Accepted values: text, json.")
	s.Output = NewStringEnum([]string{"text", "json", "jsonl", "none"}, "text")
	s.Command.PersistentFlags().VarP(&s.Output, "output", "o", "Non-logging data output format. Accepted values: text, json, jsonl, none.")
	s.TimeFormat = NewStringEnum([]string{"relative", "iso", "raw"}, "relative")
	s.Command.PersistentFlags().Var(&s.TimeFormat, "time-format", "Time format. Accepted values: relative, iso, raw.")
	s.Color = NewStringEnum([]string{"always", "never", "auto"}, "auto")
	s.Command.PersistentFlags().Var(&s.Color, "color", "Output coloring. Accepted values: always, never, auto.")
	s.Command.PersistentFlags().BoolVar(&s.NoJsonShorthandPayloads, "no-json-shorthand-payloads", false, "Raw payload output, even if the JSON option was used.")
	s.CommandTimeout = 0
	s.Command.PersistentFlags().Var(&s.CommandTimeout, "command-timeout", "The command execution timeout. 0s means no timeout.")
	s.ClientConnectTimeout = 0
	s.Command.PersistentFlags().Var(&s.ClientConnectTimeout, "client-connect-timeout", "The client connection timeout. 0s means no timeout.")
	s.Command.PersistentFlags().StringVar(&s.ConfigDir, "config-dir", "", "The directory to store the config into.")
	s.Command.PersistentFlags().BoolVar(&s.DisablePopUp, "disable-pop-up", false, "Disable browser pop-up.")
	s.Command.PersistentFlags().StringVar(&s.ApiKey, "api-key", "", "The api key to use for auth.")
	s.initCommand(cctx)
	return &s
}

type CloudLoginCommand struct {
	Parent   *CloudCommand
	Command  cobra.Command
	Domain   string
	Audience string
	ClientId string
	Reset    bool
}

func NewCloudLoginCommand(cctx *CommandContext, parent *CloudCommand) *CloudLoginCommand {
	var s CloudLoginCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "login [flags]"
	s.Command.Short = "Log into temporal cloud"
	s.Command.Long = "Log into temporal cloud."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Domain, "domain", "login.tmprl-test.cloud", "The domain to log into.")
	s.Command.Flags().StringVar(&s.Audience, "audience", "https://saas-api.tmprl-test.cloud", "Used for login.")
	s.Command.Flags().StringVar(&s.ClientId, "client-id", "CKpwBvLaP1nScTHfNip3smJMkzXzJsur", "Used for login.")
	s.Command.Flags().BoolVar(&s.Reset, "reset", false, "Reset stored login configuration.")
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
	s.Command.Short = "Manage namespaces"
	s.Command.Long = "Commands for managing namespaces."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewCloudNamespaceApplyCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceEditCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewCloudNamespaceGetCommand(cctx, &s).Command)
	return &s
}

type CloudNamespaceApplyCommand struct {
	Parent           *CloudNamespaceCommand
	Command          cobra.Command
	Spec             string
	DryRun           bool
	AsyncOperationId string
	Idemptotent      bool
	Async            bool
}

func NewCloudNamespaceApplyCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceApplyCommand {
	var s CloudNamespaceApplyCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "apply [flags]"
	s.Command.Short = "Create or update a namespace"
	if hasHighlighting {
		s.Command.Long = "Apply a namespace configuration to Temporal Cloud. This command creates a\nnew namespace or updates an existing one based on the provided specification.\n\nYou can specify the namespace configuration using a JSON specification,\nprovided either inline or as a file path.\n\nExample with inline JSON:\n\n\x1b[1mcloud namespace apply --spec '{\"name\": \"namespace-name\", \"region\": \"us-west-2\", \"retention_days\": 7}'\x1b[0m\n\nExample with file path:\n\n\x1b[1mcloud namespace apply --spec @namespace-spec.json\x1b[0m"
	} else {
		s.Command.Long = "Apply a namespace configuration to Temporal Cloud. This command creates a\nnew namespace or updates an existing one based on the provided specification.\n\nYou can specify the namespace configuration using a JSON specification,\nprovided either inline or as a file path.\n\nExample with inline JSON:\n\n```\ncloud namespace apply --spec '{\"name\": \"namespace-name\", \"region\": \"us-west-2\", \"retention_days\": 7}'\n```\n\nExample with file path:\n\n```\ncloud namespace apply --spec @namespace-spec.json\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Spec, "spec", "s", "", "JSON specification for the namespace configuration. Can be provided as inline JSON or as a file path. If the value starts with '@', it will be treated as a file path (e.g., '@config.json'). Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "spec")
	s.Command.Flags().BoolVar(&s.DryRun, "dry-run", false, "Validate the configuration without applying changes. Shows what would be created or updated.")
	s.Command.Flags().StringVarP(&s.AsyncOperationId, "async-operation-id", "a", "", "The async operation id to use for the request, optional.")
	s.Command.Flags().BoolVarP(&s.Idemptotent, "idemptotent", "i", false, "Determines whether the command should error if there's nothing that has changed.")
	s.Command.Flags().BoolVarP(&s.Async, "async", "c", false, "Determines whether the command should return immediately with the async operation or wait until it completes.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceEditCommand struct {
	Parent           *CloudNamespaceCommand
	Command          cobra.Command
	Namespace        string
	AsyncOperationId string
	Idemptotent      bool
	Async            bool
}

func NewCloudNamespaceEditCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceEditCommand {
	var s CloudNamespaceEditCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "edit [flags]"
	s.Command.Short = "Edit a namespace"
	if hasHighlighting {
		s.Command.Long = "Edit a namespace spec on Temporal Cloud. This command updates a namespace with changes\nspecified by the user in an edit operation.\n\nExample:\n\n\x1b[1mcloud namespace edit --namespace my-namespace.my-account\x1b[0m"
	} else {
		s.Command.Long = "Edit a namespace spec on Temporal Cloud. This command updates a namespace with changes\nspecified by the user in an edit operation.\n\nExample:\n\n```\ncloud namespace edit --namespace my-namespace.my-account\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "The namespace to get, including the account. For example my-namespace.my-account. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace")
	s.Command.Flags().StringVarP(&s.AsyncOperationId, "async-operation-id", "a", "", "The async operation id to use for the request, optional.")
	s.Command.Flags().BoolVarP(&s.Idemptotent, "idemptotent", "i", false, "Determines whether the command should error if there's nothing that has changed.")
	s.Command.Flags().BoolVarP(&s.Async, "async", "c", false, "Determines whether the command should return immediately with the async operation or wait until it completes.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type CloudNamespaceGetCommand struct {
	Parent    *CloudNamespaceCommand
	Command   cobra.Command
	Namespace string
}

func NewCloudNamespaceGetCommand(cctx *CommandContext, parent *CloudNamespaceCommand) *CloudNamespaceGetCommand {
	var s CloudNamespaceGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get a namespace"
	s.Command.Long = "Get a namespace from temporal cloud."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Namespace, "namespace", "n", "", "The namespace to get, including the account. For example my-namespace.my-account. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "namespace")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}
