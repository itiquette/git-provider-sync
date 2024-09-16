// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"context"
	"errors"
	"fmt"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/archive"
	"itiquette/git-provider-sync/internal/provider/directory"
	"itiquette/git-provider-sync/internal/provider/github"
	"itiquette/git-provider-sync/internal/provider/gitlab"
)

var ErrNonSupportedProvider = errors.New("unsupported provider")

//nolint:ireturn
func NewGitProviderClient(ctx context.Context, option model.GitProviderClientOption) (interfaces.GitProvider, error) {
	var provider interfaces.GitProvider

	var err error

	switch option.Provider {
	// case configuration.GITEA:
	// 	provider, err = gitea.NewGiteaClient(ctx, option)
	case configuration.GITHUB:
		provider, err = github.NewGitHubClient(ctx, option)
	case configuration.GITLAB:
		provider, err = gitlab.NewGitLabClient(ctx, option)
	case configuration.ARCHIVE:
		provider = archive.Client{}
	case configuration.DIRECTORY:
		provider = directory.Client{}
	default:
		return nil, ErrNonSupportedProvider
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialized client: %s: %w", option, err)
	}

	return provider, nil
}
