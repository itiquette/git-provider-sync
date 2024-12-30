// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"

	"github.com/rs/zerolog"
)

// AppConfiguration represents the entire application configuration.
type AppConfiguration struct {
	GitProviderSyncConfs map[string]Environment `koanf:"gitprovidersync"`
}

// Environment represents a configuration environment (production, staging, etc).
type Environment map[string]SyncConfig

// BaseConfig holds common configuration fields for both source and mirror.
type BaseConfig struct {
	Auth         AuthConfig `koanf:"auth"`
	Domain       string     `koanf:"domain"`
	Owner        string     `koanf:"owner"`
	OwnerType    string     `koanf:"owner_type"`
	ProviderType string     `koanf:"provider_type"`
}

// SyncConfig represents a source configuration with its mirrors and backups.
type SyncConfig struct {
	BaseConfig      `koanf:",squash"`
	ActiveFromLimit string             `koanf:"active_from_limit"`
	IncludeForks    bool               `koanf:"include_forks"`
	Repositories    RepositoriesOption `koanf:"repositories"`
	UseGitBinary    bool               `koanf:"use_git_binary"`

	Mirrors map[string]MirrorConfig `koanf:"mirrors"`
}

// AuthConfig combines HTTP and SSH configurations.
type AuthConfig struct {
	CertDirPath       string `koanf:"cert_dir_path"`
	HTTPScheme        string `koanf:"http_scheme"`
	Token             string `koanf:"token"`
	Protocol          string `koanf:"method"`
	ProxyURL          string `koanf:"proxy_url"`
	SSHCommand        string `koanf:"ssh_command"`
	SSHURLRewriteFrom string `koanf:"ssh_url_rewrite_from"`
	SSHURLRewriteTo   string `koanf:"ssh_url_rewrite_to"`
}

// MirrorConfig represents a mirror target configuration.
type MirrorConfig struct {
	BaseConfig `koanf:",squash"`
	Path       string         `koanf:"path"`
	Settings   MirrorSettings `koanf:"settings"`
}

// MirrorSettings represents mirror-specific settings.
type MirrorSettings struct {
	ASCIIName         bool   `koanf:"ascii_name"`
	DescriptionPrefix string `koanf:"description_prefix"`
	Disabled          bool   `koanf:"disabled"`
	ForcePush         bool   `koanf:"force_push"`
	GitHubUploadURL   string `koanf:"github_uploadurl"`
	IgnoreInvalidName bool   `koanf:"ignore_invalid_name"`
	Visibility        string `koanf:"visibility"`
}

// String methods for logging.
func (a AuthConfig) String() string {
	return fmt.Sprintf("AuthConfig: Protocol: %s, HTTPScheme: %s, ProxyURL: %s, CertDirPath: %s, SSHCommand: %s",
		a.Protocol, a.HTTPScheme, a.ProxyURL, a.CertDirPath, a.SSHCommand)
}

func (s SyncConfig) String() string {
	return fmt.Sprintf("SourceConfig: ProviderType: %s, Domain: %s, Owner: %s, OwnerType: %s",
		s.ProviderType, s.Domain, s.Owner, s.OwnerType)
}

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
}

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

func (s *SyncConfig) FillDefaults() {
	s.BaseConfig.FillDefaults()

	// Loop over map with key for updating
	for name, mirror := range s.Mirrors {
		mirror.FillDefaults()
		s.Mirrors[name] = mirror // Update the map value
	}
}

func (a *AppConfiguration) FillDefaults() {
	for envName, env := range a.GitProviderSyncConfs {
		for sourceName, source := range env {
			source.FillDefaults()
			a.GitProviderSyncConfs[envName][sourceName] = source
		}
	}
}

func (m *MirrorConfig) FillDefaults() {
	m.BaseConfig.FillDefaults()

	m.Settings.Disabled = true
}

func (s SyncConfig) IsGroup() bool {
	return s.OwnerType == "group"
}

func (s SyncConfig) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	event := logger.Debug(). //nolint
					Str("type", s.ProviderType).
					Str("domain", s.GetDomain()).
					Str("owner", s.Owner).
					Str("ownerType", s.OwnerType).
					Interface("repositories", s.Repositories).
					Interface("auth", s.Auth.String())

	for i, mirror := range s.Mirrors {
		event.Interface(fmt.Sprintf("mirror_%-v", i), mirror)
	}

	return event
}

func (a AppConfiguration) DebugLog(logger *zerolog.Logger) {
	for envName, env := range a.GitProviderSyncConfs {
		logger.Debug().Msgf("Environment: %s", envName)

		for sourceName, source := range env {
			source.DebugLog(logger).Msgf("Source: %s", sourceName)
		}
	}
}

func (m MirrorConfig) IsArchive() bool {
	return m.ProviderType == "archive"
}

func (m MirrorConfig) IsDirectory() bool {
	return m.ProviderType == "directory"
}
