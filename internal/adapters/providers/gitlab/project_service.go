// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/api/client-go"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// ProjectService provides GitLab-specific project operations.
// This restores the sophisticated project service functionality from main branch.
type ProjectService struct {
	client *gitlab.Client
	logger ports.Logger
}

// NewProjectService creates a new GitLab project service.
func NewProjectService(client *gitlab.Client, logger ports.Logger) *ProjectService {
	return &ProjectService{
		client: client,
		logger: logger,
	}
}

// CreateProjectRequest contains parameters for creating a GitLab project.
type CreateProjectRequest struct {
	Owner                    string
	Name                     string
	Description              string
	Visibility               string // "public", "private", "internal"
	DefaultBranch            string
	AutoInit                 bool
	IsOrganization           bool // For GitLab groups/namespaces
	DisableFeatures          bool
	IssuesEnabled            *bool
	WikiEnabled              *bool
	SnippetsEnabled          *bool
	MergeRequestsEnabled     *bool
	JobsEnabled              *bool
	ContainerRegistryEnabled *bool
	SharedRunnersEnabled     *bool
	PackagesEnabled          *bool
	PagesAccessLevel         string
	ForksEnabled             *bool
	RequestAccessEnabled     *bool
	ImportURL                string
	InitializeWithReadme     bool
}

// CreateProject creates a new GitLab repository with sophisticated options.
func (ps *ProjectService) CreateProject(ctx context.Context, request CreateProjectRequest) (*entities.Repository, error) {
	ps.logger.Debug(ctx, "Creating GitLab project", map[string]interface{}{
		"owner":      request.Owner,
		"name":       request.Name,
		"visibility": request.Visibility,
		"is_org":     request.IsOrganization,
	})

	// Build project options
	createOpts := ps.buildProjectOptions(request)

	// Create project
	project, _, err := ps.client.Projects.CreateProject(createOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab project %s: %w", request.Name, err)
	}

	// Apply disabled settings if requested
	if request.DisableFeatures {
		if err := ps.applyDisabledSettings(ctx, project.ID, request.Name); err != nil {
			ps.logger.Warn(ctx, "Failed to apply disabled settings", map[string]interface{}{
				"owner":      request.Owner,
				"name":       request.Name,
				"project_id": project.ID,
				"error":      err.Error(),
			})
		}
	}

	// Convert to domain entity
	repository, err := ps.convertToRepository(project)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created project: %w", err)
	}

	ps.logger.Info(ctx, "GitLab project created successfully", map[string]interface{}{
		"owner":      request.Owner,
		"name":       request.Name,
		"project_id": project.ID,
		"url":        repository.HTTPSURL(),
	})

	return repository, nil
}

// GetProjectInfo retrieves GitLab project information.
func (ps *ProjectService) GetProjectInfo(ctx context.Context, owner, name string) (*entities.Repository, error) {
	ps.logger.Debug(ctx, "Getting GitLab project info", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	projectPath := owner + "/" + name
	project, _, err := ps.client.Projects.GetProject(projectPath, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to get GitLab project %s: %w", projectPath, err)
	}

	return ps.convertToRepository(project)
}

// UpdateProject updates an existing GitLab repository.
func (ps *ProjectService) UpdateProject(ctx context.Context, owner, name string, updates ports.UpdateRepositoryOptions) error {
	ps.logger.Debug(ctx, "Updating GitLab project", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	projectPath := owner + "/" + name

	// Get project ID first
	project, _, err := ps.client.Projects.GetProject(projectPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get project for update: %w", err)
	}

	editOpts := &gitlab.EditProjectOptions{}

	if updates.Description != nil {
		editOpts.Description = gitlab.Ptr(*updates.Description)
	}

	if updates.Visibility != nil {
		visibility := ps.convertToGitLabVisibility(*updates.Visibility)
		editOpts.Visibility = &visibility
	}

	if updates.DefaultBranch != nil {
		editOpts.DefaultBranch = gitlab.Ptr(*updates.DefaultBranch)
	}

	_, _, err = ps.client.Projects.EditProject(project.ID, editOpts)
	if err != nil {
		return fmt.Errorf("failed to update GitLab project %s: %w", projectPath, err)
	}

	ps.logger.Info(ctx, "GitLab project updated successfully", map[string]interface{}{
		"owner": owner,
		"name":  name,
	})

	return nil
}

// ValidateProjectName validates a GitLab project name.
func (ps *ProjectService) ValidateProjectName(name string) error {
	if name == "" {
		return errors.New("project name cannot be empty")
	}

	if len(name) > 100 {
		return errors.New("project name too long (max 100 characters)")
	}

	// GitLab project naming rules
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return errors.New("project name cannot start or end with a period")
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return errors.New("project name cannot start or end with a hyphen")
	}

	// Check for invalid characters
	for _, char := range name {
		if !isValidGitLabProjectChar(char) {
			return fmt.Errorf("project name contains invalid character: %c", char)
		}
	}

	// Reserved names
	reservedNames := []string{".", "..", ".git", ".well-known", "admin", "api", "root", "help"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved) {
			return fmt.Errorf("project name '%s' is reserved", name)
		}
	}

	return nil
}

// TransformProjectName transforms a project name according to GitLab conventions.
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

// buildProjectOptions builds GitLab project options from request.
func (ps *ProjectService) buildProjectOptions(request CreateProjectRequest) *gitlab.CreateProjectOptions {
	opts := &gitlab.CreateProjectOptions{
		Name:        &request.Name,
		Description: &request.Description,
	}

	// Handle namespace (organization/group)
	if request.IsOrganization {
		opts.NamespaceID = ps.getNamespaceID(request.Owner)
	}

	// Set visibility
	visibility := ps.convertToGitLabVisibility(request.Visibility)
	opts.Visibility = &visibility

	// Set default branch
	if request.DefaultBranch != "" {
		opts.DefaultBranch = &request.DefaultBranch
	}

	// Initialize with README
	opts.InitializeWithReadme = &request.InitializeWithReadme

	// Import URL
	if request.ImportURL != "" {
		opts.ImportURL = &request.ImportURL
	}

	// Feature settings
	if request.IssuesEnabled != nil {
		if *request.IssuesEnabled {
			opts.IssuesAccessLevel = gitlab.Ptr(gitlab.EnabledAccessControl)
		} else {
			opts.IssuesAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
		}
	}

	if request.WikiEnabled != nil {
		if *request.WikiEnabled {
			opts.WikiAccessLevel = gitlab.Ptr(gitlab.EnabledAccessControl)
		} else {
			opts.WikiAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
		}
	}

	if request.SnippetsEnabled != nil {
		if *request.SnippetsEnabled {
			opts.SnippetsAccessLevel = gitlab.Ptr(gitlab.EnabledAccessControl)
		} else {
			opts.SnippetsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
		}
	}

	if request.MergeRequestsEnabled != nil {
		if *request.MergeRequestsEnabled {
			opts.MergeRequestsAccessLevel = gitlab.Ptr(gitlab.EnabledAccessControl)
		} else {
			opts.MergeRequestsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
		}
	}

	if request.JobsEnabled != nil {
		if *request.JobsEnabled {
			opts.BuildsAccessLevel = gitlab.Ptr(gitlab.EnabledAccessControl)
		} else {
			opts.BuildsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
		}
	}

	if request.ContainerRegistryEnabled != nil {
		if *request.ContainerRegistryEnabled {
			opts.ContainerRegistryAccessLevel = gitlab.Ptr(gitlab.EnabledAccessControl)
		} else {
			opts.ContainerRegistryAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
		}
	}

	if request.SharedRunnersEnabled != nil {
		opts.SharedRunnersEnabled = request.SharedRunnersEnabled
	}

	if request.PackagesEnabled != nil {
		opts.PackagesEnabled = request.PackagesEnabled
	}

	// ForksEnabled not available in this client version
	_ = request.ForksEnabled

	if request.RequestAccessEnabled != nil {
		opts.RequestAccessEnabled = request.RequestAccessEnabled
	}

	// Pages access level
	// PagesAccessLevel not available in this client version
	_ = request.PagesAccessLevel

	return opts
}

// applyDisabledSettings disables features on a GitLab project.
func (ps *ProjectService) applyDisabledSettings(ctx context.Context, projectID int, projectName string) error {
	ps.logger.Debug(ctx, "Applying disabled settings", map[string]interface{}{
		"project_id":   projectID,
		"project_name": projectName,
	})

	// Create edit options with all features disabled
	editOpts := &gitlab.EditProjectOptions{
		IssuesEnabled:            gitlab.Ptr(false),
		WikiEnabled:              gitlab.Ptr(false),
		SnippetsEnabled:          gitlab.Ptr(false),
		MergeRequestsEnabled:     gitlab.Ptr(false),
		JobsEnabled:              gitlab.Ptr(false),
		ContainerRegistryEnabled: gitlab.Ptr(false),
		SharedRunnersEnabled:     gitlab.Ptr(false),
		PackagesEnabled:          gitlab.Ptr(false),
		// ForksEnabled not available in this client version
		RequestAccessEnabled: gitlab.Ptr(false),
	}

	_, _, err := ps.client.Projects.EditProject(projectID, editOpts)
	if err != nil {
		return fmt.Errorf("failed to apply disabled settings: %w", err)
	}

	return nil
}

// convertToRepository converts GitLab project to domain entity.
func (ps *ProjectService) convertToRepository(project *gitlab.Project) (*entities.Repository, error) {
	builder := entities.NewRepositoryBuilder()

	var err error

	builder, err = builder.WithName(project.Name)
	if err != nil {
		return nil, err
	}

	builder, err = builder.WithHTTPSURL(project.HTTPURLToRepo)
	if err != nil {
		return nil, err
	}

	builder, err = builder.WithSSHURL(project.SSHURLToRepo)
	if err != nil {
		return nil, err
	}

	builder, err = builder.WithDefaultBranch(project.DefaultBranch)
	if err != nil {
		return nil, err
	}

	builder = builder.WithDescription(project.Description)

	// Convert GitLab visibility to standard format
	visibility := ps.convertFromGitLabVisibility(project.Visibility)
	builder = builder.WithVisibility(visibility)

	if project.LastActivityAt != nil {
		builder = builder.WithLastActivityAt(*project.LastActivityAt)
	}

	builder = builder.WithProviderType("gitlab")
	builder = builder.WithPrivate(project.Visibility == gitlab.PrivateVisibility)
	builder = builder.WithFork(project.ForkedFromProject != nil)
	builder = builder.WithArchived(project.Archived)

	builtRepo, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &builtRepo, nil
}

// getNamespaceID resolves namespace ID from owner name.
func (ps *ProjectService) getNamespaceID(owner string) *int {
	// This would require an API call to resolve the namespace
	// For now, we'll return nil and let GitLab handle the default
	return nil
}

// convertToGitLabVisibility converts standard visibility to GitLab visibility.
func (ps *ProjectService) convertToGitLabVisibility(visibility string) gitlab.VisibilityValue {
	switch strings.ToLower(visibility) {
	case "private":
		return gitlab.PrivateVisibility
	case "internal":
		return gitlab.InternalVisibility
	case "public":
		return gitlab.PublicVisibility
	default:
		return gitlab.PublicVisibility
	}
}

// convertFromGitLabVisibility converts GitLab visibility to standard format.
func (ps *ProjectService) convertFromGitLabVisibility(visibility gitlab.VisibilityValue) string {
	switch visibility {
	case gitlab.PrivateVisibility:
		return "private"
	case gitlab.InternalVisibility:
		return "internal"
	case gitlab.PublicVisibility:
		return "public"
	default:
		return "public"
	}
}

// Note: Access level constants not available in this client version.
// func (ps *ProjectService) convertToGitLabAccessLevel(level string) string {
// 	switch strings.ToLower(level) {
// 	case "disabled":
// 		return "disabled"
// 	case "private":
// 		return "private"
// 	case "enabled":
// 		return "enabled"
// 	case "public":
// 		return "public"
// 	default:
// 		return "private"
// 	}
// }

// isValidGitLabProjectChar checks if a character is valid in GitLab project names.
func isValidGitLabProjectChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' ||
		char == '_' ||
		char == '.'
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
		name = "project"
	}

	return name
}
