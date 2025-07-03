// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gitlab-org/api/client-go"

	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Adapter implements the RepositoryProvider interface for GitLab.
type Adapter struct {
	client *gitlab.Client
	domain string
}

// Config contains GitLab adapter configuration.
type Config struct {
	Token      string
	HTTPClient *http.Client
	BaseURL    string
	UserAgent  string
}

// New creates a new GitLab adapter.
func New(token, domain string) (*Adapter, error) {
	if domain == "" {
		domain = "gitlab.com"
	}

	baseURL := "https://" + domain

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &Adapter{
		client: client,
		domain: domain,
	}, nil
}

// NewWithConfig creates a new GitLab adapter with advanced configuration.
func NewWithConfig(ctx context.Context, config Config) (*Adapter, error) {
	var options []gitlab.ClientOptionFunc

	if config.BaseURL != "" {
		options = append(options, gitlab.WithBaseURL(config.BaseURL))
	}

	if config.HTTPClient != nil {
		options = append(options, gitlab.WithHTTPClient(config.HTTPClient))
	}

	client, err := gitlab.NewClient(config.Token, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Extract domain from base URL
	domain := "gitlab.com"

	if config.BaseURL != "" {
		if strings.Contains(config.BaseURL, "://") {
			parts := strings.Split(config.BaseURL, "://")
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

// ListRepositories lists all repositories for a user/group.
func (a *Adapter) ListRepositories(ctx context.Context, config ports.ProviderConfig) ([]entities.Repository, error) {
	opts := &gitlab.ListProjectsOptions{
		Owned:       gitlab.Ptr(true),
		Membership:  gitlab.Ptr(true),
		OrderBy:     gitlab.Ptr("updated_at"),
		Sort:        gitlab.Ptr("desc"),
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	var allRepos []entities.Repository

	for {
		projects, resp, err := a.client.Projects.ListProjects(opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range projects {
			repo, err := a.convertToRepository(project)
			if err != nil {
				continue // Skip projects that can't be converted
			}

			allRepos = append(allRepos, repo)
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetRepository gets a specific repository.
func (a *Adapter) GetRepository(ctx context.Context, config ports.ProviderConfig, name string) (entities.Repository, error) {
	projectPath := config.Owner + "/" + name

	project, _, err := a.client.Projects.GetProject(projectPath, nil, gitlab.WithContext(ctx))
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to get project %s: %w", name, err)
	}

	return a.convertToRepository(project)
}

// RepositoryExists checks if a repository exists.
func (a *Adapter) RepositoryExists(ctx context.Context, request ports.RepositoryExistsRequest) (bool, string, error) {
	projectPath := request.Owner + "/" + request.Name

	project, resp, err := a.client.Projects.GetProject(projectPath, nil, gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil
		}

		return false, "", fmt.Errorf("failed to check project existence: %w", err)
	}

	return true, strconv.Itoa(project.ID), nil
}

// CreateRepository creates a new repository.
func (a *Adapter) CreateRepository(ctx context.Context, config ports.ProviderConfig, options ports.CreateRepositoryOptions) (entities.Repository, error) {
	visibility := gitlab.PublicVisibility
	if options.Visibility == constants.VisibilityPrivate {
		visibility = gitlab.PrivateVisibility
	}

	project := &gitlab.CreateProjectOptions{
		Name:                 gitlab.Ptr(options.Name),
		Description:          gitlab.Ptr(options.Description),
		Visibility:           &visibility,
		InitializeWithReadme: gitlab.Ptr(options.AutoInit),
		DefaultBranch:        gitlab.Ptr(options.DefaultBranch),
	}

	// Set namespace for group/organization
	if config.Owner != "" {
		// Try to get namespace ID
		namespaces, _, err := a.client.Namespaces.ListNamespaces(&gitlab.ListNamespacesOptions{
			Search: gitlab.Ptr(config.Owner),
		}, gitlab.WithContext(ctx))
		if err == nil && len(namespaces) > 0 {
			for _, ns := range namespaces {
				if ns.Path == config.Owner {
					project.NamespaceID = gitlab.Ptr(ns.ID)

					break
				}
			}
		}
	}

	createdProject, _, err := a.client.Projects.CreateProject(project, gitlab.WithContext(ctx))
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to create project: %w", err)
	}

	return a.convertToRepository(createdProject)
}

// UpdateRepository updates an existing repository.
func (a *Adapter) UpdateRepository(ctx context.Context, config ports.ProviderConfig, name string, options ports.UpdateRepositoryOptions) error {
	projectPath := config.Owner + "/" + name
	project := &gitlab.EditProjectOptions{}

	if options.Description != nil {
		project.Description = gitlab.Ptr(*options.Description)
	}

	if options.Visibility != nil {
		visibility := gitlab.PublicVisibility
		if *options.Visibility == constants.VisibilityPrivate {
			visibility = gitlab.PrivateVisibility
		}

		project.Visibility = &visibility
	}

	if options.DefaultBranch != nil {
		project.DefaultBranch = gitlab.Ptr(*options.DefaultBranch)
	}

	_, _, err := a.client.Projects.EditProject(projectPath, project, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// DeleteRepository deletes a repository.
func (a *Adapter) DeleteRepository(ctx context.Context, config ports.ProviderConfig, name string) error {
	projectPath := config.Owner + "/" + name

	_, err := a.client.Projects.DeleteProject(projectPath, &gitlab.DeleteProjectOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// Note: Repository filtering is handled by domain.FilterRepositoriesUseCase
// This adapter focuses only on GitLab API interactions

// ValidateRepositoryName validates a repository name for GitLab.
func (a *Adapter) ValidateRepositoryName(name string) error {
	if name == "" {
		return errors.New("repository name cannot be empty")
	}

	if len(name) > 255 {
		return errors.New("repository name too long (max 255 characters)")
	}

	// GitLab naming rules
	validName := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validName.MatchString(name) {
		return errors.New("repository name contains invalid characters")
	}

	return nil
}

// TransformRepositoryName transforms a name according to options.
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
		result = result + options.Suffix
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
	projectPath := config.Owner + "/" + repoName

	protection, _, err := a.client.ProtectedBranches.GetProtectedBranch(projectPath, branch, gitlab.WithContext(ctx))
	if err != nil {
		return ports.BranchProtection{}, fmt.Errorf("failed to get branch protection: %w", err)
	}

	return a.convertBranchProtection(protection), nil
}

// SetBranchProtection sets branch protection settings.
func (a *Adapter) SetBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string, protection ports.BranchProtection) error {
	projectPath := config.Owner + "/" + repoName
	opts := a.convertToGitLabProtection(protection)

	_, _, err := a.client.ProtectedBranches.ProtectRepositoryBranches(projectPath, opts, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to set branch protection: %w", err)
	}

	return nil
}

// RemoveBranchProtection removes branch protection.
func (a *Adapter) RemoveBranchProtection(ctx context.Context, config ports.ProviderConfig, repoName, branch string) error {
	projectPath := config.Owner + "/" + repoName

	_, err := a.client.ProtectedBranches.UnprotectRepositoryBranches(projectPath, branch, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to remove branch protection: %w", err)
	}

	return nil
}

// ListProtectedBranches lists protected branches.
func (a *Adapter) ListProtectedBranches(ctx context.Context, config ports.ProviderConfig, repoName string) ([]string, error) {
	projectPath := config.Owner + "/" + repoName

	branches, _, err := a.client.ProtectedBranches.ListProtectedBranches(projectPath, &gitlab.ListProtectedBranchesOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list protected branches: %w", err)
	}

	names := make([]string, 0, len(branches))
	for _, branch := range branches {
		names = append(names, branch.Name)
	}

	return names, nil
}

// GetProviderInfo returns information about GitLab.
func (a *Adapter) GetProviderInfo() ports.ProviderInfo {
	return ports.ProviderInfo{
		Name:       "GitLab",
		Type:       "gitlab",
		Domain:     a.domain,
		APIVersion: "v4",
		Features: []ports.ProviderFeature{
			ports.FeatureRepositoryCreation,
			ports.FeatureRepositoryDeletion,
			ports.FeatureBranchProtection,
			ports.FeatureRequiredStatusChecks,
			ports.FeatureMergeRequests,
			ports.FeatureIssueTracking,
			ports.FeaturePipelines,
			ports.FeatureGroupManagement,
			ports.FeatureWebhooks,
			ports.FeatureGitLFS,
			ports.FeatureTopics,
			ports.FeatureReleases,
		},
		Capabilities: ports.ProviderCapabilities{
			MaxRepositoriesPerRequest: 100,
			MaxFileSize:               100 * 1024 * 1024,       // 100MB
			MaxRepositorySize:         10 * 1024 * 1024 * 1024, // 10GB
			RateLimitPerHour:          2000,
			SupportsSSH:               true,
			SupportsHTTPS:             true,
			SupportsPrivateRepos:      true,
			SupportsOrganizations:     true,
		},
	}
}

// SupportsFeature checks if GitLab supports a specific feature.
func (a *Adapter) SupportsFeature(feature ports.ProviderFeature) bool {
	info := a.GetProviderInfo()
	for _, f := range info.Features {
		if f == feature {
			return true
		}
	}

	return false
}

// Helper methods

func (a *Adapter) convertToRepository(project *gitlab.Project) (entities.Repository, error) {
	if project == nil {
		return entities.Repository{}, errors.New("project is nil")
	}

	builder := entities.NewRepositoryBuilder()

	builder, err := builder.WithName(project.Name)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository field: %w", err)
	}

	builder, err = builder.WithHTTPSURL(project.HTTPURLToRepo)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository field: %w", err)
	}

	builder, err = builder.WithSSHURL(project.SSHURLToRepo)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository field: %w", err)
	}

	builder, err = builder.WithDefaultBranch(project.DefaultBranch)
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to set repository field: %w", err)
	}

	builder = builder.WithDescription(project.Description)

	// Convert visibility
	visibility := constants.VisibilityPublic

	switch project.Visibility {
	case gitlab.PrivateVisibility:
		visibility = constants.VisibilityPrivate
	case gitlab.InternalVisibility:
		visibility = constants.VisibilityInternal
	case gitlab.PublicVisibility:
		visibility = constants.VisibilityPublic
	}

	builder = builder.WithVisibility(visibility)

	if project.LastActivityAt != nil {
		builder = builder.WithLastActivityAt(*project.LastActivityAt)
	}

	builder = builder.WithProjectID(strconv.Itoa(project.ID))
	builder = builder.WithProviderType("gitlab")
	builder = builder.WithPrivate(project.Visibility == gitlab.PrivateVisibility)
	builder = builder.WithFork(project.ForkedFromProject != nil)
	builder = builder.WithArchived(project.Archived)

	builtRepo, err := builder.Build()
	if err != nil {
		return entities.Repository{}, fmt.Errorf("failed to build repository from GitLab data: %w", err)
	}

	return builtRepo, nil
}

// matchesFilter removed - filtering logic moved to domain.FilterRepositoriesUseCase

func (a *Adapter) convertBranchProtection(_ *gitlab.ProtectedBranch) ports.BranchProtection {
	result := ports.BranchProtection{
		Protected: true,
	}

	// GitLab has different protection model than GitHub
	// Map the available fields

	return result
}

func (a *Adapter) convertToGitLabProtection(_ ports.BranchProtection) *gitlab.ProtectRepositoryBranchesOptions {
	return &gitlab.ProtectRepositoryBranchesOptions{
		Name:             gitlab.Ptr("*"),
		PushAccessLevel:  gitlab.Ptr(gitlab.MaintainerPermissions),
		MergeAccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions),
	}
}

// CreateRepositoryForPush creates a repository specifically for push operations.
// This restores the main branch provider.Push functionality for GitLab.
func (a *Adapter) CreateRepositoryForPush(ctx context.Context, request ports.CreateRepositoryRequest) (string, error) {
	visibility := gitlab.PublicVisibility
	if request.Private {
		visibility = gitlab.PrivateVisibility
	}

	options := &gitlab.CreateProjectOptions{
		Name:        gitlab.Ptr(request.Name),
		Description: gitlab.Ptr(request.Description),
		Visibility:  gitlab.Ptr(visibility),
	}

	if request.DefaultBranch != "" {
		options.DefaultBranch = gitlab.Ptr(request.DefaultBranch)
	}

	project, _, err := a.client.Projects.CreateProject(options)
	if err != nil {
		return "", fmt.Errorf("failed to create repository for push: %w", err)
	}

	return strconv.Itoa(project.ID), nil
}

// ProjectExists checks if a project exists and returns its ID.
// This restores the main branch provider.ProjectExists functionality for GitLab.
func (a *Adapter) ProjectExists(ctx context.Context, owner, repo string) (bool, string, error) {
	projectPath := fmt.Sprintf("%s/%s", owner, repo)

	project, resp, err := a.client.Projects.GetProject(projectPath, &gitlab.GetProjectOptions{})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check project existence: %w", err)
	}

	return true, strconv.Itoa(project.ID), nil
}

// Protect enables branch protection for GitLab.
// This restores the main branch provider.Protect functionality.
func (a *Adapter) Protect(ctx context.Context, owner string, defaultBranch string, projectIDstr string) error {
	projectID, err := strconv.Atoi(projectIDstr)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	options := &gitlab.ProtectRepositoryBranchesOptions{
		Name:             gitlab.Ptr(defaultBranch),
		PushAccessLevel:  gitlab.Ptr(gitlab.MaintainerPermissions),
		MergeAccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions),
	}

	_, _, err = a.client.ProtectedBranches.ProtectRepositoryBranches(projectID, options)
	if err != nil {
		return fmt.Errorf("failed to protect branch: %w", err)
	}

	return nil
}

// Unprotect disables branch protection for GitLab.
// This restores the main branch provider.Unprotect functionality.
func (a *Adapter) Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error {
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	_, err = a.client.ProtectedBranches.UnprotectRepositoryBranches(projectID, defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to unprotect branch: %w", err)
	}

	return nil
}

// SetDefaultBranch sets the default branch for a GitLab repository.
// This restores the main branch provider.SetDefaultBranch functionality.
func (a *Adapter) SetDefaultBranch(ctx context.Context, owner, name, branch string) error {
	projectPath := fmt.Sprintf("%s/%s", owner, name)

	options := &gitlab.EditProjectOptions{
		DefaultBranch: gitlab.Ptr(branch),
	}

	_, _, err := a.client.Projects.EditProject(projectPath, options)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

// IsValidProjectName validates a project name for GitLab.
// This restores the main branch provider.IsValidProjectName functionality.
func (a *Adapter) IsValidProjectName(ctx context.Context, name string) bool {
	return a.ValidateRepositoryName(name) == nil
}
