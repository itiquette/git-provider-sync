// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"errors"
	"fmt"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/target"
	"net"
	"net/url"
	"os"
	"slices"
	"strings"

	"golang.org/x/crypto/ssh/agent"
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
	ErrNoGitBinaryFound             = errors.New("failed to find git binary")
)

var ValidGitProviders = []string{config.GITHUB, config.GITLAB, config.ARCHIVE, config.GITEA, config.DIRECTORY}

var ValidProtocolTypes = []string{"", config.HTTPS, config.SSHAGENT}

var ValidSchemeTypes = []string{"", config.HTTPS, config.HTTP}

// validateConfiguration checks the entire ProvidersConfig for validity.
func validateConfiguration(providersConfig config.ProvidersConfig) error {
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
func validateSourceProvider(provider config.ProviderConfig) error {
	if !slices.Contains(ValidGitProviders, provider.ProviderType) {
		return fmt.Errorf("source provider: must be one of %v: %w", ValidGitProviders, ErrUnsupportedProvider)
	}

	if strings.EqualFold(provider.ProviderType, config.ARCHIVE) {
		return ErrUnsupportedArchiveProvider
	}

	if strings.EqualFold(provider.ProviderType, config.DIRECTORY) {
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

	if err := validateSSHClient(provider); err != nil {
		return err
	}

	if err := validateRepositoryLists(provider); err != nil {
		return err
	}

	return validateAdditional(provider.ProviderType, provider.Additional)
}
func validateSSHClient(configuration config.ProviderConfig) error {
	if strings.EqualFold(configuration.Git.Type, config.SSHAGENT) {
		return checkSSHAgent()
	}
	// if config.HTTPClient.ProxyURL != "" {
	// 	_, err := url.Parse(config.HTTPClient.ProxyURL)
	// 	if err != nil {
	// 		return fmt.Errorf("gitinfo proxyurl is set but an invalid url: %w", err)
	// 	}
	// }

	// if config.HTTPClient.CertDirPath != "" {
	// 	if _, err := os.Stat(config.HTTPClient.CertDirPath); os.IsNotExist(err) {
	// 		return fmt.Errorf("CertDirPath is set but is not accessible: %w", err)
	// 	}
	// }

	return nil
}

func checkSSHAgent() error {
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock == "" {
		return errors.New("SSH_AUTH_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", sshAuthSock)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH agent: %w", err)
	}
	defer conn.Close()

	agentClient := agent.NewClient(conn)

	keys, err := agentClient.List()
	if err != nil {
		return fmt.Errorf("failed to list keys from SSH agent: %w", err)
	}

	if len(keys) == 0 {
		return errors.New("SSH agent is running but has no keys")
	}

	return nil
}

// validateTargetProvider checks the validity of a target provider configuration.
func validateTargetProvider(providerConfig config.ProviderConfig) error {
	if len(providerConfig.ProviderType) == 0 || !slices.Contains(ValidGitProviders, providerConfig.ProviderType) {
		return fmt.Errorf("target provider: must be one of %v: %w", ValidGitProviders, ErrUnsupportedProvider)
	}

	if !strings.EqualFold(providerConfig.ProviderType, config.ARCHIVE) && !strings.EqualFold(providerConfig.ProviderType, config.DIRECTORY) {
		if err := validateStandardProvider(providerConfig); err != nil {
			return err
		}
	}

	if providerConfig.Git.UseGitBinary {
		if _, err := target.ValidateGitBinary(); err != nil {
			return ErrNoGitBinaryFound
		}
	}

	if len(providerConfig.SSHClient.ProxyCommand) > 0 && !providerConfig.Git.UseGitBinary {
		return errors.New("target Provider: using proxy command requires Git.UseGitBinary true due to restrictions in underlying go-git library")
	}

	return validateAdditional(providerConfig.ProviderType, providerConfig.Additional)
}

// validateStandardProvider checks the validity of standard (non-archive, non-directory) providers.
func validateStandardProvider(config config.ProviderConfig) error {
	if len(config.Domain) == 0 {
		return ErrNoTargetDomain
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

	if err := validateSSHClient(config); err != nil {
		return err
	}

	if err := validateHTTPInfo(config); err != nil {
		return err
	}

	return nil
}

// validateGroupAndUser checks the validity of group and user settings.
func validateGroupAndUser(config config.ProviderConfig) error {
	if len(config.Group) == 0 && len(config.User) == 0 {
		return ErrNoSourceGroupOrUser
	}

	if len(config.Group) > 0 && len(config.User) > 0 {
		return ErrBothGroupAndUser
	}

	return nil
}

// validateGitOption checks the validity of the protocol setting.
func validateGitOption(providerConfig config.ProviderConfig) error {
	if !slices.Contains(ValidProtocolTypes, providerConfig.Git.Type) {
		return fmt.Errorf("gitinfo type: must be one of %v: %w", ValidProtocolTypes, ErrUnsupportedProtocolType)
	}

	return nil
}

// validateHTTP checks the validity of the http setting.
func validateHTTPInfo(config config.ProviderConfig) error {
	if !slices.Contains(ValidSchemeTypes, config.HTTPClient.Scheme) {
		return fmt.Errorf("source provider: must be one of %v: %w", ValidSchemeTypes, ErrUnsupportedScheme)
	}

	if len(config.HTTPClient.Token) == 0 {
		return ErrNoTargetToken
	}

	if config.HTTPClient.ProxyURL != "" {
		_, err := url.Parse(config.HTTPClient.ProxyURL)
		if err != nil {
			return fmt.Errorf("gitinfo proxyurl is set but an invalid url: %w", err)
		}
	}

	if config.HTTPClient.CertDirPath != "" {
		if _, err := os.Stat(config.HTTPClient.CertDirPath); os.IsNotExist(err) {
			return fmt.Errorf("CertDirPath is set but is not accessible: %w", err)
		}
	}

	return nil
}

// validateRepositoryLists checks the validity of include and exclude repository lists.
func validateRepositoryLists(config config.ProviderConfig) error {
	if len(config.Repositories.Exclude) > 0 && len(config.Repositories.ExcludedRepositories()) < 1 {
		return ErrExcludeIsConfiguredButEmpty
	}

	if len(config.Repositories.Include) > 0 && len(config.Repositories.IncludedRepositories()) < 1 {
		return ErrIncludeIsConfiguredButEmpty
	}

	return nil
}

// validateAdditional checks provider-specific configuration.
func validateAdditional(name string, additional map[string]string) error {
	switch name {
	case config.ARCHIVE:
		return ValidateArchiveAdditional(additional)
	case config.DIRECTORY:
		return ValidateDirectoryAdditional(additional)
	default:
		return nil
	}
}

// ValidateArchiveAdditional checks the configuration specific to archive providers.
func ValidateArchiveAdditional(configuration map[string]string) error {
	if len(configuration["archivetargetdir"]) == 0 {
		return ErrArchiveMissingTargetPath
	}

	return nil
}

// ValidateDirectoryAdditional checks the configuration specific to directory providers.
func ValidateDirectoryAdditional(configuration map[string]string) error {
	if len(configuration["directorytargetdir"]) == 0 {
		return ErrDirectoryMissingTargetPath
	}

	return nil
}
