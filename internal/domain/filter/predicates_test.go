// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package filter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/filter"
)

// Helper to create a test repository (without test assertions).
func createTestRepo(name string, isPrivate, isFork, isArchived bool, lastActivity time.Time) entities.Repository {
	builder := entities.NewRepositoryBuilder()

	builder, _ = builder.WithName(name)
	builder, _ = builder.WithHTTPSURL("https://example.com/" + name + ".git")
	builder, _ = builder.WithSSHURL("git@example.com:" + name + ".git")

	builder = builder.
		WithPrivate(isPrivate).
		WithFork(isFork).
		WithArchived(isArchived).
		WithLastActivityAt(lastActivity)

	repo, _ := builder.Build()

	return repo
}

// Helper to create a test repository with test assertions.
func createTestRepoT(t *testing.T, name string, isPrivate, isFork, isArchived bool, lastActivity time.Time) entities.Repository {
	t.Helper()

	builder := entities.NewRepositoryBuilder()

	builder, err := builder.WithName(name)
	require.NoError(t, err)

	// Add required URLs for valid repository
	builder, err = builder.WithHTTPSURL("https://example.com/" + name + ".git")
	require.NoError(t, err)

	builder, err = builder.WithSSHURL("git@example.com:" + name + ".git")
	require.NoError(t, err)

	builder = builder.
		WithPrivate(isPrivate).
		WithFork(isFork).
		WithArchived(isArchived).
		WithLastActivityAt(lastActivity)

	repo, err := builder.Build()
	require.NoError(t, err)

	return repo
}

func TestPredicateComposition(t *testing.T) {
	t.Parallel()

	now := time.Now()
	oldActivity := now.Add(-365 * 24 * time.Hour)

	repos := []entities.Repository{
		createTestRepoT(t, "public-active", false, false, false, now),
		createTestRepoT(t, "private-active", true, false, false, now),
		createTestRepoT(t, "public-fork", false, true, false, now),
		createTestRepoT(t, "archived", false, false, true, now),
		createTestRepoT(t, "old-repo", false, false, false, oldActivity),
	}

	tests := []struct {
		name      string
		predicate filter.Predicate
		expected  []string
	}{
		{
			name:      "all public repos",
			predicate: filter.IsPublic,
			expected:  []string{"public-active", "public-fork", "archived", "old-repo"},
		},
		{
			name:      "all private repos",
			predicate: filter.IsPrivate,
			expected:  []string{"private-active"},
		},
		{
			name:      "not archived",
			predicate: filter.NotArchived,
			expected:  []string{"public-active", "private-active", "public-fork", "old-repo"},
		},
		{
			name:      "not fork",
			predicate: filter.NotFork,
			expected:  []string{"public-active", "private-active", "archived", "old-repo"},
		},
		{
			name:      "public AND not fork",
			predicate: filter.All(filter.IsPublic, filter.NotFork),
			expected:  []string{"public-active", "archived", "old-repo"},
		},
		{
			name:      "public AND not fork AND not archived",
			predicate: filter.All(filter.IsPublic, filter.NotFork, filter.NotArchived),
			expected:  []string{"public-active", "old-repo"},
		},
		{
			name:      "public OR private fork",
			predicate: filter.Any(filter.IsPublic, filter.All(filter.IsPrivate, filter.IsFork)),
			expected:  []string{"public-active", "public-fork", "archived", "old-repo"},
		},
		{
			name:      "active within 30 days",
			predicate: filter.ActiveWithin(30 * 24 * time.Hour),
			expected:  []string{"public-active", "private-active", "public-fork", "archived"},
		},
		{
			name:      "NOT (fork OR archived)",
			predicate: filter.Not(filter.Any(filter.IsFork, filter.IsArchived)),
			expected:  []string{"public-active", "private-active", "old-repo"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			filtered := filter.Filter(repos, testCase.predicate)

			names := make([]string, len(filtered))
			for i, repo := range filtered {
				names[i] = repo.Name()
			}

			assert.ElementsMatch(t, testCase.expected, names)
		})
	}
}

func TestNamePredicates(t *testing.T) {
	t.Parallel()

	repos := []entities.Repository{
		createTestRepoT(t, "api-service", false, false, false, time.Now()),
		createTestRepoT(t, "web-app", false, false, false, time.Now()),
		createTestRepoT(t, "lib-utils", false, false, false, time.Now()),
		createTestRepoT(t, "test-runner", false, false, false, time.Now()),
	}

	tests := []struct {
		name      string
		predicate filter.Predicate
		expected  []string
	}{
		{
			name:      "exact name match",
			predicate: filter.HasName("web-app"),
			expected:  []string{"web-app"},
		},
		{
			name:      "name contains 'app'",
			predicate: filter.NameContains("app"),
			expected:  []string{"web-app"},
		},
		{
			name:      "name prefix 'lib'",
			predicate: filter.NamePrefix("lib"),
			expected:  []string{"lib-utils"},
		},
		{
			name:      "name suffix 'service'",
			predicate: filter.NameSuffix("service"),
			expected:  []string{"api-service"},
		},
		{
			name:      "name in list",
			predicate: filter.NameIn([]string{"web-app", "test-runner"}),
			expected:  []string{"web-app", "test-runner"},
		},
		{
			name:      "name not in list",
			predicate: filter.NameNotIn([]string{"web-app", "test-runner"}),
			expected:  []string{"api-service", "lib-utils"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			filtered := filter.Filter(repos, testCase.predicate)

			names := make([]string, len(filtered))
			for i, repo := range filtered {
				names[i] = repo.Name()
			}

			assert.ElementsMatch(t, testCase.expected, names)
		})
	}
}

func TestFilterFunction(t *testing.T) {
	t.Parallel()

	repos := []entities.Repository{
		createTestRepoT(t, "repo1", false, false, false, time.Now()),
		createTestRepoT(t, "repo2", true, false, false, time.Now()),
		createTestRepoT(t, "repo3", false, true, false, time.Now()),
	}

	t.Run("nil predicate returns copy", func(t *testing.T) {
		t.Parallel()

		filtered := filter.Filter(repos, nil)
		assert.Len(t, filtered, 3)
		assert.Equal(t, repos, filtered)
		// Verify it's a copy, not the same slice
		assert.NotSame(t, &repos[0], &filtered[0])
	})

	t.Run("filter public repos", func(t *testing.T) {
		t.Parallel()

		filtered := filter.Filter(repos, filter.IsPublic)
		assert.Len(t, filtered, 2)
		assert.Equal(t, "repo1", filtered[0].Name())
		assert.Equal(t, "repo3", filtered[1].Name())
	})
}

func TestCountFunction(t *testing.T) {
	t.Parallel()

	repos := []entities.Repository{
		createTestRepoT(t, "public1", false, false, false, time.Now()),
		createTestRepoT(t, "public2", false, true, false, time.Now()),
		createTestRepoT(t, "private1", true, false, false, time.Now()),
	}

	assert.Equal(t, 3, filter.Count(repos, nil), "nil predicate should count all")
	assert.Equal(t, 2, filter.Count(repos, filter.IsPublic), "should count public repos")
	assert.Equal(t, 1, filter.Count(repos, filter.IsPrivate), "should count private repos")
	assert.Equal(t, 1, filter.Count(repos, filter.IsFork), "should count forks")
}

func TestPartitionFunction(t *testing.T) {
	t.Parallel()

	repos := []entities.Repository{
		createTestRepoT(t, "public1", false, false, false, time.Now()),
		createTestRepoT(t, "public2", false, false, false, time.Now()),
		createTestRepoT(t, "private1", true, false, false, time.Now()),
		createTestRepoT(t, "private2", true, false, false, time.Now()),
	}

	matching, notMatching := filter.Partition(repos, filter.IsPublic)

	require.Len(t, matching, 2)
	require.Len(t, notMatching, 2)

	assert.Equal(t, "public1", matching[0].Name())
	assert.Equal(t, "public2", matching[1].Name())
	assert.Equal(t, "private1", notMatching[0].Name())
	assert.Equal(t, "private2", notMatching[1].Name())
}

func TestComplexPredicateComposition(t *testing.T) {
	t.Parallel()

	// Create a complex predicate:
	// (Public AND (NOT Fork OR Has Description)) OR (Private AND Active in last 7 days)

	now := time.Now()
	weekOld := now.AddDate(0, 0, -8)

	repos := []entities.Repository{
		createTestRepoT(t, "public-notfork", false, false, false, now), // matches: public AND not fork
		createTestRepoT(t, "public-fork", false, true, false, now),     // no match: public but fork without desc
		createTestRepoT(t, "private-recent", true, false, false, now),  // matches: private AND recent
		createTestRepoT(t, "private-old", true, false, false, weekOld), // no match: private but old
		createTestRepoT(t, "public-archived", false, false, true, now), // matches: public AND not fork (archived doesn't matter)
	}

	predicate := filter.Any(
		filter.All(
			filter.IsPublic,
			filter.Any(
				filter.NotFork,
				filter.HasDescription, // none have descriptions in our test
			),
		),
		filter.All(
			filter.IsPrivate,
			filter.ActiveWithin(7*24*time.Hour),
		),
	)

	filtered := filter.Filter(repos, predicate)

	expectedNames := []string{"public-notfork", "private-recent", "public-archived"}

	actualNames := make([]string, len(filtered))
	for i, repo := range filtered {
		actualNames[i] = repo.Name()
	}

	assert.ElementsMatch(t, expectedNames, actualNames)
}

func TestNonePredicateComposition(t *testing.T) {
	t.Parallel()

	repos := []entities.Repository{
		createTestRepoT(t, "normal", false, false, false, time.Now()),
		createTestRepoT(t, "fork", false, true, false, time.Now()),
		createTestRepoT(t, "archived", false, false, true, time.Now()),
		createTestRepoT(t, "fork-archived", false, true, true, time.Now()),
	}

	// None(fork, archived) = NOT(fork OR archived)
	predicate := filter.None(filter.IsFork, filter.IsArchived)

	filtered := filter.Filter(repos, predicate)

	require.Len(t, filtered, 1)
	assert.Equal(t, "normal", filtered[0].Name())
}

func BenchmarkPredicates(b *testing.B) {
	// Create test data
	repos := make([]entities.Repository, 1000)
	for idx := range 1000 {
		repos[idx] = createTestRepo(
			fmt.Sprintf("repo-%d", idx),
			idx%2 == 0,                     // half private
			idx%3 == 0,                     // third are forks
			idx%10 == 0,                    // tenth are archived
			time.Now().AddDate(0, 0, -idx), // varying activity
		)
	}

	b.Run("simple predicate", func(b *testing.B) {
		for range b.N {
			_ = filter.Filter(repos, filter.IsPublic)
		}
	})

	b.Run("composite AND predicate", func(b *testing.B) {
		predicate := filter.All(filter.IsPublic, filter.NotFork, filter.NotArchived)

		b.ResetTimer()

		for range b.N {
			_ = filter.Filter(repos, predicate)
		}
	})

	b.Run("complex nested predicate", func(b *testing.B) {
		predicate := filter.Any(
			filter.All(filter.IsPublic, filter.NotFork),
			filter.All(filter.IsPrivate, filter.ActiveWithin(30*24*time.Hour)),
		)

		b.ResetTimer()

		for range b.N {
			_ = filter.Filter(repos, predicate)
		}
	})
}
