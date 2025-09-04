// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package entities contains core domain entities with business logic and behavior
// These entities are pure domain models with no external dependencies
package entities

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Domain errors for repository operations.
var (
	ErrInvalidRepositoryName = errors.New("invalid repository name")
	ErrEmptyRepositoryName   = errors.New("repository name cannot be empty")
	ErrRepositoryNameTooLong = errors.New("repository name is too long")
	ErrInvalidURL            = errors.New("invalid repository URL")
	ErrEmptyDefaultBranch    = errors.New("default branch cannot be empty")
)

// Repository represents a repository in the domain with immutable properties and behavior.
type Repository struct {
	name           string
	cleanName      string
	httpsURL       string
	sshURL         string
	defaultBranch  string
	description    string
	visibility     string
	lastActivityAt time.Time
	projectID      string
	providerType   string
	isPrivate      bool
	isFork         bool
	isArchived     bool
}

// RepositoryBuilder builds Repository instances.
type RepositoryBuilder struct {
	repo Repository
}

// NewRepositoryBuilder creates a new repository builder.
func NewRepositoryBuilder() RepositoryBuilder {
	return RepositoryBuilder{
		repo: Repository{
			visibility:    "private", // Secure default
			defaultBranch: "main",    // Modern default
		},
	}
}

// WithName sets the repository name and validates it.
func (b RepositoryBuilder) WithName(name string) (RepositoryBuilder, error) {
	if err := ValidateRepositoryName(name); err != nil {
		return b, fmt.Errorf("invalid name: %w", err)
	}

	b.repo.name = name
	b.repo.cleanName = CleanRepositoryName(name)

	return b, nil
}

// WithHTTPSURL sets the HTTPS URL and validates it.
func (b RepositoryBuilder) WithHTTPSURL(url string) (RepositoryBuilder, error) {
	if err := validateURL(url); err != nil {
		return b, fmt.Errorf("invalid HTTPS URL: %w", err)
	}

	b.repo.httpsURL = url

	return b, nil
}

// WithSSHURL sets the SSH URL and validates it.
func (b RepositoryBuilder) WithSSHURL(url string) (RepositoryBuilder, error) {
	if err := validateURL(url); err != nil {
		return b, fmt.Errorf("invalid SSH URL: %w", err)
	}

	b.repo.sshURL = url

	return b, nil
}

// WithDefaultBranch sets the default branch.
func (b RepositoryBuilder) WithDefaultBranch(branch string) (RepositoryBuilder, error) {
	if strings.TrimSpace(branch) == "" {
		return b, ErrEmptyDefaultBranch
	}

	b.repo.defaultBranch = strings.TrimSpace(branch)

	return b, nil
}

// WithDescription sets the repository description.
func (b RepositoryBuilder) WithDescription(description string) RepositoryBuilder {
	b.repo.description = strings.TrimSpace(description)

	return b
}

// WithVisibility sets the repository visibility.
func (b RepositoryBuilder) WithVisibility(visibility string) RepositoryBuilder {
	normalizedVisibility := strings.ToLower(strings.TrimSpace(visibility))
	if isValidVisibility(normalizedVisibility) {
		b.repo.visibility = normalizedVisibility
		b.repo.isPrivate = normalizedVisibility == "private"
	}

	return b
}

// WithLastActivityAt sets the last activity timestamp.
func (b RepositoryBuilder) WithLastActivityAt(timestamp time.Time) RepositoryBuilder {
	b.repo.lastActivityAt = timestamp

	return b
}

// WithProjectID sets the provider-specific project ID.
func (b RepositoryBuilder) WithProjectID(id string) RepositoryBuilder {
	b.repo.projectID = strings.TrimSpace(id)

	return b
}

// WithProviderType sets the git provider type.
func (b RepositoryBuilder) WithProviderType(providerType string) RepositoryBuilder {
	b.repo.providerType = strings.ToLower(strings.TrimSpace(providerType))

	return b
}

// WithFlags sets boolean flags for the repository.
func (b RepositoryBuilder) WithFlags(isPrivate, isFork, isArchived bool) RepositoryBuilder {
	b.repo.isPrivate = isPrivate
	b.repo.isFork = isFork
	b.repo.isArchived = isArchived

	return b
}

// WithPrivate sets the private flag for the repository.
func (b RepositoryBuilder) WithPrivate(isPrivate bool) RepositoryBuilder {
	b.repo.isPrivate = isPrivate

	return b
}

// WithFork sets the fork flag for the repository.
func (b RepositoryBuilder) WithFork(isFork bool) RepositoryBuilder {
	b.repo.isFork = isFork

	return b
}

// WithArchived sets the archived flag for the repository.
func (b RepositoryBuilder) WithArchived(isArchived bool) RepositoryBuilder {
	b.repo.isArchived = isArchived

	return b
}

// Build creates the final repository after validation.
func (b RepositoryBuilder) Build() (Repository, error) {
	if b.repo.name == "" {
		return Repository{}, ErrEmptyRepositoryName
	}

	if b.repo.httpsURL == "" && b.repo.sshURL == "" {
		return Repository{}, ErrInvalidURL
	}

	return b.repo, nil
}

// Repository accessor methods (immutable)

// Name returns the repository name.
func (r Repository) Name() string {
	return r.name
}

// CleanName returns the cleaned repository name (alphanumeric and hyphens only).
func (r Repository) CleanName() string {
	return r.cleanName
}

// HTTPSURL returns the HTTPS clone URL.
func (r Repository) HTTPSURL() string {
	return r.httpsURL
}

// SSHURL returns the SSH clone URL.
func (r Repository) SSHURL() string {
	return r.sshURL
}

// DefaultBranch returns the default branch name.
func (r Repository) DefaultBranch() string {
	return r.defaultBranch
}

// Description returns the repository description.
func (r Repository) Description() string {
	return r.description
}

// Visibility returns the repository visibility (public, internal, private).
func (r Repository) Visibility() string {
	return r.visibility
}

// LastActivityAt returns the last activity timestamp.
func (r Repository) LastActivityAt() time.Time {
	return r.lastActivityAt
}

// ProjectID returns the provider-specific project ID.
func (r Repository) ProjectID() string {
	return r.projectID
}

// ProviderType returns the git provider type.
func (r Repository) ProviderType() string {
	return r.providerType
}

// IsPrivate returns true if the repository is private.
func (r Repository) IsPrivate() bool {
	return r.isPrivate
}

// IsFork returns true if the repository is a fork.
func (r Repository) IsFork() bool {
	return r.isFork
}

// IsArchived returns true if the repository is archived.
func (r Repository) IsArchived() bool {
	return r.isArchived
}

// Repository behavior methods

// PreferredCloneURL returns the preferred URL for cloning based on availability.
func (r Repository) PreferredCloneURL() string {
	if r.httpsURL != "" {
		return r.httpsURL
	}

	return r.sshURL
}

// IsActive returns true if the repository has recent activity (within 6 months).
func (r Repository) IsActive() bool {
	if r.lastActivityAt.IsZero() {
		return true // Assume active if no activity data
	}

	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	return r.lastActivityAt.After(sixMonthsAgo)
}

// ShouldIncludeInSync determines if repository should be included in sync based on filters.
func (r Repository) ShouldIncludeInSync(includeForks, includeArchived bool) bool {
	if r.isFork && !includeForks {
		return false
	}

	if r.isArchived && !includeArchived {
		return false
	}

	return true
}

// ValidateForProvider validates repository name against provider-specific rules.
func (r Repository) ValidateForProvider(providerType string) error {
	switch strings.ToLower(providerType) {
	case "github":
		return validateGitHubName(r.name)
	case "gitlab":
		return validateGitLabName(r.name)
	case "gitea":
		return validateGiteaName(r.name)
	default:
		return ValidateRepositoryName(r.name) // Generic validation
	}
}

// WithUpdatedBranch returns a new repository with updated default branch.
func (r Repository) WithUpdatedBranch(branch string) (Repository, error) {
	if strings.TrimSpace(branch) == "" {
		return r, ErrEmptyDefaultBranch
	}

	updated := r
	updated.defaultBranch = strings.TrimSpace(branch)

	return updated, nil
}

// WithUpdatedDescription returns a new repository with updated description.
func (r Repository) WithUpdatedDescription(description string) Repository {
	updated := r
	updated.description = strings.TrimSpace(description)

	return updated
}

// Domain validation functions

// ValidateRepositoryName validates a repository name using generic rules.
func ValidateRepositoryName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrEmptyRepositoryName
	}

	if len(name) > 100 {
		return ErrRepositoryNameTooLong
	}

	// Basic validation: alphanumeric, hyphens, underscores, dots
	validNameRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("%w: contains invalid characters", ErrInvalidRepositoryName)
	}

	// Cannot start or end with special characters
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("%w: cannot start or end with period", ErrInvalidRepositoryName)
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("%w: cannot start or end with hyphen", ErrInvalidRepositoryName)
	}

	return nil
}

// CleanRepositoryName converts a repository name to alphanumeric with hyphens only.
func CleanRepositoryName(name string) string {
	// Replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
	cleaned := reg.ReplaceAllString(name, "-")

	// Remove consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	cleaned = reg.ReplaceAllString(cleaned, "-")

	// Remove leading/trailing hyphens
	cleaned = strings.Trim(cleaned, "-")

	// Ensure not empty
	if cleaned == "" {
		cleaned = "repository"
	}

	return cleaned
}

// Provider-specific validation functions

func validateGitHubName(name string) error {
	if err := ValidateRepositoryName(name); err != nil {
		return err
	}

	// GitHub specific rules
	if len(name) > 100 {
		return fmt.Errorf("%w: GitHub names must be ≤100 characters", ErrRepositoryNameTooLong)
	}

	return nil
}

func validateGitLabName(name string) error {
	if err := ValidateRepositoryName(name); err != nil {
		return err
	}

	// GitLab specific rules
	if len(name) > 255 {
		return fmt.Errorf("%w: GitLab names must be ≤255 characters", ErrRepositoryNameTooLong)
	}

	// GitLab allows spaces and plus signs
	validGitLabRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.+\- ]*$`)
	if !validGitLabRegex.MatchString(name) {
		return fmt.Errorf("%w: invalid GitLab name format", ErrInvalidRepositoryName)
	}

	// Check GitLab reserved names - these conflict with UI routes and Git internals
	// See: https://docs.gitlab.com/ee/user/reserved_names.html
	reservedNames := map[string]bool{
		"badges": true, "blame": true, "blob": true, "builds": true,
		"commits": true, "create": true, "edit": true, "files": true,
		"new": true, "raw": true, "refs": true, "tree": true, "wikis": true,
	}

	if reservedNames[strings.ToLower(name)] {
		return fmt.Errorf("%w: '%s' is reserved in GitLab", ErrInvalidRepositoryName, name)
	}

	return nil
}

func validateGiteaName(name string) error {
	if err := ValidateRepositoryName(name); err != nil {
		return err
	}

	// Gitea specific rules (similar to GitHub but more restrictive)
	if len(name) > 100 {
		return fmt.Errorf("%w: Gitea names must be ≤100 characters", ErrRepositoryNameTooLong)
	}

	return nil
}

// Helper functions

func validateURL(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return ErrInvalidURL
	}

	// Basic URL validation
	if !strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "git@") {
		return fmt.Errorf("%w: must start with http://, https://, or git@", ErrInvalidURL)
	}

	return nil
}

func isValidVisibility(visibility string) bool {
	validVisibilities := map[string]bool{
		"public":   true,
		"internal": true,
		"private":  true,
	}

	return validVisibilities[visibility]
}
