// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package github

import (
	"strings"

	"github.com/google/go-github/v66/github"
)

type ProjectOptionsBuilder struct {
	opts *github.Repository
}

func NewProjectOptionsBuilder() *ProjectOptionsBuilder {
	builder := &ProjectOptionsBuilder{
		opts: &github.Repository{},
	}

	return builder
}

func (ProjectOptionsBuilder) BasicOpts(builder *ProjectOptionsBuilder, visibility, name, description, defaultBranch string) *ProjectOptionsBuilder {
	isPrivate := false
	if strings.EqualFold(visibility, "private") {
		isPrivate = true
	}

	builder.opts.Name = &name
	builder.opts.AllowForking = github.Bool(true)
	builder.opts.Private = &isPrivate
	builder.opts.DefaultBranch = &defaultBranch
	builder.opts.Description = &description

	return builder
}

func (ProjectOptionsBuilder) DisableFeatures(builder *ProjectOptionsBuilder) *ProjectOptionsBuilder {
	builder.opts.HasIssues = github.Bool(false)
	builder.opts.HasWiki = github.Bool(false)
	builder.opts.HasPages = github.Bool(false)
	builder.opts.HasProjects = github.Bool(false)
	builder.opts.HasDownloads = github.Bool(false)
	builder.opts.AllowSquashMerge = github.Bool(false)
	builder.opts.AllowMergeCommit = github.Bool(false)
	builder.opts.AllowRebaseMerge = github.Bool(false)
	builder.opts.DeleteBranchOnMerge = github.Bool(false)
	builder.opts.AllowAutoMerge = github.Bool(false)
	builder.opts.AllowUpdateBranch = github.Bool(false)

	return builder
}
