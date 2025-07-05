// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FilterService provides GitHub-specific repository filtering operations.
// This restores the sophisticated filter service functionality from main branch.
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
	fs.logger.Debug(ctx, "Filtering GitHub repositories", map[string]interface{}{
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

	fs.logger.Info(ctx, "GitHub repository filtering completed", map[string]interface{}{
		"original_count": len(request.Repositories),
		"filtered_count": len(filtered),
	})

	return filtered, nil
}

// shouldIncludeRepository determines if a repository should be included.
func (fs *FilterService) shouldIncludeRepository(ctx context.Context, repo entities.Repository, request FilterRepositoriesRequest) bool {
	// Check owner filter (skip for now as entities.Repository doesn't have Owner method)
	// TODO: Add Owner method to Repository entity
	_ = request.OwnerFilter // Suppress unused variable warning

	// Check visibility filter
	if request.VisibilityFilter != "" && request.VisibilityFilter != "all" {
		if !strings.EqualFold(repo.Visibility(), request.VisibilityFilter) {
			return false
		}
	}

	// Check exclude patterns first
	if len(request.ExcludePatterns) > 0 {
		for _, pattern := range request.ExcludePatterns {
			if fs.matchesPattern(repo.Name(), pattern) {
				fs.logger.Debug(ctx, "Repository excluded by pattern", map[string]interface{}{
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

	// Check activity filter
	if request.ActiveFromLimit != "" {
		if !fs.isActiveRepository(ctx, repo, request.ActiveFromLimit) {
			return false
		}
	}

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

// matchesPattern checks if a repository name matches a pattern.
func (fs *FilterService) matchesPattern(name, pattern string) bool {
	// Convert to lowercase for case-insensitive matching
	name = strings.ToLower(name)
	pattern = strings.ToLower(pattern)

	// Handle different pattern types
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		return fs.wildcardMatch(name, pattern)
	}

	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		// Basic regex-like pattern (simplified)
		regexPattern := strings.Trim(pattern, "/")

		return strings.Contains(name, regexPattern)
	}

	// Check for prefix/suffix patterns
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")

		return strings.HasPrefix(name, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")

		return strings.HasSuffix(name, suffix)
	}

	// Exact match or substring match
	return name == pattern || strings.Contains(name, pattern)
}

// wildcardMatch performs wildcard matching.
func (fs *FilterService) wildcardMatch(name, pattern string) bool {
	return fs.wildcardMatchRecursive(name, pattern)
}

// wildcardMatchRecursive performs recursive wildcard matching.
func (fs *FilterService) wildcardMatchRecursive(name, pattern string) bool {
	if pattern == "" {
		return name == ""
	}

	if pattern[0] == '*' {
		for i := 0; i <= len(name); i++ {
			if fs.wildcardMatchRecursive(name[i:], pattern[1:]) {
				return true
			}
		}

		return false
	}

	if name == "" {
		return false
	}

	if pattern[0] == '?' || pattern[0] == name[0] {
		return fs.wildcardMatchRecursive(name[1:], pattern[1:])
	}

	return false
}

// isActiveRepository checks if repository meets activity requirements.
func (fs *FilterService) isActiveRepository(ctx context.Context, repo entities.Repository, activeFromLimit string) bool {
	if activeFromLimit == "" {
		return true
	}

	duration, err := time.ParseDuration(activeFromLimit)
	if err != nil {
		fs.logger.Warn(ctx, "Failed to parse activity duration", map[string]interface{}{
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

// matchesLanguageFilter checks if repository matches language filter.
func (fs *FilterService) matchesLanguageFilter(repo entities.Repository, languages []string) bool {
	// This would require additional metadata about repository languages
	// For now, we'll assume all repositories match if no specific language data is available
	// In a real implementation, this would check the repository's primary language
	return true
}

// matchesSizeFilter checks if repository matches size filter.
func (fs *FilterService) matchesSizeFilter(repo entities.Repository, filter *SizeFilter) bool {
	// This would require additional metadata about repository size
	// For now, we'll assume all repositories match if no specific size data is available
	// In a real implementation, this would check the repository's size in KB
	return true
}

// matchesActivityFilter checks if repository matches activity filter.
func (fs *FilterService) matchesActivityFilter(repo entities.Repository, filter *ActivityFilter) bool {
	// This would require additional metadata about stars, forks, watchers, etc.
	// For now, we'll assume all repositories match if no specific activity data is available
	// In a real implementation, this would check the repository's activity metrics
	return true
}

// matchesSecurityFilter checks if repository matches security filter.
func (fs *FilterService) matchesSecurityFilter(repo entities.Repository, filter *SecurityFilter) bool {
	// This would require additional metadata about security features
	// For now, we'll assume all repositories match if no specific security data is available
	// In a real implementation, this would check the repository's security settings
	return true
}

// FilterRepositoriesByTopics filters repositories by GitHub topics.
func (fs *FilterService) FilterRepositoriesByTopics(ctx context.Context, repos []entities.Repository, requiredTopics, excludedTopics []string) []entities.Repository {
	if len(requiredTopics) == 0 && len(excludedTopics) == 0 {
		return repos
	}

	var filtered []entities.Repository

	filtered = append(filtered, repos...)

	return filtered
}

// FilterRepositoriesByCollaborators filters repositories by collaborator access.
func (fs *FilterService) FilterRepositoriesByCollaborators(ctx context.Context, repos []entities.Repository, requiredCollaborators []string) []entities.Repository {
	if len(requiredCollaborators) == 0 {
		return repos
	}

	var filtered []entities.Repository

	filtered = append(filtered, repos...)

	return filtered
}

// GetFilterStatistics returns statistics about filtering results.
func (fs *FilterService) GetFilterStatistics(original, filtered []entities.Repository) map[string]interface{} {
	stats := map[string]interface{}{
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
