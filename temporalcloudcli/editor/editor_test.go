package editor_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/editor"
)

// writeScript creates a temporary executable shell script and returns its path.
func writeScript(t *testing.T, body string) string {
	t.Helper()
	f, err := os.CreateTemp("", "editor-test-*.sh")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	_, err = fmt.Fprintf(f, "#!/bin/sh\n%s\n", body)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.NoError(t, os.Chmod(f.Name(), 0o755))
	return f.Name()
}

// setEditor sets VISUAL to empty and EDITOR to the given value for the duration of the test.
func setEditor(t *testing.T, editorPath string) {
	t.Helper()
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", editorPath)
}

func TestEditProto_Success(t *testing.T) {
	script := writeScript(t, `printf '{"displayName":"updated-key"}' > "$1"`)
	setEditor(t, script)

	existing := &identityv1.ApiKeySpec{DisplayName: "original-key"}
	result, err := editor.NewEditor().EditProto(existing)

	require.NoError(t, err)
	updated, ok := result.(*identityv1.ApiKeySpec)
	require.True(t, ok)
	assert.Equal(t, "updated-key", updated.DisplayName)
	// original must not be mutated
	assert.Equal(t, "original-key", existing.DisplayName)
}

func TestEditProto_NoChanges(t *testing.T) {
	// script exits without touching the file
	script := writeScript(t, `# no-op`)
	setEditor(t, script)

	existing := &identityv1.ApiKeySpec{DisplayName: "original-key"}
	_, err := editor.NewEditor().EditProto(existing)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no changes detected")
}

func TestEditProto_EditorFails(t *testing.T) {
	script := writeScript(t, `exit 1`)
	setEditor(t, script)

	existing := &identityv1.ApiKeySpec{DisplayName: "original-key"}
	_, err := editor.NewEditor().EditProto(existing)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error executing")
}

func TestEditProto_EditorNotFound(t *testing.T) {
	setEditor(t, "/nonexistent/editor-binary")

	existing := &identityv1.ApiKeySpec{DisplayName: "original-key"}
	_, err := editor.NewEditor().EditProto(existing)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error executing")
}

func TestEditProto_InvalidJSON(t *testing.T) {
	script := writeScript(t, `printf 'not valid json' > "$1"`)
	setEditor(t, script)

	existing := &identityv1.ApiKeySpec{DisplayName: "original-key"}
	_, err := editor.NewEditor().EditProto(existing)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to convert updated json to object")
}

func TestEditProto_VISUALTakesPrecedenceOverEDITOR(t *testing.T) {
	visualScript := writeScript(t, `printf '{"displayName":"from-visual"}' > "$1"`)
	editorScript := writeScript(t, `exit 1`)
	t.Setenv("VISUAL", visualScript)
	t.Setenv("EDITOR", editorScript)

	existing := &identityv1.ApiKeySpec{DisplayName: "original-key"}
	result, err := editor.NewEditor().EditProto(existing)

	require.NoError(t, err)
	updated, ok := result.(*identityv1.ApiKeySpec)
	require.True(t, ok)
	assert.Equal(t, "from-visual", updated.DisplayName)
}

func TestEditProto_DeprecatedFieldsStrippedBeforeEditing(t *testing.T) {
	// The script exits 1 if any JSON key ending with "Deprecated" is found in the
	// editor file, which would indicate the deprecated-field stripping did not run.
	// On success it writes back a minimal valid modification.
	script := writeScript(t, `grep -q 'Deprecated' "$1" && exit 1; printf '{"id":"key-1","spec":{"displayName":"updated"}}' > "$1"`)
	setEditor(t, script)

	existing := &identityv1.ApiKey{
		Id:              "key-1",
		StateDeprecated: "active", //nolint:staticcheck
		Spec:            &identityv1.ApiKeySpec{DisplayName: "original"},
	}
	result, err := editor.NewEditor().EditProto(existing)

	require.NoError(t, err)
	updated, ok := result.(*identityv1.ApiKey)
	require.True(t, ok)
	assert.Equal(t, "updated", updated.Spec.DisplayName)
}

func TestEditProto_OptionDeprecatedFieldsStrippedBeforeEditing(t *testing.T) {
	// `regions` and `customSearchAttributes` on NamespaceSpec are marked
	// [deprecated = true] but their JSON names don't end in "Deprecated".
	// The script exits 1 if either appears in the editor file.
	script := writeScript(t, `(grep -q '"regions"' "$1" || grep -q '"customSearchAttributes"' "$1") && exit 1; printf '{"name":"my-ns","retentionDays":14}' > "$1"`)
	setEditor(t, script)

	existing := &namespacev1.NamespaceSpec{
		Name:          "my-ns",
		RetentionDays: 7,
	}
	result, err := editor.NewEditor().EditProto(existing)

	require.NoError(t, err)
	updated, ok := result.(*namespacev1.NamespaceSpec)
	require.True(t, ok)
	assert.Equal(t, int32(14), updated.RetentionDays)
}
