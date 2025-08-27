// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitGoRemoteOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		remoteName     string
		urls           []string
		isMirror       bool
		expectedConfig config.RemoteConfig
	}{
		{
			name:       "create remote config with single URL",
			remoteName: "origin",
			urls:       []string{"https://github.com/user/repo.git"},
			isMirror:   false,
			expectedConfig: config.RemoteConfig{
				Name:   "origin",
				URLs:   []string{"https://github.com/user/repo.git"},
				Mirror: false,
			},
		},
		{
			name:       "create remote config with multiple URLs",
			remoteName: "upstream",
			urls:       []string{"https://github.com/upstream/repo.git", "git@github.com:upstream/repo.git"},
			isMirror:   false,
			expectedConfig: config.RemoteConfig{
				Name:   "upstream",
				URLs:   []string{"https://github.com/upstream/repo.git", "git@github.com:upstream/repo.git"},
				Mirror: false,
			},
		},
		{
			name:       "create mirror remote config",
			remoteName: "mirror",
			urls:       []string{"https://gitlab.com/user/repo.git"},
			isMirror:   true,
			expectedConfig: config.RemoteConfig{
				Name:   "mirror",
				URLs:   []string{"https://gitlab.com/user/repo.git"},
				Mirror: true,
			},
		},
		{
			name:       "create remote config with empty URL slice",
			remoteName: "empty",
			urls:       []string{},
			isMirror:   false,
			expectedConfig: config.RemoteConfig{
				Name:   "empty",
				URLs:   []string{},
				Mirror: false,
			},
		},
		{
			name:       "create remote config with nil URL slice",
			remoteName: "nil-urls",
			urls:       nil,
			isMirror:   true,
			expectedConfig: config.RemoteConfig{
				Name:   "nil-urls",
				URLs:   nil,
				Mirror: true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := NewGitGoRemoteOption(testCase.remoteName, testCase.urls, testCase.isMirror)

			assert.Equal(t, testCase.expectedConfig.Name, result.Name)
			assert.Equal(t, testCase.expectedConfig.URLs, result.URLs)
			assert.Equal(t, testCase.expectedConfig.Mirror, result.Mirror)
		})
	}
}

func TestRepositoryErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		errorVar      error
		expectedError string
	}{
		{
			name:          "ErrInvalidLengthURL error",
			errorVar:      ErrInvalidLengthURL,
			expectedError: "remote has no URL",
		},
		{
			name:          "ErrNullRepositoryPtr error",
			errorVar:      ErrNullRepositoryPtr,
			expectedError: "parameter repositoryPtr was null",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expectedError, testCase.errorVar.Error())
		})
	}
}

func TestRepository_FieldAccess(t *testing.T) {
	t.Parallel()

	// Note: These tests focus on the structure and interface methods without requiring
	// actual git repositories, as creating temporary git repos in parallel tests
	// can be complex and resource-intensive.

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Repository struct field access",
			testFunc: func(t *testing.T) {
				t.Helper()
				projectInfo := &ProjectInfo{
					OriginalName: "test-repo",
					HTTPSURL:     "https://github.com/user/test-repo.git",
				}

				// We can't easily create a real git.Repository in a unit test without
				// setting up actual git repositories, so we test the struct design
				repo := Repository{
					goGitRepository: nil, // Would be a real repository in practice
					ProjectMetaInfo: projectInfo,
				}

				// Test ProjectInfo method
				actualProjectInfo := repo.ProjectInfo()
				require.Equal(t, projectInfo, actualProjectInfo)
				assert.Equal(t, "test-repo", actualProjectInfo.OriginalName)

				// Test GoGitRepository method
				assert.Nil(t, repo.GoGitRepository()) // nil in this test case
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestRepository_NilValues_HandlesGracefully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "repository with nil project info",
			testFunc: func(t *testing.T) {
				t.Helper()
				repo := Repository{
					goGitRepository: nil,
					ProjectMetaInfo: nil,
				}

				// Should handle nil project info gracefully
				assert.Nil(t, repo.ProjectInfo())
				assert.Nil(t, repo.GoGitRepository())
			},
		},
		{
			name: "repository error handling patterns",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test the error constants are properly defined
				require.Error(t, ErrInvalidLengthURL)
				require.Error(t, ErrNullRepositoryPtr)
			},
		},
		{
			name: "remote config with special characters",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Test remote config creation with various URL types
				specialURLs := []string{
					"https://user:pass@github.com/owner/repo.git",
					"git@gitlab.com:owner/repo-with-special@chars.git",
					"https://github.com/owner/repo-with-unicode-项目.git",
				}

				config := NewGitGoRemoteOption("special", specialURLs, true)

				assert.Equal(t, "special", config.Name)
				assert.Equal(t, specialURLs, config.URLs)
				assert.True(t, config.Mirror)
			},
		},
		{
			name: "remote config with very long names and URLs",
			testFunc: func(t *testing.T) {
				t.Helper()
				longName := "very-long-remote-name-that-exceeds-normal-length-limits"
				longURL := "https://very-long-domain-name.example.com/organization-with-very-long-name/repository-with-extremely-long-name.git"

				config := NewGitGoRemoteOption(longName, []string{longURL}, false)

				assert.Equal(t, longName, config.Name)
				assert.Equal(t, []string{longURL}, config.URLs)
				assert.False(t, config.Mirror)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestRepositoryStructComposition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectInfo *ProjectInfo
		expectNil   bool
	}{
		{
			name: "repository with complete project metadata",
			projectInfo: &ProjectInfo{
				OriginalName:  "complete-repo",
				HTTPSURL:      "https://github.com/user/complete-repo.git",
				SSHURL:        "git@github.com:user/complete-repo.git",
				DefaultBranch: "main",
				Description:   "Repository with complete metadata",
			},
			expectNil: false,
		},
		{
			name: "repository with minimal project metadata",
			projectInfo: &ProjectInfo{
				OriginalName: "minimal-repo",
				HTTPSURL:     "https://example.com/minimal-repo.git",
			},
			expectNil: false,
		},
		{
			name:        "repository with nil project metadata",
			projectInfo: nil,
			expectNil:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Create repository with test project info
			repo := Repository{
				goGitRepository: nil, // Domain model shouldn't require git implementation
				ProjectMetaInfo: test.projectInfo,
			}

			// Act & Assert: Verify struct composition behavior
			retrievedInfo := repo.ProjectInfo()
			gitRepo := repo.GoGitRepository()

			if test.expectNil {
				assert.Nil(t, retrievedInfo)
			} else {
				assert.Equal(t, test.projectInfo, retrievedInfo)
				assert.Equal(t, test.projectInfo.OriginalName, retrievedInfo.OriginalName)

				if test.projectInfo.DefaultBranch != "" {
					assert.Equal(t, test.projectInfo.DefaultBranch, retrievedInfo.DefaultBranch)
				}
			}

			// GoGitRepository should be nil in these unit tests (no git dependency)
			assert.Nil(t, gitRepo)
		})
	}
}

func TestRepositoryCollectionManagement(t *testing.T) {
	t.Parallel()

	// Test managing multiple repository instances
	testRepositories := []struct {
		name        string
		httpsURL    string
		description string
	}{
		{"user-repo", "https://github.com/user/user-repo.git", "User repository"},
		{"org-repo", "https://github.com/org/org-repo.git", "Organization repository"},
		{"test-repo", "https://gitlab.com/test/test-repo.git", "Test repository"},
	}

	// Arrange: Create collection of repositories
	repos := make([]Repository, len(testRepositories))
	for index, testRepo := range testRepositories {
		repos[index] = Repository{
			goGitRepository: nil,
			ProjectMetaInfo: &ProjectInfo{
				OriginalName: testRepo.name,
				HTTPSURL:     testRepo.httpsURL,
				Description:  testRepo.description,
			},
		}
	}

	// Act & Assert: Verify collection operations
	require.Len(t, repos, len(testRepositories))

	for index, repo := range repos {
		expected := testRepositories[index]
		projectInfo := repo.ProjectInfo()

		require.NotNil(t, projectInfo)
		assert.Equal(t, expected.name, projectInfo.OriginalName)
		assert.Equal(t, expected.httpsURL, projectInfo.HTTPSURL)
		assert.Equal(t, expected.description, projectInfo.Description)
	}

	// Test repository filtering by name
	filteredRepos := make([]Repository, 0)

	for _, repo := range repos {
		if strings.Contains(repo.ProjectInfo().OriginalName, "repo") {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	assert.Len(t, filteredRepos, 3) // All test repositories contain "repo"
}
