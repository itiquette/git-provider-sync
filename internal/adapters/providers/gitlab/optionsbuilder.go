// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"gitlab.com/gitlab-org/api/client-go"
)

const (
	visibilityPublic = "public"
)

// ProjectOptionsBuilder provides GitLab repository creation options.
type ProjectOptionsBuilder struct {
	opts *gitlab.CreateProjectOptions
}

// NewProjectOptionsBuilder creates a new GitLab project options builder.
func NewProjectOptionsBuilder() *ProjectOptionsBuilder {
	return &ProjectOptionsBuilder{
		opts: &gitlab.CreateProjectOptions{},
	}
}

// WithBasicOpts sets the fundamental repository options with namespace support.
func (b *ProjectOptionsBuilder) WithBasicOpts(visibility, name, description, defaultBranch string, namespaceID int) {
	b.opts.DefaultBranch = gitlab.Ptr(defaultBranch)
	b.opts.Description = gitlab.Ptr(description)
	b.opts.Name = gitlab.Ptr(name)
	b.opts.Visibility = gitlab.Ptr(b.toVisibility(visibility))

	if namespaceID != 0 {
		b.opts.NamespaceID = gitlab.Ptr(namespaceID)
	}
}

// WithDisabledFeatures disables all GitLab features for a minimalist repository
//
//	feature disabling .
func (b *ProjectOptionsBuilder) WithDisabledFeatures() {
	b.configureAllFeatures(false)
}

// WithEnabledFeatures enables all GitLab features for a full-featured repository.
func (b *ProjectOptionsBuilder) WithEnabledFeatures() {
	b.configureAllFeatures(true)
}

// WithDevOpsFeatures configures CI/CD and DevOps-related features.
func (b *ProjectOptionsBuilder) WithDevOpsFeatures(enabled bool) {
	b.opts.AutoDevopsEnabled = gitlab.Ptr(enabled)
	b.opts.BuildsAccessLevel = b.accessLevel(enabled)
	b.opts.GroupRunnersEnabled = gitlab.Ptr(enabled)
	b.opts.SharedRunnersEnabled = gitlab.Ptr(enabled)
}

// WithCollaborationFeatures configures collaboration-related features.
func (b *ProjectOptionsBuilder) WithCollaborationFeatures(enabled bool) {
	b.opts.IssuesAccessLevel = b.accessLevel(enabled)
	b.opts.MergeRequestsAccessLevel = b.accessLevel(enabled)
	b.opts.WikiAccessLevel = b.accessLevel(enabled)
	b.opts.SnippetsAccessLevel = b.accessLevel(enabled)
}

// WithPackagingFeatures configures package and container registry features.
func (b *ProjectOptionsBuilder) WithPackagingFeatures(enabled bool) {
	b.opts.ContainerRegistryAccessLevel = b.accessLevel(enabled)
	b.opts.PackagesEnabled = gitlab.Ptr(enabled)
}

// WithSecurityFeatures configures security and compliance features.
func (b *ProjectOptionsBuilder) WithSecurityFeatures(enabled bool) {
	b.opts.SecurityAndComplianceAccessLevel = b.accessLevel(enabled)
	b.opts.RequirementsAccessLevel = b.accessLevel(enabled)
}

// WithDeploymentFeatures configures deployment and environment features.
func (b *ProjectOptionsBuilder) WithDeploymentFeatures(enabled bool) {
	b.opts.ReleasesAccessLevel = b.accessLevel(enabled)
	b.opts.EnvironmentsAccessLevel = b.accessLevel(enabled)
	b.opts.InfrastructureAccessLevel = b.accessLevel(enabled)
}

// WithExperimentalFeatures configures experimental features.
func (b *ProjectOptionsBuilder) WithExperimentalFeatures(enabled bool) {
	b.opts.FeatureFlagsAccessLevel = b.accessLevel(enabled)
	b.opts.ModelExperimentsAccessLevel = b.accessLevel(enabled)
}

// WithStorageFeatures configures storage-related features.
func (b *ProjectOptionsBuilder) WithStorageFeatures(enabled bool) {
	b.opts.LFSEnabled = gitlab.Ptr(enabled)
}

// WithAccessFeatures configures access and permission features.
func (b *ProjectOptionsBuilder) WithAccessFeatures(enabled bool) {
	b.opts.RequestAccessEnabled = gitlab.Ptr(enabled)
}

// WithMonitoringFeatures configures monitoring and observability features.
func (b *ProjectOptionsBuilder) WithMonitoringFeatures(enabled bool) {
	b.opts.MonitorAccessLevel = b.accessLevel(enabled)
}

// WithPagesFeatures configures GitLab Pages features.
func (b *ProjectOptionsBuilder) WithPagesFeatures(enabled bool) {
	b.opts.PagesAccessLevel = b.accessLevel(enabled)
}

// WithInitializationOptions configures repository initialization options.
func (b *ProjectOptionsBuilder) WithInitializationOptions(autoInit bool, readme, license, gitignore string) {
	b.opts.InitializeWithReadme = gitlab.Ptr(autoInit)

	if readme != "" {
		// Note: Readme content would be set via separate API call
		b.opts.InitializeWithReadme = gitlab.Ptr(true)
	}

	// Note: License and gitignore templates would be handled via separate API calls
	_ = license
	_ = gitignore
}

// WithImportOptions configures import settings.
func (b *ProjectOptionsBuilder) WithImportOptions(importURL string) {
	if importURL != "" {
		b.opts.ImportURL = gitlab.Ptr(importURL)
	}
}

// Build returns the configured project options.
func (b *ProjectOptionsBuilder) Build() *gitlab.CreateProjectOptions {
	return b.opts
}

// Reset clears all options and starts fresh.
func (b *ProjectOptionsBuilder) Reset() {
	b.opts = &gitlab.CreateProjectOptions{}
}

// ConfigureAllFeatures configures all GitLab features to enabled or disabled state.
func (b *ProjectOptionsBuilder) configureAllFeatures(enabled bool) {
	// Core DevOps features
	b.opts.AutoDevopsEnabled = gitlab.Ptr(enabled)
	b.opts.BuildsAccessLevel = b.accessLevel(enabled)
	b.opts.GroupRunnersEnabled = gitlab.Ptr(enabled)
	b.opts.SharedRunnersEnabled = gitlab.Ptr(enabled)

	// Repository features
	b.opts.IssuesAccessLevel = b.accessLevel(enabled)
	b.opts.MergeRequestsAccessLevel = b.accessLevel(enabled)
	b.opts.WikiAccessLevel = b.accessLevel(enabled)
	b.opts.SnippetsAccessLevel = b.accessLevel(enabled)

	// Package and container registry
	b.opts.ContainerRegistryAccessLevel = b.accessLevel(enabled)
	b.opts.PackagesEnabled = gitlab.Ptr(enabled)

	// Pages and documentation
	b.opts.PagesAccessLevel = b.accessLevel(enabled)

	// Monitoring and observability
	b.opts.MonitorAccessLevel = b.accessLevel(enabled)

	// Security and compliance
	b.opts.SecurityAndComplianceAccessLevel = b.accessLevel(enabled)
	b.opts.RequirementsAccessLevel = b.accessLevel(enabled)

	// Release and deployment
	b.opts.ReleasesAccessLevel = b.accessLevel(enabled)
	b.opts.EnvironmentsAccessLevel = b.accessLevel(enabled)
	b.opts.InfrastructureAccessLevel = b.accessLevel(enabled)

	// Feature flags and experimentation
	b.opts.FeatureFlagsAccessLevel = b.accessLevel(enabled)
	b.opts.ModelExperimentsAccessLevel = b.accessLevel(enabled)

	// Git LFS and access requests
	b.opts.LFSEnabled = gitlab.Ptr(enabled)
	b.opts.RequestAccessEnabled = gitlab.Ptr(enabled)
}

// AccessLevel helper function to convert boolean to access level.
func (b *ProjectOptionsBuilder) accessLevel(enabled bool) *gitlab.AccessControlValue {
	if enabled {
		return gitlab.Ptr(gitlab.EnabledAccessControl)
	}

	return gitlab.Ptr(gitlab.DisabledAccessControl)
}

// ToVisibility converts string visibility to GitLab visibility value.
func (b *ProjectOptionsBuilder) toVisibility(visibility string) gitlab.VisibilityValue {
	switch visibility {
	case visibilityPrivate:
		return gitlab.PrivateVisibility
	case visibilityInternal:
		return gitlab.InternalVisibility
	case visibilityPublic:
		return gitlab.PublicVisibility
	default:
		return gitlab.InternalVisibility
	}
}

// GetVisibilityString converts GitLab visibility to string.
func GetVisibilityString(vis gitlab.VisibilityValue) string {
	switch vis {
	case gitlab.PublicVisibility:
		return visibilityPublic
	case gitlab.PrivateVisibility:
		return visibilityPrivate
	case gitlab.InternalVisibility:
		return visibilityInternal
	default:
		return visibilityPublic
	}
}
