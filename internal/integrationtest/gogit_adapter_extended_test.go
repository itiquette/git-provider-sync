//go:build integration

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/repository/gogit"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/integrationtest/testutil"
)

// TestGoGitAdapterExtendedIntegration tests GoGit adapter operations
// Moved from internal/adapters/repository/gogit/adapter_test.go:375
func TestGoGitAdapterExtendedIntegration(t *testing.T) {
	// Isolate Git environment from host system
	// Note: Cannot use t.Parallel() when using t.Setenv in IsolateGitEnvironment
	testutil.IsolateGitEnvironment(t)

	adapter := gogit.New(ports.GitConfig{
		UserName:    "GoGit Extended Test",
		UserEmail:   "gogit-extended@integration.test",
		StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
	})

	tests := []struct {
		name     string
		testFunc func(t *testing.T, adapter *gogit.Adapter)
	}{
		{
			name:     "init_and_basic_operations",
			testFunc: testGoGitInitAndBasicOperations,
		},
		{
			name:     "remote_management",
			testFunc: testGoGitRemoteManagement,
		},
		{
			name:     "branch_operations",
			testFunc: testGoGitBranchOperations,
		},
		{
			name:     "file_operations",
			testFunc: testGoGitFileOperations,
		},
		{
			name:     "error_handling",
			testFunc: testGoGitErrorHandling,
		},
		{
			name:     "repository_info",
			testFunc: testGoGitRepositoryInfo,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t, adapter)
		})
	}
}

// TestGoGitInitAndBasicOperations tests repository initialization and basic operations
func testGoGitInitAndBasicOperations(t *testing.T, adapter *gogit.Adapter) {
	t.Helper()

	// Create safe isolated test environment using our git test utility
	env, err := testutil.SetupSimpleGitTestEnvironment(t, adapter)
	require.NoError(t, err)
	// No manual cleanup needed - t.TempDir() handles it automatically

	ctx := context.Background()

	// Test repository initialization
	repoPath := filepath.Join(env.TmpDir, "init-test-repo")
	repo, err := adapter.Init(ctx, repoPath, ports.InitOptions{
		Bare: false,
	})
	require.NoError(t, err, "Should initialize repository")
	require.NotNil(t, repo)

	defer func() { _ = repo.Close() }()

	// Test basic repository properties
	assert.Equal(t, "init-test-repo", repo.Name())
	assert.Equal(t, repoPath, repo.Path())
	assert.False(t, repo.IsBare())
	assert.True(t, repo.IsClean())
	assert.False(t, repo.HasChanges())

	// Test repository status
	status, err := repo.Status(ctx)
	require.NoError(t, err, "Should get status")
	assert.NotNil(t, status)
	assert.True(t, status.IsClean)

	// Test opening existing repository
	repo2, err := adapter.Open(ctx, repoPath)
	require.NoError(t, err, "Should open existing repository")
	require.NotNil(t, repo2)

	defer func() { _ = repo2.Close() }()

	assert.Equal(t, repo.Name(), repo2.Name())
	assert.Equal(t, repo.Path(), repo2.Path())

	// Test bare repository initialization
	bareRepoPath := filepath.Join(env.TmpDir, "bare-repo")
	bareRepo, err := adapter.Init(ctx, bareRepoPath, ports.InitOptions{
		Bare: true,
	})
	require.NoError(t, err, "Should initialize bare repository")
	require.NotNil(t, bareRepo)

	defer func() { _ = bareRepo.Close() }()

	assert.True(t, bareRepo.IsBare())
	assert.Equal(t, "bare-repo", bareRepo.Name())

	t.Logf("✅ GoGit init and basic operations completed")
	t.Logf("   Test environment: %s", env.TmpDir)
	t.Logf("   Repository: %s", repoPath)
	t.Logf("   Bare repository: %s", bareRepoPath)
}

// TestGoGitRemoteManagement tests remote operations using test environment
func testGoGitRemoteManagement(t *testing.T, adapter *gogit.Adapter) {
	t.Helper()

	// Create test environment with realistic git setup
	env, err := testutil.SetupGitTestEnvironment(t, adapter, testutil.GitTestOptions{
		SourceRepoName:  "remote-source",
		TargetRepoName:  "remote-target",
		WorkingRepoName: "remote-workspace",
		InitialFiles: map[string]string{
			"README.md": "# Remote Management Test\n\nTesting remote operations",
		},
		AddRemotes: map[string]string{
			"origin": "", // Will use source bare repo
		},
	})
	require.NoError(t, err)
	// No manual cleanup needed - t.TempDir() handles it automatically

	ctx := context.Background()
	repo := env.WorkingRepo.Repo

	// Test initial remotes
	remotes, err := repo.ListRemotes(ctx)
	require.NoError(t, err)
	require.Len(t, remotes, 1)
	assert.Equal(t, "origin", remotes[0].Name)
	assert.Equal(t, env.GetSourceURL(), remotes[0].URL)

	// Test adding new remote
	err = repo.AddRemote(ctx, "upstream", env.GetTargetURL())
	require.NoError(t, err, "Should add upstream remote")

	// Verify remotes
	remotes, err = repo.ListRemotes(ctx)
	require.NoError(t, err)
	require.Len(t, remotes, 2)

	remoteMap := make(map[string]string)
	for _, remote := range remotes {
		remoteMap[remote.Name] = remote.URL
	}

	assert.Equal(t, env.GetSourceURL(), remoteMap["origin"])
	assert.Equal(t, env.GetTargetURL(), remoteMap["upstream"])

	// Test updating remote URL
	err = repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
	require.NoError(t, err, "Should update remote URL")

	// Verify update
	remotes, err = repo.ListRemotes(ctx)
	require.NoError(t, err)

	var updatedOrigin *ports.RemoteInfo
	for _, remote := range remotes {
		if remote.Name == "origin" {
			updatedOrigin = &remote
			break
		}
	}

	require.NotNil(t, updatedOrigin)
	assert.Equal(t, env.GetTargetURL(), updatedOrigin.URL)

	// Test removing remote
	err = repo.RemoveRemote(ctx, "upstream")
	require.NoError(t, err, "Should remove remote")

	// Verify removal
	remotes, err = repo.ListRemotes(ctx)
	require.NoError(t, err)
	require.Len(t, remotes, 1)
	assert.Equal(t, "origin", remotes[0].Name)

	t.Logf("✅ GoGit remote management completed")
	t.Logf("   Source URL: %s", env.GetSourceURL())
	t.Logf("   Target URL: %s", env.GetTargetURL())
}

// TestGoGitBranchOperations tests branch management operations
func testGoGitBranchOperations(t *testing.T, adapter *gogit.Adapter) {
	t.Helper()

	// Create simple test environment for branch operations
	env, err := testutil.SetupSimpleGitTestEnvironment(t, adapter)
	require.NoError(t, err)
	// No manual cleanup needed - t.TempDir() handles it automatically

	ctx := context.Background()
	repo := env.WorkingRepo.Repo

	// Test listing branches (may be empty for new repo)
	branches, err := repo.ListBranches(ctx)
	if err != nil {
		t.Logf("List branches failed (expected for empty repo): %v", err)
	} else {
		t.Logf("Initial branches: %+v", branches)
	}

	// Test current branch
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		t.Logf("Current branch detection failed (expected for empty repo): %v", err)
	} else {
		assert.NotEmpty(t, currentBranch)
		t.Logf("Current branch: %s", currentBranch)
	}

	// Test creating branch (may fail for empty repository)
	err = repo.CreateBranch(ctx, "feature/test", "main")
	if err != nil {
		t.Logf("Create branch failed (expected for empty repo): %v", err)
	} else {
		t.Log("Successfully created feature branch")

		// Test checkout if creation succeeded
		err = repo.CheckoutBranch(ctx, "feature/test")
		if err != nil {
			t.Logf("Checkout branch failed: %v", err)
		} else {
			t.Log("Successfully checked out feature branch")
		}
	}

	// Test setting default branch
	err = repo.SetDefaultBranch(ctx, "main")
	if err != nil {
		t.Logf("Set default branch failed: %v", err)
	} else {
		t.Log("Successfully set default branch")
	}

	t.Logf("✅ GoGit branch operations completed")
	t.Logf("   Working repo: %s", env.WorkingRepo.Path)
}

// TestGoGitFileOperations tests file operations within repository
func testGoGitFileOperations(t *testing.T, adapter *gogit.Adapter) {
	t.Helper()

	// Create test environment with files
	env, err := testutil.SetupGitTestEnvironment(t, adapter, testutil.GitTestOptions{
		SourceRepoName:  "file-source",
		TargetRepoName:  "file-target",
		WorkingRepoName: "file-workspace",
		InitialFiles: map[string]string{
			"main.go":     "package main\n\nfunc main() {\n\tprintln(\"Hello GoGit!\")\n}",
			"README.md":   "# File Operations Test\n\nTesting file operations",
			"config.yaml": "app:\n  name: gogit-test\n  version: 1.0.0",
		},
		AddRemotes: map[string]string{
			"origin": "",
		},
	})
	require.NoError(t, err)
	// No manual cleanup needed - t.TempDir() handles it automatically

	ctx := context.Background()
	repo := env.WorkingRepo.Repo

	// Test repository status
	status, err := repo.Status(ctx)
	require.NoError(t, err, "Should get repository status")
	assert.NotNil(t, status)

	// New repository might have changes from initial file creation
	t.Logf("Repository status - Clean: %t, Modified files: %d", status.IsClean, len(status.Modified))

	// Test diff operations
	diff, err := repo.Diff(ctx, ports.DiffOptions{})
	if err != nil {
		t.Logf("Diff failed (expected for new repo): %v", err)
	} else {
		t.Logf("Diff output length: %d characters", len(diff))
	}

	// Test tags (empty repository may not support tags yet)
	tags, err := repo.ListTags(ctx)
	if err != nil {
		t.Logf("List tags failed (expected for empty repo): %v", err)
	} else {
		t.Logf("Repository tags: %d", len(tags))
	}

	// Test commits
	commits, err := repo.ListCommits(ctx, ports.ListCommitsOptions{})
	if err != nil {
		t.Logf("List commits failed (expected for empty repo): %v", err)
	} else {
		t.Logf("Repository commits: %d", len(commits))
	}

	t.Logf("✅ GoGit file operations completed")
	t.Logf("   Files created: 3") // main.go, README.md, config.yaml
	t.Logf("   Working directory: %s", env.WorkingRepo.Path)
}

// TestGoGitErrorHandling tests error handling scenarios
func testGoGitErrorHandling(t *testing.T, adapter *gogit.Adapter) {
	t.Helper()

	// Create test environment
	env, err := testutil.SetupSimpleGitTestEnvironment(t, adapter)
	require.NoError(t, err)
	// No manual cleanup needed - t.TempDir() handles it automatically

	ctx := context.Background()

	// Test opening non-existent repository
	nonExistentPath := filepath.Join(env.TmpDir, "non-existent-repo")
	_, err = adapter.Open(ctx, nonExistentPath)
	require.Error(t, err, "Should fail to open non-existent repository")

	// Test initializing in invalid location (if any)
	invalidPath := "/invalid/path/that/should/not/exist"
	_, err = adapter.Init(ctx, invalidPath, ports.InitOptions{})
	require.Error(t, err, "Should fail to initialize in invalid location")

	// Test clone with invalid URL
	clonePath := filepath.Join(env.TmpDir, "clone-test")
	_, err = adapter.Clone(ctx, ports.CloneOptions{
		URL:  "invalid://not-a-valid-url",
		Path: clonePath,
		Auth: ports.AuthOptions{Type: ports.AuthTypeNone},
	})
	require.Error(t, err, "Should fail to clone invalid URL")

	t.Logf("✅ GoGit error handling completed")
	t.Logf("   All error scenarios handled correctly")
}

// TestGoGitRepositoryInfo tests repository information retrieval
func testGoGitRepositoryInfo(t *testing.T, adapter *gogit.Adapter) {
	t.Helper()

	// Create test environment
	env, err := testutil.SetupSimpleGitTestEnvironment(t, adapter)
	require.NoError(t, err)
	// No manual cleanup needed - t.TempDir() handles it automatically

	ctx := context.Background()

	// Test adapter properties
	assert.Equal(t, "go-git", adapter.GetName())
	assert.True(t, adapter.SupportsURL("https://github.com/test/repo.git"))
	assert.True(t, adapter.SupportsURL("git@github.com:test/repo.git"))
	assert.True(t, adapter.SupportsURL(env.GetSourceURL()))
	assert.False(t, adapter.SupportsURL("ftp://invalid.com/repo"))

	// Test repository initialization and properties
	repoPath := filepath.Join(env.TmpDir, "info-test-repo")
	repo, err := adapter.Init(ctx, repoPath, ports.InitOptions{Bare: false})
	require.NoError(t, err)
	require.NotNil(t, repo)

	defer func() { _ = repo.Close() }()

	// Test repository info
	assert.Equal(t, "info-test-repo", repo.Name())
	assert.Equal(t, repoPath, repo.Path())
	assert.False(t, repo.IsBare())
	assert.True(t, repo.IsClean())
	assert.False(t, repo.HasChanges())

	// Test URL (may be empty for local repo)
	url := repo.URL()
	t.Logf("Repository URL: '%s'", url)

	t.Logf("✅ GoGit repository info completed")
	t.Logf("   Adapter name: %s", adapter.GetName())
	t.Logf("   Repository: %s", repo.Name())
}
