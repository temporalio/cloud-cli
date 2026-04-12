//go:build integration
// +build integration

package temporalcloudcli_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	"go.temporal.io/cloud-sdk/cloudclient"
	"google.golang.org/grpc"

	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type CommandHarness struct {
	*require.Assertions
	t       *testing.T
	Options temporalcloudcli.CommandOptions
	// Defaults to a context closed on close or test complete
	Context context.Context
	// Can be used to cancel context given to commands (simulating interrupt)
	CancelContext context.CancelFunc
	Stdin         bytes.Buffer
}

func NewCommandHarness(t *testing.T) *CommandHarness {
	h := &CommandHarness{Assertions: require.New(t), t: t}
	h.Context, h.CancelContext = context.WithCancel(context.Background())
	t.Cleanup(h.Close)
	return h
}

// Reentrant, called after test by default, cancels context
func (h *CommandHarness) Close() {
	// Cancel context
	if h.CancelContext != nil {
		h.CancelContext()
	}
}

// Pieces must appear in order on the line and not overlap
func (h *CommandHarness) ContainsOnSameLine(text string, pieces ...string) {
	h.NoError(AssertContainsOnSameLine(text, pieces...))
}

func AssertContainsOnSameLine(text string, pieces ...string) error {
	// Build regex pattern based on pieces
	pattern := ""
	for _, piece := range pieces {
		if pattern != "" {
			pattern += ".*"
		}
		pattern += regexp.QuoteMeta(piece)
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	// Split into lines, then check each piece is present
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if regex.MatchString(line) {
			return nil
		}
	}
	return fmt.Errorf("pieces not found in order on any line together")
}

func TestAssertContainsOnSameLine(t *testing.T) {
	require.Error(t, AssertContainsOnSameLine("a b c", "b", "a"))
	require.Error(t, AssertContainsOnSameLine("a\nb c", "a", "b"))
	require.NoError(t, AssertContainsOnSameLine("aba", "b", "a"))
	require.NoError(t, AssertContainsOnSameLine("a b a", "b", "a"))
	require.NoError(t, AssertContainsOnSameLine("axb", "a", "b"))
	require.NoError(t, AssertContainsOnSameLine("a a", "a", "a"))
}

func (h *CommandHarness) Eventually(
	condition func() bool,
	waitFor time.Duration,
	tick time.Duration,
	msgAndArgs ...interface{},
) {
	h.t.Helper()
	// We cannot use require.Eventually because it was poorly developed to run the
	// condition function in a goroutine which means it can run after complete or
	// have other race conditions. Don't even need a complicated ticker because it
	// doesn't need to be interruptible.
	for start := time.Now(); time.Since(start) < waitFor; {
		if condition() {
			return
		}
		time.Sleep(tick)
	}
	h.Fail("condition did not evaluate to true within timeout", msgAndArgs...)
}

func (h *CommandHarness) T() *testing.T {
	return h.t
}

type CommandResult struct {
	Err    error
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

func (h *CommandHarness) Execute(args ...string) *CommandResult {
	// Copy options, update as needed
	res := &CommandResult{}
	options := h.Options
	// Set stdio
	options.Stdin = &h.Stdin
	options.Stdout = &res.Stdout
	options.Stderr = &res.Stderr
	// Set args
	options.Args = args
	// Capture error
	options.Fail = func(err error) {
		if res.Err != nil {
			panic("fail called twice, just failed with " + err.Error())
		}
		res.Err = err
	}

	// Run
	ctx, cancel := context.WithCancel(h.Context)
	h.t.Cleanup(cancel)
	defer cancel()
	h.t.Logf("Calling: %v", strings.Join(args, " "))
	temporalcloudcli.Execute(ctx, options)
	if res.Stdout.Len() > 0 {
		h.t.Logf("Stdout:\n-----\n%s\n-----", &res.Stdout)
	}
	if res.Stderr.Len() > 0 {
		h.t.Logf("Stderr:\n-----\n%s\n-----", &res.Stderr)
	}
	return res
}

type EnvLookupMap map[string]string

func (e EnvLookupMap) Environ() []string {
	ret := make([]string, 0, len(e))
	for k := range e {
		ret = append(ret, k)
	}
	return ret
}

func (e EnvLookupMap) LookupEnv(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}

// Run shared server suite
func TestSharedServerSuite(t *testing.T) {
	suite.Run(t, new(SharedServerSuite))
}

type SharedServerSuite struct {
	// Replaced each test
	*CommandHarness

	Suite suite.Suite

	apiKey string
	server string
}

func (s *SharedServerSuite) SetupSuite() {
	s.apiKey = os.Getenv("TEMPORAL_API_KEY")
	s.Suite.Require().NotEmpty(s.apiKey, "Could not load TEMPORAL_API_KEY. Are you running with `mise run test` and have you filled out your .env? See README.md for details.")
	s.server = os.Getenv("TEMPORAL_CLOUD_SERVER")
	s.Suite.Require().NotEmpty(s.server, "Could not load TEMPORAL_CLOUD_SERVER. Are you running with `mise run test` and have you filled out your .env? See README.md for details.")
}

func (s *SharedServerSuite) TearDownSuite() {
}

func (s *SharedServerSuite) SetupTest() {
	// Create new command harness
	s.CommandHarness = NewCommandHarness(s.Suite.T())
}

func (s *SharedServerSuite) TearDownTest() {
	if s.CommandHarness != nil {
		s.CommandHarness.Close()
	}
	s.CommandHarness = nil
}

func (s *SharedServerSuite) T() *testing.T                 { return s.Suite.T() }
func (s *SharedServerSuite) SetT(t *testing.T)             { s.Suite.SetT(t) }
func (s *SharedServerSuite) SetS(suite suite.TestingSuite) { s.Suite.SetS(suite) }

func (s *SharedServerSuite) generateRandomID() string {
	letters := "abcdefghijklmnopqrstuvwxyz123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *SharedServerSuite) getCloudClient() *cloudclient.Client {
	opts := cloudclient.Options{
		APIKey:   s.apiKey,
		HostPort: s.server,
	}

	cloudClient, err := cloudclient.New(opts)
	s.Suite.Require().NoError(err)
	return cloudClient
}

// pollAsyncOperation polls an async operation until it reaches a terminal state.
// This is a test wrapper around the PollAsyncOperation function.
func (s *SharedServerSuite) pollAsyncOperation(
	cloudClient *cloudclient.Client,
	operationID string,
) error {
	// Create a minimal CommandContext for testing (with discard printer to skip output)
	cctx := &temporalcloudcli.CommandContext{
		Context: s.Context,
		Printer: &printer.Printer{
			Output: io.Discard, // Discard all output for tests
		},
	}

	poller := temporalcloudcli.AsyncOperationPoller{CloudClient: cloudClient.CloudService()}
	return poller.PollAsyncOperation(cctx, operationID, "")
}

// --- clearDeprecatedFieldsInterceptor ---

func TestClearDeprecatedFieldsInterceptor(t *testing.T) {
	type nonProto struct{ Name string }

	tests := []struct {
		name        string
		reply       any
		invoker     grpc.UnaryInvoker
		expectedErr string
		verify      func(t *testing.T, reply any)
	}{
		{
			name:  "ClearsDeprecatedFields",
			reply: &identityv1.ApiKey{},
			invoker: func(_ context.Context, _ string, _, rep any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
				key := rep.(*identityv1.ApiKey)
				key.Id = "key-1"
				key.StateDeprecated = "active"
				return nil
			},
			verify: func(t *testing.T, reply any) {
				key := reply.(*identityv1.ApiKey)
				assert.Equal(t, "key-1", key.Id)
				assert.Empty(t, key.StateDeprecated)
			},
		},
		{
			name:  "InvokerError",
			reply: &identityv1.ApiKey{},
			invoker: func(_ context.Context, _ string, _, rep any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
				rep.(*identityv1.ApiKey).StateDeprecated = "active"
				return errors.New("rpc failed")
			},
			expectedErr: "rpc failed",
			verify: func(t *testing.T, reply any) {
				// Fields are NOT cleared when the invoker fails.
				assert.Equal(t, "active", reply.(*identityv1.ApiKey).StateDeprecated)
			},
		},
		{
			name:  "NonProtoReply",
			reply: &nonProto{Name: "ignored"},
			invoker: func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
				return nil
			},
			verify: func(t *testing.T, reply any) {
				assert.Equal(t, "ignored", reply.(*nonProto).Name)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := temporalcloudcli.ClearDeprecatedFieldsInterceptor(context.Background(), "/test.Service/Method", nil, tt.reply, nil, tt.invoker)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			if tt.verify != nil {
				tt.verify(t, tt.reply)
			}
		})
	}
}
