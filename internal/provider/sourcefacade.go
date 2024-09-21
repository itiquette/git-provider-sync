// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package provider handles operations related to Git providers and repositories.
package provider

import (
	"context"
	"fmt"
	"path/filepath"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
)

// Clone clones multiple repositories based on their metadata.
// It takes a context, a SourceReader interface for cloning operations,
// and a slice of RepositoryMetainfo containing information about the repositories to clone.
// It returns a slice of GitRepository interfaces representing the cloned repositories and any error encountered.
func Clone(ctx context.Context, reader interfaces.SourceReader, gitInfo model.GitOption, metainfos []model.RepositoryMetainfo, httpClient model.HTTPClientOption) ([]interfaces.GitRepository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Cloning repositories")

	repositories := make([]interfaces.GitRepository, 0, len(metainfos))
	tmpDirPath, err := model.GetTmpDirPath(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to clone, could not get tmp dir path: %s %w", tmpDirPath, err)
	}

	for _, metainfo := range metainfos {
		targetPath := filepath.Join(tmpDirPath, metainfo.OriginalName)

		option := model.NewCloneOption(ctx, metainfo, true, targetPath, gitInfo, httpClient)

		resultRepo, err := reader.Clone(ctx, option)
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository %s: %w", metainfo.OriginalName, err)
		}

		resultRepo.Meta = metainfo
		repositories = append(repositories, resultRepo)
	}

	return repositories, nil
}

// FetchMetainfo retrieves metadata information for repositories from a Git provider.
// It takes a context, provider configuration, and a GitProvider interface.
// It returns a slice of RepositoryMetainfo containing the fetched metadata and any error encountered.
func FetchMetainfo(ctx context.Context, config configuration.ProviderConfig, gitProvider interfaces.GitProvider) ([]model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)

	// Log the metadata fetching operation
	logger.Info().
		Str("domain", config.Domain).
		Str("provider", gitProvider.Name()).
		Msg("Fetching source meta info")

	// Fetch the metadata from the Git provider
	// The 'true' parameter likely indicates that all available metadata should be fetched
	metainfo, err := gitProvider.Metainfos(ctx, config, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository meta information: %w", err)
	}

	return metainfo, nil
}
