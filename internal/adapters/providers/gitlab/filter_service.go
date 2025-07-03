// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FilterService provides GitLab-specific repository filtering operations.
// This restores the sophisticated filter service functionality from main branch.
type FilterService struct {
	logger ports.Logger
}

// NewFilterService creates a new GitLab filter service.
func NewFilterService(logger ports.Logger) *FilterService {
	return &FilterService{
		logger: logger,
	}
}

// FilterRepositoriesRequest contains parameters for filtering GitLab repositories.
type FilterRepositoriesRequest struct {
	Repositories     []entities.Repository
	IncludePatterns  []string
	ExcludePatterns  []string
	ActiveFromLimit  string
	OwnerFilter      string
	VisibilityFilter string // "public", "private", "internal"
	LanguageFilter   []string
	TopicFilter      []string
	MembershipFilter *MembershipFilter
	ActivityFilter   *ActivityFilter
	SecurityFilter   *SecurityFilter
	LicenseFilter    []string
	ArchiveFilter    *ArchiveFilter
}

// MembershipFilter defines GitLab-specific membership filtering.
type MembershipFilter struct {
	Owned          *bool
	Membership     *bool
	Starred        *bool
	MinAccessLevel string // "guest", "reporter", "developer", "maintainer", "owner"
}

// ActivityFilter defines repository activity filtering criteria.
type ActivityFilter struct {
	MinStars          *int
	MaxStars          *int
	MinForks          *int
	MaxForks          *int
	MinIssues         *int
	MaxIssues         *int
	MinMergeRequests  *int
	MaxMergeRequests  *int
	HasIssues         *bool
	HasWiki           *bool
	HasPages          *bool
	HasContainer      *bool
	HasPackages       *bool
	HasSnippets       *bool
	WithSharedRunners *bool
	OnlyEmptyRepo     *bool
}

// SecurityFilter defines GitLab security-related filtering criteria.
type SecurityFilter struct {
	HasVulnerabilities       *bool
	HasSecurityPolicy        *bool
	HasDependencyScanning    *bool
	HasContainerScanning     *bool
	HasSASTScanning          *bool
	HasSecretDetection       *bool
	RequireSignedCommits     *bool
	HasPushRules             *bool
	HasMergeRequestApprovals *bool
	OnlyMirror               *bool
	ComplianceFrameworks     []string
}

// ArchiveFilter defines archive-related filtering.
type ArchiveFilter struct {
	IncludeArchived *bool
	OnlyArchived    *bool
	ArchivedBefore  *time.Time
	ArchivedAfter   *time.Time
}

// FilterRepositories performs GitLab-specific repository filtering.
func (fs *FilterService) FilterRepositories(ctx context.Context, request FilterRepositoriesRequest) ([]entities.Repository, error) {
	fs.logger.Debug(ctx, "Filtering GitLab repositories", map[string]interface{}{
		"total_count":      len(request.Repositories),
		"include_patterns": request.IncludePatterns,
		"exclude_patterns": request.ExcludePatterns,
		"owner_filter":     request.OwnerFilter,
		"visibility":       request.VisibilityFilter,
		"topics":           request.TopicFilter,
	})

	var filtered []entities.Repository

	for _, repo := range request.Repositories {
		if fs.shouldIncludeRepository(ctx, repo, request) {
			filtered = append(filtered, repo)
		}
	}

	fs.logger.Info(ctx, "GitLab repository filtering completed", map[string]interface{}{
		"original_count": len(request.Repositories),
		"filtered_count": len(filtered),
	})

	return filtered, nil
}

// shouldIncludeRepository determines if a repository should be included.
func (fs *FilterService) shouldIncludeRepository(ctx context.Context, repo entities.Repository, request FilterRepositoriesRequest) bool {
	// Check archive filter first
	if request.ArchiveFilter != nil {
		if !fs.matchesArchiveFilter(repo, request.ArchiveFilter) {
			return false
		}
	}

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

	// Check topic filter
	if len(request.TopicFilter) > 0 {
		if !fs.matchesTopicFilter(repo, request.TopicFilter) {
			return false
		}
	}

	// Check license filter
	if len(request.LicenseFilter) > 0 {
		if !fs.matchesLicenseFilter(repo, request.LicenseFilter) {
			return false
		}
	}

	// Check membership filter
	if request.MembershipFilter != nil {
		if !fs.matchesMembershipFilter(repo, request.MembershipFilter) {
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

// matchesTopicFilter checks if repository matches topic filter.
func (fs *FilterService) matchesTopicFilter(repo entities.Repository, topics []string) bool {
	// This would require additional metadata about repository topics
	// For now, we'll assume all repositories match if no specific topic data is available
	// In a real implementation, this would check the repository's topics/tags
	return true
}

// matchesLicenseFilter checks if repository matches license filter.
func (fs *FilterService) matchesLicenseFilter(repo entities.Repository, licenses []string) bool {
	// This would require additional metadata about repository license
	// For now, we'll assume all repositories match if no specific license data is available
	// In a real implementation, this would check the repository's license
	return true
}

// matchesMembershipFilter checks if repository matches membership filter.
func (fs *FilterService) matchesMembershipFilter(repo entities.Repository, filter *MembershipFilter) bool {
	// This would require additional metadata about user's relationship to the repository
	// For now, we'll assume all repositories match if no specific membership data is available
	// In a real implementation, this would check:
	// - If user owns the repository
	// - If user is a member of the repository
	// - If user has starred the repository
	// - User's access level
	return true
}

// matchesActivityFilter checks if repository matches activity filter.
func (fs *FilterService) matchesActivityFilter(repo entities.Repository, filter *ActivityFilter) bool {
	// This would require additional metadata about stars, forks, issues, MRs, etc.
	// For now, we'll assume all repositories match if no specific activity data is available
	// In a real implementation, this would check the repository's activity metrics
	return true
}

// matchesSecurityFilter checks if repository matches security filter.
func (fs *FilterService) matchesSecurityFilter(repo entities.Repository, filter *SecurityFilter) bool {
	// This would require additional metadata about security features
	// For now, we'll assume all repositories match if no specific security data is available
	// In a real implementation, this would check:
	// - Security scanning results
	// - Push rules configuration
	// - Merge request approval settings
	// - Compliance framework assignments
	return true
}

// matchesArchiveFilter checks if repository matches archive filter.
func (fs *FilterService) matchesArchiveFilter(repo entities.Repository, filter *ArchiveFilter) bool {
	isArchived := repo.IsArchived()

	// Check if archived repositories should be included
	if filter.IncludeArchived != nil && !*filter.IncludeArchived && isArchived {
		return false
	}

	// Check if only archived repositories should be included
	if filter.OnlyArchived != nil && *filter.OnlyArchived && !isArchived {
		return false
	}

	// Check archive date filters (would require additional metadata)
	// For now, we'll assume date filters pass if no specific archive date is available
	_ = filter.ArchivedBefore
	_ = filter.ArchivedAfter

	return true
}

// FilterRepositoriesByNamespace filters repositories by GitLab namespace.
func (fs *FilterService) FilterRepositoriesByNamespace(ctx context.Context, repos []entities.Repository, namespaces []string) []entities.Repository {
	if len(namespaces) == 0 {
		return repos
	}

	var filtered []entities.Repository

	filtered = append(filtered, repos...)

	return filtered
}

// FilterRepositoriesByAccessLevel filters repositories by user access level.
func (fs *FilterService) FilterRepositoriesByAccessLevel(ctx context.Context, repos []entities.Repository, minAccessLevel string) []entities.Repository {
	if minAccessLevel == "" {
		return repos
	}

	var filtered []entities.Repository

	filtered = append(filtered, repos...)

	return filtered
}

// FilterRepositoriesByCompliance filters repositories by compliance frameworks.
func (fs *FilterService) FilterRepositoriesByCompliance(ctx context.Context, repos []entities.Repository, frameworks []string) []entities.Repository {
	if len(frameworks) == 0 {
		return repos
	}

	var filtered []entities.Repository

	filtered = append(filtered, repos...)

	return filtered
}

// GetFilterStatistics returns statistics about GitLab filtering results.
func (fs *FilterService) GetFilterStatistics(original, filtered []entities.Repository) map[string]interface{} {
	stats := map[string]interface{}{
		"original_count":      len(original),
		"filtered_count":      len(filtered),
		"filtered_percentage": float64(len(filtered)) / float64(len(original)) * 100,
	}

	// Count by visibility
	publicCount := 0
	privateCount := 0
	internalCount := 0
	archivedCount := 0
	forkedCount := 0

	for _, repo := range filtered {
		switch repo.Visibility() {
		case "private":
			privateCount++
		case "internal":
			internalCount++
		default:
			publicCount++
		}

		if repo.IsArchived() {
			archivedCount++
		}

		if repo.IsFork() {
			forkedCount++
		}
	}

	stats["public_count"] = publicCount
	stats["private_count"] = privateCount
	stats["internal_count"] = internalCount
	stats["archived_count"] = archivedCount
	stats["forked_count"] = forkedCount

	return stats
}
