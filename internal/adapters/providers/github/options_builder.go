// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"strings"

	"github.com/google/go-github/v71/github"
)

// ProjectOptionsBuilder provides sophisticated GitHub repository creation options.
// This restores the advanced options builder functionality from main branch.
type ProjectOptionsBuilder struct {
	opts *github.Repository
}

// NewProjectOptionsBuilder creates a new GitHub project options builder.
func NewProjectOptionsBuilder() *ProjectOptionsBuilder {
	return &ProjectOptionsBuilder{
		opts: &github.Repository{},
	}
}

// BasicOpts sets the fundamental repository options.
func (b *ProjectOptionsBuilder) BasicOpts(visibility, name, description, defaultBranch string) {
	isPrivate := strings.EqualFold(visibility, "private")

	b.opts.AllowAutoMerge = github.Ptr(true)
	b.opts.AllowForking = github.Ptr(true)
	b.opts.AllowRebaseMerge = github.Ptr(true)
	b.opts.AllowSquashMerge = github.Ptr(true)
	b.opts.DefaultBranch = &defaultBranch
	b.opts.Description = &description
	b.opts.Name = &name
	b.opts.Private = &isPrivate
}

// DisableFeatures disables various GitHub features for a cleaner repository.
func (b *ProjectOptionsBuilder) DisableFeatures() {
	b.opts.AllowAutoMerge = github.Ptr(false)
	b.opts.AllowMergeCommit = github.Ptr(false)
	b.opts.AllowSquashMerge = github.Ptr(false)
	b.opts.AllowUpdateBranch = github.Ptr(false)
	b.opts.DeleteBranchOnMerge = github.Ptr(false)
	b.opts.HasDownloads = github.Ptr(false)
	b.opts.HasIssues = github.Ptr(false)
	b.opts.HasPages = github.Ptr(false)
	b.opts.HasProjects = github.Ptr(false)
	b.opts.HasWiki = github.Ptr(false)
}

// EnableSecurityFeatures enables security-related repository features.
func (b *ProjectOptionsBuilder) EnableSecurityFeatures() {
	// Note: Some security features might require API calls after creation
	b.opts.HasIssues = github.Ptr(true) // Required for security advisories
}

// SetVisibility configures repository visibility.
func (b *ProjectOptionsBuilder) SetVisibility(visibility string) {
	isPrivate := strings.EqualFold(visibility, "private")
	b.opts.Private = &isPrivate
}

// SetTemplate configures template repository settings.
func (b *ProjectOptionsBuilder) SetTemplate(isTemplate bool) {
	b.opts.IsTemplate = github.Ptr(isTemplate)
}

// SetAutoInit configures automatic initialization with README.
func (b *ProjectOptionsBuilder) SetAutoInit(autoInit bool) {
	b.opts.AutoInit = github.Ptr(autoInit)
}

// SetGitIgnoreTemplate sets the gitignore template.
func (b *ProjectOptionsBuilder) SetGitIgnoreTemplate(template string) {
	if template != "" {
		b.opts.GitignoreTemplate = github.Ptr(template)
	}
}

// SetLicenseTemplate sets the license template.
func (b *ProjectOptionsBuilder) SetLicenseTemplate(license string) {
	if license != "" {
		b.opts.LicenseTemplate = github.Ptr(license)
	}
}

// SetMergeSettings configures merge-related settings.
func (b *ProjectOptionsBuilder) SetMergeSettings(allowMergeCommit, allowSquashMerge, allowRebaseMerge bool) {
	b.opts.AllowMergeCommit = github.Ptr(allowMergeCommit)
	b.opts.AllowSquashMerge = github.Ptr(allowSquashMerge)
	b.opts.AllowRebaseMerge = github.Ptr(allowRebaseMerge)
}

// SetBranchSettings configures branch-related settings.
func (b *ProjectOptionsBuilder) SetBranchSettings(deleteBranchOnMerge, allowUpdateBranch bool) {
	b.opts.DeleteBranchOnMerge = github.Ptr(deleteBranchOnMerge)
	b.opts.AllowUpdateBranch = github.Ptr(allowUpdateBranch)
}

// SetCollaborationSettings configures collaboration features.
func (b *ProjectOptionsBuilder) SetCollaborationSettings(allowForking, allowAutoMerge bool) {
	b.opts.AllowForking = github.Ptr(allowForking)
	b.opts.AllowAutoMerge = github.Ptr(allowAutoMerge)
}

// EnableAllFeatures enables all GitHub repository features.
func (b *ProjectOptionsBuilder) EnableAllFeatures() {
	b.opts.HasIssues = github.Ptr(true)
	b.opts.HasProjects = github.Ptr(true)
	b.opts.HasWiki = github.Ptr(true)
	b.opts.HasPages = github.Ptr(true)
	b.opts.HasDownloads = github.Ptr(true)
}

// Build returns the configured repository options.
func (b *ProjectOptionsBuilder) Build() *github.Repository {
	return b.opts
}

// Reset clears all options and starts fresh.
func (b *ProjectOptionsBuilder) Reset() {
	b.opts = &github.Repository{}
}

// Clone creates a copy of the current options builder.
func (b *ProjectOptionsBuilder) Clone() *ProjectOptionsBuilder {
	clone := &ProjectOptionsBuilder{
		opts: &github.Repository{},
	}

	// Copy all the pointer values
	if b.opts.Name != nil {
		clone.opts.Name = github.Ptr(*b.opts.Name)
	}

	if b.opts.Description != nil {
		clone.opts.Description = github.Ptr(*b.opts.Description)
	}

	if b.opts.Private != nil {
		clone.opts.Private = github.Ptr(*b.opts.Private)
	}

	if b.opts.DefaultBranch != nil {
		clone.opts.DefaultBranch = github.Ptr(*b.opts.DefaultBranch)
	}

	if b.opts.AutoInit != nil {
		clone.opts.AutoInit = github.Ptr(*b.opts.AutoInit)
	}

	if b.opts.HasIssues != nil {
		clone.opts.HasIssues = github.Ptr(*b.opts.HasIssues)
	}

	if b.opts.HasProjects != nil {
		clone.opts.HasProjects = github.Ptr(*b.opts.HasProjects)
	}

	if b.opts.HasWiki != nil {
		clone.opts.HasWiki = github.Ptr(*b.opts.HasWiki)
	}

	if b.opts.HasPages != nil {
		clone.opts.HasPages = github.Ptr(*b.opts.HasPages)
	}

	if b.opts.HasDownloads != nil {
		clone.opts.HasDownloads = github.Ptr(*b.opts.HasDownloads)
	}

	if b.opts.AllowMergeCommit != nil {
		clone.opts.AllowMergeCommit = github.Ptr(*b.opts.AllowMergeCommit)
	}

	if b.opts.AllowSquashMerge != nil {
		clone.opts.AllowSquashMerge = github.Ptr(*b.opts.AllowSquashMerge)
	}

	if b.opts.AllowRebaseMerge != nil {
		clone.opts.AllowRebaseMerge = github.Ptr(*b.opts.AllowRebaseMerge)
	}

	if b.opts.AllowAutoMerge != nil {
		clone.opts.AllowAutoMerge = github.Ptr(*b.opts.AllowAutoMerge)
	}

	if b.opts.AllowForking != nil {
		clone.opts.AllowForking = github.Ptr(*b.opts.AllowForking)
	}

	if b.opts.AllowUpdateBranch != nil {
		clone.opts.AllowUpdateBranch = github.Ptr(*b.opts.AllowUpdateBranch)
	}

	if b.opts.DeleteBranchOnMerge != nil {
		clone.opts.DeleteBranchOnMerge = github.Ptr(*b.opts.DeleteBranchOnMerge)
	}

	if b.opts.IsTemplate != nil {
		clone.opts.IsTemplate = github.Ptr(*b.opts.IsTemplate)
	}

	if b.opts.GitignoreTemplate != nil {
		clone.opts.GitignoreTemplate = github.Ptr(*b.opts.GitignoreTemplate)
	}

	if b.opts.LicenseTemplate != nil {
		clone.opts.LicenseTemplate = github.Ptr(*b.opts.LicenseTemplate)
	}

	return clone
}
