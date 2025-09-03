// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package filter_test

import (
	"fmt"
	"time"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/filter"
)

// Helper to create a test repository for examples.
func createExampleRepo(name string, isPrivate, isFork, isArchived bool, lastActivity time.Time) entities.Repository {
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

// ExampleFilter demonstrates basic filtering.
func ExampleFilter() {
	repos := []entities.Repository{
		createExampleRepo("public-repo", false, false, false, time.Now()),
		createExampleRepo("private-repo", true, false, false, time.Now()),
		createExampleRepo("archived-repo", false, false, true, time.Now()),
		createExampleRepo("fork-repo", false, true, false, time.Now()),
	}

	// Filter to get only public repos
	publicRepos := filter.Filter(repos, filter.IsPublic)
	fmt.Printf("Public repos: %d\n", len(publicRepos))
	// Output: Public repos: 3
}

// ExampleAll demonstrates combining predicates with AND logic.
func ExampleAll() {
	repos := []entities.Repository{
		createExampleRepo("active-public", false, false, false, time.Now()),
		createExampleRepo("active-private", true, false, false, time.Now()),
		createExampleRepo("archived-public", false, false, true, time.Now()),
		createExampleRepo("old-public", false, false, false, time.Now().AddDate(-1, 0, 0)),
	}

	// Get public repos that are not archived and active within 30 days
	predicate := filter.All(
		filter.IsPublic,
		filter.NotArchived,
		filter.ActiveWithin(30*24*time.Hour),
	)

	filtered := filter.Filter(repos, predicate)
	fmt.Printf("Active public repos: %d\n", len(filtered))
	// Output: Active public repos: 1
}

// ExampleAny demonstrates combining predicates with OR logic.
func ExampleAny() {
	repos := []entities.Repository{
		createExampleRepo("api-service", false, false, false, time.Now()),
		createExampleRepo("web-app", false, false, false, time.Now()),
		createExampleRepo("cli-tool", false, false, false, time.Now()),
		createExampleRepo("docs", false, false, false, time.Now()),
	}

	// Get repos that are either named "api-service" or contain "app"
	predicate := filter.Any(
		filter.HasName("api-service"),
		filter.NameContains("app"),
	)

	filtered := filter.Filter(repos, predicate)
	fmt.Printf("Matching repos: %d\n", len(filtered))
	// Output: Matching repos: 2
}

// ExamplePartition demonstrates splitting repos into two groups.
func ExamplePartition() {
	repos := []entities.Repository{
		createExampleRepo("active-1", false, false, false, time.Now()),
		createExampleRepo("active-2", false, false, false, time.Now()),
		createExampleRepo("old-1", false, false, false, time.Now().AddDate(-1, 0, 0)),
		createExampleRepo("old-2", false, false, false, time.Now().AddDate(-1, 0, 0)),
	}

	// Split into active and inactive repos
	active, inactive := filter.Partition(repos, filter.ActiveWithin(30*24*time.Hour))

	fmt.Printf("Active: %d, Inactive: %d\n", len(active), len(inactive))
	// Output: Active: 2, Inactive: 2
}
