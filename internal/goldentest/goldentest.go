// Package goldentest provides shared infrastructure for golden-file tests.
package goldentest

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// update denotes if the golden file contents should be updated or not.
//
// The -update-golden flag is registered when this package is
// imported. Imports happen only from _test.go files, so the flag never lands
// in production binaries.
var update = flag.Bool("update-golden", false, "update golden files")

// AssertJSON compares got against the JSON golden file at path, ignoring key
// ordering. When -update-golden is set, got is written to the path.
func AssertJSON(t *testing.T, path string, got []byte) {
	t.Helper()
	if *update {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, got, 0o644))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.JSONEq(t, string(want), string(got), "golden mismatch at %s; rerun with -update-golden", path)
}
