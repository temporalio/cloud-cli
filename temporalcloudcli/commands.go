package temporalcloudcli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/temporalproto"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"go.temporal.io/cloud-sdk/api/resource/v1"
	"go.temporal.io/cloud-sdk/cloudclient"
	"go.temporal.io/sdk/contrib/envconfig"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// Version is the value put as the default command version. This is often
// replaced at build time via ldflags.
var Version = "0.0.0-DEV"

type CommandContext struct {
	// This context is closed on interrupt
	context.Context
	Options                   CommandOptions
	DeprecatedEnvConfigValues map[string]map[string]string
	FlagsWithEnvVars          []*pflag.Flag

	// These values may not be available until after pre-run of main command
	Printer               *printer.Printer
	Logger                *slog.Logger
	JSONOutput            bool
	JSONShorthandPayloads bool

	// Is set to true if any command actually started running. This is a hack to workaround the fact
	// that cobra does not properly exit nonzero if an unknown command/subcommand is given.
	ActuallyRanCommand bool

	// Root/current command only set inside of pre-run
	RootCommand    *CloudCommand
	CurrentCommand *cobra.Command

	NamespaceClient NamespaceClient
	Poller          Poller
}

type NamespaceClient interface {
	AddCACerts(context.Context, namespace.AddCACertsParams) (*operation.AsyncOperation, error)
	ListCACerts(context.Context, string) ([]cert.CACert, error)
	DeleteCACerts(context.Context, namespace.DeleteCACertsParams) (*operation.AsyncOperation, error)
	AddCertFilters(context.Context, namespace.AddCertFiltersParams) (*operation.AsyncOperation, error)
	ListCertFilters(context.Context, string) ([]*namespacev1.CertificateFilterSpec, error)
	DeleteCertFilters(context.Context, namespace.DeleteCertFiltersParams) (*operation.AsyncOperation, error)
	GetNamespace(context.Context, string) (*namespacev1.Namespace, error)
}

type Poller interface {
	PollAsyncOperation(*CommandContext, string, string) error
}

type CommandOptions struct {
	// If empty, assumed to be os.Args[1:]
	Args []string
	// If nil, [envconfig.EnvLookupOS] is used.
	EnvLookup envconfig.EnvLookup

	// These three fields below default to OS values
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Defaults to logging error then os.Exit(1)
	Fail func(error)

	AdditionalClientGRPCDialOptions []grpc.DialOption
	ClientConnectTimeout            time.Duration
}

// NewCommandContext creates a CommandContext for use by the rest of the CLI.
// Among other things, this parses the env config file and modifies
// options/flags according to the parameters set there.
//
// A CommandContext and CancelFunc are always returned, even in the event of an
// error; this is so the CommandContext can be used to print an appropriate
// error message.
func NewCommandContext(ctx context.Context, options CommandOptions) (*CommandContext, context.CancelFunc, error) {
	cctx := &CommandContext{Context: ctx, Options: options}
	if err := cctx.preprocessOptions(); err != nil {
		return cctx, func() {}, err
	}

	// Setup interrupt handler
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	cctx.Context = ctx
	return cctx, stop, nil
}

// BuildCloudOptions creates CloudOptions from the command's ClientOptions.
// It uses the RootCommand's CommonOptions and the CommandContext's logger.
//
// This method encapsulates the CloudOptionsBuilder pattern and should be used
// by all commands that need CloudOptions.
//
// AIDEV-NOTE: This is the standard way for commands to create CloudOptions.
// It automatically uses cctx.RootCommand.CommonOptions regardless of command depth.
func (cctx *CommandContext) BuildCloudOptions(clientOpts ClientOptions) (*CloudOptions, error) {
	builder := CloudOptionsBuilder{
		ClientOptions: clientOpts,
		CommonOptions: cctx.RootCommand.CommonOptions,
		Logger:        cctx.Logger,
		EnvLookup:     envconfig.EnvLookupOS,
	}
	return builder.Build(cctx.Context)
}

// BuildCloudClient creates a CloudClient from the command's ClientOptions.
// It builds CloudOptions internally and then creates the client.
//
// This is a convenience method for commands that need the CloudClient directly
// without needing to keep a reference to CloudOptions.
//
// AIDEV-NOTE: Use this method in command run functions instead of manually
// creating CloudOptions and CloudClient separately.
func (cctx *CommandContext) BuildCloudClient(clientOpts ClientOptions) (*cloudclient.Client, error) {
	cloudOpts, err := cctx.BuildCloudOptions(clientOpts)
	if err != nil {
		return nil, err
	}
	opts := cloudclient.Options{
		UserAgent: fmt.Sprintf("temporalio-cloud-cli/%s", VersionString()),
	}
	if cloudOpts.Server != "" {
		opts.HostPort = cloudOpts.Server
	}
	if cloudOpts.ApiKey != "" {
		// an explicit api key was provided, use it
		opts.APIKey = cloudOpts.ApiKey
	} else {
		// fallaback to the oauth based sso token provider
		opts.APIKeyReader = cloudOpts
	}

	cloudClient, err := cloudclient.New(opts)
	if err != nil {
		return nil, err
	}
	return cloudClient, nil
}

func (c *CommandContext) preprocessOptions() error {
	if len(c.Options.Args) == 0 {
		c.Options.Args = os.Args[1:]
	}
	if c.Options.EnvLookup == nil {
		c.Options.EnvLookup = envconfig.EnvLookupOS
	}

	if c.Options.Stdin == nil {
		c.Options.Stdin = os.Stdin
	}
	if c.Options.Stdout == nil {
		c.Options.Stdout = os.Stdout
	}
	if c.Options.Stderr == nil {
		c.Options.Stderr = os.Stderr
	}

	// Setup default fail callback
	if c.Options.Fail == nil {
		c.Options.Fail = func(err error) {
			// If context is closed, say that the program was interrupted and ignore
			// the actual error
			if c.Err() != nil {
				err = fmt.Errorf("program interrupted")
			}
			if c.Logger != nil {
				c.Logger.Error(err.Error())
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(1)
		}
	}

	return nil
}

const flagEnvVarAnnotation = "__temporal_env_var"

func (c *CommandContext) BindFlagEnvVar(flag *pflag.Flag, envVar string) {
	if flag.Annotations == nil {
		flag.Annotations = map[string][]string{}
	}
	flag.Annotations[flagEnvVarAnnotation] = []string{envVar}
	c.FlagsWithEnvVars = append(c.FlagsWithEnvVars, flag)
}

func (c *CommandContext) MarshalFriendlyJSONPayloads(m *common.Payloads) (json.RawMessage, error) {
	if m == nil {
		return []byte("null"), nil
	}
	// Use one if there's one, otherwise just serialize whole thing
	if p := m.GetPayloads(); len(p) == 1 {
		return c.MarshalProtoJSON(p[0])
	}
	return c.MarshalProtoJSON(m)
}

// Starts with newline
func (c *CommandContext) MarshalFriendlyFailureBodyText(f *failure.Failure, indent string) (s string) {
	for f != nil {
		s += "\n" + indent + "Message: " + f.Message
		if f.StackTrace != "" {
			s += "\n" + indent + "StackTrace:\n" + indent + "    " +
				strings.Join(strings.Split(f.StackTrace, "\n"), "\n"+indent+"    ")
		}
		if f = f.Cause; f != nil {
			s += "\n" + indent + "Cause:"
			indent += "    "
		}
	}
	return
}

// Takes payload shorthand into account, can use
// MarshalProtoJSONNoPayloadShorthand if needed
func (c *CommandContext) MarshalProtoJSON(m proto.Message) ([]byte, error) {
	return c.MarshalProtoJSONWithOptions(m, c.JSONShorthandPayloads)
}

func (c *CommandContext) MarshalProtoJSONWithOptions(m proto.Message, jsonShorthandPayloads bool) ([]byte, error) {
	opts := temporalproto.CustomJSONMarshalOptions{Indent: c.Printer.JSONIndent}
	if jsonShorthandPayloads {
		opts.Metadata = map[string]any{common.EnablePayloadShorthandMetadataKey: true}
	}
	return opts.Marshal(m)
}

func (c *CommandContext) UnmarshalProtoJSON(b []byte, m proto.Message) error {
	return UnmarshalProtoJSONWithOptions(b, m, c.JSONShorthandPayloads)
}

func UnmarshalProtoJSONWithOptions(b []byte, m proto.Message, jsonShorthandPayloads bool) error {
	opts := temporalproto.CustomJSONUnmarshalOptions{DiscardUnknown: true}
	if jsonShorthandPayloads {
		opts.Metadata = map[string]any{common.EnablePayloadShorthandMetadataKey: true}
	}
	return opts.Unmarshal(b, m)
}

// Set flag values from environment file & variables. Returns a callback to log anything interesting
// since logging will not yet be initialized when this runs.
func (c *CommandContext) populateFlagsFromEnv(flags *pflag.FlagSet) (func(*slog.Logger), error) {
	if flags == nil {
		return func(logger *slog.Logger) {}, nil
	}
	var logCalls []func(*slog.Logger)
	var flagErr error
	flags.VisitAll(func(flag *pflag.Flag) {
		// If the flag was already changed by the user, we don't overwrite
		if flagErr != nil || flag.Changed {
			return
		}

		if anns := flag.Annotations[flagEnvVarAnnotation]; len(anns) == 1 {
			if envVal, _ := c.Options.EnvLookup.LookupEnv(anns[0]); envVal != "" {
				if err := flag.Value.Set(envVal); err != nil {
					flagErr = fmt.Errorf("failed setting flag %v with env name %v and value %v: %w",
						flag.Name, anns[0], envVal, err)
					return
				}
				if flag.Changed {
					logCalls = append(logCalls, func(l *slog.Logger) {
						l.Info("Env var overrode --env setting", "env_var", anns[0], "flag", flag.Name)
					})
				}
				flag.Changed = true
			}
		}
	})
	logFn := func(logger *slog.Logger) {
		for _, call := range logCalls {
			call(logger)
		}
	}
	return logFn, flagErr
}

// Returns error if JSON output enabled
func (c *CommandContext) promptYes(message string, autoConfirm bool) (bool, error) {
	if c.JSONOutput && !autoConfirm {
		return false, fmt.Errorf("must bypass prompts when using JSON output")
	}
	c.Printer.Print(message, " ")
	if autoConfirm {
		c.Printer.Println("yes")
		return true, nil
	}
	line, _ := bufio.NewReader(c.Options.Stdin).ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}

// Returns error if JSON output enabled
func (c *CommandContext) promptString(message string, expected string, autoConfirm bool) (bool, error) {
	if c.JSONOutput && !autoConfirm {
		return false, fmt.Errorf("must bypass prompts when using JSON output")
	}
	c.Printer.Print(message, " ")
	if autoConfirm {
		c.Printer.Println(expected)
		return true, nil
	}
	line, _ := bufio.NewReader(c.Options.Stdin).ReadString('\n')
	line = strings.TrimSpace(line)
	return line == expected, nil
}

// Execute runs the Temporal CLI with the given context and options. This
// intentionally does not return an error but rather invokes Fail on the
// options.
func Execute(ctx context.Context, options CommandOptions) {
	// Create context and run. We always get a context and cancel func back even
	// if an error was returned. This is so we can use the context to print an
	// error message using the appropriate Fail() method, regardless of why the
	// failure occurred.
	//
	// (In most cases, an error here likely means a problem with the user's env
	// config file, or some other issue in their environment.)
	cctx, cancel, err := NewCommandContext(ctx, options)
	defer cancel()

	if err == nil {
		// We have a context; let's actually run the command.
		cmd := NewCloudCommand(cctx)
		cmd.Command.SetArgs(cctx.Options.Args)
		err = cmd.Command.ExecuteContext(cctx)
	}

	if err != nil {
		// Either we failed to create the context, OR the command itself failed.
		// Either way, we need to print an error message.
		cctx.Options.Fail(err)
	}

	// If no command ever actually got run, exit nonzero with an error.  This is
	// an ugly hack to make sure that iff the user explicitly asked for help, we
	// exit with a zero error code.  (The other situation in which help is
	// printed is when the user invokes an unknown command--we still want a
	// non-zero exit in that case.)  We should revisit this if/when the
	// following Cobra issues get fixed:
	//
	// - https://github.com/spf13/cobra/issues/1156
	// - https://github.com/spf13/cobra/issues/706
	if !cctx.ActuallyRanCommand {
		zeroExitArgs := []string{"--help", "-h", "--version", "-v", "help"}
		if slices.ContainsFunc(cctx.Options.Args, func(a string) bool {
			return slices.Contains(zeroExitArgs, a)
		}) {
			return
		}
		cctx.Options.Fail(fmt.Errorf("unknown command"))
	}
}

// getUsageTemplate returns a custom usage template with proper flag wrapping
// The default template can be found here: https://github.com/spf13/cobra/blob/v1.9.1/command.go#L1937-L1966
func getUsageTemplate() string {
	// Get terminal width, default to 80 if unable to determine
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	// Use width - 1 for wrapping to avoid edge cases
	flagWidth := width - 1

	// Custom template that uses FlagUsagesWrapped for proper indentation
	return fmt.Sprintf(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsagesWrapped %d | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsagesWrapped %d | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`, flagWidth, flagWidth)
}

func (c *CloudCommand) initCommand(cctx *CommandContext) {
	c.Command.Version = VersionString()

	// Set custom usage template with proper flag wrapping
	c.Command.SetUsageTemplate(getUsageTemplate())

	// Unfortunately color is a global option, so we can set in pre-run but we
	// must unset in post-run
	origNoColor := color.NoColor
	// AIDEV-NOTE: Store cancel function for command timeout context to prevent resource leak
	var timeoutCancel context.CancelFunc
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Set command
		cctx.CurrentCommand = cmd
		// Populate environ. We will make the error return here which will cause
		// usage to be printed.
		logCalls, err := cctx.populateFlagsFromEnv(cmd.Flags())
		if err != nil {
			return err
		}

		// Default color.NoColor global is equivalent to "auto" so only override if
		// never or always
		if c.Color.Value == "never" || c.Color.Value == "always" {
			color.NoColor = c.Color.Value == "never"
		}

		res := c.preRun(cctx, &timeoutCancel)

		logCalls(cctx.Logger)

		// Always disable color if JSON output is on (must be run after preRun so JSONOutput is set)
		if cctx.JSONOutput {
			color.NoColor = true
		}
		cctx.ActuallyRanCommand = true

		return res
	}
	c.Command.PersistentPostRun = func(*cobra.Command, []string) {
		// AIDEV-NOTE: Clean up command timeout context to prevent resource leak
		if timeoutCancel != nil {
			timeoutCancel()
		}
		color.NoColor = origNoColor
	}
}

var buildInfo string

func VersionString() string {
	// To add build-time information to the version string, use
	// go build -ldflags "-X github.com/temporalio/cloud-cli/temporalcloudcli.buildInfo=<MyString>"
	bi := buildInfo
	if bi != "" {
		bi = fmt.Sprintf(", %s", bi)
	}
	return fmt.Sprintf("%s%s", Version, bi)
}

func registerKnownPrinterEnumToStringConverters(p *printer.Printer) {
	// Register any enum converters for known types here.
	printer.RegisterEnumToStringConverter[resource.ResourceState](p, "RESOURCE_STATE_", resource.ResourceState_name)
	printer.RegisterEnumToStringConverter[operation.AsyncOperation_State](p, "STATE_", operation.AsyncOperation_State_name)
}

func (c *CloudCommand) preRun(cctx *CommandContext, timeoutCancel *context.CancelFunc) error {
	// Set this command as the root
	cctx.RootCommand = c

	// Configure logger if not already on context
	if cctx.Logger == nil {
		// If level is never, make noop logger
		if c.LogLevel.Value == "never" {
			cctx.Logger = newNopLogger()
		} else {
			var level slog.Level
			if err := level.UnmarshalText([]byte(c.LogLevel.Value)); err != nil {
				return fmt.Errorf("invalid log level %q: %w", c.LogLevel.Value, err)
			}
			var handler slog.Handler
			switch c.LogFormat.Value {
			// We have a "pretty" alias for compatibility
			case "", "text", "pretty":
				handler = slog.NewTextHandler(cctx.Options.Stderr, &slog.HandlerOptions{
					Level: level,
					// Remove the TZ
					ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
						if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
							a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000"))
						}
						return a
					},
				})
			case "json":
				handler = slog.NewJSONHandler(cctx.Options.Stderr, &slog.HandlerOptions{Level: level})
			default:
				return fmt.Errorf("invalid log format %q", c.LogFormat.Value)
			}
			cctx.Logger = slog.New(handler)
		}
	}

	// Configure printer if not already on context
	cctx.JSONOutput = c.Output.Value == "json" || c.Output.Value == "jsonl"
	// Only indent JSON if not jsonl
	var jsonIndent string
	if c.Output.Value == "json" {
		jsonIndent = "  "
	}
	if cctx.Printer == nil {
		printerOutput := cctx.Options.Stdout
		// Disable printer by making writer noop if "none" chosen
		if c.Output.Value == "none" {
			printerOutput = nopWriter{}
		}
		cctx.Printer = &printer.Printer{
			Output:               printerOutput,
			JSON:                 cctx.JSONOutput,
			JSONIndent:           jsonIndent,
			JSONPayloadShorthand: !c.NoJsonShorthandPayloads,
		}
		switch c.TimeFormat.Value {
		case "iso":
			cctx.Printer.FormatTime = func(t time.Time) string { return t.Format(time.RFC3339) }
		case "raw":
			cctx.Printer.FormatTime = func(t time.Time) string { return fmt.Sprintf("%v", t) }
		case "relative":
			cctx.Printer.FormatTime = humanize.Time
		default:
			return fmt.Errorf("invalid time format %q", c.TimeFormat.Value)
		}
		registerKnownPrinterEnumToStringConverters(cctx.Printer)
	}
	cctx.JSONShorthandPayloads = !c.NoJsonShorthandPayloads
	if c.CommandTimeout.Duration() > 0 {
		// AIDEV-NOTE: Store cancel function to prevent timeout goroutine leak
		cctx.Context, *timeoutCancel = context.WithTimeoutCause(
			cctx.Context,
			c.CommandTimeout.Duration(),
			fmt.Errorf("command timed out after %v", c.CommandTimeout.Duration()),
		)
	}
	if c.ClientConnectTimeout.Duration() > 0 {
		cctx.Options.ClientConnectTimeout = c.ClientConnectTimeout.Duration()
	}

	return nil
}

func newNopLogger() *slog.Logger { return slog.New(discardLogHandler{}) }

type discardLogHandler struct{}

func (discardLogHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardLogHandler) Handle(context.Context, slog.Record) error { return nil }
func (d discardLogHandler) WithAttrs([]slog.Attr) slog.Handler      { return d }
func (d discardLogHandler) WithGroup(string) slog.Handler           { return d }

type nopWriter struct{}

func (nopWriter) Write(b []byte) (int, error) { return len(b), nil }
