// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/target"

	"golang.org/x/crypto/ssh/agent"
)

// Constants for configuration values.
const (
	// Environment variables.
	sshAuthSockEnv = "SSH_AUTH_SOCK"

	// Provider types.
	providerGitHub    = config.GITHUB
	providerGitLab    = config.GITLAB
	providerGitea     = config.GITEA
	providerArchive   = config.ARCHIVE
	providerDirectory = config.DIRECTORY

	// Protocol types.
	protocolHTTPS = config.HTTPS
	protocolHTTP  = config.HTTP
	protocolSSH   = config.SSHAGENT

	// Validation constants.
	maxDescriptionLength = 1000
	maxRepoNameLength    = 255
	minRepoNameLength    = 1
)

// Define error variables for various configuration validation scenarios.
var (
	// Provider Type Errors.
	ErrUnsupportedProvider          = errors.New("unsupported provider")
	ErrUnsupportedArchiveProvider   = errors.New("source provider: does not support reading from archive")
	ErrUnsupportedDirectoryProvider = errors.New("source provider: does not support reading from directory")
	ErrUnsupportedProviderURL       = errors.New("unsupported Git provider URL")

	// Configuration Errors.
	ErrNoSourceDomain    = errors.New("source provider: no domain configured")
	ErrNoTargetDomain    = errors.New("target provider: no domain configured")
	ErrNoTargetProviders = errors.New("no target provider/s configured")
	ErrNoHTTPToken       = errors.New("no httpclient token set")
	ErrInvalidDuration   = errors.New("invalid duration format")

	// Authentication Errors.
	ErrTokenAuth        = errors.New("target provider currently only supports token auth")
	ErrNoGitBinaryFound = errors.New("failed to find git binary")
	ErrInvalidToken     = errors.New("invalid token format")

	// Protocol Errors.
	ErrUnsupportedScheme       = errors.New("unsupported scheme")
	ErrUnsupportedProtocolType = errors.New("unsupported protocol type")
	ErrHasNoHTTPPrefix         = errors.New("target provider currently only supports http/s")
	ErrTargetURLValidFormat    = errors.New("target url must be a Git provider URL")

	// User/Group Errors.
	ErrNoSourceGroupOrUser = errors.New("source provider: no group path or user configured")
	ErrNoTargetGroupOrUser = errors.New("target provider: no group path or user configured")
	ErrBothGroupAndUser    = errors.New("provider: group path and user configured, only one is allowed")
	ErrInvalidGroupName    = errors.New("invalid group name")
	ErrInvalidUserName     = errors.New("invalid username")

	// Repository Errors.
	ErrExcludeIsConfiguredButEmpty = errors.New("exclude is configured but 'repositories:' contains no repository names")
	ErrIncludeIsConfiguredButEmpty = errors.New("include is configured but 'repositories:' contains no repository names")
	ErrInvalidRepoName             = errors.New("invalid repository name")
	ErrInvalidDescription          = errors.New("invalid repository description")

	// Path Errors.
	ErrArchiveMissingTargetPath   = errors.New("archive target provider: missing property archivetargetdir")
	ErrDirectoryMissingTargetPath = errors.New("directory target provider: missing property directorytargetdir")
	ErrInvalidPath                = errors.New("invalid file path")
)

// Valid provider and protocol configurations.
var (
	ValidGitProviders  = []string{providerGitHub, providerGitLab, providerGitea, providerArchive, providerDirectory}
	ValidProtocolTypes = []string{"", protocolHTTPS, protocolSSH}
	ValidSchemeTypes   = []string{"", protocolHTTPS, protocolHTTP}
)

// validateConfiguration performs validation of the entire ProvidersConfig.
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

	return validateCrossConfiguration(providersConfig)
}

// validateSourceProvider validates the source provider configuration.
func validateSourceProvider(provider config.ProviderConfig) error {
	if !isValidProviderType(provider.ProviderType) {
		return fmt.Errorf("source provider: must be one of %v: %w", ValidGitProviders, ErrUnsupportedProvider)
	}

	if strings.EqualFold(provider.ProviderType, config.ARCHIVE) {
		return ErrUnsupportedArchiveProvider
	}

	if strings.EqualFold(provider.ProviderType, config.DIRECTORY) {
		return ErrUnsupportedDirectoryProvider
	}

	if err := validateDomainName(provider.Domain); err != nil {
		return fmt.Errorf("%w %w", ErrNoSourceDomain, err)
	}

	if err := validateGroupAndUser(provider); err != nil {
		return err
	}

	if err := validateHTTPInfo(provider); err != nil {
		return err
	}

	if err := validateSSHClient(provider); err != nil {
		return err
	}

	if len(provider.SSHClient.SSHCommand) > 0 && !provider.Git.UseGitBinary {
		return errors.New("source Provider: using proxy command requires Git.UseGitBinary true due to restrictions in underlying go-git library")
	}

	if err := validateGitOption(provider); err != nil {
		return err
	}

	if err := validateRepositoryLists(provider); err != nil {
		return err
	}

	// Original sync run checks - using the actual fields from your config
	if provider.SyncRun.ActiveFromLimit != "" {
		if _, err := time.ParseDuration(provider.SyncRun.ActiveFromLimit); err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
	}

	return nil
}

// validateTargetProvider with original sync run validation preserved.
func validateTargetProvider(providerConfig config.ProviderConfig) error {
	if !isValidProviderType(providerConfig.ProviderType) {
		return fmt.Errorf("target provider: must be one of %v: %w", ValidGitProviders, ErrUnsupportedProvider)
	}

	if providerConfig.ProviderType != config.ARCHIVE && providerConfig.ProviderType != config.DIRECTORY {
		if len(providerConfig.Domain) == 0 {
			return ErrNoSourceDomain
		}

		if err := validateGroupAndUser(providerConfig); err != nil {
			return err
		}

		if err := validateGitOption(providerConfig); err != nil {
			return err
		}

		// Original git binary check
		if providerConfig.Git.UseGitBinary {
			if _, err := target.ValidateGitBinary(); err != nil {
				return ErrNoGitBinaryFound
			}
		}

		if err := validateHTTPInfo(providerConfig); err != nil {
			return err
		}

		if err := validateSSHClient(providerConfig); err != nil {
			return err
		}

		if len(providerConfig.SSHClient.SSHCommand) > 0 && !providerConfig.Git.UseGitBinary {
			return errors.New("target Provider: using proxy command requires Git.UseGitBinary true due to restrictions in underlying go-git library")
		}

		if err := validateSyncRunConfig(providerConfig.SyncRun); err != nil {
			return fmt.Errorf("invalid syncrun: %w", err)
		}
	}

	if err := validateAdditional(providerConfig.ProviderType, providerConfig.Additional); err != nil {
		return fmt.Errorf("invalid additional: %w", err)
	}

	return nil
}

// validateGroupAndUser validates group and user settings.
func validateGroupAndUser(config config.ProviderConfig) error {
	if len(config.Group) == 0 && len(config.User) == 0 {
		return ErrNoSourceGroupOrUser
	}

	if len(config.Group) > 0 && len(config.User) > 0 {
		return ErrBothGroupAndUser
	}

	// Additional validation for group and user names
	if config.Group != "" {
		if err := validateGroupName(config.Group); err != nil {
			return err
		}
	}

	if config.User != "" {
		if err := validateUsername(config.User); err != nil {
			return err
		}
	}

	return nil
}

// validateGitOption validates the Git protocol configuration.
func validateGitOption(providerConfig config.ProviderConfig) error {
	if !isValidProtocolType(providerConfig.Git.Type) {
		return fmt.Errorf("gitinfo type: must be one of %v: %w", ValidProtocolTypes, ErrUnsupportedProtocolType)
	}

	//TODO add more checks:
	return nil
}

// validateHTTPInfo validates HTTP client configuration.
func validateHTTPInfo(config config.ProviderConfig) error {
	if !isValidSchemeType(config.HTTPClient.Scheme) {
		return fmt.Errorf("source provider: must be one of %v: %w", ValidSchemeTypes, ErrUnsupportedScheme)
	}

	if config.HTTPClient.ProxyURL != "" {
		if err := validateURL(config.HTTPClient.ProxyURL); err != nil {
			return fmt.Errorf("gitinfo proxyurl is set but an invalid url: %w", err)
		}
	}

	if config.HTTPClient.CertDirPath != "" {
		if err := validatePathExists(config.HTTPClient.CertDirPath); err != nil {
			return fmt.Errorf("CertDirPath is set but is not accessible: %w", err)
		}
	}

	return nil
}

// validateSSHClient validates SSH client configuration.
func validateSSHClient(configuration config.ProviderConfig) error {
	if strings.EqualFold(configuration.Git.Type, config.SSHAGENT) {
		return checkSSHAgent()
	}

	if configuration.SSHClient.SSHCommand != "" {
		if err := validateSSHCommand(configuration.SSHClient.SSHCommand); err != nil {
			return err
		}
	}

	return nil
}

// checkSSHAgent validates SSH agent configuration and connectivity.
func checkSSHAgent() error {
	sshAuthSock := os.Getenv(sshAuthSockEnv)
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

// validateRepositoryLists validates repository include/exclude lists.
func validateRepositoryLists(config config.ProviderConfig) error {
	if len(config.Repositories.Exclude) > 0 && len(config.Repositories.ExcludedRepositories()) < 1 {
		return ErrExcludeIsConfiguredButEmpty
	}

	if len(config.Repositories.Include) > 0 && len(config.Repositories.IncludedRepositories()) < 1 {
		return ErrIncludeIsConfiguredButEmpty
	}

	// Handle nullable description field
	if config.Repositories.Description != nil {
		if err := validateRepoDescription(*config.Repositories.Description); err != nil {
			return err
		}
	}

	return nil
}

func validateRepoDescription(description string) error {
	if len(description) > maxDescriptionLength {
		return fmt.Errorf("%w: description exceeds maximum length of %d characters",
			ErrInvalidDescription, maxDescriptionLength)
	}

	return nil
}

// validateAdditional validates provider-specific configuration.
func validateAdditional(providerType string, additional map[string]string) error {
	switch providerType {
	case config.ARCHIVE:
		return validateArchiveAdditional(additional)
	case config.DIRECTORY:
		return validateDirectoryAdditional(additional)
	default:
		return nil
	}
}

// ValidateArchiveAdditional validates archive-specific configuration.
func validateArchiveAdditional(configuration map[string]string) error {
	path, exists := configuration["archivetargetdir"]
	if !exists || path == "" {
		return ErrArchiveMissingTargetPath
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("%w: path must be absolute: %s", ErrInvalidPath, path)
	}

	return nil
}

// ValidateDirectoryAdditional validates directory-specific configuration.
func validateDirectoryAdditional(configuration map[string]string) error {
	path, exists := configuration["directorytargetdir"]
	if !exists || path == "" {
		return ErrDirectoryMissingTargetPath
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("%w: path must be absolute: %s", ErrInvalidPath, path)
	}

	return nil
}

// Helper functions for additional validation

func validateUsername(username string) error {
	if len(username) < 1 || len(username) > 39 {
		return fmt.Errorf("%w: length must be between 1 and 39 characters", ErrInvalidUserName)
	}
	// Add additional username validation rules if needed
	return nil
}

func validateGroupName(groupName string) error {
	if len(groupName) < 1 || len(groupName) > 255 {
		return fmt.Errorf("%w: length must be between 1 and 255 characters", ErrInvalidGroupName)
	}
	// Add additional group name validation rules if needed
	return nil
}

// Cross-configuration validation.
func validateCrossConfiguration(config config.ProvidersConfig) error {
	// Validate unique target combinations
	seen := make(map[string]bool)

	for _, target := range config.ProviderTargets {
		key := fmt.Sprintf("%s-%s-%s-%s",
			target.ProviderType, target.Domain, target.Group, target.User)
		if seen[key] {
			return fmt.Errorf("duplicate target configuration found for: %s", key)
		}

		seen[key] = true
	}

	// Add more cross-configuration validations as needed
	return nil
}

// SyncRun configuration validation.
func validateSyncRunConfig(syncRun config.SyncRunOption) error {
	if syncRun.ActiveFromLimit != "" {
		if _, err := time.ParseDuration(syncRun.ActiveFromLimit); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidDuration, syncRun.ActiveFromLimit)
		}
	}

	// Additional sync run validations
	if syncRun.IgnoreInvalidName && !syncRun.CleanupInvalidName {
		return errors.New("when IgnoreInvalidName is true, CleanupInvalidName must also be true")
	}

	return nil
}

// validatePathExists verifies that a file path exists and is accessible.
func validatePathExists(path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("%w: path must be absolute: %s", ErrInvalidPath, path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%w: path does not exist: %s", ErrInvalidPath, path)
	}

	return nil
}

// validateURL checks if a URL string is valid.
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("%w: empty URL", ErrUnsupportedProviderURL)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnsupportedProviderURL, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https", ErrUnsupportedProviderURL)
	}

	return nil
}

// validateSSHCommand validates the SSH command configuration.
func validateSSHCommand(command string) error {
	if command == "" {
		return nil // Empty command is valid
	}

	// Basic validation of SSH command format
	if !strings.HasPrefix(command, "ssh ") {
		return errors.New("SSH command must start with 'ssh'")
	}

	return nil
}

// validateDomainName checks if a domain name is valid.
func validateDomainName(domain string) error {
	if domain == "" {
		return errors.New("domain name cannot be empty")
	}

	// Add more domain validation rules if needed
	if strings.Contains(domain, "://") {
		return errors.New("domain should not include protocol scheme")
	}

	return nil
}

// Additional utility functions

// isValidProviderType checks if a provider type is supported.
func isValidProviderType(providerType string) bool {
	return slices.Contains(ValidGitProviders, providerType)
}

// isValidProtocolType checks if a protocol type is supported.
func isValidProtocolType(protocolType string) bool {
	return slices.Contains(ValidProtocolTypes, protocolType)
}

// isValidSchemeType checks if a scheme type is supported.
func isValidSchemeType(schemeType string) bool {
	return slices.Contains(ValidSchemeTypes, schemeType)
}
