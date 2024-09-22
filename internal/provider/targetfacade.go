// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
	"itiquette/git-provider-sync/internal/target"
)

// Error variables for common failure scenarios.
var (
	ErrTargetRepositoryName = errors.New("failed target repository name validation")
	ErrCreateRepository     = errors.New("failed to create repository")
	ErrPushChanges          = errors.New("failed to push changes")
)

// Push handles the process of pushing changes to a Git provider.
// It checks if the repository exists, creates it if necessary, and then pushes the changes.
//
// Parameters:
//   - ctx: The context for the operation
//   - config: Configuration for the provider
//   - provider: The Git provider interface
//   - writer: Interface for writing to the target
//   - repository: The Git repository interface
//
// Returns an error if any step in the process fails.
func Push(ctx context.Context, config configuration.ProviderConfig, provider interfaces.GitProvider, writer interfaces.TargetWriter, repository interfaces.GitRepository, sourceGitOption model.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Push:")
	config.DebugLog(logger).Msg("Push:")

	if _, _, err := exists(ctx, config, provider, repository); err != nil {
		return fmt.Errorf("failed to check if the repository exists at provider: %w", err)
	}

	cliOptions := model.CLIOptions(ctx)

	pushOption := getPushOption(ctx, config, repository, cliOptions.ForcePush)

	if err := writer.Push(ctx, pushOption, sourceGitOption, config.Git); err != nil {
		return fmt.Errorf("%w: %w", ErrPushChanges, err)
	}

	return nil
}

// getPushOption determines the appropriate PushOption based on the provider configuration.
// It handles different scenarios for archive, directory, and remote Git providers.
func getPushOption(ctx context.Context, config configuration.ProviderConfig, repository interfaces.GitRepository, forcePush bool) model.PushOption {
	switch strings.ToLower(config.ProviderType) {
	case configuration.ARCHIVE:
		name := repository.Metainfo().Name(ctx)

		return model.NewPushOption(target.ArchiveTargetPath(name, config.ArchiveTargetDir()), false, false, model.HTTPClientOption{})
	case configuration.DIRECTORY:
		return model.NewPushOption(config.DirectoryTargetDir(), false, false, model.HTTPClientOption{})
	default:
		return model.NewPushOption(toGitURL(ctx, config, repository), false, forcePush, config.HTTPClient)
	}
}

// create attempts to create a new repository on the Git provider.
// It builds the repository description and uses the provider's Create method.
func create(ctx context.Context, config configuration.ProviderConfig, provider interfaces.GitProvider, repository interfaces.GitRepository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering create:")
	config.DebugLog(logger).Msg("create:")

	gpsUpstreamRemote, err := repository.Remote(configuration.GPSUPSTREAM)
	if err != nil || gpsUpstreamRemote.URL == "" {
		return fmt.Errorf("failed to get gpsupstream remote: %w", err)
	}

	description := buildDescription(gpsUpstreamRemote, repository, config.Repositories.Description)
	name := repository.Metainfo().Name(ctx)

	option := model.NewCreateOption(name, repository.Metainfo().Visibility, description, repository.Metainfo().DefaultBranch)

	if err := provider.Create(ctx, config, option); err != nil {
		return fmt.Errorf("%w: %s. err: %w", ErrCreateRepository, name, err)
	}

	return nil
}

// buildDescription creates a description for the repository, combining the upstream URL and existing description.
func buildDescription(gpsUpstreamRemote model.Remote, repository interfaces.GitRepository, userDescription *string) string {
	var description string
	if userDescription != nil {
		description = *userDescription
	} else {
		description = "Git Provider Sync cloned this from: " + gpsUpstreamRemote.URL + ": "
	}

	if repository.Metainfo().Description != "" {
		description += repository.Metainfo().Description
	}

	return stringconvert.RemoveLinebreaks(description)
}

// exists checks if a repository already exists on the Git provider.
// If it doesn't exist, it attempts to create it.
func exists(ctx context.Context, config configuration.ProviderConfig, provider interfaces.GitProvider, repository interfaces.GitRepository) (bool, context.Context, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering exists:")
	config.DebugLog(logger).Msg("exists:")

	if isArchiveOrDirectory(config.ProviderType) {
		return false, ctx, nil
	}

	cliOption := model.CLIOptions(ctx)
	repositoryName := repository.Metainfo().Name(ctx)

	repoExists := repositoryExists(ctx, config, provider, repositoryName)

	if !repoExists {
		logger.Debug().Str("name", repositoryName).Msg("Repository - Did not exist")

		if err := create(ctx, config, provider, repository); err != nil {
			return false, ctx, err
		}

		cliOption.ForcePush = true

		ctx = model.WithCLIOption(ctx, cliOption)
	}

	logger.Debug().Str("domain", config.Domain).Str("name", repositoryName).Msg("Repository - Exists:")

	return true, ctx, nil
}

// isArchiveOrDirectory checks if the provider is of type ARCHIVE or DIRECTORY.
func isArchiveOrDirectory(provider string) bool {
	return strings.EqualFold(provider, configuration.ARCHIVE) || strings.EqualFold(provider, configuration.DIRECTORY)
}

// repositoryExists checks if a repository with the given name exists on the provider.
func repositoryExists(ctx context.Context, config configuration.ProviderConfig, provider interfaces.GitProvider, repositoryName string) bool {
	logger := log.Logger(ctx)
	metainfos, err := provider.Metainfos(ctx, config, false)

	if err != nil {
		logger.Error().Msgf("failed to get repository meta information. Aborting run. err: %s", err.Error())
		panic(2)
	}

	for _, metainfo := range metainfos {
		if strings.EqualFold(repositoryName, metainfo.OriginalName) {
			return true
		}
	}

	return false
}

// SetGPSUpstreamRemoteFromOrigin sets the GPSUPSTREAM remote to match the ORIGIN remote.
// This ensures that the upstream remote is correctly set for syncing operations.
func SetGPSUpstreamRemoteFromOrigin(ctx context.Context, remote interfaces.GitRemote) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering SetGPSUpstreamRemoteFromOrigin")

	originRemote, err := remote.Remote(configuration.ORIGIN)
	if err != nil {
		return fmt.Errorf("failed to get origin remote: %w", err)
	}

	if err := remote.DeleteRemote(configuration.GPSUPSTREAM); err != nil {
		return fmt.Errorf("failed to delete gpsupstream remote: %w", err)
	}

	if err := remote.CreateRemote(configuration.GPSUPSTREAM, originRemote.URL, true); err != nil {
		return fmt.Errorf("failed to create gpsupstream remote: %w", err)
	}

	gpsUpstreamRemote, _ := remote.Remote(configuration.GPSUPSTREAM)
	if gpsUpstreamRemote.URL != originRemote.URL {
		return errors.New("mismatch in gpsupstream vs origin remote")
	}

	return nil
}

// toGitURL constructs a Git provider URL.
// This URL can be used for authenticated Git operations.
func toGitURL(ctx context.Context, config configuration.ProviderConfig, repository interfaces.GitRepository) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering toGitURL:")

	repositoryName := repository.Metainfo().Name(ctx)

	trimmedProviderConfigURL := strings.TrimRight(config.Domain, "/")
	projectPath := getProjectPath(config, repositoryName)

	scheme := config.Scheme
	if len(scheme) > 0 {
		return fmt.Sprintf("%s://%s/%s", scheme, trimmedProviderConfigURL, projectPath)
	}

	return fmt.Sprintf("https://%s/%s", trimmedProviderConfigURL, projectPath)
}

// getProjectPath constructs the project path based on whether it's a group or user repository.
func getProjectPath(config configuration.ProviderConfig, repositoryName string) string {
	if config.IsGroup() {
		return fmt.Sprintf("%s/%s", config.Group, repositoryName)
	}

	return fmt.Sprintf("%s/%s", config.User, repositoryName)
}
