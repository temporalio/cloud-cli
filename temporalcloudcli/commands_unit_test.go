package temporalcloudcli_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// ExecuteAsSubprocess is primarily usedful for testing default Fail() behavior of commands,
// since Excecute sets its own Fail handler.
func (h *CommandHarness) ExecuteAsSubprocess(args ...string) *CommandResult {
	res := &CommandResult{}
	testName := h.t.Name()
	if os.Getenv("GO_TEST_FUNC") == testName {
		options := h.Options
		options.Args = args
		// Run
		ctx, cancel := context.WithCancel(h.Context)
		h.t.Cleanup(cancel)
		defer cancel()
		h.t.Logf("Calling: %v\n", strings.Join(args, " "))
		temporalcloudcli.Execute(ctx, options)
		return nil
	}

	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=^%s$", testName))
	cmd.Env = append(os.Environ(), fmt.Sprintf("GO_TEST_FUNC=%s", testName))

	stdoutPipe, err := cmd.StdoutPipe()
	assert.NoErrorf(h.t, err, "Could not set up stdout")
	stderrPipe, err := cmd.StderrPipe()
	assert.NoErrorf(h.t, err, "Could not set up stderr")

	err = cmd.Start()
	assert.NoErrorf(h.t, err, "Could not start subcommand")

	_, err = res.Stdout.ReadFrom(stdoutPipe)
	assert.NoErrorf(h.t, err, "Failed reading stdout from subcommand")
	_, err = res.Stderr.ReadFrom(stderrPipe)
	assert.NoErrorf(h.t, err, "Failed reading stderr from subcommand")

	res.Err = cmd.Wait()
	return res
}

func TestUnknownCommandExitsNonzero(t *testing.T) {
	commandHarness := NewCommandHarness(t)
	res := commandHarness.Execute("blerkflow")
	assert.Contains(t, res.Err.Error(), "unknown command")
}

func TestErrorsAreNotLogMessages(t *testing.T) {
	commandHarness := NewCommandHarness(t)
	t.Setenv("TEMPORAL_API_KEY", "")
	res := commandHarness.ExecuteAsSubprocess("whoami")

	require.Error(t, res.Err)
	stdout := res.Stdout.String()
	stderr := res.Stderr.String()
	assert.NotContains(t, stdout, "level=")
	assert.NotContains(t, stderr, "level=")
	assert.Contains(t, stderr, "no login session found, please run `temporal cloud login`")
}

func TestErrorsAreNotLogMessagesJSON(t *testing.T) {
	commandHarness := NewCommandHarness(t)
	t.Setenv("TEMPORAL_API_KEY", "")
	res := commandHarness.ExecuteAsSubprocess("whoami", "--output", "json")

	require.Error(t, res.Err)
	stdout := res.Stdout.String()
	stderr := res.Stderr.String()
	assert.NotContains(t, stdout, "level=")
	assert.NotContains(t, stderr, "level=")
	assert.Contains(t, stderr, "no login session found, please run `temporal cloud login`")
}
