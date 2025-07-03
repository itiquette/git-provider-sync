// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// ProjectService provides Gitea-specific project operations.
// This restores the sophisticated project service functionality from main branch.
type ProjectService struct {
	client *gitea.Client
	logger ports.Logger
}

// NewProjectService creates a new Gitea project service.
func NewProjectService(client *gitea.Client, logger ports.Logger) *ProjectService {
	return &ProjectService{
		client: client,
		logger: logger,
	}
}

// CreateProjectRequest contains parameters for creating a Gitea project.
type CreateProjectRequest struct {
	Owner           string
	Name            string
	Description     string
	Visibility      string // "public", "private"
	DefaultBranch   string
	AutoInit        bool
	IsOrganization  bool
	DisableFeatures bool
}

// CreateProject creates a new Gitea repository with options.
func (ps *ProjectService) CreateProject(ctx context.Context, request CreateProjectRequest) (*entities.Repository, error) {
	ps.logger.Debug(ctx, "Creating Gitea project", map[string]interface{}{
		"owner":  request.Owner,
		"name":   request.Name,
		"is_org": request.IsOrganization,
	})

	// Build repository options
	createOpts := gitea.CreateRepoOption{
		Name:        request.Name,
		Description: request.Description,
		Private:     request.Visibility == "private",
		AutoInit:    request.AutoInit,
	}

	if request.DefaultBranch != "" {
		createOpts.DefaultBranch = request.DefaultBranch
	}

	var createdRepo *gitea.Repository

	var err error

	// Create in organization or user account
	if request.IsOrganization {
		createdRepo, _, err = ps.client.CreateOrgRepo(request.Owner, createOpts)
	} else {
		createdRepo, _, err = ps.client.CreateRepo(createOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea project %s: %w", request.Name, err)
	}

	// Apply disabled settings if requested
	if request.DisableFeatures {
		if err := ps.applyDisabledSettings(ctx, createdRepo.Owner.UserName, request.Name); err != nil {
			ps.logger.Warn(ctx, "Failed to apply disabled settings", map[string]interface{}{
				"owner": request.Owner,
				"name":  request.Name,
				"error": err.Error(),
			})
		}
	}

	// Convert to domain entity
	repository, err := ps.convertToRepository(createdRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created repository: %w", err)
	}

	ps.logger.Info(ctx, "Gitea project created successfully", map[string]interface{}{
		"owner": request.Owner,
		"name":  request.Name,
		"url":   repository.HTTPSURL(),
	})

	return repository, nil
}

// GetProjectInfo retrieves project information.
func (ps *ProjectService) GetProjectInfo(ctx context.Context, owner, name string) (*entities.Repository, error) {
	ps.logger.Debug(ctx, "Getting Gitea project info", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	repo, _, err := ps.client.GetRepo(owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get Gitea project %s/%s: %w", owner, name, err)
	}

	return ps.convertToRepository(repo)
}

// UpdateProject updates an existing Gitea repository.
func (ps *ProjectService) UpdateProject(ctx context.Context, owner, name string, updates ports.UpdateRepositoryOptions) error {
	ps.logger.Debug(ctx, "Updating Gitea project", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	editOpts := gitea.EditRepoOption{}

	if updates.Description != nil {
		editOpts.Description = updates.Description
	}

	if updates.Visibility != nil {
		private := *updates.Visibility == "private"
		editOpts.Private = &private
	}

	if updates.DefaultBranch != nil {
		editOpts.DefaultBranch = updates.DefaultBranch
	}

	_, _, err := ps.client.EditRepo(owner, name, editOpts)
	if err != nil {
		return fmt.Errorf("failed to update Gitea project %s/%s: %w", owner, name, err)
	}

	ps.logger.Info(ctx, "Gitea project updated successfully", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	return nil
}

// ValidateProjectName validates a Gitea repository name.
func (ps *ProjectService) ValidateProjectName(name string) error {
	if name == "" {
		return errors.New("repository name cannot be empty")
	}

	if len(name) > 100 {
		return errors.New("repository name too long (max 100 characters)")
	}

	// Gitea repository naming rules
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return errors.New("repository name cannot start or end with a period")
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return errors.New("repository name cannot start or end with a hyphen")
	}

	// Check for invalid characters
	for _, char := range name {
		if !isValidGiteaRepoChar(char) {
			return fmt.Errorf("repository name contains invalid character: %c", char)
		}
	}

	// Reserved names
	reservedNames := []string{".", "..", ".git"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved) {
			return fmt.Errorf("repository name '%s' is reserved", name)
		}
	}

	return nil
}

// applyDisabledSettings disables features on a Gitea repository.
func (ps *ProjectService) applyDisabledSettings(ctx context.Context, owner, projectName string) error {
	ps.logger.Debug(ctx, "Applying disabled settings", map[string]interface{}{
		"owner": owner,
		"repo":  projectName,
	})

	// Create edit options with all features disabled
	editOpts := gitea.EditRepoOption{
		HasIssues:       gitea.OptionalBool(false),
		HasWiki:         gitea.OptionalBool(false),
		HasProjects:     gitea.OptionalBool(false),
		HasPullRequests: gitea.OptionalBool(false),
		HasReleases:     gitea.OptionalBool(false),
		// HasActions might not be available in all Gitea versions
	}

	_, _, err := ps.client.EditRepo(owner, projectName, editOpts)
	if err != nil {
		return fmt.Errorf("failed to apply disabled settings: %w", err)
	}

	return nil
}

// convertToRepository converts Gitea repository to domain entity.
func (ps *ProjectService) convertToRepository(repo *gitea.Repository) (*entities.Repository, error) {
	builder := entities.NewRepositoryBuilder()

	var err error

	builder, err = builder.WithName(repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to set repository name: %w", err)
	}

	builder, err = builder.WithHTTPSURL(repo.CloneURL)
	if err != nil {
		return nil, fmt.Errorf("failed to set HTTPS URL: %w", err)
	}

	builder, err = builder.WithSSHURL(repo.SSHURL)
	if err != nil {
		return nil, fmt.Errorf("failed to set SSH URL: %w", err)
	}

	builder, err = builder.WithDefaultBranch(repo.DefaultBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to set default branch: %w", err)
	}

	builder = builder.WithDescription(repo.Description)

	visibility := "public"
	if repo.Private {
		visibility = "private"
	}

	builder = builder.WithVisibility(visibility)

	if !repo.Updated.IsZero() {
		builder = builder.WithLastActivityAt(repo.Updated)
	}

	builder = builder.WithProviderType("gitea")
	builder = builder.WithPrivate(repo.Private)
	builder = builder.WithFork(repo.Fork)
	builder = builder.WithArchived(repo.Archived)

	builtRepo, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build repository entity: %w", err)
	}

	return &builtRepo, nil
}

// isValidGiteaRepoChar checks if a character is valid in Gitea repo names.
func isValidGiteaRepoChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' ||
		char == '_' ||
		char == '.'
}

// OptionalBool converts a bool to gitea.OptionalBool.
func OptionalBool(val bool) *bool {
	return &val
}
