// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// testLogger is a simple no-op logger for testing.
type testLogger struct{}

func (l testLogger) Trace(_ context.Context, _ string, _ map[string]any) {}
func (l testLogger) Debug(_ context.Context, _ string, _ map[string]any) {}
func (l testLogger) Info(_ context.Context, _ string, _ map[string]any)  {}
func (l testLogger) Warn(_ context.Context, _ string, _ map[string]any)  {}
func (l testLogger) Error(_ context.Context, _ string, _ map[string]any) {}
func (l testLogger) Fatal(_ context.Context, _ string, _ map[string]any) {}
func (l testLogger) IsLevelEnabled(_ ports.LogLevel) bool                { return true }

func TestFilterRepositoriesUseCase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		request         FilterRequest
		expectedCount   int
		expectedSuccess bool
	}{
		{
			name: "empty repositories list",
			request: FilterRequest{
				Repositories:         []entities.Repository{},
				FilterOptions:        ports.FilterOptions{},
				IncludedRepositories: []string{},
				ExcludedRepositories: []string{},
				ActiveFromLimit:      "",
			},
			expectedCount:   0,
			expectedSuccess: true,
		},
		{
			name: "no filtering applied",
			request: FilterRequest{
				Repositories: []entities.Repository{
					createFilterTestRepository("repo1", time.Now()),
					createFilterTestRepository("repo2", time.Now()),
				},
				FilterOptions: ports.FilterOptions{
					IncludeForks:    true,
					IncludeArchived: true,
					IncludePublic:   true,
					IncludePrivate:  true,
				},
				IncludedRepositories: []string{},
				ExcludedRepositories: []string{},
				ActiveFromLimit:      "",
			},
			expectedCount:   2,
			expectedSuccess: true,
		},
		{
			name: "activity filtering - recent repositories",
			request: FilterRequest{
				Repositories: []entities.Repository{
					createFilterTestRepository("repo1", time.Now().Add(-1*time.Hour)),
					createFilterTestRepository("repo2", time.Now().Add(-48*time.Hour)),
				},
				FilterOptions: ports.FilterOptions{
					IncludeForks:    true,
					IncludeArchived: true,
					IncludePublic:   true,
					IncludePrivate:  true,
				},
				IncludedRepositories: []string{},
				ExcludedRepositories: []string{},
				ActiveFromLimit:      "-24h",
			},
			expectedCount:   1,
			expectedSuccess: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewFilterRepositoriesUseCase(testLogger{})
			ctx := context.Background()

			response, err := useCase.Execute(ctx, test.request)

			require.NoError(t, err)
			assert.Equal(t, test.expectedSuccess, response.Success)
			assert.Equal(t, test.expectedCount, response.FilteredCount)
			assert.Equal(t, len(test.request.Repositories), response.OriginalCount)
		})
	}
}

func TestFilterRepositoriesUseCase_filterByActivity(t *testing.T) {
	t.Parallel()

	useCase := NewFilterRepositoriesUseCase(testLogger{})
	ctx := context.Background()

	tests := []struct {
		name             string
		repositories     []entities.Repository
		activeFromLimit  string
		expectedFiltered int
		expectedSkipped  int
	}{
		{
			name: "no activity limit",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepository("repo2", time.Now().Add(-48*time.Hour)),
			},
			activeFromLimit:  "",
			expectedFiltered: 2,
			expectedSkipped:  0,
		},
		{
			name: "filter old repositories",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now().Add(-1*time.Hour)),
				createFilterTestRepository("repo2", time.Now().Add(-48*time.Hour)),
				createFilterTestRepository("repo3", time.Now().Add(-1*time.Minute)),
			},
			activeFromLimit:  "-24h",
			expectedFiltered: 2,
			expectedSkipped:  1,
		},
		{
			name: "invalid duration format",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
			},
			activeFromLimit:  "invalid-duration",
			expectedFiltered: 1,
			expectedSkipped:  0,
		},
		{
			name: "zero time repositories are included",
			repositories: []entities.Repository{
				createFilterTestRepositoryWithZeroTime("repo1"),
				createFilterTestRepository("repo2", time.Now().Add(-48*time.Hour)),
			},
			activeFromLimit:  "-24h",
			expectedFiltered: 1,
			expectedSkipped:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filtered, skipped := useCase.filterByActivity(ctx, test.repositories, test.activeFromLimit)

			assert.Len(t, filtered, test.expectedFiltered)
			assert.Equal(t, test.expectedSkipped, skipped)
		})
	}
}

func TestFilterRepositoriesUseCase_filterByIncludeExclude(t *testing.T) {
	t.Parallel()

	useCase := NewFilterRepositoriesUseCase(testLogger{})
	ctx := context.Background()

	tests := []struct {
		name                 string
		repositories         []entities.Repository
		includedRepositories []string
		excludedRepositories []string
		expectedFiltered     int
		expectedSkippedInc   int
		expectedSkippedExc   int
	}{
		{
			name: "no include/exclude lists",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepository("repo2", time.Now()),
			},
			includedRepositories: []string{},
			excludedRepositories: []string{},
			expectedFiltered:     2,
			expectedSkippedInc:   0,
			expectedSkippedExc:   0,
		},
		{
			name: "include specific repositories",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepository("repo2", time.Now()),
				createFilterTestRepository("repo3", time.Now()),
			},
			includedRepositories: []string{"repo1", "repo3"},
			excludedRepositories: []string{},
			expectedFiltered:     2,
			expectedSkippedInc:   1,
			expectedSkippedExc:   0,
		},
		{
			name: "exclude specific repositories",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepository("repo2", time.Now()),
				createFilterTestRepository("repo3", time.Now()),
			},
			includedRepositories: []string{},
			excludedRepositories: []string{"repo2"},
			expectedFiltered:     2,
			expectedSkippedInc:   0,
			expectedSkippedExc:   1,
		},
		{
			name: "include takes precedence over exclude",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepository("repo2", time.Now()),
			},
			includedRepositories: []string{"repo1"},
			excludedRepositories: []string{"repo1", "repo2"},
			expectedFiltered:     1,
			expectedSkippedInc:   1,
			expectedSkippedExc:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filtered, skippedInc, skippedExc := useCase.filterByIncludeExclude(
				ctx,
				test.repositories,
				test.includedRepositories,
				test.excludedRepositories,
			)

			assert.Len(t, filtered, test.expectedFiltered)
			assert.Equal(t, test.expectedSkippedInc, skippedInc)
			assert.Equal(t, test.expectedSkippedExc, skippedExc)
		})
	}
}

func TestFilterRepositoriesUseCase_shouldIncludeRepository(t *testing.T) {
	t.Parallel()

	useCase := NewFilterRepositoriesUseCase(testLogger{})

	tests := []struct {
		name                 string
		repoName             string
		includedRepositories []string
		excludedRepositories []string
		expected             bool
	}{
		{
			name:                 "no lists",
			repoName:             "repo1",
			includedRepositories: []string{},
			excludedRepositories: []string{},
			expected:             true,
		},
		{
			name:                 "in include list",
			repoName:             "repo1",
			includedRepositories: []string{"repo1", "repo2"},
			excludedRepositories: []string{},
			expected:             true,
		},
		{
			name:                 "not in include list",
			repoName:             "repo3",
			includedRepositories: []string{"repo1", "repo2"},
			excludedRepositories: []string{},
			expected:             false,
		},
		{
			name:                 "in exclude list",
			repoName:             "repo1",
			includedRepositories: []string{},
			excludedRepositories: []string{"repo1"},
			expected:             false,
		},
		{
			name:                 "include overrides exclude",
			repoName:             "repo1",
			includedRepositories: []string{"repo1"},
			excludedRepositories: []string{"repo1"},
			expected:             true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := useCase.shouldIncludeRepository(
				test.repoName,
				test.includedRepositories,
				test.excludedRepositories,
			)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFilterRepositoriesUseCase_filterByPatterns(t *testing.T) {
	t.Parallel()

	useCase := NewFilterRepositoriesUseCase(testLogger{})
	ctx := context.Background()

	tests := []struct {
		name             string
		repositories     []entities.Repository
		filterOptions    ports.FilterOptions
		expectedFiltered int
		expectedSkipped  int
	}{
		{
			name: "no patterns",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepository("repo2", time.Now()),
			},
			filterOptions: ports.FilterOptions{
				IncludePatterns: []string{},
				ExcludePatterns: []string{},
			},
			expectedFiltered: 2,
			expectedSkipped:  0,
		},
		{
			name: "include pattern matches",
			repositories: []entities.Repository{
				createFilterTestRepository("test-repo1", time.Now()),
				createFilterTestRepository("prod-repo2", time.Now()),
				createFilterTestRepository("test-repo3", time.Now()),
			},
			filterOptions: ports.FilterOptions{
				IncludePatterns: []string{"test-*"},
				ExcludePatterns: []string{},
			},
			expectedFiltered: 2,
			expectedSkipped:  1,
		},
		{
			name: "exclude pattern matches",
			repositories: []entities.Repository{
				createFilterTestRepository("test-repo1", time.Now()),
				createFilterTestRepository("prod-repo2", time.Now()),
				createFilterTestRepository("backup-repo3", time.Now()),
			},
			filterOptions: ports.FilterOptions{
				IncludePatterns: []string{},
				ExcludePatterns: []string{"backup-*"},
			},
			expectedFiltered: 2,
			expectedSkipped:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filtered, skipped := useCase.filterByPatterns(ctx, test.repositories, test.filterOptions)

			assert.Len(t, filtered, test.expectedFiltered)
			assert.Equal(t, test.expectedSkipped, skipped)
		})
	}
}

func TestFilterRepositoriesUseCase_matchesPatterns(t *testing.T) {
	t.Parallel()

	useCase := NewFilterRepositoriesUseCase(testLogger{})

	tests := []struct {
		name     string
		repoName string
		patterns []string
		expected bool
	}{
		{
			name:     "no patterns",
			repoName: "repo1",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "exact match",
			repoName: "repo1",
			patterns: []string{"repo1"},
			expected: true,
		},
		{
			name:     "wildcard match",
			repoName: "test-repo1",
			patterns: []string{"test-*"},
			expected: true,
		},
		{
			name:     "no match",
			repoName: "prod-repo1",
			patterns: []string{"test-*"},
			expected: false,
		},
		{
			name:     "multiple patterns - first matches",
			repoName: "test-repo1",
			patterns: []string{"test-*", "prod-*"},
			expected: true,
		},
		{
			name:     "multiple patterns - second matches",
			repoName: "prod-repo1",
			patterns: []string{"test-*", "prod-*"},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := useCase.matchesPatterns(test.repoName, test.patterns, []string{})

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFilterRepositoriesUseCase_matchPattern(t *testing.T) {
	t.Parallel()

	useCase := NewFilterRepositoriesUseCase(testLogger{})

	tests := []struct {
		name     string
		repoName string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			repoName: "repo1",
			pattern:  "repo1",
			expected: true,
		},
		{
			name:     "wildcard prefix",
			repoName: "test-repo",
			pattern:  "test-*",
			expected: true,
		},
		{
			name:     "wildcard suffix",
			repoName: "repo-test",
			pattern:  "*-test",
			expected: true,
		},
		{
			name:     "wildcard middle",
			repoName: "test-repo-prod",
			pattern:  "test-*-prod",
			expected: true,
		},
		{
			name:     "no match",
			repoName: "prod-repo",
			pattern:  "test-*",
			expected: false,
		},
		{
			name:     "empty pattern",
			repoName: "repo1",
			pattern:  "",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := useCase.matchPattern(test.repoName, test.pattern)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFilterRepositoriesUseCase_filterByAttributes(t *testing.T) {
	t.Parallel()

	useCase := NewFilterRepositoriesUseCase(testLogger{})
	ctx := context.Background()

	tests := []struct {
		name             string
		repositories     []entities.Repository
		filterOptions    ports.FilterOptions
		expectedFiltered int
	}{
		{
			name: "no attribute filtering",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepositoryWithAttributes("repo2", time.Now(), true, false),
			},
			filterOptions: ports.FilterOptions{
				IncludeForks:    true,
				IncludeArchived: true,
				IncludePublic:   true,
			},
			expectedFiltered: 2,
		},
		{
			name: "exclude forks",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepositoryWithAttributes("repo2", time.Now(), true, false),
				createFilterTestRepositoryWithAttributes("repo3", time.Now(), false, false),
			},
			filterOptions: ports.FilterOptions{
				IncludeForks:    false,
				IncludeArchived: true,
				IncludePublic:   true,
			},
			expectedFiltered: 2,
		},
		{
			name: "exclude archived",
			repositories: []entities.Repository{
				createFilterTestRepository("repo1", time.Now()),
				createFilterTestRepositoryWithAttributes("repo2", time.Now(), false, true),
				createFilterTestRepositoryWithAttributes("repo3", time.Now(), false, false),
			},
			filterOptions: ports.FilterOptions{
				IncludeForks:    true,
				IncludeArchived: false,
				IncludePublic:   true,
			},
			expectedFiltered: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filtered := useCase.filterByAttributes(ctx, test.repositories, test.filterOptions)

			assert.Len(t, filtered, test.expectedFiltered)
		})
	}
}

func TestFilterRepositoriesUseCase_shouldIncludeByAttributes(t *testing.T) {
	t.Parallel()

	useCase := NewFilterRepositoriesUseCase(testLogger{})

	tests := []struct {
		name          string
		repo          entities.Repository
		filterOptions ports.FilterOptions
		expected      bool
	}{
		{
			name:          "include all attributes",
			repo:          createFilterTestRepositoryWithAttributes("repo1", time.Now(), true, true),
			filterOptions: ports.FilterOptions{IncludeForks: true, IncludeArchived: true, IncludePublic: true},
			expected:      true,
		},
		{
			name:          "exclude fork",
			repo:          createFilterTestRepositoryWithAttributes("repo1", time.Now(), true, false),
			filterOptions: ports.FilterOptions{IncludeForks: false, IncludeArchived: true, IncludePublic: true},
			expected:      false,
		},
		{
			name:          "exclude archived",
			repo:          createFilterTestRepositoryWithAttributes("repo1", time.Now(), false, true),
			filterOptions: ports.FilterOptions{IncludeForks: true, IncludeArchived: false, IncludePublic: true},
			expected:      false,
		},
		{
			name:          "normal repository always included",
			repo:          createFilterTestRepositoryWithAttributes("repo1", time.Now(), false, false),
			filterOptions: ports.FilterOptions{IncludeForks: false, IncludeArchived: false, IncludePublic: true},
			expected:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := useCase.shouldIncludeByAttributes(test.repo, test.filterOptions)

			assert.Equal(t, test.expected, result)
		})
	}
}

// Helper functions for creating test repositories

func createFilterTestRepository(name string, lastActivity time.Time) entities.Repository {
	builder := entities.NewRepositoryBuilder()
	builder, _ = builder.WithName(name)
	builder = builder.WithLastActivityAt(lastActivity)
	builder, _ = builder.WithHTTPSURL("https://github.com/test/" + name + ".git")
	repo, _ := builder.Build()

	return repo
}

func createFilterTestRepositoryWithZeroTime(name string) entities.Repository {
	builder := entities.NewRepositoryBuilder()
	builder, _ = builder.WithName(name)
	builder, _ = builder.WithHTTPSURL("https://github.com/test/" + name + ".git")
	// Don't set LastActivityAt, so it remains zero time
	repo, _ := builder.Build()

	return repo
}

func createFilterTestRepositoryWithAttributes(name string, lastActivity time.Time, isFork, isArchived bool) entities.Repository {
	builder := entities.NewRepositoryBuilder()
	builder, _ = builder.WithName(name)
	builder = builder.WithLastActivityAt(lastActivity)
	builder, _ = builder.WithHTTPSURL("https://github.com/test/" + name + ".git")
	builder = builder.WithFork(isFork)
	builder = builder.WithArchived(isArchived)
	builder = builder.WithPrivate(false)
	repo, _ := builder.Build()

	return repo
}
