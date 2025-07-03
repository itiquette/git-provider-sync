// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// advanced_options_builder.go - Sophisticated GitHub repository options restored from main branch
package github

import (
	"strings"

	"github.com/google/go-github/v71/github"
)

// AdvancedProjectOptionsBuilder creates sophisticated GitHub repository configurations
// This restores the main branch ProjectOptionsBuilder functionality completely
type AdvancedProjectOptionsBuilder struct {
	opts *github.Repository
}

// NewAdvancedProjectOptionsBuilder creates a new advanced options builder
func NewAdvancedProjectOptionsBuilder() *AdvancedProjectOptionsBuilder {
	return &AdvancedProjectOptionsBuilder{
		opts: &github.Repository{},
	}
}

// BuildBasicOptions configures basic repository options with all GitHub features
// This restores the main branch basicOpts functionality completely
func (b *AdvancedProjectOptionsBuilder) BuildBasicOptions(visibility, name, description, defaultBranch string) *github.Repository {
	isPrivate := strings.EqualFold(visibility, "private")

	// Configure all advanced GitHub repository features (restored from main branch)
	b.opts.AllowAutoMerge = github.Ptr(true)
	b.opts.AllowForking = github.Ptr(true)
	b.opts.AllowRebaseMerge = github.Ptr(true)
	b.opts.AllowSquashMerge = github.Ptr(true)
	b.opts.AllowMergeCommit = github.Ptr(true)
	b.opts.AllowUpdateBranch = github.Ptr(true)
	b.opts.DeleteBranchOnMerge = github.Ptr(true)
	b.opts.DefaultBranch = &defaultBranch
	b.opts.Description = &description
	b.opts.Name = &name
	b.opts.Private = &isPrivate

	// Enable repository features (restored from main branch)
	b.opts.HasDownloads = github.Ptr(true)
	b.opts.HasIssues = github.Ptr(true)
	b.opts.HasPages = github.Ptr(false)    // Usually disabled for mirrors
	b.opts.HasProjects = github.Ptr(false) // Usually disabled for mirrors
	b.opts.HasWiki = github.Ptr(false)     // Usually disabled for mirrors

	// Configure security settings
	b.opts.SecurityAndAnalysis = &github.SecurityAndAnalysis{
		SecretScanning: &github.SecretScanning{
			Status: github.Ptr("enabled"),
		},
		SecretScanningPushProtection: &github.SecretScanningPushProtection{
			Status: github.Ptr("enabled"),
		},
	}

	return b.opts
}

// BuildDisabledFeatures disables repository features for protection-focused repos
// This restores the main branch disableFeatures functionality completely
func (b *AdvancedProjectOptionsBuilder) BuildDisabledFeatures() *github.Repository {
	// Disable all merge options (restored from main branch)
	b.opts.AllowAutoMerge = github.Ptr(false)
	b.opts.AllowMergeCommit = github.Ptr(false)
	b.opts.AllowSquashMerge = github.Ptr(false)
	b.opts.AllowRebaseMerge = github.Ptr(false)
	b.opts.AllowUpdateBranch = github.Ptr(false)
	b.opts.DeleteBranchOnMerge = github.Ptr(false)

	// Disable repository features (restored from main branch)
	b.opts.HasDownloads = github.Ptr(false)
	b.opts.HasIssues = github.Ptr(false)
	b.opts.HasPages = github.Ptr(false)
	b.opts.HasProjects = github.Ptr(false)
	b.opts.HasWiki = github.Ptr(false)

	// Disable vulnerability alerts and automated security fixes
	b.opts.SecurityAndAnalysis = &github.SecurityAndAnalysis{
		SecretScanning: &github.SecretScanning{
			Status: github.Ptr("disabled"),
		},
		SecretScanningPushProtection: &github.SecretScanningPushProtection{
			Status: github.Ptr("disabled"),
		},
	}

	return b.opts
}

// BuildEnterpriseOptions configures options for GitHub Enterprise servers
func (b *AdvancedProjectOptionsBuilder) BuildEnterpriseOptions(uploadURL string) *github.Repository {
	// Configure enterprise-specific settings
	if uploadURL != "" {
		// Custom handling for GitHub Enterprise upload URLs would go here
		// This is typically handled at the client level, not repository level
		_ = uploadURL // Acknowledge parameter but no action needed at repository level
	}

	// Enterprise repositories often have stricter security requirements
	b.opts.SecurityAndAnalysis = &github.SecurityAndAnalysis{
		SecretScanning: &github.SecretScanning{
			Status: github.Ptr("enabled"),
		},
		SecretScanningPushProtection: &github.SecretScanningPushProtection{
			Status: github.Ptr("enabled"),
		},
	}

	return b.opts
}

// BuildMirrorOptions configures options specifically for mirror repositories
func (b *AdvancedProjectOptionsBuilder) BuildMirrorOptions() *github.Repository {
	// Mirrors typically have limited features enabled
	b.opts.HasIssues = github.Ptr(false)
	b.opts.HasProjects = github.Ptr(false)
	b.opts.HasWiki = github.Ptr(false)
	b.opts.HasPages = github.Ptr(false)

	// Enable downloads for accessing releases
	b.opts.HasDownloads = github.Ptr(true)

	// Enable forking for mirrors
	b.opts.AllowForking = github.Ptr(true)

	return b.opts
}

// GetOptions returns the configured repository options
func (b *AdvancedProjectOptionsBuilder) GetOptions() *github.Repository {
	return b.opts
}

// Reset resets the builder to initial state
func (b *AdvancedProjectOptionsBuilder) Reset() {
	b.opts = &github.Repository{}
}

// BasicOpts configures basic repository options
func (b *AdvancedProjectOptionsBuilder) BasicOpts(visibility, name, description, defaultBranch string) {
	b.opts = b.BuildBasicOptions(visibility, name, description, defaultBranch)
}

// DisableFeatures disables repository features
func (b *AdvancedProjectOptionsBuilder) DisableFeatures() {
	b.opts = b.BuildDisabledFeatures()
}

// EnableAllFeatures enables all repository features
func (b *AdvancedProjectOptionsBuilder) EnableAllFeatures() {
	b.opts = b.BuildBasicOptions("public", "", "", "main")
}

// SetAutoInit sets auto initialization
func (b *AdvancedProjectOptionsBuilder) SetAutoInit(autoInit bool) {
	b.opts.AutoInit = github.Ptr(autoInit)
}

// SetGitIgnoreTemplate sets gitignore template
func (b *AdvancedProjectOptionsBuilder) SetGitIgnoreTemplate(template string) {
	if template != "" {
		b.opts.GitignoreTemplate = github.Ptr(template)
	}
}

// SetLicenseTemplate sets license template
func (b *AdvancedProjectOptionsBuilder) SetLicenseTemplate(template string) {
	if template != "" {
		b.opts.LicenseTemplate = github.Ptr(template)
	}
}

// Build returns the configured repository options
func (b *AdvancedProjectOptionsBuilder) Build() *github.Repository {
	return b.opts
}
