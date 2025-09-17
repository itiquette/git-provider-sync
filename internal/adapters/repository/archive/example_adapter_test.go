// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package archive_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"itiquette/git-provider-sync/internal/testutil"
)

// Example of how tests should be written using our test helpers.
// No direct Afero usage, no os.WriteFile/ReadFile.

func TestArchiveOperations_WithTestHelpers(t *testing.T) {
	t.Parallel()

	// Simple one-liner to get everything needed
	test := testutil.NewTest(t)

	// Create test structure using helper
	test.Setup(map[string]string{
		"/source/file1.txt":     "content1",
		"/source/dir/file2.txt": "content2",
		"/source/.gitignore":    "*.tmp",
		"/archives/":            "", // empty = directory
	})

	// Verify structure was created
	test.FS.AssertFileExists("/source/file1.txt")
	test.FS.AssertFileContent("/source/file1.txt", "content1")
	test.FS.AssertDirExists("/archives")

	// The test would continue with actual archive operations...
}

func TestClone_WithTestHelpers(t *testing.T) {
	t.Parallel()

	test := testutil.NewTest(t)

	// Create a mock archive
	archivePath := "/archives/repo.tar.gz"
	test.FS.WriteFile(archivePath, "mock archive data")

	// Create destination
	destPath := "/dest/repo"
	test.FS.CreateDir(filepath.Dir(destPath))

	// Mock adapter would use the test filesystem
	// Real implementation would need adapter changes to support TestFS

	// Verify
	test.FS.AssertFileExists(archivePath)
}

func TestConfiguration_WithTestHelpers(t *testing.T) {
	t.Parallel()
	test := testutil.NewTest(t)

	// Use standard configs
	configPath := test.WriteConfig(testutil.Configs.GitHub)

	// Now configPath can be used with loaders
	test.FS.AssertFileExists(configPath)
}

func TestRepository_WithTestHelpers(t *testing.T) {
	t.Parallel()
	test := testutil.NewTest(t)

	// Create a git repository
	repoPath := test.CreateRepo("my-repo")

	// Add files to it
	test.FS.WriteFile(filepath.Join(repoPath, "main.go"), "package main")
	test.FS.WriteFile(filepath.Join(repoPath, "go.mod"), "module example")

	// Verify
	test.FS.AssertFileExists(filepath.Join(repoPath, ".git/config"))
	test.FS.AssertFileExists(filepath.Join(repoPath, "README.md"))
	test.FS.AssertFileExists(filepath.Join(repoPath, "main.go"))
}

func TestBatchOperations_WithTestHelpers(t *testing.T) {
	t.Parallel()
	test := testutil.NewTest(t)

	// Create multiple repos efficiently
	reposDir := "/repos"
	test.FS.CreateDir(reposDir)

	repos := []string{"repo1", "repo2", "repo3"}
	for _, name := range repos {
		repoPath := filepath.Join(reposDir, name)
		test.FS.CreateDir(repoPath)
		test.FS.WriteFile(filepath.Join(repoPath, "config.yaml"), "test: true")
	}

	// List and verify - files with config.yaml in each repo
	for _, name := range repos {
		repoPath := filepath.Join(reposDir, name)
		assert.True(t, test.FS.Exists(filepath.Join(repoPath, "config.yaml")))
	}
}

func TestIsolation_IsGuaranteed(t *testing.T) {
	t.Parallel()
	test := testutil.NewTest(t)

	// Everything happens in memory - no host interaction
	test.FS.WriteFile("/etc/passwd", "this would be bad on real system!")

	// But it's safe because we're in virtual filesystem
	content := test.FS.ReadFile("/etc/passwd")
	assert.Equal(t, "this would be bad on real system!", content)

	// This is completely isolated from the real /etc/passwd
}

// BenchmarkWithTestHelpers shows performance testing.
func BenchmarkWithTestHelpers(b *testing.B) {
	test := testutil.NewTest(b)

	// Setup once
	test.Setup(map[string]string{
		"/source/file.txt": "content",
	})

	b.ResetTimer()

	for range b.N {
		// Fast in-memory operations
		content := test.FS.ReadFile("/source/file.txt")
		test.FS.WriteFile("/dest/file.txt", content)
		test.FS.Remove("/dest/file.txt")
	}
}
