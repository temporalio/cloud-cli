package prompter_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	"github.com/temporalio/cloud-cli/temporalcloudcli/prompter"
)

func newTestPrompter(out *bytes.Buffer, stdinContent string, jsonMode bool, autoConfirm bool) prompter.Prompter {
	p := &printer.Printer{Output: out, JSON: jsonMode}
	stdin := bufio.NewReader(strings.NewReader(stdinContent))
	return prompter.NewPrompter(p, stdin, autoConfirm)
}

func TestPromptApply(t *testing.T) {
	oldSpec := &identityv1.ApiKeySpec{DisplayName: "old-key"}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "new-key"}

	tests := []struct {
		name           string
		json           bool
		autoConfirm    bool
		stdin          string
		expectedErr    string
		outputContains string
	}{
		{
			name:           "AutoConfirm",
			autoConfirm:    true,
			outputContains: "Apply (y/yes)? yes",
		},
		{
			name:  "UserAcceptsWithY",
			stdin: "y\n",
		},
		{
			name:  "UserAcceptsWithYes",
			stdin: "yes\n",
		},
		{
			name:        "UserDeclines",
			stdin:       "n\n",
			expectedErr: "Aborting apply.",
		},
		{
			name:        "JSONModeRequiresAutoConfirm",
			json:        true,
			autoConfirm: false,
			expectedErr: "must bypass prompts when using JSON output",
		},
		{
			name:        "JSONModeWithAutoConfirm",
			json:        true,
			autoConfirm: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			pr := newTestPrompter(&buf, tt.stdin, tt.json, tt.autoConfirm)

			err := pr.PromptApply(oldSpec, newSpec, false)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			if tt.outputContains != "" {
				assert.Contains(t, buf.String(), tt.outputContains)
			}
		})
	}
}

// TestPromptApply_JSONModePrintsDiff verifies that in JSON mode the diff is
// emitted before the prompt is evaluated.
func TestPromptApply_JSONModePrintsDiff(t *testing.T) {
	oldSpec := &identityv1.ApiKeySpec{DisplayName: "before"}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "after"}

	var buf bytes.Buffer
	pr := newTestPrompter(&buf, "", true, true)

	err := pr.PromptApply(oldSpec, newSpec, false)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "before")
	assert.Contains(t, out, "after")
}

// TestPromptApply_VerboseFlag passes verbose=true and confirms no error is
// returned; verbose only affects which diff lines are shown.
func TestPromptApply_VerboseFlag(t *testing.T) {
	spec := &identityv1.ApiKeySpec{DisplayName: "key"}

	var buf bytes.Buffer
	pr := newTestPrompter(&buf, "", false, true)

	err := pr.PromptApply(spec, spec, true)
	require.NoError(t, err)
}
