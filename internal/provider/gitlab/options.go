// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlab

import (
	"github.com/xanzy/go-gitlab"
)

type ProjectOptionsBuilder struct {
	opts *gitlab.CreateProjectOptions
}

func NewProjectOptionsBuilder() *ProjectOptionsBuilder {
	builder := &ProjectOptionsBuilder{
		opts: &gitlab.CreateProjectOptions{},
	}

	return builder
}

func (ProjectOptionsBuilder) BasicOpts(builder *ProjectOptionsBuilder, visibility, name, description, defaultBranch string, namespaceID int) *ProjectOptionsBuilder {
	builder.opts.Name = gitlab.Ptr(name)
	builder.opts.Description = gitlab.Ptr(description)
	builder.opts.DefaultBranch = gitlab.Ptr(defaultBranch)
	builder.opts.Visibility = gitlab.Ptr(toVisibility(visibility))

	if namespaceID != 0 {
		builder.opts.NamespaceID = gitlab.Ptr(namespaceID)
	}

	return builder
}

func (ProjectOptionsBuilder) DisableOpts(builder *ProjectOptionsBuilder) *ProjectOptionsBuilder {
	builder.opts.BuildsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.AutoDevopsEnabled = gitlab.Ptr(false)
	builder.opts.ContainerRegistryAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.EnvironmentsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.FeatureFlagsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.GroupRunnersEnabled = gitlab.Ptr(false)
	builder.opts.InfrastructureAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.IssuesAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.LFSEnabled = gitlab.Ptr(false)
	builder.opts.MergeRequestsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.MonitorAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.PackagesEnabled = gitlab.Ptr(false)
	builder.opts.PublicBuilds = gitlab.Ptr(false)
	builder.opts.ReleasesAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.RequestAccessEnabled = gitlab.Ptr(false)
	builder.opts.RequirementsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.SecurityAndComplianceAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.SharedRunnersEnabled = gitlab.Ptr(false)
	builder.opts.SnippetsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.WikiAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.ContainerRegistryAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.PackagesEnabled = gitlab.Ptr(false)
	builder.opts.PagesAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)
	builder.opts.ModelExperimentsAccessLevel = gitlab.Ptr(gitlab.DisabledAccessControl)

	return builder
}
