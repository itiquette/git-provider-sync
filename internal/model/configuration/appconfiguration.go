// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package model defines configuration structures for git provider sync operations
// Including application settings, sync options, authentication, and mirror targets
package model

import (
	"fmt"

	"github.com/rs/zerolog"
)

// AppConfiguration represents the entire application configuration.
type AppConfiguration struct {
	GitProviderSyncConfs map[string]Environment `json:"gitprovidersync" koanf:"gitprovidersync"`
}

// Environment represents a configuration environment (production, staging, etc).
type Environment map[string]SyncConfig

// BaseConfig holds common configuration fields for both source and mirror.
type BaseConfig struct {
	Auth         AuthConfig `koanf:"auth"`           // Authentication configuration including tokens and certificates
	Domain       string     `koanf:"domain"`         // Provider domain (e.g., github.com, gitlab.com, gitea.example.com)
	Owner        string     `koanf:"owner"`          // Repository owner name (user or organization)
	OwnerType    string     `koanf:"owner_type"`     // Owner type: "user", "group", or "org"
	ProviderType string     `koanf:"provider_type"`  // Git provider: "github", "gitlab", "gitea", "directory", or "archive"
	UseGitBinary bool       `koanf:"use_git_binary"` // Use native git binary instead of go-git library for operations
}

// SyncConfig represents a source configuration with its mirrors and backups.
type SyncConfig struct {
	BaseConfig      `koanf:",squash"`
	ActiveFromLimit string             `koanf:"active_from_limit"` // Filter repositories by activity period (e.g., "30d", "1y")
	IncludeForks    bool               `koanf:"include_forks"`     // Include forked repositories in synchronization
	Repositories    RepositoriesOption `koanf:"repositories"`      // Repository selection criteria (include/exclude patterns)

	Mirrors map[string]MirrorConfig `koanf:"mirrors"` // Named mirror configurations for this source
}

// AuthConfig combines HTTP and SSH configurations
// CONFIGURATION EXAMPLES:
// GitHub with token authentication:
//
//	auth:
//	  token: "ghp_xxxxxxxxxxxxxxxxxxxx"
//	  http_scheme: "https"
//	  request_timeout: 30
//
// GitLab with custom domain and proxy:
//
//	auth:
//	  token: "glpat-xxxxxxxxxxxxxxxxxxxx"
//	  http_scheme: "https"
//	  proxy_url: "http://proxy.example.com:8080"
//	  cert_dir_path: "/etc/ssl/certs"
//	  request_timeout: 60
//
// SSH-based authentication for GitHub:
//
//	auth:
//	  protocol: "ssh"
//	  ssh_command: "ssh -i /path/to/key -o StrictHostKeyChecking=no"
//	  ssh_url_rewrite_from: "https://github.com/"
//	  ssh_url_rewrite_to: "git@github.com:"
//
// Corporate environment with certificates:
//
//	auth:
//	  token: "company_token"
//	  http_scheme: "https"
//	  cert_dir_path: "/opt/company/certs"
//	  proxy_url: "https://corporate-proxy:8443"
//	  request_timeout: 120
type AuthConfig struct {
	CertDirPath       string `koanf:"cert_dir_path"`        // Directory containing SSL/TLS certificates for HTTPS
	HTTPScheme        string `koanf:"http_scheme"`          // HTTP scheme: "http" or "https" (default: "https")
	RequestTimeout    int    `koanf:"request_timeout"`      // HTTP request timeout in seconds (default: 30)
	GitTimeout        int    `koanf:"git_timeout"`          // Git operation timeout in seconds (default: 300, 5 minutes)
	HTTPTimeout       int    `koanf:"http_timeout"`         // HTTP/API timeout in seconds (default: 30)
	Token             string `koanf:"token"`                // API token for provider authentication (never logged for security)
	TokenFile         string `koanf:"token_file"`           // Path to file containing API token (more secure than token field)
	Protocol          string `koanf:"protocol"`             // Transport protocol: "tls", "tcp", or "ssh" (default: "tls")
	ProxyURL          string `koanf:"proxy_url"`            // HTTP proxy URL for requests (optional)
	SSHCommand        string `koanf:"ssh_command"`          // Custom SSH command for git operations (optional)
	SSHURLRewriteFrom string `koanf:"ssh_url_rewrite_from"` // Pattern to rewrite SSH URLs from (optional)
	SSHURLRewriteTo   string `koanf:"ssh_url_rewrite_to"`   // Pattern to rewrite SSH URLs to (optional)
}

// MirrorConfig represents a mirror target configuration.
type MirrorConfig struct {
	BaseConfig `koanf:",squash"`
	Path       string         `koanf:"path"`     // Target path for directory/archive mirrors
	Settings   MirrorSettings `koanf:"settings"` // Mirror-specific behavior settings
}

// MirrorSettings represents mirror-specific settings.
type MirrorSettings struct {
	AlphaNumHyphName  bool   `koanf:"alphanumhyph_name"`   // Transform repository names to alphanumeric-hyphen format
	DescriptionPrefix string `koanf:"description_prefix"`  // Prefix added to mirrored repository descriptions
	Disabled          bool   `koanf:"disabled"`            // Whether this mirror is disabled (default: true for safety)
	ForcePush         bool   `koanf:"force_push"`          // Allow force pushes to overwrite target repositories
	GitHubUploadURL   string `koanf:"github_uploadurl"`    // Custom GitHub Enterprise upload URL
	IgnoreInvalidName bool   `koanf:"ignore_invalid_name"` // Skip repositories with invalid names instead of failing
	Visibility        string `koanf:"visibility"`          // Repository visibility: "public", "private", or "internal"
}

// String formats authentication configuration for structured logging.
func (a AuthConfig) String() string {
	return fmt.Sprintf("AuthConfig: Protocol: %s, HTTPScheme: %s, ProxyURL: %s, CertDirPath: %s, SSHCommand: %s",
		a.Protocol, a.HTTPScheme, a.ProxyURL, a.CertDirPath, a.SSHCommand)
}

// String formats sync configuration for structured logging.
func (s SyncConfig) String() string {
	return fmt.Sprintf("SyncConfig: ProviderType: %s, Domain: %s, Owner: %s, OwnerType: %s",
		s.ProviderType, s.Domain, s.Owner, s.OwnerType)
}

// FillDefaults applies provider-specific defaults.
func (b *BaseConfig) FillDefaults() {
	if b.Domain == "" {
		b.Domain = b.GetDomain()
	}

	if b.OwnerType == "" {
		b.OwnerType = GROUP
	}

	if b.Auth.HTTPScheme == "" {
		b.Auth.HTTPScheme = HTTPS
	}

	if b.Auth.Protocol == "" {
		b.Auth.Protocol = TLS
	}

	if b.Auth.RequestTimeout == 0 {
		b.Auth.RequestTimeout = 30
	}

	// Set default git timeout to 5 minutes if not specified
	if b.Auth.GitTimeout == 0 {
		b.Auth.GitTimeout = 300
	}

	// Set default HTTP timeout to 30 seconds if not specified
	if b.Auth.HTTPTimeout == 0 {
		b.Auth.HTTPTimeout = 30
	}
}

// GetDomain resolves provider domain with fallback to standard service endpoints.
func (b BaseConfig) GetDomain() string {
	if b.Domain == "" {
		switch b.ProviderType {
		case "gitea":
			return "gitea.com"
		case "github":
			return "github.com"
		case "gitlab":
			return "gitlab.com"
		default:
			return ""
		}
	}

	return b.Domain
}

// FillDefaults propagates configuration defaults through sync configuration hierarchy.
func (s *SyncConfig) FillDefaults() {
	s.BaseConfig.FillDefaults()

	// Loop over map with key for updating
	for name, mirror := range s.Mirrors {
		mirror.FillDefaults()
		s.Mirrors[name] = mirror // Update the map value
	}
}

// FillDefaults cascades configuration defaults through entire application configuration tree.
func (a *AppConfiguration) FillDefaults() {
	for envName, env := range a.GitProviderSyncConfs {
		for sourceName, source := range env {
			source.FillDefaults()
			a.GitProviderSyncConfs[envName][sourceName] = source
		}
	}
}

// FillDefaults initializes mirror configuration with security-first defaults (disabled by default).
func (m *MirrorConfig) FillDefaults() {
	m.BaseConfig.FillDefaults()

	m.Settings.Disabled = true
}

// IsGroup returns true if the sync configuration is for a group owner.
func (s SyncConfig) IsGroup() bool {
	return s.OwnerType == "group"
}

// DebugLog constructs structured debug log event with configuration context and mirror details.
func (s SyncConfig) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	event := logger.Debug(). //nolint:zerologlint // Event returned for caller to dispatch
					Str("type", s.ProviderType).
					Str("domain", s.GetDomain()).
					Str("owner", s.Owner).
					Str("ownerType", s.OwnerType).
					Interface("repositories", s.Repositories).
					Str("auth", s.Auth.String())

	// Handle mirrors safely - create separate dict for each mirror
	for name, mirror := range s.Mirrors {
		mirrorDict := zerolog.Dict().
			Str("provider_type", mirror.ProviderType).
			Str("domain", mirror.GetDomain()).
			Str("owner", mirror.Owner).
			Str("owner_type", mirror.OwnerType).
			Str("path", mirror.Path).
			Str("auth", mirror.Auth.String())
		event = event.Dict("mirror_"+name, mirrorDict)
	}

	return event
}

// DebugLog renders complete application configuration hierarchy for diagnostic purposes.
func (a AppConfiguration) DebugLog(logger *zerolog.Logger) {
	for envName, env := range a.GitProviderSyncConfs {
		logger.Debug().Msgf("Environment: %s", envName)

		for sourceName, source := range env {
			// Create a new event for each source to avoid race conditions
			event := source.DebugLog(logger)
			event.Msgf("Source: %s", sourceName)
		}
	}
}

// IsArchive returns true if the mirror configuration is for archive provider.
func (m MirrorConfig) IsArchive() bool {
	return m.ProviderType == "archive"
}

// IsDirectory returns true if the mirror configuration is for directory provider.
func (m MirrorConfig) IsDirectory() bool {
	return m.ProviderType == "directory"
}
