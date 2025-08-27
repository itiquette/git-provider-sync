// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v71/github"
	"golang.org/x/oauth2"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Adapter implements the RepositoryProvider interface for GitHub.
type Adapter struct {
	client *github.Client
}

// Config contains GitHub adapter configuration.
type Config struct {
	Token      string
	HTTPClient *http.Client
	BaseURL    string
	UploadURL  string
	UserAgent  string
}

// New constructs GitHub adapter with optional token authentication.
func New(ctx context.Context, token string) *Adapter {
	var client *github.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &Adapter{client: client}
}

// NewWithConfig constructs GitHub adapter with comprehensive configuration including custom endpoints.
func NewWithConfig(ctx context.Context, config Config) *Adapter {
	var client *github.Client

	// Create base client
	switch {
	case config.HTTPClient != nil:
		client = github.NewClient(config.HTTPClient)
	case config.Token != "":
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
		client = github.NewClient(oauth2.NewClient(ctx, ts))
	default:
		client = github.NewClient(nil)
	}

	// Configure URLs first (must be done before other configurations)
	if config.BaseURL != "" {
		if baseURL, err := url.Parse(config.BaseURL); err == nil {
			client.BaseURL = baseURL
		}
	}

	if config.UploadURL != "" {
		if uploadURL, err := url.Parse(config.UploadURL); err == nil {
			client.UploadURL = uploadURL
		}
	}

	// Apply token authentication if provided
	if config.Token != "" {
		client = client.WithAuthToken(config.Token)
	}

	return &Adapter{client: client}
}

// ListRepositories retrieves paginated repository collection for specified owner with metadata conversion.
func (a *Adapter) ListRepositories(ctx context.Context, config ports.ProviderConfig) ([]entities.Repository, error) {
	opts := &github.RepositoryListByUserOptions{
		Type:        "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []entities.Repository

	for {
		repos, resp, err := a.client.Repositories.ListByUser(ctx, config.Owner, opts)
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

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetRepository fetches individual repository with complete metadata and error handling.
func (a *Adapter) GetRepository(ctx context.Context, config ports.ProviderConfig, name string) (entities.Repository, error) {
	repo, _, err := a.client.Repositories.Get(ctx, config.Owner, name)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to get repository %s: %w", name, err)
	}

	return a.convertToRepository(repo)
}

// RepositoryExists validates repository existence with proper 404 handling and ID extraction.
func (a *Adapter) RepositoryExists(ctx context.Context, request ports.RepositoryExistsRequest) (bool, string, error) {
	repository, resp, err := a.client.Repositories.Get(ctx, request.Owner, request.Name)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil
		}

		return false, "", fmt.Errorf("failed to check repository existence: %w", err)
	}

	if repository.ID != nil {
		return true, strconv.FormatInt(*repository.ID, 10), nil
	}

	return true, "", nil
}

// CreateRepository establishes new repository with organization detection and metadata initialization.
func (a *Adapter) CreateRepository(ctx context.Context, config ports.ProviderConfig, options ports.CreateRepositoryOptions) (entities.Repository, error) {
	repo := &github.Repository{
		Name:        github.Ptr(options.Name),
		Description: github.Ptr(options.Description),
		Private:     github.Ptr(options.Visibility == constants.VisibilityPrivate),
		AutoInit:    github.Ptr(options.AutoInit),
	}

	if options.DefaultBranch != "" {
		repo.DefaultBranch = github.Ptr(options.DefaultBranch)
	}

	var createdRepo *github.Repository

	var err error

	// Check if we're creating in an organization
	if a.isOrganization(ctx, config.Owner) {
		createdRepo, _, err = a.client.Repositories.Create(ctx, config.Owner, repo)
	} else {
		createdRepo, _, err = a.client.Repositories.Create(ctx, "", repo)
	}

	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to create repository: %w", err)
	}

	return a.convertToRepository(createdRepo)
}

// UpdateRepository modifies repository properties with selective field updates and validation.
func (a *Adapter) UpdateRepository(ctx context.Context, config ports.ProviderConfig, name string, options ports.UpdateRepositoryOptions) error {
	repo := &github.Repository{}

	if options.Description != nil {
		repo.Description = github.Ptr(*options.Description)
	}

	if options.Visibility != nil {
		repo.Private = github.Ptr(*options.Visibility == constants.VisibilityPrivate)
	}

	if options.DefaultBranch != nil {
		repo.DefaultBranch = github.Ptr(*options.DefaultBranch)
	}

	_, _, err := a.client.Repositories.Edit(ctx, config.Owner, name, repo)
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}

	return nil
}

// DeleteRepository removes repository with proper error handling and confirmation.
func (a *Adapter) DeleteRepository(ctx context.Context, config ports.ProviderConfig, name string) error {
	_, err := a.client.Repositories.Delete(ctx, config.Owner, name)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	return nil
}

// Note: Repository filtering is handled by domain.FilterRepositoriesUseCase
// This adapter focuses only on GitHub API interactions

// ValidateRepositoryName validates a repository name for GitHub.
func (a *Adapter) ValidateRepositoryName(name string) error {
	if name == "" {
		return domain.ErrRepositoryNameEmpty
	}

	if len(name) > 100 {
		return domain.ErrRepositoryNameTooLong
	}

	// GitHub naming rules - alphanumeric, dots, underscores, and hyphens only
	// Stricter 100 character limit compared to GitLab's 255 character limit
	validName := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validName.MatchString(name) {
		return domain.ErrRepositoryNameInvalidChars
	}

	return nil
}

// TransformRepositoryName transforms a repository name according to transformation rules.
// The options parameter supports prefix/suffix addition, case conversion,
// character replacement, and length constraints for GitHub provider compatibility.
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
func (a *Adapter) GetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) (ports.BranchProtection, error) {
	protection, _, err := a.client.Repositories.GetBranchProtection(ctx, config.Owner, repoName, branch)
	if err != nil {
		return ports.BranchProtection{}, fmt.Errorf("failed to get branch protection: %w", err)
	}

	return a.convertBranchProtection(protection), nil
}

// SetBranchProtection sets branch protection settings.
func (a *Adapter) SetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string, protection ports.BranchProtection) error {
	req := a.convertToGitHubProtection(protection)

	_, _, err := a.client.Repositories.UpdateBranchProtection(ctx, config.Owner, repoName, branch, req)
	if err != nil {
		return fmt.Errorf("failed to set branch protection: %w", err)
	}

	return nil
}

// RemoveBranchProtection removes branch protection.
func (a *Adapter) RemoveBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) error {
	_, err := a.client.Repositories.RemoveBranchProtection(ctx, config.Owner, repoName, branch)
	if err != nil {
		return fmt.Errorf("failed to remove branch protection: %w", err)
	}

	return nil
}

// ListProtectedBranches lists protected branches.
func (a *Adapter) ListProtectedBranches(ctx context.Context, config ports.ProviderConfig, repoName string) ([]string, error) {
	branches, _, err := a.client.Repositories.ListBranches(ctx, config.Owner, repoName, &github.BranchListOptions{
		Protected: github.Ptr(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list protected branches: %w", err)
	}

	var names []string

	for _, branch := range branches {
		if branch.Name != nil {
			names = append(names, *branch.Name)
		}
	}

	return names, nil
}

// GetProviderInfo returns GitHub API capabilities, rate limits, and supported features for provider selection.
func (a *Adapter) GetProviderInfo() ports.ProviderInfo {
	return ports.ProviderInfo{
		Name:       "GitHub",
		Type:       "github",
		Domain:     "github.com",
		APIVersion: "v3",
		Features: []ports.ProviderFeature{
			ports.FeatureRepositoryCreation,
			ports.FeatureRepositoryDeletion,
			ports.FeatureBranchProtection,
			ports.FeatureRequiredStatusChecks,
			ports.FeatureRequiredReviews,
			ports.FeaturePullRequests,
			ports.FeatureIssueTracking,
			ports.FeatureActions,
			ports.FeatureSecurityAdvisories,
			ports.FeatureOrganizations,
			ports.FeatureWebhooks,
			ports.FeatureGitLFS,
			ports.FeatureTopics,
			ports.FeatureReleases,
		},
		Capabilities: ports.ProviderCapabilities{
			MaxRepositoriesPerRequest: 100,
			MaxFileSize:               100 * 1024 * 1024,  // 100MB
			MaxRepositorySize:         1024 * 1024 * 1024, // 1GB
			RateLimitPerHour:          5000,
			SupportsSSH:               true,
			SupportsHTTPS:             true,
			SupportsPrivateRepos:      true,
			SupportsOrganizations:     true,
		},
	}
}

// SupportsFeature checks if GitHub supports a specific feature.
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
func (a *Adapter) CreateRepositoryForPush(ctx context.Context, request ports.CreateRepositoryRequest) (string, error) {
	repo := &github.Repository{
		Name:        github.Ptr(request.Name),
		Description: github.Ptr(request.Description),
		Private:     github.Ptr(request.Private),
	}

	if request.DefaultBranch != "" {
		repo.DefaultBranch = github.Ptr(request.DefaultBranch)
	}

	var createdRepo *github.Repository

	var err error

	// Create repository for authenticated user (empty string means user's repositories)
	createdRepo, _, err = a.client.Repositories.Create(ctx, "", repo)
	if err != nil {
		return "", fmt.Errorf("failed to create repository for push: %w", err)
	}

	if createdRepo.ID != nil {
		return strconv.FormatInt(*createdRepo.ID, 10), nil
	}

	return "", domain.ErrRepositoryNoID
}

// ProjectExists checks if a project exists and returns its ID.
func (a *Adapter) ProjectExists(ctx context.Context, owner, repo string) (bool, string, error) {
	repository, resp, err := a.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil
		}

		return false, "", fmt.Errorf("failed to check project existence: %w", err)
	}

	if repository.ID != nil {
		return true, strconv.FormatInt(*repository.ID, 10), nil
	}

	return true, "", nil
}

// Protect enables branch protection for the specified repository.
func (a *Adapter) Protect(ctx context.Context, owner string, defaultBranch string, projectIDstr string) error {
	// Convert projectID from string to int64
	projectID, err := strconv.ParseInt(projectIDstr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Get repository name from ID (GitHub API limitation)
	repo, _, err := a.client.Repositories.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get repository by ID: %w", err)
	}

	if repo.Name == nil {
		return domain.ErrRepositoryNoName
	}

	// Create basic protection rules
	protection := &github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Strict: true,
		},
		EnforceAdmins: true,
	}

	_, _, err = a.client.Repositories.UpdateBranchProtection(ctx, owner, *repo.Name, defaultBranch, protection)
	if err != nil {
		return fmt.Errorf("failed to protect branch: %w", err)
	}

	return nil
}

// Unprotect disables branch protection for the specified repository.
func (a *Adapter) Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error {
	// Convert projectID from string to int64
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Get repository name from ID (GitHub API limitation)
	repo, _, err := a.client.Repositories.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get repository by ID: %w", err)
	}

	if repo.Name == nil {
		return domain.ErrRepositoryNoName
	}

	if repo.Owner == nil || repo.Owner.Login == nil {
		return domain.ErrRepositoryNoOwner
	}

	_, err = a.client.Repositories.RemoveBranchProtection(ctx, *repo.Owner.Login, *repo.Name, defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to unprotect branch: %w", err)
	}

	return nil
}

// SetDefaultBranch sets the default branch for a repository.
func (a *Adapter) SetDefaultBranch(ctx context.Context, owner, name, branch string) error {
	repo := &github.Repository{
		DefaultBranch: github.Ptr(branch),
	}

	_, _, err := a.client.Repositories.Edit(ctx, owner, name, repo)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

// IsValidProjectName validates a project name for GitHub.
func (a *Adapter) IsValidProjectName(_ context.Context, name string) bool {
	return a.ValidateRepositoryName(name) == nil
}

func (a *Adapter) convertToGitHubProtection(protection ports.BranchProtection) *github.ProtectionRequest {
	req := &github.ProtectionRequest{}

	if len(protection.RequiredStatusChecks.Contexts) > 0 {
		req.RequiredStatusChecks = &github.RequiredStatusChecks{
			Strict:   protection.RequiredStatusChecks.Strict,
			Contexts: &protection.RequiredStatusChecks.Contexts,
		}
	}

	if protection.RequiredPullRequestReviews.RequiredApprovingReviewCount > 0 {
		req.RequiredPullRequestReviews = &github.PullRequestReviewsEnforcementRequest{
			RequiredApprovingReviewCount: protection.RequiredPullRequestReviews.RequiredApprovingReviewCount,
			DismissStaleReviews:          protection.RequiredPullRequestReviews.DismissStaleReviews,
			RequireCodeOwnerReviews:      protection.RequiredPullRequestReviews.RequireCodeOwnerReviews,
		}
	}

	if protection.EnforceAdmins {
		req.EnforceAdmins = true
	}

	return req
}

func (a *Adapter) convertToRepository(repo *github.Repository) (entities.Repository, error) {
	if repo == nil {
		return entities.Repository{}, domain.ErrRepositoryNil
	}

	builder := entities.NewRepositoryBuilder()

	if err := a.setRepositoryStringFields(repo, &builder); err != nil {
		return entities.Repository{}, err
	}

	a.setRepositoryMetadata(repo, &builder)
	a.setRepositoryFlags(repo, &builder)

	builtRepo, err := builder.Build()
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to build repository from GitHub data: %w", err)
	}

	return builtRepo, nil
}

// setRepositoryStringFields uses the common helper to set string fields.
func (a *Adapter) setRepositoryStringFields(repo *github.Repository, builder *entities.RepositoryBuilder) error {
	return setRepositoryStringFields(repo, builder)
}

func (a *Adapter) setRepositoryMetadata(repo *github.Repository, builder *entities.RepositoryBuilder) {
	if repo.Description != nil {
		*builder = builder.WithDescription(*repo.Description)
	}

	visibility := constants.VisibilityPublic
	if repo.Private != nil && *repo.Private {
		visibility = constants.VisibilityPrivate
	}

	*builder = builder.WithVisibility(visibility)

	if repo.UpdatedAt != nil {
		*builder = builder.WithLastActivityAt(repo.UpdatedAt.Time)
	}

	if repo.ID != nil {
		*builder = builder.WithProjectID(strconv.FormatInt(*repo.ID, 10))
	}

	*builder = builder.WithProviderType("github")
}

func (a *Adapter) setRepositoryFlags(repo *github.Repository, builder *entities.RepositoryBuilder) {
	if repo.Private != nil {
		*builder = builder.WithPrivate(*repo.Private)
	}

	if repo.Fork != nil {
		*builder = builder.WithFork(*repo.Fork)
	}

	if repo.Archived != nil {
		*builder = builder.WithArchived(*repo.Archived)
	}
}

func (a *Adapter) isOrganization(ctx context.Context, owner string) bool {
	_, _, err := a.client.Organizations.Get(ctx, owner)

	return err == nil
}

func (a *Adapter) convertBranchProtection(protection *github.Protection) ports.BranchProtection {
	result := ports.BranchProtection{
		Protected: true,
	}

	if protection.RequiredStatusChecks != nil {
		var contexts []string
		if protection.RequiredStatusChecks.Contexts != nil {
			contexts = *protection.RequiredStatusChecks.Contexts
		}

		result.RequiredStatusChecks = ports.RequiredStatusChecks{
			Strict:   protection.RequiredStatusChecks.Strict,
			Contexts: contexts,
		}
	}

	if protection.RequiredPullRequestReviews != nil {
		result.RequiredPullRequestReviews = ports.RequiredPullRequestReviews{
			RequiredApprovingReviewCount: protection.RequiredPullRequestReviews.RequiredApprovingReviewCount,
			DismissStaleReviews:          protection.RequiredPullRequestReviews.DismissStaleReviews,
			RequireCodeOwnerReviews:      protection.RequiredPullRequestReviews.RequireCodeOwnerReviews,
		}
	}

	if protection.EnforceAdmins != nil {
		result.EnforceAdmins = protection.EnforceAdmins.Enabled
	}

	return result
}
