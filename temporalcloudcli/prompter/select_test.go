package prompter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/temporalio/cloud-cli/temporalcloudcli/prompter"
)

func TestSelect(t *testing.T) {
	options := []string{"Google", "Microsoft", "Email + Password"}

	tests := []struct {
		name          string
		input         string
		expectedIndex int
		expectedErr   error
	}{
		{name: "enter picks the first option", input: "\r", expectedIndex: 0},
		{name: "down arrow", input: "\x1b[B\r", expectedIndex: 1},
		{name: "application mode down arrow", input: "\x1bOB\r", expectedIndex: 1},
		{name: "down twice", input: "\x1b[B\x1b[B\r", expectedIndex: 2},
		{name: "down wraps to the top", input: "\x1b[B\x1b[B\x1b[B\r", expectedIndex: 0},
		{name: "up wraps to the bottom", input: "\x1b[A\r", expectedIndex: 2},
		{name: "j and k", input: "jjk\r", expectedIndex: 1},
		{name: "other keys are ignored", input: "x\x1b[C\n", expectedIndex: 0},
		{name: "ctrl-c cancels", input: "\x1b[B\x03", expectedErr: prompter.ErrSelectCanceled},
		{name: "q cancels", input: "q", expectedErr: prompter.ErrSelectCanceled},
		{name: "eof cancels", input: "\x1b[B", expectedErr: prompter.ErrSelectCanceled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer

			index, err := prompter.Select(strings.NewReader(tt.input), &out, "Sign up with:", options)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedIndex, index)
			assert.Contains(t, out.String(), "Sign up with: "+options[tt.expectedIndex])
		})
	}
}

func TestSelect_NoOptions(t *testing.T) {
	_, err := prompter.Select(strings.NewReader("\r"), &bytes.Buffer{}, "Pick:", nil)

	require.EqualError(t, err, "no options to select from")
}
