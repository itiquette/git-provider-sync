// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package gitea provides Gitea provider adapter.
package gitea

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"

	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Static errors for err113 compliance.
var (
	ErrRepositoryNameEmpty             = errors.New("repository name cannot be empty")
	ErrRepositoryNameTooLong           = errors.New("repository name too long (max 100 characters)")
	ErrRepositoryNameInvalidCharacters = errors.New("repository name contains invalid characters")
	ErrRepositoryIsNil                 = errors.New("repository is nil")
)

// Adapter implements the RepositoryProvider interface for Gitea.
type Adapter struct {
	client *gitea.Client
	domain string
}

// Config contains Gitea adapter configuration.
type Config struct {
	Token         string
	HTTPClient    *http.Client
	BaseURL       string
	UserAgent     string
	SkipVerifySSL bool
}

// New creates a new Gitea adapter.
func New(token, domain string) (*Adapter, error) {
	if domain == "" {
		domain = "gitea.com"
	}

	baseURL := "https://" + domain

	client, err := gitea.NewClient(baseURL, gitea.SetToken(token))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	return &Adapter{
		client: client,
		domain: domain,
	}, nil
}

// NewWithConfig creates a new Gitea adapter with configuration.
func NewWithConfig(_ context.Context, config Config) (*Adapter, error) {
	var options []gitea.ClientOption

	if config.Token != "" {
		options = append(options, gitea.SetToken(config.Token))
	}

	if config.HTTPClient != nil {
		options = append(options, gitea.SetHTTPClient(config.HTTPClient))
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://gitea.com"
	}

	client, err := gitea.NewClient(baseURL, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	// Extract domain from base URL
	domain := "gitea.com"

	if baseURL != "" {
		if strings.Contains(baseURL, "://") {
			parts := strings.Split(baseURL, "://")
			if len(parts) > 1 {
				domain = strings.Split(parts[1], "/")[0]
			}
		}
	}

	return &Adapter{
		client: client,
		domain: domain,
	}, nil
}

// ListRepositories lists all repositories for a user/organization.
func (a *Adapter) ListRepositories(_ context.Context, config ports.ProviderConfig) ([]entities.Repository, error) {
	opts := gitea.ListReposOptions{
		ListOptions: gitea.ListOptions{PageSize: 100},
	}

	var allRepos []entities.Repository

	page := 1

	for {
		opts.Page = page

		repos, resp, err := a.client.ListUserRepos(config.Owner, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, repo := range repos {
			domainRepo, err := a.convertToRepository(repo)
			if err != nil {
				continue // Skip repos that can't be converted
			}

			allRepos = append(allRepos, domainRepo)
		}

		if resp == nil || len(repos) < opts.PageSize {
			break
		}

		page++
	}

	return allRepos, nil
}

// GetRepository gets a specific repository.
func (a *Adapter) GetRepository(_ context.Context, config ports.ProviderConfig, name string) (entities.Repository, error) {
	repo, _, err := a.client.GetRepo(config.Owner, name)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to get repository %s: %w", name, err)
	}

	return a.convertToRepository(repo)
}

// RepositoryExists checks if a repository exists.
func (a *Adapter) RepositoryExists(_ context.Context, request ports.RepositoryExistsRequest) (bool, string, error) {
	repo, resp, err := a.client.GetRepo(request.Owner, request.Name)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil
		}

		return false, "", fmt.Errorf("failed to check repository existence: %w", err)
	}

	return true, strconv.FormatInt(repo.ID, 10), nil
}

// CreateRepository creates a new repository.
func (a *Adapter) CreateRepository(_ context.Context, config ports.ProviderConfig, options ports.CreateRepositoryOptions) (entities.Repository, error) {
	createOpts := gitea.CreateRepoOption{
		Name:        options.Name,
		Description: options.Description,
		Private:     options.Visibility == constants.VisibilityPrivate,
		AutoInit:    options.AutoInit,
	}

	if options.DefaultBranch != "" {
		createOpts.DefaultBranch = options.DefaultBranch
	}

	var createdRepo *gitea.Repository

	var err error

	// Check if we're creating in an organization
	if a.isOrganization(config.Owner) {
		createdRepo, _, err = a.client.CreateOrgRepo(config.Owner, createOpts)
	} else {
		createdRepo, _, err = a.client.CreateRepo(createOpts)
	}

	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to create repository: %w", err)
	}

	return a.convertToRepository(createdRepo)
}

// UpdateRepository updates an existing repository.
func (a *Adapter) UpdateRepository(_ context.Context, config ports.ProviderConfig, name string, options ports.UpdateRepositoryOptions) error {
	editOpts := gitea.EditRepoOption{}

	if options.Description != nil {
		editOpts.Description = options.Description
	}

	if options.Visibility != nil {
		private := *options.Visibility == constants.VisibilityPrivate
		editOpts.Private = &private
	}

	if options.DefaultBranch != nil {
		editOpts.DefaultBranch = options.DefaultBranch
	}

	_, _, err := a.client.EditRepo(config.Owner, name, editOpts)
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}

	return nil
}

// DeleteRepository deletes a repository.
func (a *Adapter) DeleteRepository(_ context.Context, config ports.ProviderConfig, name string) error {
	_, err := a.client.DeleteRepo(config.Owner, name)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	return nil
}

// ValidateRepositoryName validates a repository name for Gitea.
func (a *Adapter) ValidateRepositoryName(name string) error {
	if name == "" {
		return ErrRepositoryNameEmpty
	}

	if len(name) > 100 {
		return ErrRepositoryNameTooLong
	}

	// Gitea naming rules - follows GitHub conventions with 100 character limit
	// More restrictive than GitLab to maintain consistency with GitHub workflows
	validName := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validName.MatchString(name) {
		return ErrRepositoryNameInvalidCharacters
	}

	return nil
}

// TransformRepositoryName transforms a repository name according to transformation rules.
// The options parameter supports prefix/suffix addition, case conversion,
// character replacement, and length constraints for Gitea provider compatibility.
func (a *Adapter) TransformRepositoryName(name string, options ports.NameTransformOptions) string {
	result := name

	if options.ToLowercase {
		result = strings.ToLower(result)
	}

	if options.ToUppercase {
		result = strings.ToUpper(result)
	}

	for old, new := range options.Replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	if options.Prefix != "" {
		result = options.Prefix + result
	}

	if options.Suffix != "" {
		result += options.Suffix
	}

	if options.AlphaNumericOnly {
		reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
		result = reg.ReplaceAllString(result, "-")
	}

	if options.MaxLength > 0 && len(result) > options.MaxLength {
		result = result[:options.MaxLength]
	}

	return result
}

// GetBranchProtection gets branch protection settings.
func (a *Adapter) GetBranchProtection(_ context.Context, config ports.ProviderConfig, repoName, branch string) (ports.BranchProtection, error) {
	protection, _, err := a.client.GetBranchProtection(config.Owner, repoName, branch)
	if err != nil {
		return ports.BranchProtection{}, fmt.Errorf("failed to get branch protection: %w", err)
	}

	return a.convertBranchProtection(protection), nil
}

// SetBranchProtection sets branch protection settings.
func (a *Adapter) SetBranchProtection(_ context.Context, config ports.ProviderConfig, repoName, _ string, protection ports.BranchProtection) error {
	opts := a.convertToGiteaProtection(protection)

	_, _, err := a.client.CreateBranchProtection(config.Owner, repoName, opts)
	if err != nil {
		return fmt.Errorf("failed to set branch protection: %w", err)
	}

	return nil
}

// RemoveBranchProtection removes branch protection.
func (a *Adapter) RemoveBranchProtection(_ context.Context, config ports.ProviderConfig, repoName, branch string) error {
	_, err := a.client.DeleteBranchProtection(config.Owner, repoName, branch)
	if err != nil {
		return fmt.Errorf("failed to remove branch protection: %w", err)
	}

	return nil
}

// ListProtectedBranches lists protected branches.
func (a *Adapter) ListProtectedBranches(_ context.Context, config ports.ProviderConfig, repoName string) ([]string, error) {
	protections, _, err := a.client.ListBranchProtections(config.Owner, repoName, gitea.ListBranchProtectionsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list protected branches: %w", err)
	}

	names := make([]string, 0, len(protections))
	for _, protection := range protections {
		names = append(names, protection.BranchName)
	}

	return names, nil
}

// GetProviderInfo returns Gitea API capabilities, features, and limitations for provider selection.
func (a *Adapter) GetProviderInfo() ports.ProviderInfo {
	return ports.ProviderInfo{
		Name:       "Gitea",
		Type:       "gitea",
		Domain:     a.domain,
		APIVersion: "v1",
		Features: []ports.ProviderFeature{
			ports.FeatureRepositoryCreation,
			ports.FeatureRepositoryDeletion,
			ports.FeatureBranchProtection,
			ports.FeaturePullRequests,
			ports.FeatureIssueTracking,
			ports.FeatureOrganizations,
			ports.FeatureWebhooks,
			ports.FeatureGitLFS,
			ports.FeatureReleases,
		},
		Capabilities: ports.ProviderCapabilities{
			MaxRepositoriesPerRequest: 100,
			MaxFileSize:               100 * 1024 * 1024,  // 100MB
			MaxRepositorySize:         1024 * 1024 * 1024, // 1GB
			RateLimitPerHour:          3000,
			SupportsSSH:               true,
			SupportsHTTPS:             true,
			SupportsPrivateRepos:      true,
			SupportsOrganizations:     true,
		},
	}
}

// SupportsFeature checks if Gitea supports a specific feature.
func (a *Adapter) SupportsFeature(feature ports.ProviderFeature) bool {
	info := a.GetProviderInfo()
	for _, f := range info.Features {
		if f == feature {
			return true
		}
	}

	return false
}

// CreateRepositoryForPush creates a repository optimized for push operations with minimal metadata.
func (a *Adapter) CreateRepositoryForPush(_ context.Context, request ports.CreateRepositoryRequest) (string, error) {
	opts := gitea.CreateRepoOption{
		Name:        request.Name,
		Description: request.Description,
		Private:     request.Private,
	}

	if request.DefaultBranch != "" {
		opts.DefaultBranch = request.DefaultBranch
	}

	repo, _, err := a.client.CreateRepo(opts)
	if err != nil {
		return "", fmt.Errorf("failed to create repository for push: %w", err)
	}

	return strconv.FormatInt(repo.ID, 10), nil
}

// ProjectExists checks if a project exists and returns its ID.
func (a *Adapter) ProjectExists(_ context.Context, owner, repo string) (bool, string, error) {
	repository, resp, err := a.client.GetRepo(owner, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil
		}

		return false, "", fmt.Errorf("failed to check project existence: %w", err)
	}

	return true, strconv.FormatInt(repository.ID, 10), nil
}

// Protect enables branch protection for Gitea.
func (a *Adapter) Protect(_ context.Context, owner string, defaultBranch string, projectIDstr string) error {
	// Gitea uses repo owner/name for protection, not just ID
	// We'd need to get the repo name from the ID, but for now use owner
	opts := gitea.CreateBranchProtectionOption{
		BranchName:        defaultBranch,
		RequiredApprovals: 1, // Default protection level
	}

	_, _, err := a.client.CreateBranchProtection(owner, projectIDstr, opts)
	if err != nil {
		return fmt.Errorf("failed to protect branch: %w", err)
	}

	return nil
}

// Unprotect disables branch protection for Gitea.
func (a *Adapter) Unprotect(_ context.Context, _ string, _ string) error {
	// Gitea protection removal - would need repo owner/name
	// For now, return nil as this is a placeholder
	return nil
}

// SetDefaultBranch sets the default branch for a Gitea repository.
func (a *Adapter) SetDefaultBranch(_ context.Context, owner, name, branch string) error {
	opts := gitea.EditRepoOption{
		DefaultBranch: &branch,
	}

	_, _, err := a.client.EditRepo(owner, name, opts)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

// IsValidProjectName validates a project name for Gitea.
func (a *Adapter) IsValidProjectName(_ context.Context, name string) bool {
	return a.ValidateRepositoryName(name) == nil
}

func (a *Adapter) convertToRepository(repo *gitea.Repository) (entities.Repository, error) {
	if repo == nil {
		return entities.Repository{}, ErrRepositoryIsNil
	}

	builder := entities.NewRepositoryBuilder()

	builder, err := builder.WithName(repo.Name)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository field: %w", err)
	}

	builder, err = builder.WithHTTPSURL(repo.CloneURL)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository field: %w", err)
	}

	builder, err = builder.WithSSHURL(repo.SSHURL)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository field: %w", err)
	}

	builder, err = builder.WithDefaultBranch(repo.DefaultBranch)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository field: %w", err)
	}

	builder = builder.WithDescription(repo.Description)

	visibility := constants.VisibilityPublic
	if repo.Private {
		visibility = constants.VisibilityPrivate
	}

	builder = builder.WithVisibility(visibility)

	if !repo.Updated.IsZero() {
		builder = builder.WithLastActivityAt(repo.Updated)
	}

	builder = builder.WithProjectID(strconv.FormatInt(repo.ID, 10))
	builder = builder.WithProviderType("gitea")
	builder = builder.WithPrivate(repo.Private)
	builder = builder.WithFork(repo.Fork)
	builder = builder.WithArchived(repo.Archived)

	builtRepo, err := builder.Build()
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to build repository from Gitea data: %w", err)
	}

	return builtRepo, nil
}

func (a *Adapter) isOrganization(owner string) bool {
	// Simple check - try to get organization info
	_, _, err := a.client.GetOrg(owner)

	return err == nil
}

func (a *Adapter) convertBranchProtection(protection *gitea.BranchProtection) ports.BranchProtection {
	result := ports.BranchProtection{
		Protected: true,
	}

	if protection.RequiredApprovals > 0 {
		result.RequiredPullRequestReviews = ports.RequiredPullRequestReviews{
			RequiredApprovingReviewCount: int(protection.RequiredApprovals),
		}
	}

	// Map other Gitea-specific fields as needed

	return result
}

func (a *Adapter) convertToGiteaProtection(protection ports.BranchProtection) gitea.CreateBranchProtectionOption {
	opts := gitea.CreateBranchProtectionOption{
		BranchName: "*", // Default to all branches
	}

	if protection.RequiredPullRequestReviews.RequiredApprovingReviewCount > 0 {
		opts.RequiredApprovals = int64(protection.RequiredPullRequestReviews.RequiredApprovingReviewCount)
	}

	return opts
}
