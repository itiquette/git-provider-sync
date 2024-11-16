// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitea

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/log"

	"code.gitea.io/sdk/gitea"
)

type ProjectOptionsBuilder struct {
	opts *gitea.CreateRepoOption
}

func NewProjectOptionsBuilder() *ProjectOptionsBuilder {
	builder := &ProjectOptionsBuilder{
		opts: &gitea.CreateRepoOption{},
	}

	return builder
}

func (ProjectOptionsBuilder) BasicOpts(builder *ProjectOptionsBuilder, _, name, description, defaultBranch string) *ProjectOptionsBuilder {
	builder.opts.Name = name
	builder.opts.Description = description
	builder.opts.DefaultBranch = defaultBranch
	//	builder.opts.Private = toVisibility(/* visibility */)

	return builder
}

func (p ProjectService) ApplyDisabledSettings(ctx context.Context, owner, repo string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering gitea:ApplyDisabledSettings")
	logger.Debug().Str("owner", owner).Str("repo", repo).Msg("Entering gitea:ApplyDisabledSettings")

	// These settings can only be applied after repository creation
	editOpts := gitea.EditRepoOption{
		HasIssues:       new(bool),
		HasWiki:         new(bool),
		HasProjects:     new(bool),
		HasPullRequests: new(bool),
		HasReleases:     new(bool),
		HasActions:      new(bool),
	}

	// Disable all features
	*editOpts.HasIssues = false
	*editOpts.HasWiki = false
	*editOpts.HasProjects = false
	*editOpts.HasPullRequests = false
	*editOpts.HasReleases = false
	*editOpts.HasActions = false

	_, _, err := p.client.EditRepo(owner, repo, editOpts)
	if err != nil {
		return fmt.Errorf("failed to edit repository settings: owner: %s, repo: %s, err: %w", owner, repo, err)
	}

	return nil
}
