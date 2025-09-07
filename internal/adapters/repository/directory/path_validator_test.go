// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDirectoryPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		// Unix/Linux critical paths
		{
			name:        "root directory",
			path:        "/",
			shouldError: true,
		},
		{
			name:        "etc directory",
			path:        "/etc",
			shouldError: true,
		},
		{
			name:        "etc subdirectory",
			path:        "/etc/nginx",
			shouldError: true,
		},
		{
			name:        "usr directory",
			path:        "/usr",
			shouldError: true,
		},
		{
			name:        "bin directory",
			path:        "/bin",
			shouldError: true,
		},
		{
			name:        "var directory",
			path:        "/var",
			shouldError: true,
		},
		{
			name:        "home root directory",
			path:        "/home",
			shouldError: true,
		},
		{
			name:        "user home subdirectory is allowed",
			path:        "/home/username/backups",
			shouldError: false,
		},
		{
			name:        "opt directory",
			path:        "/opt",
			shouldError: true,
		},
		{
			name:        "mnt directory",
			path:        "/mnt",
			shouldError: true,
		},
		{
			name:        "media directory",
			path:        "/media",
			shouldError: true,
		},
		{
			name:        "srv directory",
			path:        "/srv",
			shouldError: true,
		},

		// Safe paths
		{
			name:        "normal backup directory",
			path:        "/backup/repos",
			shouldError: false,
		},
		{
			name:        "tmp subdirectory",
			path:        "/tmp/backup",
			shouldError: false,
		},
		{
			name:        "data directory",
			path:        "/data/git-backups",
			shouldError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateDirectoryPath(test.path)

			if test.shouldError {
				require.Error(t, err, "Expected error for dangerous path: %s", test.path)
				assert.Contains(t, err.Error(), "refusing")
			} else {
				require.NoError(t, err, "Expected no error for safe path: %s", test.path)
			}
		})
	}
}

func TestValidateDirectoryPath_EdgeCases(t *testing.T) {
	t.Parallel()

	// Test relative path resolution
	t.Run("relative paths", func(t *testing.T) {
		t.Parallel()
		// These will resolve to absolute paths, so the safety check
		// will depend on where they resolve to
		err := validateDirectoryPath("./backup")
		// Should work unless we're running tests in a system directory
		assert.NoError(t, err)
	})

	t.Run("parent directory traversal", func(t *testing.T) {
		t.Parallel()
		// This should resolve and then be checked
		err := validateDirectoryPath("/home/user/../..")
		// This resolves to "/" which should be blocked
		assert.Error(t, err)
	})
}
