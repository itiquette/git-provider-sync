//go:build integration

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package testutil provides utilities for setting up isolated git environments for integration testing.
package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// GitTestEnvironment represents an isolated git testing environment with bare repositories
// that can simulate git providers and working repositories for testing git operations.
type GitTestEnvironment struct {
	// TmpDir is the base temporary directory containing all test repositories
	TmpDir string

	// SourceBareRepo represents a "source" git provider (e.g., GitHub)
	SourceBareRepo GitTestRepo

	// TargetBareRepo represents a "target" git provider (e.g., GitLab)
	TargetBareRepo GitTestRepo

	// WorkingRepo is a clone of SourceBareRepo that can be used for git operations
	WorkingRepo GitTestRepo

	// GitOps is the git operations interface used to create repositories
	GitOps ports.GitOperations

	// Context for git operations
	Ctx context.Context
}

// GitTestRepo represents a repository in the test environment
type GitTestRepo struct {
	// Name is the repository name
	Name string

	// Path is the filesystem path to the repository
	Path string

	// URL is the git URL that can be used for clone/push operations
	URL string

	// Repo is the actual git repository interface (nil for bare repos)
	Repo ports.GitRepository

	// IsBare indicates if this is a bare repository
	IsBare bool
}

// GitTestOptions configures the git test environment
type GitTestOptions struct {
	// SourceRepoName is the name of the source repository (default: "source-repo")
	SourceRepoName string

	// TargetRepoName is the name of the target repository (default: "target-repo")
	TargetRepoName string

	// WorkingRepoName is the name of the working repository (default: "working-repo")
	WorkingRepoName string

	// InitialFiles are files to add to the source repository
	InitialFiles map[string]string

	// InitialBranch is the initial branch name (default: "main")
	InitialBranch string

	// AddRemotes configures remote URLs in the working repository
	AddRemotes map[string]string
}

// DefaultGitTestOptions returns default options for git test environment
func DefaultGitTestOptions() GitTestOptions {
	return GitTestOptions{
		SourceRepoName:  "source-repo",
		TargetRepoName:  "target-repo", 
		WorkingRepoName: "working-repo",
		InitialBranch:   "main",
		InitialFiles: map[string]string{
			"README.md":      "# Test Repository\n\nThis is a test repository for integration tests.\n",
			"src/main.go":    "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
			".gitignore":     "*.log\n*.tmp\n.DS_Store\n",
			"docs/setup.md":  "# Setup\n\nSetup instructions for the project.\n",
		},
		AddRemotes: map[string]string{
			"origin": "", // Will be set to source bare repo URL
		},
	}
}

// SetupGitTestEnvironment creates an isolated git test environment with bare repositories
// that can simulate git providers and a working repository for testing git operations.
//
// The environment includes:
// - A source bare repository (simulates GitHub/source provider)
// - A target bare repository (simulates GitLab/target provider)  
// - A working repository cloned from source with test content
//
// All repositories are created in isolated temporary directories and can be used
// to test real git operations including clone, push, pull, remote management, etc.
//
// Example usage:
//
//	env, err := testutil.SetupGitTestEnvironment(t, gitOps, testutil.DefaultGitTestOptions())
//	require.NoError(t, err)
//	// No manual cleanup needed - t.TempDir() handles it automatically
//
//	// Test git operations
//	err = env.WorkingRepo.Repo.Push(ctx, ports.PushOptions{
//		Remote: "origin",
//		Auth:   ports.AuthOptions{Type: ports.AuthTypeNone},
//	})
//	require.NoError(t, err)
func SetupGitTestEnvironment(t *testing.T, gitOps ports.GitOperations, opts GitTestOptions) (*GitTestEnvironment, error) {
	t.Helper()

	ctx := context.Background()

	// Create base temporary directory using t.TempDir() for automatic cleanup
	tmpDir := t.TempDir()

	env := &GitTestEnvironment{
		TmpDir: tmpDir,
		GitOps: gitOps,
		Ctx:    ctx,
	}

	// Apply defaults
	if opts.SourceRepoName == "" {
		opts.SourceRepoName = "source-repo"
	}
	if opts.TargetRepoName == "" {
		opts.TargetRepoName = "target-repo"
	}
	if opts.WorkingRepoName == "" {
		opts.WorkingRepoName = "working-repo"
	}
	if opts.InitialBranch == "" {
		opts.InitialBranch = "main"
	}

	// Setup source bare repository (simulates GitHub/source provider)
	if err := env.setupSourceBareRepo(opts); err != nil {
		return nil, fmt.Errorf("failed to setup source bare repo: %w", err)
	}

	// Setup target bare repository (simulates GitLab/target provider)
	if err := env.setupTargetBareRepo(opts); err != nil {
		return nil, fmt.Errorf("failed to setup target bare repo: %w", err)
	}

	// Setup working repository with initial content
	if err := env.setupWorkingRepo(opts); err != nil {
		return nil, fmt.Errorf("failed to setup working repo: %w", err)
	}

	return env, nil
}

// setupSourceBareRepo creates a bare repository that simulates the source git provider
func (env *GitTestEnvironment) setupSourceBareRepo(opts GitTestOptions) error {
	sourcePath := filepath.Join(env.TmpDir, opts.SourceRepoName+".git")
	
	// Create bare repository
	sourceRepo, err := env.GitOps.Init(env.Ctx, sourcePath, ports.InitOptions{Bare: true})
	if err != nil {
		return fmt.Errorf("failed to init source bare repo: %w", err)
	}
	defer sourceRepo.Close() // Close immediately since bare repos don't need to stay open

	env.SourceBareRepo = GitTestRepo{
		Name:   opts.SourceRepoName,
		Path:   sourcePath,
		URL:    sourcePath, // Local path as URL for file:// protocol
		Repo:   nil,        // Bare repos don't have a working interface
		IsBare: true,
	}

	return nil
}

// setupTargetBareRepo creates a bare repository that simulates the target git provider
func (env *GitTestEnvironment) setupTargetBareRepo(opts GitTestOptions) error {
	targetPath := filepath.Join(env.TmpDir, opts.TargetRepoName+".git")
	
	// Create bare repository
	targetRepo, err := env.GitOps.Init(env.Ctx, targetPath, ports.InitOptions{Bare: true})
	if err != nil {
		return fmt.Errorf("failed to init target bare repo: %w", err)
	}
	defer targetRepo.Close() // Close immediately since bare repos don't need to stay open

	env.TargetBareRepo = GitTestRepo{
		Name:   opts.TargetRepoName,
		Path:   targetPath,
		URL:    targetPath, // Local path as URL for file:// protocol
		Repo:   nil,        // Bare repos don't have a working interface
		IsBare: true,
	}

	return nil
}

// setupWorkingRepo creates a working repository with initial content and pushes to source bare repo
func (env *GitTestEnvironment) setupWorkingRepo(opts GitTestOptions) error {
	workingPath := filepath.Join(env.TmpDir, opts.WorkingRepoName)
	
	// Create working repository
	workingRepo, err := env.GitOps.Init(env.Ctx, workingPath, ports.InitOptions{Bare: false})
	if err != nil {
		return fmt.Errorf("failed to init working repo: %w", err)
	}

	env.WorkingRepo = GitTestRepo{
		Name:   opts.WorkingRepoName,
		Path:   workingPath,
		URL:    workingPath,
		Repo:   workingRepo,
		IsBare: false,
	}

	// Add source bare repo as origin remote
	originURL := env.SourceBareRepo.URL
	if remoteURL, exists := opts.AddRemotes["origin"]; exists && remoteURL != "" {
		originURL = remoteURL
	}
	
	if err := workingRepo.AddRemote(env.Ctx, "origin", originURL); err != nil {
		return fmt.Errorf("failed to add origin remote: %w", err)
	}

	// Add additional remotes
	for remoteName, remoteURL := range opts.AddRemotes {
		if remoteName == "origin" {
			continue // Already handled above
		}
		if remoteURL == "" {
			continue // Skip empty URLs
		}
		
		if err := workingRepo.AddRemote(env.Ctx, remoteName, remoteURL); err != nil {
			return fmt.Errorf("failed to add remote %s: %w", remoteName, err)
		}
	}

	// Create initial files if specified
	if len(opts.InitialFiles) > 0 {
		if err := env.createInitialFiles(opts); err != nil {
			return fmt.Errorf("failed to create initial files: %w", err)
		}
	}

	return nil
}

// createInitialFiles creates the initial test files in the working repository
func (env *GitTestEnvironment) createInitialFiles(opts GitTestOptions) error {
	for filePath, content := range opts.InitialFiles {
		fullPath := filepath.Join(env.WorkingRepo.Path, filePath)
		
		// Create directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		
		// Write file content
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
	}

	return nil
}

// PushToSource pushes the working repository content to the source bare repository.
// This simulates the initial state where content exists in the source provider.
//
// Note: This requires git binary operations to add, commit, and push.
// For now, this is a placeholder that would need git binary adapter implementation.
func (env *GitTestEnvironment) PushToSource() error {
	// TODO: Implement git add, commit, push operations
	// This would require extending the GitOperations interface or using git binary adapter
	return fmt.Errorf("PushToSource not implemented - requires git binary operations for add/commit/push")
}

// Clone creates a fresh clone of the source repository for testing
func (env *GitTestEnvironment) Clone(name string) (*GitTestRepo, error) {
	clonePath := filepath.Join(env.TmpDir, name)
	
	cloneRepo, err := env.GitOps.Clone(env.Ctx, ports.CloneOptions{
		URL:    env.SourceBareRepo.URL,
		Path:   clonePath,
		Branch: "main",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	testRepo := &GitTestRepo{
		Name:   name,
		Path:   clonePath,
		URL:    clonePath,
		Repo:   cloneRepo,
		IsBare: false,
	}

	return testRepo, nil
}

// NOTE: No manual cleanup method needed - t.TempDir() automatically handles cleanup

// GetSourceURL returns the URL for the source bare repository
func (env *GitTestEnvironment) GetSourceURL() string {
	return env.SourceBareRepo.URL
}

// GetTargetURL returns the URL for the target bare repository
func (env *GitTestEnvironment) GetTargetURL() string {
	return env.TargetBareRepo.URL
}

// SetupSimpleGitTestEnvironment is a convenience function that creates a simple git test environment
// with default options and test content. This is useful for most integration tests.
func SetupSimpleGitTestEnvironment(t *testing.T, gitOps ports.GitOperations) (*GitTestEnvironment, error) {
	t.Helper()
	return SetupGitTestEnvironment(t, gitOps, DefaultGitTestOptions())
}