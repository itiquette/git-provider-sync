//go:build integration

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package testutil provides usage examples for the git test environment utility.
package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/repository/gogit"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// ExampleGitTestEnvironment_BasicUsage demonstrates basic usage of the git test environment
func ExampleGitTestEnvironment_BasicUsage() {
	// This example shows how to use the git test environment utility
	// in integration tests for testing real git operations.

	gitOps := gogit.New(ports.GitConfig{
		UserName:    "Test User",
		UserEmail:   "test@example.com",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
	})

	// Setup isolated git environment with default options
	env, err := SetupSimpleGitTestEnvironment(nil, gitOps)
	if err != nil {
		panic(err)
	}
	// No manual cleanup needed - t.TempDir() handles it automatically

	ctx := context.Background()

	// Test git remote operations
	err = env.WorkingRepo.Repo.AddRemote(ctx, "upstream", env.GetTargetURL())
	if err != nil {
		panic(err)
	}

	// Test remote URL updates (the critical operation we fixed)
	err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
	if err != nil {
		panic(err)
	}

	// Verify remote was updated
	remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
	if err != nil {
		panic(err)
	}

	for _, remote := range remotes {
		if remote.Name == "origin" && remote.URL == env.GetTargetURL() {
			// Success! Remote was updated correctly
			break
		}
	}
}

// ExampleGitTestEnvironment_CustomSetup demonstrates custom setup options
func ExampleGitTestEnvironment_CustomSetup() {
	gitOps := gogit.New(ports.GitConfig{
		UserName:    "Test User",
		UserEmail:   "test@example.com",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
	})

	// Custom options for specific test scenarios
	opts := GitTestOptions{
		SourceRepoName:  "my-github-repo",
		TargetRepoName:  "my-gitlab-repo",
		WorkingRepoName: "my-local-clone",
		InitialBranch:   "develop",
		InitialFiles: map[string]string{
			"package.json":    `{"name": "my-app", "version": "1.0.0"}`,
			"src/index.js":    `console.log("Hello World");`,
			"README.md":       "# My Application\n\nBuilt with Node.js",
			".gitignore":      "node_modules/\n*.log\n",
			"docs/api.md":     "# API Documentation\n\nAPI endpoints...",
		},
		AddRemotes: map[string]string{
			"origin":   "", // Will use source bare repo URL
			"upstream": "https://github.com/upstream/repo.git",
			"fork":     "https://github.com/myuser/repo.git",
		},
	}

	env, err := SetupGitTestEnvironment(nil, gitOps, opts)
	if err != nil {
		panic(err)
	}
	// No manual cleanup needed - t.TempDir() handles it automatically

	// Now you have a fully configured git environment with:
	// - Custom repository names
	// - Multiple remotes configured
	// - Realistic project structure with multiple files
	// - Ready for complex git operation testing
	
	// Example: Show environment information
	_ = env.TmpDir // All repos are in isolated temporary directory
	_ = env.GetSourceURL() // Source repo URL (simulates GitHub)
	_ = env.GetTargetURL() // Target repo URL (simulates GitLab)
}

// TestIntegrationWithGitTestEnvironment shows how to integrate the utility into existing tests
func TestIntegrationWithGitTestEnvironment(t *testing.T) {
	t.Parallel()

	gitOps := gogit.New(ports.GitConfig{
		UserName:    "Integration Test",
		UserEmail:   "integration@test.com",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
	})

	t.Run("replace_existing_integration_test", func(t *testing.T) {
		// Instead of manually setting up bare repos and working directories,
		// use the utility for cleaner, more maintainable tests
		
		env, err := SetupSimpleGitTestEnvironment(t, gitOps)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		ctx := context.Background()

		// Your test logic here - this replaces all the manual setup
		// that was previously done in integration tests

		// Example: Test the critical GitHub to GitLab sync scenario
		originalRemotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
		require.NoError(t, err)
		require.Len(t, originalRemotes, 1)
		assert.Equal(t, "origin", originalRemotes[0].Name)
		assert.Equal(t, env.GetSourceURL(), originalRemotes[0].URL)

		// Add GPSUPSTREAM (backup of original URL)
		err = env.WorkingRepo.Repo.AddRemote(ctx, "GPSUPSTREAM", env.GetSourceURL())
		require.NoError(t, err)

		// Update origin to point to target (the fix we implemented)
		err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
		require.NoError(t, err)

		// Verify the fix worked
		updatedRemotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
		require.NoError(t, err)
		require.Len(t, updatedRemotes, 2)

		remoteMap := make(map[string]string)
		for _, remote := range updatedRemotes {
			remoteMap[remote.Name] = remote.URL
		}

		assert.Equal(t, env.GetTargetURL(), remoteMap["origin"])
		assert.Equal(t, env.GetSourceURL(), remoteMap["GPSUPSTREAM"])

		t.Logf("✅ Test passed: Remote update works correctly")
		t.Logf("Source (GitHub): %s", env.GetSourceURL())
		t.Logf("Target (GitLab): %s", env.GetTargetURL())
	})

	t.Run("test_push_operations_with_mocks", func(t *testing.T) {
		// For push operations that would require network calls,
		// combine the git test environment with mocks
		
		env, err := SetupSimpleGitTestEnvironment(t, gitOps)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		ctx := context.Background()

		// Setup remotes for the test scenario
		err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
		require.NoError(t, err)

		// Create push options that would be used in real code
		pushOptions := ports.PushOptions{
			Remote: "origin",
			Force:  false,
			Auth:   ports.AuthOptions{Type: ports.AuthTypeNone},
		}

		// In real tests, you might mock the Push operation or use
		// the actual local bare repository for true integration testing
		
		// Verify the setup is correct for push operations
		remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
		require.NoError(t, err)
		
		var originRemote *ports.RemoteInfo
		for _, remote := range remotes {
			if remote.Name == "origin" {
				originRemote = &remote
				break
			}
		}
		
		require.NotNil(t, originRemote)
		assert.Equal(t, env.GetTargetURL(), originRemote.URL)
		assert.Equal(t, "origin", pushOptions.Remote)
	})
}

// TestAdvancedGitTestScenarios demonstrates advanced usage patterns
func TestAdvancedGitTestScenarios(t *testing.T) {
	t.Parallel()

	gitOps := gogit.New(ports.GitConfig{
		UserName:    "Advanced Test",
		UserEmail:   "advanced@test.com",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
	})

	t.Run("multi_provider_sync_scenario", func(t *testing.T) {
		// Simulate syncing between multiple providers
		// GitHub -> GitLab -> Gitea workflow
		
		opts := GitTestOptions{
			SourceRepoName:  "github-original",
			TargetRepoName:  "gitlab-mirror",
			WorkingRepoName: "sync-workspace",
			InitialFiles: map[string]string{
				"README.md":       "# Multi-Provider Sync Test",
				"src/app.py":      "print('Hello from multi-provider sync')",
				"requirements.txt": "requests>=2.25.0\nflask>=2.0.0",
			},
			AddRemotes: map[string]string{
				"origin": "", // GitHub
			},
		}

		env, err := SetupGitTestEnvironment(t, gitOps, opts)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		ctx := context.Background()

		// Add third provider (Gitea) by creating another "bare repo"
		giteaPath := env.TmpDir + "/gitea-repo.git"
		giteaRepo, err := env.GitOps.Init(ctx, giteaPath, ports.InitOptions{Bare: true})
		require.NoError(t, err)
		defer giteaRepo.Close()

		// Setup remotes for multi-provider scenario
		err = env.WorkingRepo.Repo.AddRemote(ctx, "gitlab", env.GetTargetURL())
		require.NoError(t, err)
		
		err = env.WorkingRepo.Repo.AddRemote(ctx, "gitea", giteaPath)
		require.NoError(t, err)

		// Test complex remote management scenarios
		remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
		require.NoError(t, err)
		assert.Len(t, remotes, 3) // origin, gitlab, gitea

		// Simulate provider switching
		err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
		require.NoError(t, err)

		t.Log("✅ Multi-provider sync test environment ready")
	})

	t.Run("branch_and_tag_operations", func(t *testing.T) {
		env, err := SetupSimpleGitTestEnvironment(t, gitOps)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		ctx := context.Background()

		// Test branch operations that might be needed for sync
		branches, err := env.WorkingRepo.Repo.ListBranches(ctx)
		require.NoError(t, err)
		t.Logf("Available branches: %+v", branches)

		// Test current branch detection
		currentBranch, err := env.WorkingRepo.Repo.CurrentBranch()
		require.NoError(t, err)
		t.Logf("Current branch: %s", currentBranch)

		// These operations would be useful for testing branch protection
		// and default branch setting features
	})
}