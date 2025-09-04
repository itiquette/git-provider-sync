// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package filter

import (
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// Predicate is a pure function that tests if a repository matches certain criteria
// Pure functions with no side effects make testing and composition easier.
type Predicate func(entities.Repository) bool

// All returns a predicate that matches when ALL provided predicates match (AND logic)
// Example: All(IsPublic, NotArchived, HasActivity).
func All(predicates ...Predicate) Predicate {
	return func(repo entities.Repository) bool {
		for _, p := range predicates {
			if !p(repo) {
				return false
			}
		}

		return true
	}
}

// Any returns a predicate that matches when ANY provided predicate matches (OR logic)
// Example: Any(HasTopic("go"), HasTopic("golang")).
func Any(predicates ...Predicate) Predicate {
	return func(repo entities.Repository) bool {
		for _, p := range predicates {
			if p(repo) {
				return true
			}
		}

		return false
	}
}

// Not returns a predicate that inverts the result of the given predicate
// Example: Not(IsArchived).
func Not(predicate Predicate) Predicate {
	return func(repo entities.Repository) bool {
		return !predicate(repo)
	}
}

// None returns a predicate that matches when NONE of the predicates match
// Equivalent to Not(Any(...)) but more explicit.
func None(predicates ...Predicate) Predicate {
	return Not(Any(predicates...))
}

// IsPublic returns true if the repository is public.
func IsPublic(repo entities.Repository) bool {
	return !repo.IsPrivate()
}

// IsPrivate returns true if the repository is private.
func IsPrivate(repo entities.Repository) bool {
	return repo.IsPrivate()
}

// IsArchived returns true if the repository is archived.
func IsArchived(repo entities.Repository) bool {
	return repo.IsArchived()
}

// NotArchived returns true if the repository is not archived.
func NotArchived(repo entities.Repository) bool {
	return !repo.IsArchived()
}

// IsFork returns true if the repository is a fork.
func IsFork(repo entities.Repository) bool {
	return repo.IsFork()
}

// NotFork returns true if the repository is not a fork.
func NotFork(repo entities.Repository) bool {
	return !repo.IsFork()
}

// HasName returns a predicate that checks if the repository has the exact name.
func HasName(name string) Predicate {
	return func(repo entities.Repository) bool {
		return repo.Name() == name
	}
}

// NameContains returns a predicate that checks if the repository name contains the substring.
func NameContains(substring string) Predicate {
	return func(repo entities.Repository) bool {
		return strings.Contains(repo.Name(), substring)
	}
}

// NamePrefix returns a predicate that checks if the repository name starts with the prefix.
func NamePrefix(prefix string) Predicate {
	return func(repo entities.Repository) bool {
		return strings.HasPrefix(repo.Name(), prefix)
	}
}

// NameSuffix returns a predicate that checks if the repository name ends with the suffix.
func NameSuffix(suffix string) Predicate {
	return func(repo entities.Repository) bool {
		return strings.HasSuffix(repo.Name(), suffix)
	}
}

// ActiveSince returns a predicate that checks if the repository was active since the given time.
func ActiveSince(since time.Time) Predicate {
	return func(repo entities.Repository) bool {
		lastActivity := repo.LastActivityAt()

		return lastActivity.IsZero() || lastActivity.After(since) || lastActivity.Equal(since)
	}
}

// ActiveWithin returns a predicate that checks if the repository was active within the duration.
func ActiveWithin(duration time.Duration) Predicate {
	return ActiveSince(time.Now().Add(-duration))
}

// InactiveSince returns a predicate that checks if the repository has been inactive since the given time.
func InactiveSince(since time.Time) Predicate {
	return func(repo entities.Repository) bool {
		lastActivity := repo.LastActivityAt()

		return !lastActivity.IsZero() && lastActivity.Before(since)
	}
}

// HasDescription returns a predicate that checks if the repository has a non-empty description.
func HasDescription(repo entities.Repository) bool {
	return repo.Description() != ""
}

// DescriptionContains returns a predicate that checks if the description contains the substring.
func DescriptionContains(substring string) Predicate {
	return func(repo entities.Repository) bool {
		return strings.Contains(repo.Description(), substring)
	}
}

// NameIn returns a predicate that checks if the repository name is in the provided list.
func NameIn(names []string) Predicate {
	// Create a map for O(1) lookup
	nameSet := make(map[string]bool, len(names))
	for _, name := range names {
		nameSet[name] = true
	}

	return func(repo entities.Repository) bool {
		return nameSet[repo.Name()]
	}
}

// NameNotIn returns a predicate that checks if the repository name is NOT in the provided list.
func NameNotIn(names []string) Predicate {
	return Not(NameIn(names))
}

// Filter applies the predicate to filter a slice of repositories
// is a pure function that returns a new slice without modifying the input.
func Filter(repos []entities.Repository, predicate Predicate) []entities.Repository {
	if predicate == nil {
		// Return a copy to maintain immutability
		result := make([]entities.Repository, len(repos))
		copy(result, repos)

		return result
	}

	result := make([]entities.Repository, 0, len(repos))

	for _, repo := range repos {
		if predicate(repo) {
			result = append(result, repo)
		}
	}

	return result
}

// Count returns the number of repositories that match the predicate
// is a pure function with no side effects.
func Count(repos []entities.Repository, predicate Predicate) int {
	if predicate == nil {
		return len(repos)
	}

	count := 0

	for _, repo := range repos {
		if predicate(repo) {
			count++
		}
	}

	return count
}

// Partition splits repositories into two slices: matching and not matching the predicate
// is a pure function that returns new slices without modifying the input.
func Partition(repos []entities.Repository, predicate Predicate) ([]entities.Repository, []entities.Repository) {
	matching := make([]entities.Repository, 0, len(repos))
	notMatching := make([]entities.Repository, 0, len(repos))

	for _, repo := range repos {
		if predicate(repo) {
			matching = append(matching, repo)
		} else {
			notMatching = append(notMatching, repo)
		}
	}

	return matching, notMatching
}
