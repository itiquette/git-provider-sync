// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// mirror.go - Target-specific operations and writers
package synccmd

import (
	"context"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/mirror/archive"
	"itiquette/git-provider-sync/internal/mirror/directory"
	"itiquette/git-provider-sync/internal/mirror/gitbinary"
	"itiquette/git-provider-sync/internal/mirror/gitlib"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
)

func toMirror(ctx context.Context, syncCfg gpsconfig.SyncConfig, mirrorCfg gpsconfig.MirrorConfig, repositories []interfaces.GitRepository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering toTarget")
	//	targetCfg.DebugLog(logger)

	ctx = initMirrorSync(ctx, syncCfg, mirrorCfg, repositories)

	client, err := createMirrorProviderClient(ctx, mirrorCfg, syncCfg)
	if err != nil {
		return fmt.Errorf("failed to create mirror provider client: %w", err)
	}

	for _, repo := range repositories {
		repo.ProjectInfo().Name(ctx)

		if err := processRepository(ctx, mirrorCfg, client, repo, syncCfg); err != nil {
			return fmt.Errorf("failed to process repository: %w", err)
		}
	}

	summary(ctx, syncCfg)

	return nil
}

func processRepository(ctx context.Context, mirrorCfg gpsconfig.MirrorConfig, client interfaces.GitProvider, repo interfaces.GitRepository, sourceCfg gpsconfig.SyncConfig) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering processRepository")
	repo.ProjectInfo().DebugLog(logger).Msg("processRepository")

	if repo.ProjectInfo().OriginalName == "" {
		return ErrEmptyMetainfo
	}

	ignoreRepository, err := validateRepository(ctx, client, repo, mirrorCfg)
	if err != nil {
		return err
	}

	if ignoreRepository {
		logger.Warn().Str("name", repo.ProjectInfo().Name(ctx)).Msg("Ignoring invalid repository")

		return nil
	}

	if err := prepareRepository(ctx, mirrorCfg, repo); err != nil {
		return fmt.Errorf("failed to prepare repository: %w", err)
	}

	if err := pushRepository(ctx, sourceCfg, mirrorCfg, client, repo); err != nil {
		return fmt.Errorf("failed to push repository: %w", err)
	}

	if mirrorCfg.ProviderType == gpsconfig.DIRECTORY {
		gitHandler := directory.NewGitHandler(gitlib.NewService())
		storageHandler := directory.NewStorageHandler()
		dirService := directory.NewService(gitHandler, storageHandler)

		if err := dirService.Pull(ctx, sourceCfg, mirrorCfg.Path, repo); err != nil {
			return fmt.Errorf("failed to pull repository for directory target: %w", err)
		}
	}

	return nil
}

func validateRepository(ctx context.Context, client interfaces.GitProvider, repo interfaces.GitRepository, mirrorCfg gpsconfig.MirrorConfig) (bool, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering validateRepository")

	cliOpts := model.CLIOptions(ctx)

	name := repo.ProjectInfo().OriginalName
	if cliOpts.ASCIIName || mirrorCfg.Settings.ASCIIName {
		name = stringconvert.RemoveNonAlphaNumericChars(ctx, name)
		p := repo.ProjectInfo()
		p.SetCleanName(name)
	}

	ignoreRepository := false
	if client.IsValidProjectName(ctx, name) {
		return ignoreRepository, nil
	}

	if meta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(*model.SyncRunMetainfo); ok {
		(*meta.Fail)["invalid"] = append((*meta.Fail)["invalid"], name)

		if cliOpts.IgnoreInvalidName || mirrorCfg.Settings.IgnoreInvalidName {
			return true, nil
		}
	}

	if !cliOpts.IgnoreInvalidName && !mirrorCfg.Settings.IgnoreInvalidName {
		return ignoreRepository, fmt.Errorf("%w: %s", ErrInvalidRepoName, name)
	}

	log.Logger(ctx).Debug().
		Str("name", name).
		Bool("ignoreInvalidName", cliOpts.IgnoreInvalidName).
		Msg("invalid repository name, ignoring")

	return ignoreRepository, nil
}
func prepareRepository(ctx context.Context, mirrorCfg gpsconfig.MirrorConfig, repo interfaces.GitRepository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering prepareRepository")

	if mirrorCfg.ProviderType == gpsconfig.ARCHIVE {
		return nil
	}

	if err := provider.SetGPSUpstreamRemoteFromOrigin(ctx, repo); err != nil {
		return fmt.Errorf("create gpsupstream remote: %w", err)
	}

	return nil
}

func createMirrorProviderClient(ctx context.Context, mirrorCfg gpsconfig.MirrorConfig, syncCfg gpsconfig.SyncConfig) (interfaces.GitProvider, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering createProviderClient")

	client, err := provider.NewGitProviderClient(ctx, model.GitProviderClientOption{
		ProviderType: mirrorCfg.ProviderType,
		AuthCfg:      mirrorCfg.Auth,
		Domain:       mirrorCfg.GetDomain(),
		Repositories: syncCfg.Repositories,
		UploadURL:    mirrorCfg.Settings.GitHubUploadURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider client: %w", err)
	}

	return client, nil
}

func pushRepository(ctx context.Context, sourceCfg gpsconfig.SyncConfig, mirrorCfg gpsconfig.MirrorConfig, client interfaces.GitProvider, repo interfaces.GitRepository) error {
	writer, err := getMirrorWriter(mirrorCfg, sourceCfg.UseGitBinary)
	if err != nil {
		return fmt.Errorf("get mirror writer: %w", err)
	}

	if err := provider.Push(ctx, mirrorCfg, client, writer, repo, sourceCfg); err != nil {
		return fmt.Errorf("failed to push to mirror target: %w", err)
	}

	incrementSyncCount(ctx)

	return nil
}

func getMirrorWriter(mirrorCfg gpsconfig.MirrorConfig, useGitBin bool) (interfaces.MirrorWriter, error) {
	switch strings.ToLower(mirrorCfg.ProviderType) {
	case gpsconfig.ARCHIVE:
		gitHandler := archive.NewGitHandler(gitlib.NewService())
		storageHandler := archive.NewStorageHandler()
		archiverHandler := archive.NewHandler()

		return archive.NewService(*gitHandler, storageHandler, archiverHandler), nil
	case gpsconfig.DIRECTORY:
		gitHandler := directory.NewGitHandler(gitlib.NewService())
		storageHandler := directory.NewStorageHandler()

		return directory.NewService(gitHandler, storageHandler), nil
	default:
		if useGitBin {
			writer, err := gitbinary.NewService()
			if err != nil {
				return nil, fmt.Errorf("create git binary writer: %w", err)
			}

			return writer, nil
		}

		return gitlib.NewService(), nil
	}
}

func incrementSyncCount(ctx context.Context) {
	if meta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
		meta.Total++
	}
}
