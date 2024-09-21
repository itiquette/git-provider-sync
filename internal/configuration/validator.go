// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"errors"
	"fmt"
	"itiquette/git-provider-sync/internal/model"
	"net/url"
	"os"
	"slices"
	"strings"
)

// Define error variables for various configuration validation scenarios.
var (
	ErrUnsupportedScheme            = errors.New("unsupported scheme")
	ErrUnsupportedProvider          = errors.New("unsupported provider")
	ErrUnsupportedProtocolType      = errors.New("unsupported protocol type")
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
	ErrInvalidSSHKeyPath            = errors.New("ssh key path invalid")
)

var ValidGitProviders = []string{GITHUB, GITLAB, ARCHIVE, GITEA, DIRECTORY}

var ValidProtocolTypes = []string{"", model.HTTPS, model.SSHAGENT, model.SSHKEY}

var ValidSchemeTypes = []string{"", model.HTTPS, model.HTTP}

// validateConfiguration checks the entire ProvidersConfig for validity.
func validateConfiguration(providersConfig ProvidersConfig) error {
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
	if !slices.Contains(ValidGitProviders, provider.ProviderType) {
		return fmt.Errorf("source provider: must be one of %v: %w", ValidGitProviders, ErrUnsupportedProvider)
	}

	if strings.EqualFold(provider.ProviderType, ARCHIVE) {
		return ErrUnsupportedArchiveProvider
	}

	if strings.EqualFold(provider.ProviderType, DIRECTORY) {
		return ErrUnsupportedDirectoryProvider
	}

	if len(provider.Domain) == 0 {
		return ErrNoSourceDomain
	}

	if err := validateGroupAndUser(provider); err != nil {
		return err
	}

	if err := validateGitOption(provider); err != nil {
		return err
	}

	if err := validateRepositoryLists(provider); err != nil {
		return err
	}

	return validateadditional(provider.ProviderType, provider.Additional)
}

// validateTargetProvider checks the validity of a target provider configuration.
func validateTargetProvider(config ProviderConfig) error {
	if len(config.ProviderType) == 0 || !slices.Contains(ValidGitProviders, config.ProviderType) {
		return fmt.Errorf("target provider: must be one of %v: %w", ValidGitProviders, ErrUnsupportedProvider)
	}

	if !strings.EqualFold(config.ProviderType, ARCHIVE) && !strings.EqualFold(config.ProviderType, DIRECTORY) {
		if err := validateStandardProvider(config); err != nil {
			return err
		}
	}

	return validateadditional(config.ProviderType, config.Additional)
}

// validateStandardProvider checks the validity of standard (non-archive, non-directory) providers.
func validateStandardProvider(config ProviderConfig) error {
	if len(config.Domain) == 0 {
		return ErrNoTargetDomain
	}

	if !slices.Contains(ValidSchemeTypes, config.Scheme) {
		return fmt.Errorf("source provider: must be one of %v: %w", ValidSchemeTypes, ErrUnsupportedScheme)
	}

	if err := validateGroupAndUser(config); err != nil {
		return err
	}

	if err := validateRepositoryLists(config); err != nil {
		return err
	}

	if err := validateGitOption(config); err != nil {
		return err
	}

	if err := validateHTTPInfo(config); err != nil {
		return err
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

// validateGitOption checks the validity of the protocol setting.
func validateGitOption(config ProviderConfig) error {
	if !slices.Contains(ValidProtocolTypes, config.Git.Type) {
		return fmt.Errorf("gitinfo type: must be one of %v: %w", ValidProtocolTypes, ErrUnsupportedProtocolType)
	}

	if strings.EqualFold(config.Git.Type, model.SSHKEY) {
		if len(config.Git.SSHPrivateKeyPath) == 0 {
			return fmt.Errorf("gitinfo type was sshkey, but sshprivatekeypath was empty. err: %w", ErrInvalidSSHKeyPath)
		}

		_, err := os.Stat(config.Git.SSHPrivateKeyPath)
		if err != nil {
			return fmt.Errorf("gitinfo type was sshkey, but keyfile: %s could not be read. err: %w", config.Git.SSHPrivateKeyPath, ErrInvalidSSHKeyPath)
		}
	}

	return nil
}

// validateHTTP checks the validity of the http setting.
func validateHTTPInfo(config ProviderConfig) error {
	if len(config.HTTPClient.Token) == 0 {
		return ErrNoTargetToken
	}

	if config.HTTPClient.ProxyURL != "" {
		_, err := url.Parse(config.HTTPClient.ProxyURL)
		if err != nil {
			return fmt.Errorf("gitinfo proxyurl is set but an invalid url: %w", err)
		}
	}

	return nil
}

// validateRepositoryLists checks the validity of include and exclude repository lists.
func validateRepositoryLists(config ProviderConfig) error {
	if len(config.Repositories.Exclude) > 0 && len(config.Repositories.ExcludedRepositories()) < 1 {
		return ErrExcludeIsConfiguredButEmpty
	}

	if len(config.Repositories.Include) > 0 && len(config.Repositories.IncludedRepositories()) < 1 {
		return ErrIncludeIsConfiguredButEmpty
	}

	return nil
}

// validateadditional checks provider-specific configuration.
func validateadditional(name string, additional map[string]string) error {
	switch name {
	case ARCHIVE:
		return ValidateArchiveadditional(additional)
	case DIRECTORY:
		return ValidateDirectoryadditional(additional)
	default:
		return nil
	}
}

// ValidateArchiveadditional checks the configuration specific to archive providers.
func ValidateArchiveadditional(configuration map[string]string) error {
	if len(configuration["archivetargetdir"]) == 0 {
		return ErrArchiveMissingTargetPath
	}

	return nil
}

// ValidateDirectoryadditional checks the configuration specific to directory providers.
func ValidateDirectoryadditional(configuration map[string]string) error {
	if len(configuration["directorytargetdir"]) == 0 {
		return ErrDirectoryMissingTargetPath
	}

	return nil
}
