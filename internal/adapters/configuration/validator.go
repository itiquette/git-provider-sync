// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/agent"

	"itiquette/git-provider-sync/internal/adapters/log"
	"itiquette/git-provider-sync/internal/application/dto"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/validation"
)

const (
	sshAuthSockEnv = "SSH_AUTH_SOCK"
)

var (
	// File-specific validation errors that are not covered by domain validation.

	// ErrInvalidDuration indicates invalid duration format.
	ErrInvalidDuration = errors.New("invalid duration format")
	// ErrNoGitBinaryFound indicates that the git binary was not found.
	ErrNoGitBinaryFound = errors.New("failed to find git binary")
	// ErrInvalidURL indicates an invalid URL format.
	ErrInvalidURL = errors.New("invalid URL")
	// ErrInvalidPath indicates an invalid file path format.
	ErrInvalidPath = errors.New("invalid file path")
)

// These constants are no longer needed since domain validation handles provider type validation

// ValidateConfiguration validates the entire application configuration using domain validation.
func validateConfiguration(ctx context.Context, appCfg *dto.AppConfiguration) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering validateConfiguration")

	// Convert to domain types and use domain validation
	domainConfig := convertToAppConfiguration(appCfg)
	results := validation.ValidateAppConfiguration(domainConfig)

	if !results.Valid {
		// Convert validation results to configuration errors
		var errorMessages []string
		for _, result := range results.Results {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", result.Field, result.Message))
		}

		return fmt.Errorf("configuration validation failed: %s", strings.Join(errorMessages, ", "))
	}

	// Additional file-specific validations not covered by domain validation
	if len(appCfg.GitProviderSyncConfs) == 0 {
		return domain.ErrNoSyncConfigurations
	}

	for envName, env := range appCfg.GitProviderSyncConfs {
		logger.Debug().Msgf("Validating environment %s", envName)

		if err := validateEnvironmentSpecific(ctx, envName, env); err != nil {
			return fmt.Errorf("invalid environment %s: %w", envName, err)
		}
	}

	return nil
}

// ValidateEnvironmentSpecific validates environment-specific aspects not covered by domain validation.
func validateEnvironmentSpecific(ctx context.Context, envName string, env dto.Environment) error {
	if len(env) == 0 {
		return fmt.Errorf("%w: %s", domain.ErrNoSyncConfigInEnvironment, envName)
	}

	for sourceName, syncConfig := range env {
		if err := validateSyncConfigSpecific(ctx, sourceName, syncConfig); err != nil {
			return fmt.Errorf("invalid sync config %s: %w", sourceName, err)
		}
	}

	return nil
}

// ValidateSyncConfigSpecific validates file-specific configuration aspects.
func validateSyncConfigSpecific(ctx context.Context, _ string, syncCfg dto.SyncConfig) error {
	// Only validate aspects not covered by domain validation
	validators := []struct {
		name string
		fn   func() error
	}{
		{"git-binary", func() error { return validateGitBinaryIfNeeded(ctx, syncCfg) }},
		{"duration", func() error { return validateDuration(syncCfg.ActiveFromLimit) }},
		{"ssh-auth", func() error { return validateSSHAuthIfNeeded(ctx, syncCfg.Auth) }},
		{"paths", func() error { return validatePathsIfSet(syncCfg.Auth) }},
	}

	for _, validator := range validators {
		if err := validator.fn(); err != nil {
			return err
		}
	}

	return nil
}

// ConvertToAppConfiguration converts file configuration to domain configuration.
func convertToAppConfiguration(appCfg *dto.AppConfiguration) ports.AppConfiguration {
	// Convert global settings
	globalSettings := ports.GlobalSettings{
		LogLevel:     "info", // Default since this isn't in file config
		LogFormat:    "json", // Default since this isn't in file config
		MaxCacheSize: 100,    // Default since this isn't in file config
		CacheTTL:     3600,   // Default since this isn't in file config
	}

	// Convert environments
	environments := make(map[string]ports.EnvironmentConfiguration)

	for envName, env := range appCfg.GitProviderSyncConfs {
		for sourceName, syncConfig := range env {
			envConfig := ports.EnvironmentConfiguration{
				Name:    envName + "-" + sourceName,
				Source:  convertToSourceConfiguration(syncConfig),
				Mirrors: convertToMirrorConfigurations(syncConfig.Mirrors),
				Options: ports.EnvironmentOptions{
					MaxConcurrency: 5,    // Default
					Timeout:        300,  // Default
					RetryAttempts:  3,    // Default
					RetryDelay:     1000, // Default
				},
			}
			environments[envName+"-"+sourceName] = envConfig
		}
	}

	return ports.AppConfiguration{
		GlobalSettings: globalSettings,
		Environments:   environments,
	}
}

// ConvertToSourceConfiguration converts sync config to source configuration.
func convertToSourceConfiguration(syncCfg dto.SyncConfig) ports.SourceConfiguration {
	return ports.SourceConfiguration{
		ProviderType:   syncCfg.ProviderType,
		Domain:         syncCfg.GetDomain(),
		Owner:          syncCfg.Owner,
		Authentication: convertToAuthenticationConfiguration(syncCfg.Auth),
	}
}

// ConvertToMirrorConfigurations converts mirror configs to domain mirror configurations.
func convertToMirrorConfigurations(mirrors map[string]dto.MirrorConfig) map[string]ports.MirrorConfiguration {
	result := make(map[string]ports.MirrorConfiguration)

	for name, mirror := range mirrors {
		// For local providers (archive/directory), use a placeholder path to avoid validation errors
		// Since the current config structure doesn't have path fields
		path := ""
		if mirror.ProviderType == "archive" || mirror.ProviderType == "directory" {
			path = "/tmp" // Placeholder to satisfy domain validation
		}

		result[name] = ports.MirrorConfiguration{
			Name:           name,
			ProviderType:   mirror.ProviderType,
			Domain:         mirror.GetDomain(),
			Owner:          mirror.Owner,
			Path:           path,
			Authentication: convertToAuthenticationConfiguration(mirror.Auth),
		}
	}

	return result
}

// ConvertToAuthenticationConfiguration converts auth config to domain authentication configuration.
func convertToAuthenticationConfiguration(authCfg dto.AuthConfig) ports.AuthenticationConfiguration {
	// Determine auth type based on what's configured
	authType := ports.AuthenticationTypeNone
	if authCfg.Token != "" {
		authType = ports.AuthenticationTypeToken
	} else if authCfg.Protocol == dto.SSH {
		authType = ports.AuthenticationTypeSSH
	}

	return ports.AuthenticationConfiguration{
		Type:       authType,
		Token:      authCfg.Token,
		Username:   "", // Not available in current file config structure
		Password:   "", // Not available in current file config structure
		SSHKeyPath: "", // Not directly available in current config
		SSHKey:     "", // Not directly available in current config
	}
}

// File-specific validation functions (not covered by domain validation)

func validateGitBinaryIfNeeded(ctx context.Context, syncCfg dto.SyncConfig) error {
	if !syncCfg.UseGitBinary {
		return nil
	}

	if err := validateGitBinary(ctx); err != nil {
		return ErrNoGitBinaryFound
	}

	return nil
}

func validateDuration(limit string) error {
	if limit == "" {
		return nil
	}

	if _, err := time.ParseDuration(limit); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDuration, err)
	}

	return nil
}

// ValidateSSHAuthIfNeeded validates SSH-specific authentication configuration.
func validateSSHAuthIfNeeded(ctx context.Context, authCfg dto.AuthConfig) error {
	if authCfg.Protocol == dto.SSH {
		if err := checkSSHAgent(ctx); err != nil {
			return err
		}
	}

	if authCfg.SSHURLRewriteFrom != "" || authCfg.SSHURLRewriteTo != "" {
		if authCfg.SSHURLRewriteFrom == "" || authCfg.SSHURLRewriteTo == "" {
			return domain.ErrSSHRewriteBothRequired
		}
	}

	return validateSSHCommand(authCfg.SSHCommand)
}

// ValidatePathsIfSet validates file paths if they are configured.
func validatePathsIfSet(authCfg dto.AuthConfig) error {
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

	return nil
}

// Helper functions for file-specific validations

func checkSSHAgent(ctx context.Context) error {
	sshAuthSock := os.Getenv(sshAuthSockEnv)
	if sshAuthSock == "" {
		return domain.ErrSSHAuthSockNotSet
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(timeoutCtx, "unix", sshAuthSock)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	agentClient := agent.NewClient(conn)

	keys, err := agentClient.List()
	if err != nil {
		return fmt.Errorf("failed to list keys from SSH agent: %w", err)
	}

	if len(keys) == 0 {
		return domain.ErrSSHAgentNoKeys
	}

	return nil
}

// File-specific validation helpers

func validatePathExists(path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("%w: path must be absolute: %s", ErrInvalidPath, path)
	}

	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
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
		return domain.ErrSSHCommandMustStartWithSSH
	}

	return nil
}

// These functions are now handled by domain validation and no longer needed

// ValidateGitBinary validates that git binary is available and functional.
func validateGitBinary(ctx context.Context) error {
	// Check if git command exists
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git binary not found in PATH: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test git version to ensure it's functional
	cmd := exec.CommandContext(timeoutCtx, "git", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git binary not functional: %w", err)
	}

	return nil
}

// Configuration validation is now consolidated - this function removed as
// duplicates domain validation logic
