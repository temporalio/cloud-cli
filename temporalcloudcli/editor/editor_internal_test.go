package editor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- stripDeprecatedFields ---

func TestStripDeprecatedFields(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string // expected JSON (semantically compared)
		expectedErr string
	}{
		{
			name:     "RemovesTopLevelDeprecatedKey",
			input:    `{"displayName":"key","stateDeprecated":"old"}`,
			expected: `{"displayName":"key"}`,
		},
		{
			name:     "PreservesNonDeprecatedFields",
			input:    `{"displayName":"key","state":"active"}`,
			expected: `{"displayName":"key","state":"active"}`,
		},
		{
			name:     "RemovesNestedDeprecatedKey",
			input:    `{"spec":{"displayName":"key","ownerTypeDeprecated":"user","ownerId":"o1"}}`,
			expected: `{"spec":{"displayName":"key","ownerId":"o1"}}`,
		},
		{
			name:     "RemovesDeprecatedKeyInArrayElement",
			input:    `{"items":[{"id":"k1","stateDeprecated":"active"},{"id":"k2"}]}`,
			expected: `{"items":[{"id":"k1"},{"id":"k2"}]}`,
		},
		{
			name:     "RemovesMultipleDeprecatedKeys",
			input:    `{"roleDeprecated":"owner","role":"admin","permissionDeprecated":"read","permission":"write"}`,
			expected: `{"role":"admin","permission":"write"}`,
		},
		{
			name:     "DoesNotRemoveKeyContainingButNotEndingWithDeprecated",
			input:    `{"deprecatedField":"value","otherField":"other"}`,
			expected: `{"deprecatedField":"value","otherField":"other"}`,
		},
		{
			name:     "EmptyObject",
			input:    `{}`,
			expected: `{}`,
		},
		{
			name:     "EmptyArray",
			input:    `{"items":[]}`,
			expected: `{"items":[]}`,
		},
		{
			name:        "InvalidJSON",
			input:       `not-valid-json`,
			expectedErr: "unable to parse json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := stripDeprecatedFields([]byte(tt.input))
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(result))
		})
	}
}
