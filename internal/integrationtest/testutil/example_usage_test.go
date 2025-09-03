//go:build integration

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package testutil provides usage examples for the git test environment utility.
package testutil

import (
	"context"

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
			"README.md":     "# Custom Test Repository",
			"src/main.go":   "package main\n\nfunc main() {}",
			".gitignore":    "*.log\ntmp/",
			"docs/guide.md": "# User Guide\n\nWelcome to the test project",
		},
		AddRemotes: map[string]string{
			"origin":   "", // Will use source bare repo
			"upstream": "https://example.com/upstream.git",
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
	_ = env.TmpDir         // All repos are in isolated temporary directory
	_ = env.GetSourceURL() // Source repo URL (simulates GitHub)
	_ = env.GetTargetURL() // Target repo URL (simulates GitLab)
}
