// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"context"
	"testing"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Test provides everything needed for a test.
// This is the main entry point - tests should use this, not Afero directly.
//
//nolint:containedctx // Test utility needs to store context for test lifecycle
type Test struct {
	T   testing.TB
	FS  *TestFS         // Filesystem operations (Afero hidden inside)
	Ctx context.Context // Context for operations
}

// NewTest creates a new test environment with in-memory filesystem.
// This is what tests should use - simple and clean.
//
//nolint:thelper,varnamelen // t is the established parameter name in this widely-used test utility
func NewTest(t testing.TB) *Test {
	t.Helper()

	// Note: We don't enforce isolation here because some tests
	// need to interact with real filesystem to test safety guards
	// Tests can explicitly call RequireTestEnvironment if needed

	return &Test{
		T:   t,
		FS:  NewTestFS(t),
		Ctx: context.Background(),
	}
}

// Parallel marks the test as parallel.
func (test *Test) Parallel() *Test {
	if t, ok := test.T.(*testing.T); ok {
		t.Parallel()
	}

	return test
}

// WriteConfig is a shortcut for writing config files.
func (test *Test) WriteConfig(content string) string {
	return test.FS.WriteConfig(content)
}

// CreateRepo is a shortcut for creating a git repo.
func (test *Test) CreateRepo(name string) string {
	return test.FS.CreateGitRepo(name)
}

// Setup creates a complete test structure from a map.
func (test *Test) Setup(structure map[string]string) {
	test.FS.CreateStructure(structure)
}

// GetFileSystem returns a ports.FileSystem implementation for dependency injection.
// This allows tests to pass a memory filesystem to production code that expects the interface.
func (test *Test) GetFileSystem() ports.FileSystem {
	return test.FS.GetFileSystem()
}

// ConfigFixture provides standard test configurations.
type ConfigFixture struct {
	Minimal string
	GitHub  string
	GitLab  string
	Mirror  string
}

// Configs provides standard test configurations.
//
//nolint:gochecknoglobals // Test fixtures are acceptable as package-level variables
var Configs = ConfigFixture{
	Minimal: `gitprovidersync:
  test:
    source:
      provider_type: github
      owner: test-owner
      owner_type: user
`,
	GitHub: `gitprovidersync:
  github:
    source:
      provider_type: github
      domain: github.com
      owner: octocat
      owner_type: user
      auth:
        token: "${GITHUB_TOKEN}"
`,
	GitLab: `gitprovidersync:
  gitlab:
    source:
      provider_type: gitlab
      domain: gitlab.com
      owner: gitlab-org
      owner_type: group
      auth:
        token: "${GITLAB_TOKEN}"
`,
	Mirror: `gitprovidersync:
  sync:
    source:
      provider_type: github
      owner: source-owner
      mirrors:
        backup:
          provider_type: gitlab
          owner: backup-owner
`,
}
