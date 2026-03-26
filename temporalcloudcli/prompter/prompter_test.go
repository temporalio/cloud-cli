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
		expectedResult bool
		expectedErr    string
		outputContains string
	}{
		{
			name:           "AutoConfirm",
			autoConfirm:    true,
			expectedResult: true,
			outputContains: "Apply (y/yes)? yes",
		},
		{
			name:           "UserAcceptsWithY",
			stdin:          "y\n",
			expectedResult: true,
		},
		{
			name:           "UserAcceptsWithYes",
			stdin:          "yes\n",
			expectedResult: true,
		},
		{
			name:           "UserDeclines",
			stdin:          "n\n",
			expectedResult: false,
			outputContains: "Aborting apply.",
		},
		{
			name:        "JSONModeRequiresAutoConfirm",
			json:        true,
			autoConfirm: false,
			expectedErr: "must bypass prompts when using JSON output",
		},
		{
			name:           "JSONModeWithAutoConfirm",
			json:           true,
			autoConfirm:    true,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			pr := newTestPrompter(&buf, tt.stdin, tt.json, tt.autoConfirm)

			yes, err := pr.PromptApply(oldSpec, newSpec, false)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, yes)
			}
			if tt.outputContains != "" {
				assert.Contains(t, buf.String(), tt.outputContains)
			}
		})
	}
}

// TestPromptApply_JSONModeSkipsDiff verifies that in JSON mode the diff is
// not emitted (text diffs are suppressed in JSON output mode).
func TestPromptApply_JSONModeSkipsDiff(t *testing.T) {
	oldSpec := &identityv1.ApiKeySpec{DisplayName: "before"}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "after"}

	var buf bytes.Buffer
	pr := newTestPrompter(&buf, "", true, true)

	yes, err := pr.PromptApply(oldSpec, newSpec, false)
	require.NoError(t, err)
	assert.True(t, yes)

	out := buf.String()
	assert.NotContains(t, out, "before")
	assert.NotContains(t, out, "after")
}

// TestPromptApply_PrintsDiff verifies that in text mode the diff is emitted
// containing the before and after display names.
func TestPromptApply_PrintsDiff(t *testing.T) {
	oldSpec := &identityv1.ApiKeySpec{DisplayName: "before"}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "after"}

	var buf bytes.Buffer
	pr := newTestPrompter(&buf, "", false, true)

	yes, err := pr.PromptApply(oldSpec, newSpec, false)
	require.NoError(t, err)
	assert.True(t, yes)

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

	_, err := pr.PromptApply(spec, spec, true)
	require.NoError(t, err)
}
