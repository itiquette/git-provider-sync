// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

//go:build integration

package integrationtest

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/repository/gogit"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/integrationtest/testutil"
)

// Integration tests with git test environment utility
func TestEnhancedGitOperationsIntegration(t *testing.T) {
	// Isolate Git environment from host system
	// Note: Cannot use t.Parallel() when using t.Setenv in IsolateGitEnvironment
	testutil.IsolateGitEnvironment(t)

	gitOps := gogit.New(ports.GitConfig{
		UserName:    "integration-test",
		UserEmail:   "test@git-provider-sync.local",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
	})

	t.Run("complete_repository_lifecycle", func(t *testing.T) {
		testCompleteRepositoryLifecycle(t, gitOps)
	})

	t.Run("large_repository_operations", func(t *testing.T) {
		testLargeRepositoryOperations(t, gitOps)
	})

	t.Run("remote_management_comprehensive", func(t *testing.T) {
		testRemoteManagementComprehensive(t, gitOps)
	})
}

// TestCompleteRepositoryLifecycle tests the complete lifecycle of a git repository
func testCompleteRepositoryLifecycle(t *testing.T, gitOps ports.GitOperations) {
	ctx := context.Background()

	// Use safe git test environment utility
	env, err := testutil.SetupGitTestEnvironment(t, gitOps, testutil.GitTestOptions{
		SourceRepoName:  "lifecycle-source",
		TargetRepoName:  "lifecycle-target",
		WorkingRepoName: "lifecycle-repo",
		InitialFiles: map[string]string{
			"README.md":          "# Lifecycle Test Repository\\n\\nTesting complete repository lifecycle",
			"src/main.go":        "package main\\n\\nfunc main() {\\n\\tprintln(\\\"Lifecycle test!\\\")\\n}",
			"docs/setup.md":      "# Setup Instructions\\n\\nHow to set up the project",
			"tests/main_test.go": "package main\\n\\nimport \\\"testing\\\"\\n\\nfunc TestMain(t *testing.T) {\\n\\tt.Log(\\\"test\\\")\\n}",
		},
		AddRemotes: map[string]string{
			"origin": "", // Will use source bare repo
		},
	})
	require.NoError(t, err)

	// Verify repository properties
	assert.Equal(t, "lifecycle-repo", env.WorkingRepo.Repo.Name())
	assert.Contains(t, env.WorkingRepo.Repo.Path(), "lifecycle-repo")
	assert.False(t, env.WorkingRepo.Repo.IsBare())

	// Step 2: Verify origin remote (configured by test utility)
	remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes")
	require.Len(t, remotes, 1, "Should have one remote")
	assert.Equal(t, "origin", remotes[0].Name)
	assert.Equal(t, env.GetSourceURL(), remotes[0].URL)

	// Step 3: Add GPSUPSTREAM remote (simulating backup of original GitHub URL)
	githubURL := "https://github.com/test/lifecycle-repo.git"
	err = env.WorkingRepo.Repo.AddRemote(ctx, "GPSUPSTREAM", githubURL)
	require.NoError(t, err, "Should add GPSUPSTREAM remote")

	// Verify GPSUPSTREAM remote was added
	remotes, err = env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes")
	require.Len(t, remotes, 2, "Should have two remotes")

	var gpsUpstreamRemote *ports.RemoteInfo
	for _, remote := range remotes {
		if remote.Name == "GPSUPSTREAM" {
			gpsUpstreamRemote = &remote
			break
		}
	}
	require.NotNil(t, gpsUpstreamRemote, "Should find GPSUPSTREAM remote")
	assert.Equal(t, githubURL, gpsUpstreamRemote.URL)

	// Step 4: Test repository status with multiple files
	status, err := env.WorkingRepo.Repo.Status(ctx)
	require.NoError(t, err, "Should get repository status")
	t.Logf("Repository has %d added files", len(status.Added))

	// Step 5: Test remote operations - update origin to point to target (simulating GitHub → GitLab sync)
	err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
	require.NoError(t, err, "Should update remote URL")

	// Verify remote was updated (origin now points to target, simulating GitLab)
	remotes, err = env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes after update")

	var updatedOrigin *ports.RemoteInfo
	for _, remote := range remotes {
		if remote.Name == "origin" {
			updatedOrigin = &remote
			break
		}
	}
	require.NotNil(t, updatedOrigin, "Should find updated origin remote")
	assert.Equal(t, env.GetTargetURL(), updatedOrigin.URL, "Remote URL should point to target")

	t.Logf("✅ Repository lifecycle test completed: origin updated from source (%s) to target (%s)",
		env.GetSourceURL(), env.GetTargetURL())
}

// TestLargeRepositoryOperations tests operations on repositories with multiple files and commits
func testLargeRepositoryOperations(t *testing.T, gitOps ports.GitOperations) {
	ctx := context.Background()

	// Create large repository with many files using safe test utility
	largeFileSet := make(map[string]string)

	// Generate multiple directories with multiple files each
	dirs := []string{"src", "docs", "tests", "config", "assets", "scripts"}
	for _, dir := range dirs {
		for i := 0; i < 5; i++ {
			fileName := fmt.Sprintf("%s/file%d.txt", dir, i)
			content := fmt.Sprintf("Content for %s\\nGenerated at test time\\nDirectory: %s\\nFile: %d", fileName, dir, i)
			largeFileSet[fileName] = content
		}
	}

	env, err := testutil.SetupGitTestEnvironment(t, gitOps, testutil.GitTestOptions{
		SourceRepoName:  "large-source",
		TargetRepoName:  "large-target",
		WorkingRepoName: "large-repo",
		InitialFiles:    largeFileSet,
		AddRemotes: map[string]string{
			"origin": "", // Will use source bare repo
		},
	})
	require.NoError(t, err)

	// Test operations on repository with many files (safely created by test utility)
	totalFiles := len(dirs) * 5 // 6 dirs * 5 files each = 30 files
	t.Logf("Testing large repository operations with %d files in %d directories", totalFiles, len(dirs))

	// Test initial remote configuration
	remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes")
	require.Len(t, remotes, 1, "Should have origin remote")
	assert.Equal(t, "origin", remotes[0].Name)
	assert.Equal(t, env.GetSourceURL(), remotes[0].URL)

	// Add backup remote (simulating additional provider)
	backupURL := "https://github.com/backup/large-repo.git"
	err = env.WorkingRepo.Repo.AddRemote(ctx, "backup", backupURL)
	require.NoError(t, err, "Should add backup remote")

	// Test multiple remotes
	remotes, err = env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes")
	require.Len(t, remotes, 2, "Should have two remotes")

	// Test remote URL update (simulating GitHub → GitLab migration)
	err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
	require.NoError(t, err, "Should update remote URL")

	// Verify remote was updated
	remotes, err = env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes after update")

	var updatedOrigin *ports.RemoteInfo
	for _, remote := range remotes {
		if remote.Name == "origin" {
			updatedOrigin = &remote
			break
		}
	}
	require.NotNil(t, updatedOrigin, "Should find updated origin remote")
	assert.Equal(t, env.GetTargetURL(), updatedOrigin.URL, "Remote URL should point to target")

	// Test repository status with many files
	status, err := env.WorkingRepo.Repo.Status(ctx)
	require.NoError(t, err, "Should get repository status")
	t.Logf("✅ Large repository test completed: %d files tested, origin updated to target", totalFiles)
	t.Logf("Repository status clean: %t", status.IsClean)
}

// TestRemoteManagementtests remote management scenarios
func testRemoteManagementComprehensive(t *testing.T, gitOps ports.GitOperations) {
	ctx := context.Background()

	// Use safe test environment utility for remote testing
	env, err := testutil.SetupGitTestEnvironment(t, gitOps, testutil.GitTestOptions{
		SourceRepoName:  "remote-source",
		TargetRepoName:  "remote-target",
		WorkingRepoName: "remote-mgmt-repo",
		InitialFiles: map[string]string{
			"README.md":     "# Remote Management Test\\n\\nTesting remote scenarios",
			"src/config.go": "package main\\n\\n// Remote management configuration",
		},
		AddRemotes: map[string]string{
			"origin": "", // Will use source bare repo
		},
	})
	require.NoError(t, err)

	// Test remote management scenarios
	remoteConfigs := []struct {
		name string
		url  string
	}{
		{"upstream", "https://github.com/upstream/remote-mgmt-repo.git"},
		{"fork", "https://github.com/fork/remote-mgmt-repo.git"},
		{"backup", "https://github.com/backup/remote-mgmt-repo.git"},
		{"GPSUPSTREAM", "https://github.com/gps/remote-mgmt-repo.git"},
	}

	// Add multiple remotes (origin already exists from setup)
	for _, config := range remoteConfigs {
		err := env.WorkingRepo.Repo.AddRemote(ctx, config.name, config.url)
		require.NoError(t, err, "Should add remote %s", config.name)
	}

	// Verify all remotes were added (origin + 4 additional = 5 total)
	remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes")
	require.Len(t, remotes, len(remoteConfigs)+1, "Should have origin + additional remotes")

	// Verify each remote exists with correct URL
	remoteMap := make(map[string]string)
	for _, remote := range remotes {
		remoteMap[remote.Name] = remote.URL
	}

	// Check origin points to source initially
	assert.Equal(t, env.GetSourceURL(), remoteMap["origin"], "Origin should point to source")

	for _, config := range remoteConfigs {
		url, exists := remoteMap[config.name]
		require.True(t, exists, "Remote %s should exist", config.name)
		assert.Equal(t, config.url, url, "Remote %s should have correct URL", config.name)
	}

	// Test critical remote update (origin → target, simulating GitHub → GitLab sync)
	err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
	require.NoError(t, err, "Should update origin remote to target")

	// Verify critical update
	remotes, err = env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes after update")

	found := false
	for _, remote := range remotes {
		if remote.Name == "origin" {
			assert.Equal(t, env.GetTargetURL(), remote.URL, "Origin should now point to target")
			found = true
			break
		}
	}
	require.True(t, found, "Should find updated origin remote")

	// Test remote removal
	err = env.WorkingRepo.Repo.RemoveRemote(ctx, "fork")
	require.NoError(t, err, "Should remove fork remote")

	// Verify removal
	remotes, err = env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes after removal")
	require.Len(t, remotes, len(remoteConfigs), "Should have origin + 3 remaining remotes")

	for _, remote := range remotes {
		assert.NotEqual(t, "fork", remote.Name, "Fork remote should be removed")
	}

	// Test that we can add target as a remote (simulating multi-provider setup)
	err = env.WorkingRepo.Repo.AddRemote(ctx, "target", env.GetTargetURL())
	require.NoError(t, err, "Should add target remote")

	t.Logf("✅ remote management test completed")
	t.Logf("   Source URL: %s", env.GetSourceURL())
	t.Logf("   Target URL: %s", env.GetTargetURL())
	t.Logf("   Total remotes tested: %d", len(remoteConfigs)+2) // origin, target, + others
}

// Helper functions moved to testutil.GitTestEnvironment
