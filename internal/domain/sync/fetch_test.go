// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestFetchSourceRepositoriesUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*SharedMockRepositoryProvider, *SharedMockGitOperations, *SharedMockLogger)
		request        FetchSourceRequest
		expectedError  bool
		expectedResult func(FetchSourceResponse) bool
	}{
		{
			name: "successful_fetch_and_clone_repositories",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				repos := []entities.Repository{
					createTestRepositoryWithActivity("repo1", false, false, time.Now()),
					createTestRepositoryWithActivity("repo2", false, false, time.Now()),
				}

				// Mock repository listing
				provider.On("ListRepositories", mock.Anything, mock.AnythingOfType("ports.ProviderConfig")).Return(repos, nil)

				// Mock git cloning for each repository
				for _, repo := range repos {
					mockRepo := &SharedMockGitRepository{}
					mockRepo.On("Name").Return(repo.Name())
					mockRepo.On("Path").Return("/tmp/" + repo.Name())
					gitOps.On("Clone", mock.Anything, mock.MatchedBy(func(opts ports.CloneOptions) bool {
						return opts.URL == repo.HTTPSURL()
					})).Return(mockRepo, nil)
				}

				// Mock logger calls (lenient)
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: FetchSourceRequest{
				ProviderConfig: createTestProviderConfig(),
				DryRun:         false,
				IncludeForks:   true,
				Filters: ports.FilterOptions{
					IncludeForks:    true,
					IncludeArchived: false,
					IncludePrivate:  true,
					IncludePublic:   true,
				},
			},
			expectedError: false,
			expectedResult: func(resp FetchSourceResponse) bool {
				return resp.Success && len(resp.ClonedRepos) == 2 && resp.ProcessedCount == 2
			},
		},
		{
			name: "dry_run_mode_skips_cloning",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				repos := []entities.Repository{
					createTestRepositoryWithActivity("dry-repo1", false, false, time.Now()),
					createTestRepositoryWithActivity("dry-repo2", false, false, time.Now()),
				}

				// Mock repository listing
				provider.On("ListRepositories", mock.Anything, mock.AnythingOfType("ports.ProviderConfig")).Return(repos, nil)

				// No clone operations should be called in dry run
				// Mock logger calls (lenient)
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: FetchSourceRequest{
				ProviderConfig: createTestProviderConfig(),
				DryRun:         true,
				IncludeForks:   true,
				Filters: ports.FilterOptions{
					IncludeForks:    true,
					IncludeArchived: false,
					IncludePrivate:  true,
					IncludePublic:   true,
				},
			},
			expectedError: false,
			expectedResult: func(resp FetchSourceResponse) bool {
				return resp.Success && len(resp.ClonedRepos) == 0 && resp.ProcessedCount == 2 && len(resp.Repositories) == 2
			},
		},
		{
			name: "repository_filtering_excludes_forks_and_archived",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				repos := []entities.Repository{
					createTestRepositoryWithActivity("normal-repo", false, false, time.Now()),
					createTestRepositoryWithActivity("fork-repo", true, false, time.Now()),     // Fork - should be excluded
					createTestRepositoryWithActivity("archived-repo", false, true, time.Now()), // Archived - should be excluded
					createTestRepositoryWithActivity("fork-archived", true, true, time.Now()),  // Both - should be excluded
				}

				// Mock repository listing
				provider.On("ListRepositories", mock.Anything, mock.AnythingOfType("ports.ProviderConfig")).Return(repos, nil)

				// Only one repository should be cloned (normal-repo)
				mockRepo := &SharedMockGitRepository{}
				mockRepo.On("Name").Return("normal-repo")
				mockRepo.On("Path").Return("/tmp/normal-repo")
				gitOps.On("Clone", mock.Anything, mock.MatchedBy(func(opts ports.CloneOptions) bool {
					return opts.URL == "https://github.com/test/normal-repo.git"
				})).Return(mockRepo, nil)

				// Mock logger calls (lenient)
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: FetchSourceRequest{
				ProviderConfig: createTestProviderConfig(),
				DryRun:         false,
				IncludeForks:   false, // Exclude forks
				Filters: ports.FilterOptions{
					IncludeForks:    false, // Double ensure forks excluded
					IncludeArchived: false, // Exclude archived
					IncludePrivate:  true,
					IncludePublic:   true,
				},
			},
			expectedError: false,
			expectedResult: func(resp FetchSourceResponse) bool {
				return resp.Success && len(resp.ClonedRepos) == 1 && resp.ProcessedCount == 4 && resp.SkippedCount == 3
			},
		},
		{
			name: "provider_listing_failure",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				// Mock repository listing failure
				provider.On("ListRepositories", mock.Anything, mock.AnythingOfType("ports.ProviderConfig")).Return(
					[]entities.Repository{}, errors.New("failed to authenticate with provider"))

				// Mock logger calls (lenient)
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: FetchSourceRequest{
				ProviderConfig: createTestProviderConfig(),
				DryRun:         false,
				IncludeForks:   true,
				Filters:        ports.FilterOptions{},
			},
			expectedError: true,
			expectedResult: func(resp FetchSourceResponse) bool {
				return !resp.Success
			},
		},
		{
			name: "partial_clone_failures",
			setupMocks: func(provider *SharedMockRepositoryProvider, gitOps *SharedMockGitOperations, logger *SharedMockLogger) {
				repos := []entities.Repository{
					createTestRepositoryWithActivity("success-repo", false, false, time.Now()),
					createTestRepositoryWithActivity("fail-repo", false, false, time.Now()),
				}

				// Mock repository listing
				provider.On("ListRepositories", mock.Anything, mock.AnythingOfType("ports.ProviderConfig")).Return(repos, nil)

				// Mock successful clone for first repo
				mockRepo := &SharedMockGitRepository{}
				mockRepo.On("Name").Return("success-repo")
				mockRepo.On("Path").Return("/tmp/success-repo")
				gitOps.On("Clone", mock.Anything, mock.MatchedBy(func(opts ports.CloneOptions) bool {
					return opts.URL == "https://github.com/test/success-repo.git"
				})).Return(mockRepo, nil)

				// Mock failed clone for second repo
				gitOps.On("Clone", mock.Anything, mock.MatchedBy(func(opts ports.CloneOptions) bool {
					return opts.URL == "https://github.com/test/fail-repo.git"
				})).Return((*SharedMockGitRepository)(nil), errors.New("clone failed: repository not found"))

				// Mock logger calls (lenient)
				logger.On("Info", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
				logger.On("Error", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			},
			request: FetchSourceRequest{
				ProviderConfig: createTestProviderConfig(),
				DryRun:         false,
				IncludeForks:   true,
				Filters: ports.FilterOptions{
					IncludeForks:    true,
					IncludeArchived: false,
					IncludePrivate:  true,
					IncludePublic:   true,
				},
			},
			expectedError: false, // Partial failures don't cause overall failure
			expectedResult: func(resp FetchSourceResponse) bool {
				return !resp.Success && len(resp.ClonedRepos) == 1 && len(resp.Errors) == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockProvider := &SharedMockRepositoryProvider{}
			mockGitOps := &SharedMockGitOperations{}
			mockLogger := &SharedMockLogger{}

			// Setup test-specific mocks
			tt.setupMocks(mockProvider, mockGitOps, mockLogger)

			// Create use case
			useCase := NewFetchSourceRepositoriesUseCase(mockProvider, mockGitOps, mockLogger)

			// Execute with temporary directory in context
			ctx := context.Background()
			ctx, err := filesystem.CreateTmpDir(ctx, "", "fetch_test")
			require.NoError(t, err)
			defer func() {
				if cleanupErr := filesystem.DeleteTmpDir(ctx); cleanupErr != nil {
					t.Logf("Failed to cleanup temp directory: %v", cleanupErr)
				}
			}()

			result, err := useCase.Execute(ctx, tt.request)

			// Verify error expectation
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify result expectations
			if tt.expectedResult != nil {
				require.True(t, tt.expectedResult(result), "Result validation failed")
			}

			// Verify all mocks were called as expected
			mockProvider.AssertExpectations(t)
			mockGitOps.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestFetchSourceRepositoriesUseCase_applyFilters(t *testing.T) {
	tests := []struct {
		name         string
		repositories []entities.Repository
		filters      ports.FilterOptions
		includeForks bool
		expected     int // Number of repositories after filtering
	}{
		{
			name: "no_filters_includes_all",
			repositories: []entities.Repository{
				createTestRepositoryWithActivity("repo1", false, false, time.Now()),
				createTestRepositoryWithActivity("repo2", true, false, time.Now()), // Fork
				createTestRepositoryWithActivity("repo3", false, true, time.Now()), // Archived
				createTestRepositoryWithActivity("repo4", true, true, time.Now()),  // Fork + Archived
			},
			filters: ports.FilterOptions{
				IncludeForks:    true,
				IncludeArchived: true,
				IncludePrivate:  true,
				IncludePublic:   true,
			},
			includeForks: true,
			expected:     4,
		},
		{
			name: "exclude_forks",
			repositories: []entities.Repository{
				createTestRepositoryWithActivity("repo1", false, false, time.Now()),
				createTestRepositoryWithActivity("repo2", true, false, time.Now()), // Fork - excluded
				createTestRepositoryWithActivity("repo3", false, true, time.Now()), // Archived
			},
			filters: ports.FilterOptions{
				IncludeForks:    false,
				IncludeArchived: true,
				IncludePrivate:  true,
				IncludePublic:   true,
			},
			includeForks: false,
			expected:     2,
		},
		{
			name: "exclude_archived",
			repositories: []entities.Repository{
				createTestRepositoryWithActivity("repo1", false, false, time.Now()),
				createTestRepositoryWithActivity("repo2", true, false, time.Now()), // Fork
				createTestRepositoryWithActivity("repo3", false, true, time.Now()), // Archived - excluded
			},
			filters: ports.FilterOptions{
				IncludeForks:    true,
				IncludeArchived: false,
				IncludePrivate:  true,
				IncludePublic:   true,
			},
			includeForks: true,
			expected:     2,
		},
		{
			name: "exclude_forks_and_archived",
			repositories: []entities.Repository{
				createTestRepositoryWithActivity("repo1", false, false, time.Now()),
				createTestRepositoryWithActivity("repo2", true, false, time.Now()), // Fork - excluded
				createTestRepositoryWithActivity("repo3", false, true, time.Now()), // Archived - excluded
				createTestRepositoryWithActivity("repo4", true, true, time.Now()),  // Fork + Archived - excluded
			},
			filters: ports.FilterOptions{
				IncludeForks:    false,
				IncludeArchived: false,
				IncludePrivate:  true,
				IncludePublic:   true,
			},
			includeForks: false,
			expected:     1, // Only repo1 should remain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create use case with minimal dependencies for testing filtering
			mockProvider := &SharedMockRepositoryProvider{}
			mockGitOps := &SharedMockGitOperations{}
			mockLogger := &SharedMockLogger{}

			useCase := NewFetchSourceRepositoriesUseCase(mockProvider, mockGitOps, mockLogger)

			// Apply filters
			filtered := useCase.applyFilters(tt.repositories, tt.filters, tt.includeForks)

			// Verify expected count
			require.Equal(t, tt.expected, len(filtered), "Filtered repository count mismatch")
		})
	}
}

func TestFetchSourceRepositoriesUseCase_matchesPatterns(t *testing.T) {
	tests := []struct {
		name            string
		repositoryName  string
		includePatterns []string
		excludePatterns []string
		expected        bool
	}{
		{
			name:            "no_patterns_includes_all",
			repositoryName:  "any-repo",
			includePatterns: []string{},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "exact_include_match",
			repositoryName:  "specific-repo",
			includePatterns: []string{"specific-repo"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "wildcard_include_match",
			repositoryName:  "any-repo",
			includePatterns: []string{"*"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "exact_exclude_match",
			repositoryName:  "excluded-repo",
			includePatterns: []string{"*"},
			excludePatterns: []string{"excluded-repo"},
			expected:        false,
		},
		{
			name:            "include_match_but_excluded",
			repositoryName:  "test-repo",
			includePatterns: []string{"test-repo"},
			excludePatterns: []string{"test-repo"},
			expected:        false, // Exclude takes precedence
		},
		{
			name:            "no_include_match",
			repositoryName:  "other-repo",
			includePatterns: []string{"specific-repo"},
			excludePatterns: []string{},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create use case with minimal dependencies for testing pattern matching
			mockProvider := &SharedMockRepositoryProvider{}
			mockGitOps := &SharedMockGitOperations{}
			mockLogger := &SharedMockLogger{}

			useCase := NewFetchSourceRepositoriesUseCase(mockProvider, mockGitOps, mockLogger)

			// Test pattern matching
			result := useCase.matchesPatterns(tt.repositoryName, tt.includePatterns, tt.excludePatterns)

			require.Equal(t, tt.expected, result, "Pattern matching result mismatch")
		})
	}
}

// Helper functions are defined in mocks_test.go to avoid duplication
