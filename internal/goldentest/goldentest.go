// Package goldentest provides shared infrastructure for golden-file tests.
//
// AIDEV-NOTE: The -update-golden flag is registered when this package is
// imported. Imports happen only from _test.go files, so the flag never lands
// in production binaries.
package goldentest

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var Update = flag.Bool("update-golden", false, "update golden files")

// AssertJSON compares got against the JSON golden file at path, ignoring key
// ordering. When -update-golden is set, got is written to path and the
// assertion is skipped.
func AssertJSON(t *testing.T, path string, got []byte) {
	t.Helper()
	if *Update {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, got, 0o644))
		return
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.JSONEq(t, string(want), string(got), "golden mismatch at %s; rerun with -update-golden", path)
}
