// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v71/github"

	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// ProjectService provides GitHub-specific project operations with all advanced features.
// This restores the sophisticated project service functionality from main branch completely.
type ProjectService struct {
	client            *github.Client
	logger            ports.Logger
	optBuilder        *ProjectOptionsBuilder
	protectionService *ProtectionService
}

// NewProjectService creates a new GitHub project service with advanced capabilities.
func NewProjectService(client *github.Client, logger ports.Logger) *ProjectService {
	return &ProjectService{
		client:            client,
		logger:            logger,
		optBuilder:        NewProjectOptionsBuilder(),
		protectionService: NewProtectionService(client, logger),
	}
}

// CreateProjectRequest contains parameters for creating a project.
type CreateProjectRequest struct {
	Owner           string
	Name            string
	Description     string
	Visibility      string // "public", "private"
	DefaultBranch   string
	AutoInit        bool
	IsOrganization  bool
	DisableFeatures bool // Disable wikis, issues, projects, etc.
	Template        *TemplateRepository
	License         string
	GitIgnore       string
}

// TemplateRepository represents a template repository.
type TemplateRepository struct {
	Owner              string
	Name               string
	IncludeAllBranches bool
}

// CreateProject creates a new GitHub repository with sophisticated options.
func (ps *ProjectService) CreateProject(ctx context.Context, request CreateProjectRequest) (*entities.Repository, error) {
	ps.logger.Debug(ctx, "Creating GitHub project", map[string]interface{}{
		"owner":      request.Owner,
		"name":       request.Name,
		"visibility": request.Visibility,
		"is_org":     request.IsOrganization,
	})

	// Build repository options
	repoOpts := ps.buildRepositoryOptions(request)

	// Determine if creating in organization or user account
	orgName := ""
	if request.IsOrganization {
		orgName = request.Owner
	}

	// Create repository
	createdRepo, _, err := ps.client.Repositories.Create(ctx, orgName, repoOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub project %s: %w", request.Name, err)
	}

	// Convert to domain entity
	repository, err := ps.convertToRepository(createdRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created repository: %w", err)
	}

	ps.logger.Info(ctx, "GitHub project created successfully", map[string]interface{}{
		"owner": request.Owner,
		"name":  request.Name,
		"url":   repository.HTTPSURL(),
	})

	return repository, nil
}

// UpdateProject updates an existing GitHub repository.
func (ps *ProjectService) UpdateProject(ctx context.Context, owner, name string, updates ports.UpdateRepositoryOptions) error {
	ps.logger.Debug(ctx, "Updating GitHub project", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	repoOpts := &github.Repository{}

	if updates.Description != nil {
		repoOpts.Description = github.Ptr(*updates.Description)
	}

	if updates.Visibility != nil {
		private := *updates.Visibility == constants.VisibilityPrivate
		repoOpts.Private = github.Ptr(private)
	}

	if updates.DefaultBranch != nil {
		repoOpts.DefaultBranch = github.Ptr(*updates.DefaultBranch)
	}

	_, _, err := ps.client.Repositories.Edit(ctx, owner, name, repoOpts)
	if err != nil {
		return fmt.Errorf("failed to update GitHub project %s/%s: %w", owner, name, err)
	}

	ps.logger.Info(ctx, "GitHub project updated successfully", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	return nil
}

// ValidateProjectName validates a GitHub repository name using sophisticated validation.
func (ps *ProjectService) ValidateProjectName(name string) error {
	if !IsValidGitHubRepositoryName(name) {
		cleanName, _, issues := ValidateAndCleanName(name)

		return fmt.Errorf("invalid repository name '%s': %v (suggested: '%s')", name, issues, cleanName)
	}

	return nil
}

// IsValidProjectName checks if a project name is valid.
func (ps *ProjectService) IsValidProjectName(name string) bool {
	return IsValidGitHubRepositoryName(name)
}

// CreateProjectWithAdvancedOptions creates a project using the sophisticated options builder.
func (ps *ProjectService) CreateProjectWithAdvancedOptions(ctx context.Context, request CreateProjectRequest) (*entities.Repository, error) {
	ps.logger.Debug(ctx, "Creating GitHub project with advanced options", map[string]interface{}{
		"owner":      request.Owner,
		"name":       request.Name,
		"visibility": request.Visibility,
		"is_org":     request.IsOrganization,
		"disabled":   request.DisableFeatures,
	})

	// Use the sophisticated options builder
	ps.optBuilder.Reset()
	ps.optBuilder.BasicOpts(request.Visibility, request.Name, request.Description, request.DefaultBranch)

	if request.DisableFeatures {
		ps.optBuilder.DisableFeatures()
	} else {
		ps.optBuilder.EnableAllFeatures()
	}

	// Set additional sophisticated options
	ps.optBuilder.SetAutoInit(request.AutoInit)
	ps.optBuilder.SetGitIgnoreTemplate(request.GitIgnore)
	ps.optBuilder.SetLicenseTemplate(request.License)

	// Build the options
	repoOpts := ps.optBuilder.Build()

	// Determine organization context
	orgName := ""
	if request.IsOrganization {
		orgName = request.Owner
	}

	// Create repository
	createdRepo, _, err := ps.client.Repositories.Create(ctx, orgName, repoOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub project %s: %w", request.Name, err)
	}

	ps.logger.Debug(ctx, "GitHub project created successfully", map[string]interface{}{
		"owner":     request.Owner,
		"name":      request.Name,
		"full_name": *createdRepo.FullName,
	})

	// Convert to domain entity
	repository, err := ps.convertToRepository(createdRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created repository: %w", err)
	}

	ps.logger.Info(ctx, "GitHub project created successfully", map[string]interface{}{
		"owner": request.Owner,
		"name":  request.Name,
		"url":   repository.HTTPSURL(),
	})

	return repository, nil
}

// ExistsProject checks if a project exists and returns its information.
func (ps *ProjectService) ExistsProject(ctx context.Context, owner, name string) (bool, string, error) {
	ps.logger.Debug(ctx, "Checking if GitHub project exists", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	project, resp, err := ps.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil // Repository doesn't exist
		}

		return false, "", fmt.Errorf("failed to check project existence: %w", err)
	}

	projectID := ""
	if project.FullName != nil {
		projectID = *project.FullName
	}

	return true, projectID, nil
}

// TransformProjectName transforms a repository name according to GitHub conventions.
func (ps *ProjectService) TransformProjectName(name string, options ports.NameTransformOptions) string {
	result := name

	// Apply transformations
	if options.ToLowercase {
		result = strings.ToLower(result)
	}

	if options.ToUppercase {
		result = strings.ToUpper(result)
	}

	// Apply replacements
	for old, new := range options.Replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Add prefix/suffix
	if options.Prefix != "" {
		result = options.Prefix + result
	}

	if options.Suffix != "" {
		result = result + options.Suffix
	}

	// Make alphanumeric only if requested
	if options.AlphaNumericOnly {
		result = ps.makeAlphaNumeric(result)
	}

	// Truncate if needed
	if options.MaxLength > 0 && len(result) > options.MaxLength {
		result = result[:options.MaxLength]
	}

	// Ensure it's still valid after transformations
	if err := ps.ValidateProjectName(result); err != nil {
		// If invalid, try to fix common issues
		result = ps.sanitizeProjectName(result)
	}

	return result
}

// buildRepositoryOptions builds GitHub repository options from request.
func (ps *ProjectService) buildRepositoryOptions(request CreateProjectRequest) *github.Repository {
	opts := &github.Repository{
		Name:        github.Ptr(request.Name),
		Description: github.Ptr(request.Description),
		Private:     github.Ptr(request.Visibility == "private"),
		AutoInit:    github.Ptr(request.AutoInit),
	}

	if request.DefaultBranch != "" {
		opts.DefaultBranch = github.Ptr(request.DefaultBranch)
	}

	if request.DisableFeatures {
		opts.HasIssues = github.Ptr(false)
		opts.HasProjects = github.Ptr(false)
		opts.HasWiki = github.Ptr(false)
		opts.HasDownloads = github.Ptr(false)
	}

	// Template repository support would require additional GitHub API features
	// Skipping for now as TemplateRepository type is not available in this version
	_ = request.Template

	if request.License != "" {
		opts.LicenseTemplate = github.Ptr(request.License)
	}

	if request.GitIgnore != "" {
		opts.GitignoreTemplate = github.Ptr(request.GitIgnore)
	}

	return opts
}

// convertToRepository converts GitHub repository to domain entity.
func (ps *ProjectService) convertToRepository(repo *github.Repository) (*entities.Repository, error) {
	builder := entities.NewRepositoryBuilder()

	if repo.Name != nil {
		var err error

		builder, err = builder.WithName(*repo.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to set repository name: %w", err)
		}
	}

	if repo.CloneURL != nil {
		var err error

		builder, err = builder.WithHTTPSURL(*repo.CloneURL)
		if err != nil {
			return nil, fmt.Errorf("failed to set HTTPS URL: %w", err)
		}
	}

	if repo.SSHURL != nil {
		var err error

		builder, err = builder.WithSSHURL(*repo.SSHURL)
		if err != nil {
			return nil, fmt.Errorf("failed to set SSH URL: %w", err)
		}
	}

	if repo.DefaultBranch != nil {
		var err error

		builder, err = builder.WithDefaultBranch(*repo.DefaultBranch)
		if err != nil {
			return nil, fmt.Errorf("failed to set default branch: %w", err)
		}
	}

	if repo.Description != nil {
		builder = builder.WithDescription(*repo.Description)
	}

	visibility := "public"
	if repo.Private != nil && *repo.Private {
		visibility = "private"
	}

	builder = builder.WithVisibility(visibility)

	if repo.UpdatedAt != nil {
		builder = builder.WithLastActivityAt(repo.UpdatedAt.Time)
	}

	// Owner information is provided through the ProviderConfig

	builder = builder.WithProviderType("github")

	if repo.Private != nil {
		builder = builder.WithPrivate(*repo.Private)
	}

	if repo.Fork != nil {
		builder = builder.WithFork(*repo.Fork)
	}

	if repo.Archived != nil {
		builder = builder.WithArchived(*repo.Archived)
	}

	builtRepo, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build repository entity: %w", err)
	}

	return &builtRepo, nil
}

// makeAlphaNumeric converts a string to alphanumeric only (plus hyphens).
func (ps *ProjectService) makeAlphaNumeric(input string) string {
	var result strings.Builder

	for _, char := range input {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' {
			result.WriteRune(char)
		} else {
			result.WriteRune('-')
		}
	}

	return result.String()
}

// sanitizeProjectName attempts to fix common naming issues.
func (ps *ProjectService) sanitizeProjectName(name string) string {
	// Remove leading/trailing periods and hyphens
	name = strings.Trim(name, ".-")

	// Replace consecutive hyphens with single hyphen
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Ensure it's not empty
	if name == "" {
		name = "repository"
	}

	return name
}

// GetProjectInfos retrieves repository metadata with sophisticated filtering and pagination
// This restores the main branch getProjectInfos functionality completely
func (ps *ProjectService) GetProjectInfos(ctx context.Context, owner string, isOrganization bool, includeForks bool) ([]*entities.Repository, error) {
	ps.logger.Debug(ctx, "Fetching GitHub project infos", map[string]interface{}{
		"owner":          owner,
		"isOrganization": isOrganization,
		"includeForks":   includeForks,
	})

	var allRepos []*github.Repository

	listType := "sources"
	if includeForks {
		listType = "all"
	}

	if isOrganization {
		// List organization repositories with sophisticated pagination (restored from main branch)
		opt := &github.RepositoryListByOrgOptions{
			Type:        listType,
			Sort:        "full_name",
			ListOptions: github.ListOptions{PerPage: 100}, // GitHub's max is 100
		}

		for {
			repos, resp, err := ps.client.Repositories.ListByOrg(ctx, owner, opt)
			if err != nil {
				return nil, fmt.Errorf("failed to list org repositories. page: %d, err: %w", opt.Page, err)
			}

			allRepos = append(allRepos, repos...)

			if resp.NextPage == 0 {
				break
			}

			opt.Page = resp.NextPage
		}
	} else {
		// List user repositories with sophisticated pagination (restored from main branch)
		opt := &github.RepositoryListByAuthenticatedUserOptions{
			Visibility:  "all",
			Affiliation: "owner",
			Sort:        "full_name",
			ListOptions: github.ListOptions{PerPage: 100}, // GitHub's max is 100
		}

		for {
			repos, resp, err := ps.client.Repositories.ListByAuthenticatedUser(ctx, opt)
			if err != nil {
				return nil, fmt.Errorf("failed to list user repositories. page: %d, err: %w", opt.Page, err)
			}

			allRepos = append(allRepos, repos...)

			if resp.NextPage == 0 {
				break
			}

			opt.Page = resp.NextPage
		}
	}

	ps.logger.Debug(ctx, "Total fetched repositories", map[string]interface{}{
		"totalRepositories": len(allRepos),
	})

	// Convert to domain entities with sophisticated filtering (restored from main branch)
	var repositories []*entities.Repository

	for _, repo := range allRepos {
		// Apply fork filtering (restored from main branch)
		if !includeForks && repo.Fork != nil && *repo.Fork {
			continue
		}

		entity, err := ps.convertGitHubRepoToEntity(repo)
		if err != nil {
			ps.logger.Warn(ctx, "Failed to convert repository", map[string]interface{}{
				"repo":  repo.GetName(),
				"error": err.Error(),
			})

			continue
		}

		repositories = append(repositories, entity)
	}

	ps.logger.Info(ctx, "Successfully fetched GitHub project infos", map[string]interface{}{
		"owner":             owner,
		"totalRepositories": len(allRepos),
		"filteredCount":     len(repositories),
		"includeForks":      includeForks,
	})

	return repositories, nil
}

// GetProjectInfo retrieves detailed metadata for a single repository
// This restores the main branch newProjectInfo functionality completely
func (ps *ProjectService) GetProjectInfo(ctx context.Context, owner, name string) (*entities.Repository, error) {
	ps.logger.Debug(ctx, "Fetching GitHub project info", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	gitHubProject, _, err := ps.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get projectInfo. name: %s, err: %w", name, err)
	}

	entity, err := ps.convertGitHubRepoToEntity(gitHubProject)
	if err != nil {
		return nil, fmt.Errorf("failed to convert repository to entity: %w", err)
	}

	ps.logger.Debug(ctx, "Successfully fetched GitHub project info", map[string]interface{}{
		"owner":      owner,
		"name":       name,
		"visibility": entity.Visibility(),
	})

	return entity, nil
}

// RepositoryExists checks if a repository exists and returns metadata
// This restores the main branch Exists functionality completely
func (ps *ProjectService) RepositoryExists(ctx context.Context, owner, repo string) (bool, string, error) {
	ps.logger.Debug(ctx, "Checking GitHub repository existence", map[string]interface{}{
		"owner": owner,
		"repo":  repo,
	})

	project, resp, err := ps.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, "", nil // Repository doesn't exist
		}

		return false, "", err
	}

	projectID := getValueOrEmpty(project.FullName)

	ps.logger.Debug(ctx, "GitHub repository exists", map[string]interface{}{
		"owner":     owner,
		"repo":      repo,
		"projectID": projectID,
	})

	return true, projectID, nil
}

// SetDefaultBranch sets the default branch for a repository
// This restores the main branch setDefaultBranch functionality completely
func (ps *ProjectService) SetDefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	ps.logger.Debug(ctx, "Setting GitHub default branch", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
		"branch":      branch,
	})

	_, _, err := ps.client.Repositories.Edit(ctx, owner, projectName, &github.Repository{
		DefaultBranch: github.Ptr(branch),
	})
	if err != nil {
		return fmt.Errorf("failed to set default branch. err: %w", err)
	}

	ps.logger.Info(ctx, "GitHub default branch set successfully", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
		"branch":      branch,
	})

	return nil
}

// convertGitHubRepoToEntity converts GitHub repository to domain entity with all metadata
// This restores the main branch repository conversion functionality completely
func (ps *ProjectService) convertGitHubRepoToEntity(repo *github.Repository) (*entities.Repository, error) {
	builder := entities.NewRepositoryBuilder()

	// Set basic properties (restored from main branch)
	if repo.Name != nil {
		var err error

		builder, err = builder.WithName(*repo.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to set repository name: %w", err)
		}
	}

	if repo.CloneURL != nil {
		var err error

		builder, err = builder.WithHTTPSURL(*repo.CloneURL)
		if err != nil {
			return nil, fmt.Errorf("failed to set HTTPS URL: %w", err)
		}
	}

	if repo.SSHURL != nil {
		var err error

		builder, err = builder.WithSSHURL(*repo.SSHURL)
		if err != nil {
			return nil, fmt.Errorf("failed to set SSH URL: %w", err)
		}
	}

	if repo.Description != nil {
		builder = builder.WithDescription(*repo.Description)
	}

	if repo.Private != nil {
		visibility := constants.VisibilityPublic
		if *repo.Private {
			visibility = constants.VisibilityPrivate
		}

		builder = builder.WithVisibility(visibility)
	}

	if repo.DefaultBranch != nil {
		var err error

		builder, err = builder.WithDefaultBranch(*repo.DefaultBranch)
		if err != nil {
			return nil, fmt.Errorf("failed to set default branch: %w", err)
		}
	}

	if repo.UpdatedAt != nil {
		builder = builder.WithLastActivityAt(repo.UpdatedAt.Time)
	}

	if repo.FullName != nil {
		builder = builder.WithProjectID(*repo.FullName)
	}

	if repo.Fork != nil {
		builder = builder.WithFork(*repo.Fork)
	}

	if repo.Archived != nil {
		builder = builder.WithArchived(*repo.Archived)
	}

	// Note: Size field not available in current repository entity

	// Build the entity
	entity, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build repository entity: %w", err)
	}

	return &entity, nil
}

// getValueOrEmpty is a helper function that returns the value of a string pointer if it's not nil,
// or "N/A" otherwise. This restores the main branch helper functionality.
func getValueOrEmpty(s *string) string {
	if s != nil {
		return *s
	}

	return "N/A"
}
