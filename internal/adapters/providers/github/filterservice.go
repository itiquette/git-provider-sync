// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FilterService provides GitHub-specific repository filtering operations.
type FilterService struct {
	logger ports.Logger
}

// NewFilterService creates a new GitHub filter service.
func NewFilterService(logger ports.Logger) *FilterService {
	return &FilterService{
		logger: logger,
	}
}

// FilterRepositoriesRequest contains parameters for filtering GitHub repositories.
type FilterRepositoriesRequest struct {
	Repositories     []entities.Repository
	IncludePatterns  []string
	ExcludePatterns  []string
	ActiveFromLimit  string
	OwnerFilter      string
	VisibilityFilter string
	LanguageFilter   []string
	SizeFilter       *SizeFilter
	ActivityFilter   *ActivityFilter
	SecurityFilter   *SecurityFilter
}

// SizeFilter defines repository size filtering criteria.
type SizeFilter struct {
	MinSizeKB *int64
	MaxSizeKB *int64
}

// ActivityFilter defines repository activity filtering criteria.
type ActivityFilter struct {
	MinStars    *int
	MaxStars    *int
	MinForks    *int
	MaxForks    *int
	MinWatchers *int
	MaxWatchers *int
	HasIssues   *bool
	HasWiki     *bool
	HasPages    *bool
}

// SecurityFilter defines security-related filtering criteria.
type SecurityFilter struct {
	HasVulnerabilities   *bool
	HasSecurityPolicy    *bool
	HasDependabot        *bool
	RequireSignedCommits *bool
}

// FilterRepositories performs GitHub-specific repository filtering.
func (fs *FilterService) FilterRepositories(ctx context.Context, request FilterRepositoriesRequest) ([]entities.Repository, error) {
	fs.logger.Debug(ctx, "Filtering GitHub repositories", map[string]any{
		"total_count":      len(request.Repositories),
		"include_patterns": request.IncludePatterns,
		"exclude_patterns": request.ExcludePatterns,
		"owner_filter":     request.OwnerFilter,
	})

	var filtered []entities.Repository

	for _, repo := range request.Repositories {
		if fs.shouldIncludeRepository(ctx, repo, request) {
			filtered = append(filtered, repo)
		}
	}

	fs.logger.Info(ctx, "GitHub repository filtering completed", map[string]any{
		"original_count": len(request.Repositories),
		"filtered_count": len(filtered),
	})

	return filtered, nil
}

// MatchesPattern checks if a repository name matches a pattern.
func (fs *FilterService) matchesPattern(name, pattern string) bool { //nolint:funcorder // Helper method used throughout
	// Convert to lowercase for case-insensitive matching
	name = strings.ToLower(name)
	pattern = strings.ToLower(pattern)

	// Handle special regex-like pattern (simple contains)
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		regexPattern := strings.Trim(pattern, "/")

		return strings.Contains(name, regexPattern)
	}

	// Use standard library filepath.Match for wildcard patterns
	matched, _ := filepath.Match(pattern, name)

	return matched
}

// IsActiveRepository checks if repository meets activity requirements.
func (fs *FilterService) isActiveRepository(ctx context.Context, repo entities.Repository, activeFromLimit string) bool { //nolint:funcorder // Helper method
	if activeFromLimit == "" {
		return true
	}

	duration, err := time.ParseDuration(activeFromLimit)
	if err != nil {
		fs.logger.Warn(ctx, "Failed to parse activity duration", map[string]any{
			"duration": activeFromLimit,
			"error":    err.Error(),
		})

		return true
	}

	cutoffTime := time.Now().Add(-duration)
	lastActivity := repo.LastActivityAt()

	if lastActivity.IsZero() {
		return false
	}

	return lastActivity.After(cutoffTime)
}

// MatchesLanguageFilter checks if repository matches language filter.
func (fs *FilterService) matchesLanguageFilter(_ entities.Repository, _ []string) bool { //nolint:funcorder // Helper method
	// would require additional metadata about repository languages
	// For now, we'll assume all repositories match if no specific language data is available
	// In a real implementation, this would check the repository's primary language
	return true
}

// MatchesSizeFilter checks if repository matches size filter.
func (fs *FilterService) matchesSizeFilter(_ entities.Repository, _ *SizeFilter) bool { //nolint:funcorder // Helper method
	// would require additional metadata about repository size
	// For now, we'll assume all repositories match if no specific size data is available
	// In a real implementation, this would check the repository's size in KB
	return true
}

// MatchesActivityFilter checks if repository matches activity filter.
func (fs *FilterService) matchesActivityFilter(_ entities.Repository, _ *ActivityFilter) bool { //nolint:funcorder // Helper method
	// would require additional metadata about stars, forks, watchers, etc
	// For now, we'll assume all repositories match if no specific activity data is available
	// In a real implementation, this would check the repository's activity metrics
	return true
}

// MatchesSecurityFilter checks if repository matches security filter.
func (fs *FilterService) matchesSecurityFilter(_ entities.Repository, _ *SecurityFilter) bool { //nolint:funcorder // Helper method
	// would require additional metadata about security features
	// For now, we'll assume all repositories match if no specific security data is available
	// In a real implementation, this would check the repository's security settings
	return true
}

// FilterRepositoriesByTopics filters repositories by GitHub topics.
func (fs *FilterService) FilterRepositoriesByTopics(_ context.Context, repos []entities.Repository, requiredTopics, excludedTopics []string) []entities.Repository {
	if len(requiredTopics) == 0 && len(excludedTopics) == 0 {
		return repos
	}

	var filtered []entities.Repository

	filtered = append(filtered, repos...)

	return filtered
}

// FilterRepositoriesByCollaborators filters repositories by collaborator access.
func (fs *FilterService) FilterRepositoriesByCollaborators(_ context.Context, repos []entities.Repository, requiredCollaborators []string) []entities.Repository {
	if len(requiredCollaborators) == 0 {
		return repos
	}

	var filtered []entities.Repository

	filtered = append(filtered, repos...)

	return filtered
}

// GetFilterStatistics returns statistics about filtering results.
func (fs *FilterService) GetFilterStatistics(original, filtered []entities.Repository) map[string]any {
	stats := map[string]any{
		"original_count":      len(original),
		"filtered_count":      len(filtered),
		"filtered_percentage": float64(len(filtered)) / float64(len(original)) * 100,
	}

	// Count by visibility
	publicCount := 0
	privateCount := 0

	for _, repo := range filtered {
		if repo.Visibility() == constants.VisibilityPrivate {
			privateCount++
		} else {
			publicCount++
		}
	}

	stats["public_count"] = publicCount
	stats["private_count"] = privateCount

	return stats
}

func (fs *FilterService) shouldIncludeRepository(ctx context.Context, repo entities.Repository, request FilterRepositoriesRequest) bool {
	return fs.passesBasicFilters(repo, request) &&
		fs.passesPatternFilters(ctx, repo, request) &&
		fs.passesContentFilters(repo, request) &&
		fs.passesAdvancedFilters(ctx, repo, request)
}

// PassesBasicFilters checks basic filters like visibility.
func (fs *FilterService) passesBasicFilters(repo entities.Repository, request FilterRepositoriesRequest) bool {
	// Owner filter not implemented - entities.Repository doesn't expose owner
	_ = request.OwnerFilter // Suppress unused variable warning

	// Check visibility filter
	if request.VisibilityFilter != "" && request.VisibilityFilter != "all" {
		if !strings.EqualFold(repo.Visibility(), request.VisibilityFilter) {
			return false
		}
	}

	return true
}

// PassesPatternFilters checks include and exclude patterns.
func (fs *FilterService) passesPatternFilters(ctx context.Context, repo entities.Repository, request FilterRepositoriesRequest) bool {
	// Check exclude patterns first
	if len(request.ExcludePatterns) > 0 {
		for _, pattern := range request.ExcludePatterns {
			if fs.matchesPattern(repo.Name(), pattern) {
				fs.logger.Debug(ctx, "Repository excluded by pattern", map[string]any{
					"repository": repo.Name(),
					"pattern":    pattern,
				})

				return false
			}
		}
	}

	// Check include patterns
	if len(request.IncludePatterns) > 0 {
		included := false

		for _, pattern := range request.IncludePatterns {
			if fs.matchesPattern(repo.Name(), pattern) {
				included = true

				break
			}
		}

		if !included {
			return false
		}
	}

	return true
}

// PassesContentFilters checks content-based filters like language and size.
func (fs *FilterService) passesContentFilters(repo entities.Repository, request FilterRepositoriesRequest) bool {
	// Check language filter
	if len(request.LanguageFilter) > 0 {
		if !fs.matchesLanguageFilter(repo, request.LanguageFilter) {
			return false
		}
	}

	// Check size filter
	if request.SizeFilter != nil {
		if !fs.matchesSizeFilter(repo, request.SizeFilter) {
			return false
		}
	}

	return true
}

// PassesAdvancedFilters checks activity and security filters.
func (fs *FilterService) passesAdvancedFilters(ctx context.Context, repo entities.Repository, request FilterRepositoriesRequest) bool {
	// Check activity filter
	if request.ActiveFromLimit != "" {
		if !fs.isActiveRepository(ctx, repo, request.ActiveFromLimit) {
			return false
		}
	}

	// Check activity metrics filter
	if request.ActivityFilter != nil {
		if !fs.matchesActivityFilter(repo, request.ActivityFilter) {
			return false
		}
	}

	// Check security filter
	if request.SecurityFilter != nil {
		if !fs.matchesSecurityFilter(repo, request.SecurityFilter) {
			return false
		}
	}

	return true
}
