// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package shared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"itiquette/git-provider-sync/internal/shared"
)

func TestRemoveAllInTempDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name:        "temp directory /tmp",
			path:        "/tmp/test-dir",
			shouldError: false,
		},
		{
			name:        "OS temp directory",
			path:        filepath.Join(os.TempDir(), "test-dir"),
			shouldError: false,
		},
		{
			name:        "home directory",
			path:        "/home/user/important",
			shouldError: true,
		},
		{
			name:        "etc directory",
			path:        "/etc/config",
			shouldError: true,
		},
		{
			name:        "var directory",
			path:        "/var/log/app",
			shouldError: true,
		},
		{
			name:        "usr directory",
			path:        "/usr/local/bin",
			shouldError: true,
		},
		{
			name:        "relative path",
			path:        "./relative/path",
			shouldError: true,
		},
		{
			name:        "root directory",
			path:        "/",
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := shared.RemoveAllInTempDir(test.path)

			if test.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not in temp directory")
			} else {
				// For temp directories, we expect no error
				// (even if directory doesn't exist, os.RemoveAll succeeds)
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveAllInTempDir_ActualDirectory(t *testing.T) {
	t.Parallel()

	// Create a real temp directory
	tempDir, err := os.MkdirTemp("", "test-removeall-*")
	require.NoError(t, err)

	// Create a file in it
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0600)
	require.NoError(t, err)

	// Verify it exists
	_, err = os.Stat(tempDir)
	require.NoError(t, err)

	// Remove it using our safe function
	err = shared.RemoveAllInTempDir(tempDir)
	require.NoError(t, err)

	// Verify it's gone
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err))
}
