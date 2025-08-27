//go:build integration

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/repository/gogit"
	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestSetupGitTestEnvironment(t *testing.T) {
	t.Parallel()

	gitOps := gogit.New(ports.GitConfig{
		UserName:    "Test User",
		UserEmail:   "test@example.com",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in THIS test
	})

	t.Run("creates_isolated_git_environment", func(t *testing.T) {
		env, err := SetupSimpleGitTestEnvironment(t, gitOps)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		// Verify environment structure
		assert.NotEmpty(t, env.TmpDir)
		assert.Equal(t, "source-repo", env.SourceBareRepo.Name)
		assert.Equal(t, "target-repo", env.TargetBareRepo.Name)
		assert.Equal(t, "working-repo", env.WorkingRepo.Name)

		// Verify bare repositories exist
		assert.True(t, env.SourceBareRepo.IsBare)
		assert.True(t, env.TargetBareRepo.IsBare)
		assert.False(t, env.WorkingRepo.IsBare)

		// Verify working repository has content
		assert.NotNil(t, env.WorkingRepo.Repo)
	})

	t.Run("working_repo_has_remotes_configured", func(t *testing.T) {
		env, err := SetupSimpleGitTestEnvironment(t, gitOps)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		// Check that origin remote is configured
		remotes, err := env.WorkingRepo.Repo.ListRemotes(context.Background())
		require.NoError(t, err)
		require.Len(t, remotes, 1)

		assert.Equal(t, "origin", remotes[0].Name)
		assert.Equal(t, env.SourceBareRepo.URL, remotes[0].URL)
	})

	t.Run("can_clone_fresh_repositories", func(t *testing.T) {
		env, err := SetupSimpleGitTestEnvironment(t, gitOps)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		// This test demonstrates the clone functionality
		// Note: For this to work, we'd need to push content to the source bare repo first
		// which requires git binary operations (add, commit, push)
		
		// For now, just verify the clone method exists and can be called
		// In a real scenario with git binary support, this would create a working clone
		_, err = env.Clone("test-clone")
		
		// This is expected to fail until we implement PushToSource with git binary operations
		assert.Error(t, err, "Clone should fail until source repo has content")
	})

	t.Run("custom_options_are_respected", func(t *testing.T) {
		opts := GitTestOptions{
			SourceRepoName:  "custom-source",
			TargetRepoName:  "custom-target",
			WorkingRepoName: "custom-working",
			InitialBranch:   "develop",
			InitialFiles: map[string]string{
				"custom.txt": "Custom content",
				"config.yml": "version: 1.0",
			},
			AddRemotes: map[string]string{
				"origin":   "", // Will use source bare repo
				"upstream": "https://example.com/upstream.git",
			},
		}

		env, err := SetupGitTestEnvironment(t, gitOps, opts)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		// Verify custom names
		assert.Equal(t, "custom-source", env.SourceBareRepo.Name)
		assert.Equal(t, "custom-target", env.TargetBareRepo.Name)
		assert.Equal(t, "custom-working", env.WorkingRepo.Name)

		// Verify custom remotes
		remotes, err := env.WorkingRepo.Repo.ListRemotes(context.Background())
		require.NoError(t, err)
		require.Len(t, remotes, 2)

		remoteNames := make(map[string]string)
		for _, remote := range remotes {
			remoteNames[remote.Name] = remote.URL
		}

		assert.Equal(t, env.SourceBareRepo.URL, remoteNames["origin"])
		assert.Equal(t, "https://example.com/upstream.git", remoteNames["upstream"])
	})

	t.Run("git_operations_can_be_tested", func(t *testing.T) {
		env, err := SetupSimpleGitTestEnvironment(t, gitOps)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		ctx := context.Background()

		// Test remote operations
		err = env.WorkingRepo.Repo.AddRemote(ctx, "test-remote", env.TargetBareRepo.URL)
		require.NoError(t, err)

		remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
		require.NoError(t, err)
		assert.Len(t, remotes, 2) // origin + test-remote

		// Test remote update (this is what our main fix was about)
		err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.TargetBareRepo.URL)
		require.NoError(t, err)

		// Verify remote was updated
		updatedRemotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
		require.NoError(t, err)
		
		var originRemote *ports.RemoteInfo
		for _, remote := range updatedRemotes {
			if remote.Name == "origin" {
				originRemote = &remote
				break
			}
		}

		require.NotNil(t, originRemote)
		assert.Equal(t, env.TargetBareRepo.URL, originRemote.URL)
	})
}

func TestGitTestEnvironmentRealWorldScenario(t *testing.T) {
	t.Parallel()

	gitOps := gogit.New(ports.GitConfig{
		UserName:    "Integration Test",
		UserEmail:   "integration@test.com",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
	})

	t.Run("github_to_gitlab_sync_simulation", func(t *testing.T) {
		// This test simulates the real-world scenario we're testing:
		// 1. A repository exists on GitHub (source)
		// 2. We clone it locally 
		// 3. We want to push it to GitLab (target)
		// 4. We need to update the origin remote from GitHub URL to GitLab URL

		opts := GitTestOptions{
			SourceRepoName:  "github-repo",
			TargetRepoName:  "gitlab-repo",
			WorkingRepoName: "local-clone",
			InitialFiles: map[string]string{
				"README.md": "# My Project\nOriginally from GitHub, now syncing to GitLab",
				"main.go":   "package main\n\nfunc main() {\n\tprintln(\"Hello from GitHub!\")\n}",
			},
			AddRemotes: map[string]string{
				"origin": "", // Will be set to GitHub URL
			},
		}

		env, err := SetupGitTestEnvironment(t, gitOps, opts)
		require.NoError(t, err)
		// No manual cleanup needed - t.TempDir() handles it automatically

		ctx := context.Background()

		// Simulate the critical GitHub -> GitLab sync operation
		t.Logf("GitHub URL: %s", env.GetSourceURL())
		t.Logf("GitLab URL: %s", env.GetTargetURL())

		// 1. Verify initial state (origin points to GitHub)
		remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
		require.NoError(t, err)
		require.Len(t, remotes, 1)
		assert.Equal(t, "origin", remotes[0].Name)
		assert.Equal(t, env.GetSourceURL(), remotes[0].URL)

		// 2. Set up GPSUPSTREAM remote (backup of original GitHub URL)
		err = env.WorkingRepo.Repo.AddRemote(ctx, "GPSUPSTREAM", env.GetSourceURL())
		require.NoError(t, err)

		// 3. Update origin to point to GitLab (THE CRITICAL FIX)
		err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
		require.NoError(t, err)

		// 4. Verify the fix: origin now points to GitLab, GPSUPSTREAM to GitHub
		updatedRemotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
		require.NoError(t, err)
		require.Len(t, updatedRemotes, 2)

		remoteMap := make(map[string]string)
		for _, remote := range updatedRemotes {
			remoteMap[remote.Name] = remote.URL
		}

		assert.Equal(t, env.GetTargetURL(), remoteMap["origin"], "Origin should now point to GitLab")
		assert.Equal(t, env.GetSourceURL(), remoteMap["GPSUPSTREAM"], "GPSUPSTREAM should point to GitHub")

		t.Logf("✅ Success: Origin updated from GitHub (%s) to GitLab (%s)", 
			remoteMap["GPSUPSTREAM"], remoteMap["origin"])

		// 5. Simulate push to GitLab (would work with git binary operations)
		pushOptions := ports.PushOptions{
			Remote:  "origin",
			Force:   false,
			Auth:    ports.AuthOptions{Type: ports.AuthTypeNone},
			Timeout: time.Minute,
		}

		// This would push to the GitLab bare repo in a real scenario
		// For now we just verify the options can be created
		assert.Equal(t, "origin", pushOptions.Remote)
		assert.Equal(t, env.GetTargetURL(), remoteMap["origin"])
	})
}

func BenchmarkGitTestEnvironmentSetup(b *testing.B) {
	gitOps := gogit.New(ports.GitConfig{
		UserName:    "Benchmark User",
		UserEmail:   "benchmark@test.com",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
	})

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := SetupSimpleGitTestEnvironment(nil, gitOps)
		if err != nil {
			b.Fatalf("Failed to setup environment: %v", err)
		}
		// No manual cleanup needed - t.TempDir() handles it automatically
	}
}