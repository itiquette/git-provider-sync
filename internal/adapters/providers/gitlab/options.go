// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"gitlab.com/gitlab-org/api/client-go"
)

// ProjectOption is a functional option for configuring GitLab project options.
type ProjectOption func(*gitlab.CreateProjectOptions)

// WithNamespace sets the namespace ID for the project.
func WithNamespace(namespaceID int) ProjectOption {
	return func(opts *gitlab.CreateProjectOptions) {
		if namespaceID != 0 {
			opts.NamespaceID = gitlab.Ptr(namespaceID)
		}
	}
}

// WithFeatures enables or disables all features.
func WithFeatures(enabled bool) ProjectOption {
	return func(opts *gitlab.CreateProjectOptions) {
		var accessLevel gitlab.AccessControlValue
		if enabled {
			accessLevel = gitlab.EnabledAccessControl
		} else {
			accessLevel = gitlab.DisabledAccessControl
		}

		// Set all access levels
		opts.IssuesAccessLevel = &accessLevel
		opts.MergeRequestsAccessLevel = &accessLevel
		opts.ForkingAccessLevel = &accessLevel
		opts.BuildsAccessLevel = &accessLevel
		opts.WikiAccessLevel = &accessLevel
		opts.SnippetsAccessLevel = &accessLevel
		opts.PagesAccessLevel = &accessLevel
		opts.OperationsAccessLevel = &accessLevel
		opts.AnalyticsAccessLevel = &accessLevel
		opts.ContainerRegistryAccessLevel = &accessLevel
		opts.ReleasesAccessLevel = &accessLevel
		opts.EnvironmentsAccessLevel = &accessLevel
		opts.FeatureFlagsAccessLevel = &accessLevel
		opts.InfrastructureAccessLevel = &accessLevel
		opts.MonitorAccessLevel = &accessLevel
		opts.ModelExperimentsAccessLevel = &accessLevel
		opts.ModelRegistryAccessLevel = &accessLevel
		opts.RequirementsAccessLevel = &accessLevel
		opts.SecurityAndComplianceAccessLevel = &accessLevel

		// Set boolean features
		opts.PackagesEnabled = gitlab.Ptr(enabled)
		opts.SharedRunnersEnabled = gitlab.Ptr(enabled)
		opts.GroupRunnersEnabled = gitlab.Ptr(enabled)
		opts.AutoDevopsEnabled = gitlab.Ptr(false) // Usually disabled by default
		opts.LFSEnabled = gitlab.Ptr(enabled)
		opts.RequestAccessEnabled = gitlab.Ptr(enabled)
		// opts.PublicJobs = gitlab.Ptr(enabled) // Field doesn't exist, use PublicBuilds
		// opts.AutoCancelPendingPipelines field type is *string not *bool
		if enabled {
			opts.AutoCancelPendingPipelines = gitlab.Ptr("enabled")
		} else {
			opts.AutoCancelPendingPipelines = gitlab.Ptr("disabled")
		}

		opts.OnlyAllowMergeIfPipelineSucceeds = gitlab.Ptr(false) // Don't enforce by default
		opts.OnlyAllowMergeIfAllDiscussionsAreResolved = gitlab.Ptr(false)
		opts.RemoveSourceBranchAfterMerge = gitlab.Ptr(enabled)
		opts.PrintingMergeRequestLinkEnabled = gitlab.Ptr(enabled)
		opts.ResolveOutdatedDiffDiscussions = gitlab.Ptr(enabled)
		// opts.RestrictUserDefinedVariables doesn't exist
		// opts.CIJobTokenScopeEnabled doesn't exist
		opts.EmailsEnabled = gitlab.Ptr(enabled)
		opts.ShowDefaultAwardEmojis = gitlab.Ptr(enabled)
	}
}

// BuildGitLabProject creates a GitLab project configuration using functional options.
// This is a pure function - creates new object each time, no mutation.
func BuildGitLabProject(visibility, name, description, defaultBranch string, opts ...ProjectOption) *gitlab.CreateProjectOptions {
	// Convert visibility string to GitLab type
	var vis gitlab.VisibilityValue

	switch visibility {
	case visibilityPrivate:
		vis = gitlab.PrivateVisibility
	case visibilityInternal:
		vis = gitlab.InternalVisibility
	default:
		vis = gitlab.PublicVisibility
	}

	// Create base project options
	project := &gitlab.CreateProjectOptions{
		Name:                 gitlab.Ptr(name),
		Description:          gitlab.Ptr(description),
		DefaultBranch:        gitlab.Ptr(defaultBranch),
		Visibility:           &vis,
		InitializeWithReadme: gitlab.Ptr(false), // Don't auto-init by default
	}

	// Apply functional options
	for _, opt := range opts {
		opt(project)
	}

	return project
}
