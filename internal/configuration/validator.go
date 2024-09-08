// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Define error variables for various configuration validation scenarios.
var (
	ErrUnsupportedProvider          = errors.New("unsupported provider")
	ErrUnsupportedProviderURL       = errors.New("unsupported Git provider URL")
	ErrNoTargetProviders            = errors.New("no target provider/s configured")
	ErrUnsupportedArchiveProvider   = errors.New("source provider: does not support reading from archive")
	ErrUnsupportedDirectoryProvider = errors.New("source provider: does not support reading from directory")
	ErrNoSourceDomain               = errors.New("source provider: no domain configured")
	ErrNoSourceGroupOrUser          = errors.New("source provider: no group path or user configured")
	ErrBothGroupAndUser             = errors.New("provider: group path and user configured, only one is allowed")
	ErrArchiveMissingTargetPath     = errors.New("archive target provider: missing property archivetargetdir")
	ErrDirectoryMissingTargetPath   = errors.New("directory target provider: missing property directorytargetdir")
	ErrExcludeIsConfiguredButEmpty  = errors.New("exclude is configured but 'repositories:' contains no repository names")
	ErrIncludeIsConfiguredButEmpty  = errors.New("include is configured but 'repositories:' contains no repository names")
	ErrNoTargetDomain               = errors.New("target provider: no domain configured")
	ErrNoTargetGroupOrUser          = errors.New("target provider: no group path or user configured")
	ErrTokenAuth                    = errors.New("target provider currently only supports token auth")
	ErrHasNoHTTPPrefix              = errors.New("target provider currently only supports http/s")
	ErrTargetURLValidFormat         = errors.New("target url must be a Git provider URL")
	ErrNoTargetToken                = errors.New("no target token set")
)

// ValidGitProviders is a slice of supported Git providers.
var ValidGitProviders = []string{GITHUB, GITLAB, ARCHIVE, GITEA, DIRECTORY}

// ValidateConfiguration checks the entire ProvidersConfig for validity.
func ValidateConfiguration(providersConfig ProvidersConfig) error {
	if err := validateSourceProvider(providersConfig.SourceProvider); err != nil {
		return err
	}

	if len(providersConfig.ProviderTargets) == 0 {
		return ErrNoTargetProviders
	}

	for _, target := range providersConfig.ProviderTargets {
		if err := validateTargetProvider(target); err != nil {
			return fmt.Errorf("failed to validate target provider: %w", err)
		}
	}

	return nil
}

// validateSourceProvider checks the validity of the source provider configuration.
func validateSourceProvider(provider ProviderConfig) error {
	if !slices.Contains(ValidGitProviders, provider.Provider) {
		return fmt.Errorf("source provider: must be one of %v: %w", ValidGitProviders, ErrUnsupportedProvider)
	}

	if strings.EqualFold(provider.Provider, ARCHIVE) {
		return ErrUnsupportedArchiveProvider
	}

	if strings.EqualFold(provider.Provider, DIRECTORY) {
		return ErrUnsupportedDirectoryProvider
	}

	if len(provider.Domain) == 0 {
		return ErrNoSourceDomain
	}

	if err := validateGroupAndUser(provider); err != nil {
		return err
	}

	if err := validateRepositoryLists(provider); err != nil {
		return err
	}

	return validateProviderSpecific(provider.Provider, provider.Providerspecific)
}

// validateTargetProvider checks the validity of a target provider configuration.
func validateTargetProvider(config ProviderConfig) error {
	if len(config.Provider) == 0 || !slices.Contains(ValidGitProviders, config.Provider) {
		return fmt.Errorf("target provider: must be one of %v: %w", ValidGitProviders, ErrUnsupportedProvider)
	}

	if !strings.EqualFold(config.Provider, ARCHIVE) && !strings.EqualFold(config.Provider, DIRECTORY) {
		if err := validateStandardProvider(config); err != nil {
			return err
		}
	}

	return validateProviderSpecific(config.Provider, config.Providerspecific)
}

// validateStandardProvider checks the validity of standard (non-archive, non-directory) providers.
func validateStandardProvider(config ProviderConfig) error {
	if len(config.Domain) == 0 {
		return ErrNoTargetDomain
	}

	if err := validateGroupAndUser(config); err != nil {
		return err
	}

	if err := validateRepositoryLists(config); err != nil {
		return err
	}

	if len(config.Token) == 0 {
		return ErrNoTargetToken
	}

	return nil
}

// validateGroupAndUser checks the validity of group and user settings.
func validateGroupAndUser(config ProviderConfig) error {
	if len(config.Group) == 0 && len(config.User) == 0 {
		return ErrNoSourceGroupOrUser
	}

	if len(config.Group) > 0 && len(config.User) > 0 {
		return ErrBothGroupAndUser
	}

	return nil
}

// validateRepositoryLists checks the validity of include and exclude repository lists.
func validateRepositoryLists(config ProviderConfig) error {
	if len(config.Exclude) > 0 && len(config.ExcludedRepositories()) < 1 {
		return ErrExcludeIsConfiguredButEmpty
	}

	if len(config.Include) > 0 && len(config.IncludedRepositories()) < 1 {
		return ErrIncludeIsConfiguredButEmpty
	}

	return nil
}

// validateProviderSpecific checks provider-specific configuration.
func validateProviderSpecific(name string, providerspecific map[string]string) error {
	switch name {
	case ARCHIVE:
		return ValidateArchiveProviderSpecific(providerspecific)
	case DIRECTORY:
		return ValidateDirectoryProviderSpecific(providerspecific)
	default:
		return nil
	}
}

// ValidateArchiveProviderSpecific checks the configuration specific to archive providers.
func ValidateArchiveProviderSpecific(configuration map[string]string) error {
	if len(configuration["archivetargetdir"]) == 0 {
		return ErrArchiveMissingTargetPath
	}

	return nil
}

// ValidateDirectoryProviderSpecific checks the configuration specific to directory providers.
func ValidateDirectoryProviderSpecific(configuration map[string]string) error {
	if len(configuration["directorytargetdir"]) == 0 {
		return ErrDirectoryMissingTargetPath
	}

	return nil
}
