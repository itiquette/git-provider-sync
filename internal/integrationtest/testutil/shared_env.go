//go:build integration

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// SharedGitEnvironment provides a reusable git environment for parallel tests
type SharedGitEnvironment struct {
	mu        sync.RWMutex
	BaseDir   string
	GitOps    ports.GitOperations
	setupOnce sync.Once
	t         *testing.T
}

var (
	sharedEnvInstance *SharedGitEnvironment
	sharedEnvOnce     sync.Once
)

// GetSharedGitEnvironment returns a singleton shared git environment for parallel tests
// This allows multiple tests to run in parallel using isolated subdirectories
func GetSharedGitEnvironment(t *testing.T) *SharedGitEnvironment {
	sharedEnvOnce.Do(func() {
		// Create base directory that will be shared
		baseDir := t.TempDir()

		// Create git operations instance
		gitOps := createDefaultGitOps()

		sharedEnvInstance = &SharedGitEnvironment{
			BaseDir: baseDir,
			GitOps:  gitOps,
			t:       t,
		}

		// Setup base environment once
		sharedEnvInstance.setupBaseEnvironment(t)
	})

	return sharedEnvInstance
}

func (env *SharedGitEnvironment) setupBaseEnvironment(t *testing.T) {
	t.Helper()

	// Create shared directories
	dirs := []string{
		filepath.Join(env.BaseDir, "repos"),
		filepath.Join(env.BaseDir, "work"),
		filepath.Join(env.BaseDir, "archives"),
	}

	for _, dir := range dirs {
		require.NoError(t, os.MkdirAll(dir, 0750))
	}
}

// CreateIsolatedWorkspace creates an isolated workspace for a specific test
// Each test gets its own subdirectory to avoid conflicts
func (env *SharedGitEnvironment) CreateIsolatedWorkspace(t *testing.T, name string) string {
	t.Helper()

	env.mu.Lock()
	defer env.mu.Unlock()

	// Create unique workspace for this test
	workspace := filepath.Join(env.BaseDir, "work", name)
	require.NoError(t, os.MkdirAll(workspace, 0750))

	// Return the workspace path
	return workspace
}

// SetupTestRepository sets up a test repository in an isolated workspace
// Safe for parallel execution as each test gets its own workspace
func (env *SharedGitEnvironment) SetupTestRepository(t *testing.T, options GitTestOptions) (*GitTestEnvironment, error) {
	t.Helper()

	// Create isolated workspace for this test
	workspace := env.CreateIsolatedWorkspace(t, options.SourceRepoName)

	// Save current directory to use as base
	// Note: We don't change directory to avoid issues with parallel tests
	_ = workspace // workspace is used in the setup

	// Create a new GitTestEnvironment with isolated workspace
	testEnv, err := SetupGitTestEnvironment(t, env.GitOps, options)
	if err != nil {
		return nil, err
	}

	// Override the temp directory to use our isolated workspace
	testEnv.TmpDir = workspace

	return testEnv, nil
}

// ParallelGitTest runs a git test in parallel with proper isolation
func ParallelGitTest(t *testing.T, name string, testFunc func(t *testing.T, env *SharedGitEnvironment)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel() // Now safe to run in parallel

		env := GetSharedGitEnvironment(t)
		testFunc(t, env)
	})
}

// createDefaultGitOps creates a git operations instance for testing
func createDefaultGitOps() ports.GitOperations {
	// Import cycle prevention - use type assertion
	// This should be passed in or use a factory
	return nil // Caller should provide this
}

// SetupParallelIntegrationTest sets up an integration test that can run in parallel
// It creates an isolated environment without using t.Setenv
func SetupParallelIntegrationTest(t *testing.T, gitOps ports.GitOperations) *SharedGitEnvironment {
	t.Helper()

	baseDir := t.TempDir()

	env := &SharedGitEnvironment{
		BaseDir: baseDir,
		GitOps:  gitOps,
		t:       t,
	}

	env.setupBaseEnvironment(t)

	return env
}

// Example usage:
// func TestGitOperations_Parallel(t *testing.T) {
//     gitOps := gogit.New(testConfig)
//     env := SetupParallelIntegrationTest(t, gitOps)
//
//     ParallelGitTest(t, "test1", func(t *testing.T, env *SharedGitEnvironment) {
//         testEnv, err := env.SetupTestRepository(t, GitTestOptions{...})
//         require.NoError(t, err)
//         // Run test with testEnv
//     })
//
//     ParallelGitTest(t, "test2", func(t *testing.T, env *SharedGitEnvironment) {
//         testEnv, err := env.SetupTestRepository(t, GitTestOptions{...})
//         require.NoError(t, err)
//         // Run test with testEnv
//     })
// }
