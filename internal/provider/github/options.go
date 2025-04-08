// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package github

import (
	"strings"

	"github.com/google/go-github/v71/github"
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

func (p *ProjectOptionsBuilder) basicOpts(visibility, name, description, defaultBranch string) {
	isPrivate := false
	if strings.EqualFold(visibility, "private") {
		isPrivate = true
	}

	p.opts.AllowAutoMerge = github.Ptr(true)
	p.opts.AllowForking = github.Ptr(true)
	p.opts.AllowRebaseMerge = github.Ptr(true)
	p.opts.AllowSquashMerge = github.Ptr(true)
	p.opts.DefaultBranch = &defaultBranch
	p.opts.Description = &description
	p.opts.Name = &name
	p.opts.Private = &isPrivate
}

func (p *ProjectOptionsBuilder) disableFeatures() {
	p.opts.AllowAutoMerge = github.Ptr(false)
	p.opts.AllowMergeCommit = github.Ptr(false)
	p.opts.AllowSquashMerge = github.Ptr(false)
	p.opts.AllowUpdateBranch = github.Ptr(false)
	p.opts.DeleteBranchOnMerge = github.Ptr(false)
	p.opts.HasDownloads = github.Ptr(false)
	p.opts.HasIssues = github.Ptr(false)
	p.opts.HasPages = github.Ptr(false)
	p.opts.HasProjects = github.Ptr(false)
	p.opts.HasWiki = github.Ptr(false)
}
