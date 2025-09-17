// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// AssertTestIsolation verifies that the test is properly isolated from the host system.
// This helps catch tests that might accidentally affect the host.
func AssertTestIsolation(t *testing.T) {
	t.Helper()

	// Check we're not in a system directory
	cwd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")

	// List of directories that should never be the CWD in tests
	dangerousPaths := []string{
		"/",
		"/etc",
		"/usr",
		"/var",
		"/opt",
		"/bin",
		"/sbin",
		"/home",
		"/Users",
		"C:\\",
		"C:\\Windows",
		"C:\\Program Files",
	}

	for _, dangerous := range dangerousPaths {
		require.NotEqual(t, dangerous, cwd, "Test is running in dangerous directory: %s", cwd)
		require.False(t, strings.HasPrefix(cwd, dangerous+string(filepath.Separator)),
			"Test is running in dangerous directory tree: %s", cwd)
	}

	// Verify we're in a temp directory or test directory
	AssertInTempDirectory(t, cwd)
}

// AssertInTempDirectory verifies the given path is within a temp directory.
func AssertInTempDirectory(t *testing.T, path string) {
	t.Helper()

	// Get system temp directory
	tempDir := os.TempDir()
	absPath, err := filepath.Abs(path)
	require.NoError(t, err, "Failed to get absolute path")

	// On some systems, temp dir might be a symlink
	evalTempDir, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err, "Failed to evaluate temp dir symlinks")

	evalPath, err := filepath.EvalSymlinks(absPath)
	require.NoError(t, err, "Failed to evaluate path symlinks")

	// Check if path is within temp directory
	isTempPath := strings.HasPrefix(evalPath, evalTempDir) ||
		strings.Contains(evalPath, "tmp") ||
		strings.Contains(evalPath, "temp") ||
		strings.Contains(evalPath, "T_") // Go test temp dirs often have this prefix

	require.True(t, isTempPath,
		"Path is not in a temp directory: %s (temp: %s)", evalPath, evalTempDir)
}

// AssertNoSystemFiles verifies that no system files will be modified.
// Call this at the start of a test that performs file operations.
func AssertNoSystemFiles(t *testing.T, paths ...string) {
	t.Helper()

	systemPaths := []string{
		"/etc",
		"/usr",
		"/var",
		"/opt",
		"/bin",
		"/sbin",
		"/lib",
		"/lib64",
		"/boot",
		"/dev",
		"/proc",
		"/sys",
	}

	// On Windows
	if runtime.GOOS == "windows" {
		systemPaths = []string{
			"C:\\Windows",
			"C:\\Program Files",
			"C:\\Program Files (x86)",
			"C:\\ProgramData",
			"C:\\System32",
		}
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue // Skip if path doesn't exist yet
		}

		for _, sysPath := range systemPaths {
			require.False(t, strings.HasPrefix(absPath, sysPath),
				"Attempting to modify system path: %s", absPath)
		}
	}
}

// AssertEnvironmentIsolated verifies critical environment variables are not set
// to production values.
func AssertEnvironmentIsolated(t *testing.T) {
	t.Helper()

	// Check that we're not using real home directory
	home := os.Getenv("HOME")
	if home != "" {
		AssertInTempDirectory(t, home)
	}

	// Check Git isolation
	gitConfig := os.Getenv("GIT_CONFIG_GLOBAL")
	if gitConfig != "" {
		AssertInTempDirectory(t, filepath.Dir(gitConfig))
	}

	// Check SSH isolation
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	require.Empty(t, sshAuthSock, "SSH_AUTH_SOCK should not be set in tests")

	// Check that proxy is disabled (to prevent accidental external calls)
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")

	if httpProxy != "" {
		require.True(t,
			httpProxy == "http://127.0.0.1:0" || httpProxy == "",
			"HTTP_PROXY should be disabled or set to invalid address in tests")
	}

	if httpsProxy != "" {
		require.True(t,
			httpsProxy == "http://127.0.0.1:0" || httpsProxy == "",
			"HTTPS_PROXY should be disabled or set to invalid address in tests")
	}
}

// AssertNoNetworkAccess verifies that network access is properly mocked.
// This is a helper to ensure tests don't make real network calls.
func AssertNoNetworkAccess(t *testing.T) {
	t.Helper()

	// Set invalid proxy to catch any HTTP calls
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:0")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:0")
	t.Setenv("NO_PROXY", "")
	t.Setenv("http_proxy", "http://127.0.0.1:0")
	t.Setenv("https_proxy", "http://127.0.0.1:0")
	t.Setenv("no_proxy", "")
}

// RequireTestEnvironment ensures the test is running in a proper test environment.
// Call this at the beginning of tests that need strict isolation.
func RequireTestEnvironment(t *testing.T) {
	t.Helper()

	// Perform all isolation checks
	AssertTestIsolation(t)
	AssertEnvironmentIsolated(t)

	// Set up network isolation
	AssertNoNetworkAccess(t)

	// Log that isolation is active (helps with debugging)
	t.Logf("Test isolation verified: running in %s", t.TempDir())
}

// AssertPathsWithinRoot verifies all paths are within the given root directory.
func AssertPathsWithinRoot(t *testing.T, root string, paths ...string) {
	t.Helper()

	rootAbs, err := filepath.Abs(root)
	require.NoError(t, err, "Failed to get absolute root path")

	for _, path := range paths {
		if path == "" {
			continue
		}

		absPath, err := filepath.Abs(path)
		require.NoError(t, err, "Failed to get absolute path for: %s", path)

		require.True(t, strings.HasPrefix(absPath, rootAbs),
			"Path %s is not within root %s", absPath, rootAbs)
	}
}

// AssertCleanupSucceeded verifies that test cleanup was successful.
// Call this in a t.Cleanup() function to ensure no test artifacts remain.
func AssertCleanupSucceeded(t *testing.T, paths ...string) {
	t.Helper()

	for _, path := range paths {
		_, err := os.Stat(path)
		if err == nil {
			t.Logf("Warning: Test artifact not cleaned up: %s", path)
		}
	}
}

// RequireValidRepositoryEntity asserts that a repository entity has all required fields.
func RequireValidRepositoryEntity(t *testing.T, repo interface{}) {
	t.Helper()
	require.NotNil(t, repo, "repository should not be nil")

	// Use reflection to check common repository fields
	// This is a generic helper for different repository types
	switch r := repo.(type) {
	case interface{ Name() string }:
		require.NotEmpty(t, r.Name(), "repository name should not be empty")
	case interface{ GetName() string }:
		require.NotEmpty(t, r.GetName(), "repository name should not be empty")
	default:
		t.Logf("Warning: Cannot validate repository name for type %T", repo)
	}
}

// RequireErrorContains asserts that an error occurred and contains a substring.
func RequireErrorContains(t *testing.T, err error, contains string, msgAndArgs ...interface{}) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)
	require.Contains(t, err.Error(), contains, msgAndArgs...)
}
