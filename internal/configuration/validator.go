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

const (
	// Environment variables.
	sshAuthSockEnv = "SSH_AUTH_SOCK"

	// Validation constants.
	maxDescriptionLength = 1000
	maxRepoNameLength    = 255
	minRepoNameLength    = 1
)

var (
	// Provider Type Errors.
	ErrUnsupportedProvider          = errors.New("unsupported provider")
	ErrUnsupportedArchiveProvider   = errors.New("source provider: does not support reading from archive")
	ErrUnsupportedDirectoryProvider = errors.New("source provider: does not support reading from directory")
	ErrInvalidURL                   = errors.New("invalid URL")

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

var (
	ValidSourceGitProviders = []string{config.GITHUB, config.GITLAB, config.GITEA}
	ValidTargetGitProviders = []string{config.GITHUB, config.GITLAB, config.GITEA, config.ARCHIVE, config.DIRECTORY}
	ValidProtocolTypes      = []string{"", config.HTTPS, config.SSHAGENT}
	ValidSchemeTypes        = []string{"", config.HTTPS, config.HTTP}
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

	return nil
}

// validateSourceProvider validates the source provider configuration.
func validateSourceProvider(provider config.ProviderConfig) error {
	if !isValidSourceProviderType(provider.ProviderType) {
		return fmt.Errorf("source provider: must be one of %v: %w", ValidSourceGitProviders, ErrUnsupportedProvider)
	}

	if err := validateDomainName(provider.Domain); err != nil {
		return fmt.Errorf("%w %w", ErrNoSourceDomain, err)
	}

	if err := validateGroupAndUser(provider); err != nil {
		return err
	}

	if err := validateHTTPClient(provider); err != nil {
		return err
	}

	if err := validateSSHClient(provider); err != nil {
		return err
	}

	if provider.Git.UseGitBinary {
		if _, err := target.ValidateGitBinary(); err != nil {
			return ErrNoGitBinaryFound
		}
	}

	if !isValidProtocolType(provider.Git.Type) {
		return fmt.Errorf("gitinfo type: must be one of %v: %w", ValidProtocolTypes, ErrUnsupportedProtocolType)
	}

	if err := validateRepositoryLists(provider); err != nil {
		return err
	}

	if provider.SyncRun.ActiveFromLimit != "" {
		if _, err := time.ParseDuration(provider.SyncRun.ActiveFromLimit); err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
	}

	if provider.Project.Description != "" {
		return errors.New("source provider does not support project.description, only target does")
	}

	if provider.SyncRun.CleanupInvalidName || provider.SyncRun.ForcePush || provider.SyncRun.IgnoreInvalidName {
		return errors.New("source provider does not support syncrun.cleanupinvalidname, forcepush, ignoreninvalid")
	}

	if provider.Additional != nil {
		return errors.New("additional is not valid for a source provider")
	}

	return nil
}

func validateTargetProvider(providerConfig config.ProviderConfig) error {
	if !isValidTargetProviderType(providerConfig.ProviderType) {
		return fmt.Errorf("target provider: must be one of %v: %w", ValidTargetGitProviders, ErrUnsupportedProvider)
	}

	if providerConfig.ProviderType != config.ARCHIVE && providerConfig.ProviderType != config.DIRECTORY {
		if len(providerConfig.Domain) == 0 {
			return ErrNoTargetDomain
		}

		if err := validateGroupAndUser(providerConfig); err != nil {
			return err
		}

		if providerConfig.Git.UseGitBinary {
			if _, err := target.ValidateGitBinary(); err != nil {
				return ErrNoGitBinaryFound
			}
		}

		if providerConfig.Git.IncludeForks {
			return errors.New("target provider: git.invalid forks is not valid here")
		}

		if !isValidProtocolType(providerConfig.Git.Type) {
			return fmt.Errorf("gitinfo type: must be one of %v: %w", ValidProtocolTypes, ErrUnsupportedProtocolType)
		}

		if err := validateHTTPClient(providerConfig); err != nil {
			return err
		}

		if err := validateSSHClient(providerConfig); err != nil {
			return err
		}

		if len(providerConfig.SSHClient.SSHCommand) > 0 && !providerConfig.Git.UseGitBinary {
			return errors.New("target Provider: using proxy command requires Git.UseGitBinary true due to restrictions in underlying go-git library")
		}

		if len(providerConfig.Repositories.Include) != 0 || len(providerConfig.Repositories.Exclude) != 0 {
			return errors.New("target provider: repositories is only valid for source provider configurations")
		}

		if providerConfig.SyncRun.ActiveFromLimit != "" {
			return errors.New("target provider: syncrun active from limit only makes sense from source provider conf")
		}

		if providerConfig.Project.Description != "" {
			if err := validateRepoDescription(providerConfig.Project.Description); err != nil {
				return err
			}
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

func validateHTTPClient(config config.ProviderConfig) error {
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

func validateSSHClient(configuration config.ProviderConfig) error {
	if len(configuration.SSHClient.SSHCommand) > 0 && !configuration.Git.UseGitBinary {
		return errors.New("using SSH-command requires Git.UseGitBinary true due to restrictions in underlying go-git library")
	}

	if strings.EqualFold(configuration.Git.Type, config.SSHAGENT) {
		return checkSSHAgent()
	}

	if err := validateSSHCommand(configuration.SSHClient.SSHCommand); err != nil {
		return err
	}

	if configuration.SSHClient.RewriteSSHURLFrom != "" || configuration.SSHClient.RewriteSSHURLTo != "" {
		if configuration.SSHClient.RewriteSSHURLFrom == "" || configuration.SSHClient.RewriteSSHURLTo == "" {
			return errors.New("if either rewritesshurlfrom or rewritesshurlto is specified, both must be provided")
		}
	}

	return nil
}

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

func validateRepositoryLists(config config.ProviderConfig) error {
	if len(config.Repositories.Exclude) > 0 && len(config.Repositories.ExcludedRepositories()) < 1 {
		return ErrExcludeIsConfiguredButEmpty
	}

	if len(config.Repositories.Include) > 0 && len(config.Repositories.IncludedRepositories()) < 1 {
		return ErrIncludeIsConfiguredButEmpty
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

func validateUsername(username string) error {
	if len(username) < 1 || len(username) > 39 {
		return fmt.Errorf("%w: length must be between 1 and 39 characters", ErrInvalidUserName)
	}

	return nil
}

func validateGroupName(groupName string) error {
	if len(groupName) < 1 || len(groupName) > 255 {
		return fmt.Errorf("%w: length must be between 1 and 255 characters", ErrInvalidGroupName)
	}

	return nil
}

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

func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("%w: empty URL", ErrInvalidURL)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https", ErrInvalidURL)
	}

	return nil
}

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

func validateDomainName(domain string) error {
	if domain == "" {
		return errors.New("domain name cannot be empty")
	}

	if strings.Contains(domain, "://") {
		return errors.New("domain should not include protocol scheme")
	}

	return nil
}

func isValidSourceProviderType(providerType string) bool {
	return slices.Contains(ValidSourceGitProviders, providerType)
}

func isValidTargetProviderType(providerType string) bool {
	return slices.Contains(ValidTargetGitProviders, providerType)
}

func isValidProtocolType(protocolType string) bool {
	return slices.Contains(ValidProtocolTypes, protocolType)
}

func isValidSchemeType(schemeType string) bool {
	return slices.Contains(ValidSchemeTypes, schemeType)
}
