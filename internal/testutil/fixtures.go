// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFixture provides a structured test environment with common paths and utilities.
type TestFixture struct {
	t       *testing.T
	TempDir string
	Paths   TestPaths
}

// TestPaths contains commonly used paths in tests.
type TestPaths struct {
	// Root test directory
	Root string

	// Configuration paths
	ConfigDir  string
	ConfigFile string

	// Git-related paths
	GitDir    string
	GitConfig string

	// SSH paths
	SSHDir    string
	SSHConfig string

	// Data directories
	DataDir  string
	CacheDir string
	LogDir   string

	// Work directories
	WorkDir    string
	ReposDir   string
	ArchiveDir string
}

// NewTestFixture creates a new test fixture with a complete directory structure.
func NewTestFixture(t *testing.T) *TestFixture {
	t.Helper()

	root := t.TempDir()

	fixture := &TestFixture{
		t:       t,
		TempDir: root,
		Paths: TestPaths{
			Root:       root,
			ConfigDir:  filepath.Join(root, "config"),
			ConfigFile: filepath.Join(root, "config", "gitprovidersync.yaml"),
			GitDir:     filepath.Join(root, "git"),
			GitConfig:  filepath.Join(root, "git", "config"),
			SSHDir:     filepath.Join(root, "ssh"),
			SSHConfig:  filepath.Join(root, "ssh", "config"),
			DataDir:    filepath.Join(root, "data"),
			CacheDir:   filepath.Join(root, "cache"),
			LogDir:     filepath.Join(root, "logs"),
			WorkDir:    filepath.Join(root, "work"),
			ReposDir:   filepath.Join(root, "repos"),
			ArchiveDir: filepath.Join(root, "archives"),
		},
	}

	// Create directory structure
	fixture.createDirectories()

	// Set up isolated environment
	fixture.isolateEnvironment()

	return fixture
}

// createDirectories creates the test directory structure.
func (f *TestFixture) createDirectories() {
	f.t.Helper()

	dirs := []string{
		f.Paths.ConfigDir,
		f.Paths.GitDir,
		f.Paths.SSHDir,
		f.Paths.DataDir,
		f.Paths.CacheDir,
		f.Paths.LogDir,
		f.Paths.WorkDir,
		f.Paths.ReposDir,
		f.Paths.ArchiveDir,
	}

	for _, dir := range dirs {
		require.NoError(f.t, os.MkdirAll(dir, 0750))
	}
}

// isolateEnvironment sets up environment isolation for the test.
func (f *TestFixture) isolateEnvironment() {
	f.t.Helper()

	// Set isolated home
	f.t.Setenv("HOME", f.Paths.Root)
	f.t.Setenv("USERPROFILE", f.Paths.Root)

	// Set XDG directories
	f.t.Setenv("XDG_CONFIG_HOME", f.Paths.ConfigDir)
	f.t.Setenv("XDG_DATA_HOME", f.Paths.DataDir)
	f.t.Setenv("XDG_CACHE_HOME", f.Paths.CacheDir)

	// Set temp directories
	f.t.Setenv("TMPDIR", f.TempDir)
	f.t.Setenv("TMP", f.TempDir)
	f.t.Setenv("TEMP", f.TempDir)

	// Isolate Git
	f.t.Setenv("GIT_CONFIG_GLOBAL", f.Paths.GitConfig)
	f.t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(f.Paths.GitDir, "system"))
	f.t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	// Isolate SSH
	f.t.Setenv("SSH_CONFIG", f.Paths.SSHConfig)
	f.t.Setenv("SSH_AUTH_SOCK", "")
	f.t.Setenv("SSH_AGENT_PID", "")
}

// WriteConfig writes a configuration file for testing.
func (f *TestFixture) WriteConfig(content string) string {
	f.t.Helper()

	require.NoError(f.t, os.WriteFile(f.Paths.ConfigFile, []byte(content), 0600))

	return f.Paths.ConfigFile
}

// WriteFile writes a file with the given content to the fixture.
func (f *TestFixture) WriteFile(relativePath, content string) string {
	f.t.Helper()

	fullPath := filepath.Join(f.Paths.Root, relativePath)
	dir := filepath.Dir(fullPath)

	require.NoError(f.t, os.MkdirAll(dir, 0750))
	require.NoError(f.t, os.WriteFile(fullPath, []byte(content), 0600))

	return fullPath
}

// CreateRepo creates a test repository directory.
func (f *TestFixture) CreateRepo(name string) string {
	f.t.Helper()

	repoPath := filepath.Join(f.Paths.ReposDir, name)
	require.NoError(f.t, os.MkdirAll(repoPath, 0750))

	// Create a basic git structure
	gitDir := filepath.Join(repoPath, ".git")
	require.NoError(f.t, os.MkdirAll(gitDir, 0750))

	// Write basic git config
	gitConfig := filepath.Join(gitDir, "config")
	config := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
`
	require.NoError(f.t, os.WriteFile(gitConfig, []byte(config), 0600))

	return repoPath
}

// CreateArchive creates a test archive file.
func (f *TestFixture) CreateArchive(name string, content []byte) string {
	f.t.Helper()

	archivePath := filepath.Join(f.Paths.ArchiveDir, name)
	require.NoError(f.t, os.WriteFile(archivePath, content, 0600))

	return archivePath
}

// Path returns an absolute path within the fixture.
func (f *TestFixture) Path(relativePath string) string {
	return filepath.Join(f.Paths.Root, relativePath)
}

// AssertFileExists verifies a file exists in the fixture.
func (f *TestFixture) AssertFileExists(relativePath string) {
	f.t.Helper()

	fullPath := filepath.Join(f.Paths.Root, relativePath)
	info, err := os.Stat(fullPath)
	require.NoError(f.t, err, "File should exist: %s", fullPath)
	require.False(f.t, info.IsDir(), "Path should be a file: %s", fullPath)
}

// AssertDirExists verifies a directory exists in the fixture.
func (f *TestFixture) AssertDirExists(relativePath string) {
	f.t.Helper()

	fullPath := filepath.Join(f.Paths.Root, relativePath)
	info, err := os.Stat(fullPath)
	require.NoError(f.t, err, "Directory should exist: %s", fullPath)
	require.True(f.t, info.IsDir(), "Path should be a directory: %s", fullPath)
}

// AssertFileContent verifies file content in the fixture.
func (f *TestFixture) AssertFileContent(relativePath, expectedContent string) {
	f.t.Helper()

	fullPath := filepath.Join(f.Paths.Root, relativePath)
	content, err := os.ReadFile(filepath.Clean(fullPath))
	require.NoError(f.t, err, "Failed to read file: %s", fullPath)
	require.Equal(f.t, expectedContent, string(content), "File content mismatch")
}
