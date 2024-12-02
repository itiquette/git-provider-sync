// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package github

import (
	"strings"

	"github.com/google/go-github/v67/github"
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

	p.opts.AllowAutoMerge = github.Bool(true)
	p.opts.AllowForking = github.Bool(true)
	p.opts.AllowRebaseMerge = github.Bool(true)
	p.opts.AllowSquashMerge = github.Bool(true)
	p.opts.DefaultBranch = &defaultBranch
	p.opts.Description = &description
	p.opts.Name = &name
	p.opts.Private = &isPrivate
}

func (p *ProjectOptionsBuilder) disableFeatures() {
	p.opts.AllowAutoMerge = github.Bool(false)
	p.opts.AllowMergeCommit = github.Bool(false)
	p.opts.AllowSquashMerge = github.Bool(false)
	p.opts.AllowUpdateBranch = github.Bool(false)
	p.opts.DeleteBranchOnMerge = github.Bool(false)
	p.opts.HasDownloads = github.Bool(false)
	p.opts.HasIssues = github.Bool(false)
	p.opts.HasPages = github.Bool(false)
	p.opts.HasProjects = github.Bool(false)
	p.opts.HasWiki = github.Bool(false)
}
