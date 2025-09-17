// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/repository/gogit"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/integrationtest/testutil"
)

// TestGitRepositoryCreationIntegration tests real git repository creation and basic operations
// TEST PURPOSE:
// Integration test validating actual git repository operations using real git commands,
// ensuring the git adapters work correctly with actual git repositories.
// SCENARIOS COVERED:
// - Repository initialization and cloning
// - Remote management (add, list, verify)
// - File operations within repositories
// - Repository properties verification
// - Multiple remote support
// LIMITATIONS:
// Cannot run in parallel due to environment variable isolation requirements (t.Setenv)
func TestGitRepositoryCreationIntegration(t *testing.T) {
	// Isolate Git environment from host system
	// Note: Cannot use t.Parallel() when using t.Setenv in IsolateGitEnvironment
	testutil.IsolateGitEnvironment(t)

	gitOps := gogit.New(ports.GitConfig{
		UserName:    "integration-test",
		UserEmail:   "test@git-provider-sync.local",
		StorageMode: ports.StorageModeMemory, // Memory mode is faster and works with bare repos
	})

	t.Run("git repository creation and file operations", func(t *testing.T) {
		testGitRepositoryCreationAndFileOperations(t, gitOps)
	})
}

// TestGitRepositoryCreationAndFileOperations tests actual git repository creation with file system
func testGitRepositoryCreationAndFileOperations(t *testing.T, gitOps ports.GitOperations) {
	ctx := context.Background()

	// Use git test environment utility for safe, isolated testing
	env, err := testutil.SetupGitTestEnvironment(t, gitOps, testutil.GitTestOptions{
		SourceRepoName:  "git-ops-test",
		WorkingRepoName: "git-ops-test",
		InitialFiles: map[string]string{
			"README.md":   "# git-ops-test\n\nTest repository for integration testing",
			"src/main.go": "package main\n\nfunc main() {\n\tprintln(\"Integration test!\")\n}",
		},
		AddRemotes: map[string]string{
			"origin": "", // Will use source bare repo URL
		},
	})
	require.NoError(t, err)

	// Test that we can perform basic git operations on real repository
	remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should be able to list remotes")
	require.Len(t, remotes, 1, "Should have origin remote")
	assert.Equal(t, "origin", remotes[0].Name)
	assert.Contains(t, remotes[0].URL, "git-ops-test.git")

	// Test repository properties
	assert.Equal(t, "git-ops-test", env.WorkingRepo.Repo.Name())
	assert.Contains(t, env.WorkingRepo.Repo.Path(), "git-ops-test")
	assert.False(t, env.WorkingRepo.Repo.IsBare())

	// Test that we can add another remote
	backupURL := "https://github.com/backup/git-ops-test.git"
	err = env.WorkingRepo.Repo.AddRemote(ctx, "backup", backupURL)
	require.NoError(t, err, "Should add backup remote")

	// Verify both remotes exist
	remotes, err = env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err, "Should list remotes after adding backup")
	require.Len(t, remotes, 2, "Should have two remotes")

	// Test that we can also use the target bare repo as a remote
	err = env.WorkingRepo.Repo.AddRemote(ctx, "target", env.GetTargetURL())
	require.NoError(t, err, "Should add target bare repo as remote")

	// Verify we now have three remotes (origin, backup, target)
	remotes, err = env.WorkingRepo.Repo.ListRemotes(ctx)
	require.NoError(t, err)
	require.Len(t, remotes, 3, "Should have three remotes now")
}

// Test repository creation moved to testutil.GitTestEnvironment

// MinimalMockProvider provides minimal implementation for integration tests
type minimalMockProvider struct{}

func (m *minimalMockProvider) ProjectExists(context.Context, string, string) (bool, string, error) {
	return false, "", nil
}
func (m *minimalMockProvider) CreateRepositoryForPush(context.Context, ports.CreateRepositoryRequest) (string, error) {
	return "", nil
}
func (m *minimalMockProvider) Protect(context.Context, string, string, string) error { return nil }
func (m *minimalMockProvider) Unprotect(context.Context, string, string) error       { return nil }
func (m *minimalMockProvider) ListRepositories(context.Context, ports.ProviderConfig) ([]entities.Repository, error) {
	return nil, nil
}
func (m *minimalMockProvider) GetRepository(context.Context, ports.ProviderConfig, string) (entities.Repository, error) {
	return entities.Repository{}, nil
}
func (m *minimalMockProvider) RepositoryExists(context.Context, ports.RepositoryExistsRequest) (bool, string, error) {
	return false, "", nil
}
func (m *minimalMockProvider) CreateRepository(context.Context, ports.ProviderConfig, ports.CreateRepositoryOptions) (entities.Repository, error) {
	return entities.Repository{}, nil
}
func (m *minimalMockProvider) UpdateRepository(context.Context, ports.ProviderConfig, string, ports.UpdateRepositoryOptions) error {
	return nil
}
func (m *minimalMockProvider) DeleteRepository(context.Context, ports.ProviderConfig, string) error {
	return nil
}
func (m *minimalMockProvider) SetDefaultBranch(context.Context, string, string, string) error {
	return nil
}
func (m *minimalMockProvider) ValidateRepositoryName(string) error             { return nil }
func (m *minimalMockProvider) IsValidProjectName(context.Context, string) bool { return true }
func (m *minimalMockProvider) TransformRepositoryName(name string, _ ports.NameTransformOptions) string {
	return name
}
func (m *minimalMockProvider) GetBranchProtection(context.Context, ports.ProviderConfig, string, string) (ports.BranchProtection, error) {
	return ports.BranchProtection{}, nil
}
func (m *minimalMockProvider) SetBranchProtection(context.Context, ports.ProviderConfig, string, string, ports.BranchProtection) error {
	return nil
}
func (m *minimalMockProvider) RemoveBranchProtection(context.Context, ports.ProviderConfig, string, string) error {
	return nil
}
func (m *minimalMockProvider) ListProtectedBranches(context.Context, ports.ProviderConfig, string) ([]string, error) {
	return nil, nil
}
func (m *minimalMockProvider) GetProviderInfo() ports.ProviderInfo        { return ports.ProviderInfo{} }
func (m *minimalMockProvider) SupportsFeature(ports.ProviderFeature) bool { return true }
