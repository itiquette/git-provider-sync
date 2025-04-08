// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"errors"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/mirror/gitbinary"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/agent"
)

const (
	sshAuthSockEnv = "SSH_AUTH_SOCK"

	maxDescriptionLength = 1000
	maxRepoNameLength    = 255
	minRepoNameLength    = 1
)

var (
	// Provider Type Errors remain the same.
	ErrUnsupportedProvider          = errors.New("unsupported provider")
	ErrUnsupportedArchiveProvider   = errors.New("source provider: does not support reading from archive")
	ErrUnsupportedDirectoryProvider = errors.New("source provider: does not support reading from directory")
	ErrInvalidURL                   = errors.New("invalid URL")

	// Configuration Errors.
	ErrNoSourceDomain  = errors.New("source provider: no domain configured")
	ErrNoTargetDomain  = errors.New("target provider: no domain configured")
	ErrNoMirrors       = errors.New("no mirror configurations provided")
	ErrNoHTTPToken     = errors.New("no http token set")
	ErrInvalidDuration = errors.New("invalid duration format")

	// Authentication Errors.
	ErrTokenAuth        = errors.New("target provider currently only supports token auth")
	ErrNoGitBinaryFound = errors.New("failed to find git binary")
	ErrInvalidToken     = errors.New("invalid token format")

	// Protocol Errors remain the same.
	ErrUnsupportedScheme       = errors.New("unsupported scheme")
	ErrUnsupportedProtocolType = errors.New("unsupported protocol type")
	ErrHasNoHTTPPrefix         = errors.New("target provider currently only supports http/s")

	// Owner Errors updated for new structure.
	ErrNoSourceOwner    = errors.New("source provider: no owner configured")
	ErrNoTargetOwner    = errors.New("target provider: no owner configured")
	ErrInvalidOwner     = errors.New("invalid owner name")
	ErrInvalidOwnerType = errors.New("invalid owner type")

	// Repository Errors.
	ErrInvalidRepoName    = errors.New("invalid repository name")
	ErrInvalidDescription = errors.New("invalid repository description")

	// Path Errors.
	ErrInvalidPath = errors.New("invalid file path")
)

var (
	ValidSourceGitProviders = []string{"github", "gitlab", "gitea"}
	ValidMirrorTargets      = []string{"github", "gitlab", "gitea", "archive", "directory"}
	ValidProtocolTypes      = []string{"", config.TLS, config.SSH}
	ValidSchemeTypes        = []string{"", config.HTTPS, config.HTTP}
	ValidOwnerTypes         = []string{"", config.USER, config.GROUP}
)

// ValidateConfiguration validates the entire application configuration.
func validateConfiguration(ctx context.Context, appCfg *config.AppConfiguration) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering validateConfiguration")

	if len(appCfg.GitProviderSyncConfs) == 0 {
		return errors.New("no git provider sync configurations found")
	}

	nrOfEnvironment := len(appCfg.GitProviderSyncConfs)
	currentEnvironmentCfg := 1

	for envName, env := range appCfg.GitProviderSyncConfs {
		logger.Debug().Msgf("Validating environment %v of %v", currentEnvironmentCfg, nrOfEnvironment)

		if err := validateEnvironment(envName, env); err != nil {
			return fmt.Errorf("invalid environment %s: %w", envName, err)
		}

		logger.Debug().Msgf("Validated environment %v of %v", currentEnvironmentCfg, nrOfEnvironment)

		currentEnvironmentCfg++
	}

	return nil
}

// validateEnvironment validates a single environment configuration.
func validateEnvironment(envName string, env config.Environment) error {
	if len(env) == 0 {
		return fmt.Errorf("environment %s has no sync configurations", envName)
	}

	for sourceName, syncConfig := range env {
		if err := validateSyncConfig(sourceName, syncConfig); err != nil {
			return fmt.Errorf("invalid sync config %s: %w", sourceName, err)
		}
	}

	return nil
}

// validateSyncConfig validates a single sync configuration.
func validateSyncConfig(_ string, syncCfg config.SyncConfig) error {
	if err := validateProviderType(syncCfg.ProviderType, ValidSourceGitProviders); err != nil {
		return err
	}

	if err := validateDomainName(syncCfg.GetDomain()); err != nil {
		return fmt.Errorf("%w: %w", ErrNoSourceDomain, err)
	}

	if err := validateOwner(syncCfg.Owner, syncCfg.OwnerType); err != nil {
		return err
	}

	if err := validateAuth(syncCfg.Auth); err != nil {
		return err
	}

	if syncCfg.UseGitBinary {
		// Note: Assuming gitbinary.ValidateGitBinary() is available
		if _, err := gitbinary.ValidateGitBinary(); err != nil {
			return ErrNoGitBinaryFound
		}
	}

	if syncCfg.ActiveFromLimit != "" {
		if _, err := time.ParseDuration(syncCfg.ActiveFromLimit); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidDuration, err)
		}
	}

	// Validate mirrors if present
	if len(syncCfg.Mirrors) > 0 {
		for _, mirror := range syncCfg.Mirrors {
			if err := validateMirrorConfig(mirror); err != nil {
				return fmt.Errorf("invalid mirror config %v: %w", mirror, err)
			}
		}
	}

	return nil
}

// validateMirrorConfig validates a mirror configuration.
func validateMirrorConfig(mirrorCfg config.MirrorConfig) error {
	if err := validateProviderType(mirrorCfg.ProviderType, ValidMirrorTargets); err != nil {
		return err
	}

	if mirrorCfg.ProviderType != "archive" && mirrorCfg.ProviderType != "directory" {
		if err := validateDomainName(mirrorCfg.GetDomain()); err != nil {
			return fmt.Errorf("%w: %w", ErrNoTargetDomain, err)
		}

		if err := validateOwner(mirrorCfg.Owner, mirrorCfg.OwnerType); err != nil {
			return err
		}

		if mirrorCfg.UseGitBinary {
			// Note: Assuming gitbinary.ValidateGitBinary() is available
			if _, err := gitbinary.ValidateGitBinary(); err != nil {
				return ErrNoGitBinaryFound
			}
		}
	}

	if err := validateAuth(mirrorCfg.Auth); err != nil {
		return err
	}

	if err := validateMirrorSettings(mirrorCfg.Settings); err != nil {
		return err
	}

	return nil
}

// validateAuth validates authentication configuration.
func validateAuth(authCfg config.AuthConfig) error {
	if !isValidSchemeType(authCfg.HTTPScheme) {
		return fmt.Errorf("invalid HTTP scheme: %w", ErrUnsupportedScheme)
	}

	if authCfg.ProxyURL != "" {
		if err := validateURL(authCfg.ProxyURL); err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
	}

	if authCfg.CertDirPath != "" {
		if err := validatePathExists(authCfg.CertDirPath); err != nil {
			return fmt.Errorf("invalid cert directory path: %w", err)
		}
	}

	if authCfg.Protocol == config.SSH {
		if err := checkSSHAgent(); err != nil {
			return err
		}
	}

	if authCfg.SSHURLRewriteFrom != "" || authCfg.SSHURLRewriteTo != "" {
		if authCfg.SSHURLRewriteFrom == "" || authCfg.SSHURLRewriteTo == "" {
			return errors.New("if either SSH URL rewrite parameter is specified, both must be provided")
		}
	}

	return validateSSHCommand(authCfg.SSHCommand)
}

// validateMirrorSettings validates mirror-specific settings.
func validateMirrorSettings(settings config.MirrorSettings) error {
	if settings.DescriptionPrefix != "" {
		if err := validateRepoDescription(settings.DescriptionPrefix); err != nil {
			return err
		}
	}

	if settings.Visibility != "" && !isValidVisibility(settings.Visibility) {
		return errors.New("invalid visibility setting")
	}

	return nil
}

// Helper functions remain largely the same with updated parameter types.
func validateOwner(owner, ownerType string) error {
	if owner == "" {
		return ErrNoSourceOwner
	}

	if !slices.Contains(ValidOwnerTypes, ownerType) {
		return ErrInvalidOwnerType
	}

	if ownerType == "user" {
		return validateUsername(owner)
	}

	return validateGroupName(owner)
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

// Existing helper functions remain the same.
func validateUsername(username string) error {
	if len(username) < 1 || len(username) > 39 {
		return fmt.Errorf("%w: length must be between 1 and 39 characters", ErrInvalidOwner)
	}

	return nil
}

func validateGroupName(groupName string) error {
	if len(groupName) < 1 || len(groupName) > 255 {
		return fmt.Errorf("%w: length must be between 1 and 255 characters", ErrInvalidOwner)
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

func validateSSHCommand(sshCommand string) error {
	if sshCommand == "" {
		return nil
	}

	if !strings.HasPrefix(sshCommand, "ssh ") {
		return errors.New("SSH command must start with 'ssh'")
	}

	return nil
}

func validateDomainName(domain string) error {
	if domain == "" {
		return errors.New("domain is required")
	}

	if strings.Contains(domain, "://") {
		return errors.New("domain should not include protocol scheme")
	}

	return nil
}

func validateProviderType(providerType string, validTypes []string) error {
	if !slices.Contains(validTypes, providerType) {
		return fmt.Errorf("%w: must be one of %v, got %s", ErrUnsupportedProvider, validTypes, providerType)
	}

	return nil
}

func isValidSchemeType(schemeType string) bool {
	return slices.Contains(ValidSchemeTypes, schemeType)
}

func isValidVisibility(visibility string) bool {
	return slices.Contains([]string{"public", "private", "internal"}, visibility)
}
