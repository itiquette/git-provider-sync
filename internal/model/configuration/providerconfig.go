// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// ProviderConfig represents the configuration for a single provider.
type ProviderConfig struct {
	ProviderType string             `koanf:"providertype"`
	Domain       string             `koanf:"domain"`
	UploadDomain string             `koanf:"uploaddomain"`
	Group        string             `koanf:"group"`
	User         string             `koanf:"user"`
	Repositories RepositoriesOption `koanf:"repositories"`
	Git          GitOption          `koanf:"git"`
	Project      ProjectOption      `koanf:"project"`
	HTTPClient   HTTPClientOption   `koanf:"httpclient"`
	SSHClient    SSHClientOption    `koanf:"sshclient"`
	SyncRun      SyncRunOption      `koanf:"syncrun"`
	Additional   map[string]string  `koanf:"additional"`
}

// String returns a string representation of ProviderConfig, masking the token.
func (p ProviderConfig) String() string {
	return fmt.Sprintf("ProviderConfig: ProviderType: %s, Domain: %s, User: %s, Group: %s, Repository: %v, Git: %v,   HTTPClient: %v, Extras: %v",
		p.ProviderType, p.Domain, p.User, p.Group, p.Repositories, p.Git, p.HTTPClient, p.Additional)
}

// DebugLog logs the ProviderConfig details at debug level.
func (p ProviderConfig) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	event := logger.Debug(). //nolint:zerologlint
					Str("provider", p.ProviderType).
					Fields(p.Repositories.Exclude).
					Fields(p.Repositories.Include)

	switch strings.ToLower(p.ProviderType) {
	case DIRECTORY:
		event.Str("target directory", p.DirectoryTargetDir())
	case ARCHIVE:
		event.Str("target directory", p.ArchiveTargetDir())
	default:
		event.Str("domain", p.Domain).
			Strs("user,group", []string{p.User, p.Group})
	}

	return event
}

// ProvidersConfig represents the configuration for source and target providers.
type ProvidersConfig struct {
	SourceProvider  ProviderConfig            `koanf:"source"`
	ProviderTargets map[string]ProviderConfig `koanf:"targets"`
}

// AppConfiguration represents the entire application configuration.
type AppConfiguration struct {
	Configurations map[string]ProvidersConfig `koanf:"configurations"`
}

// ArchiveTargetDir returns the archive target directory.
func (p ProviderConfig) ArchiveTargetDir() string {
	return p.Additional["archivetargetdir"]
}

// DirectoryTargetDir returns the directory target directory.
func (p ProviderConfig) DirectoryTargetDir() string {
	return p.Additional["directorytargetdir"]
}

// GitHubUploadURL returns the special GitHubUploadURL.
func (p ProviderConfig) GitHubUploadURL() string {
	return p.Additional["githubuploadurl"]
}

func (p ProviderConfig) GetDomain() string {

	if p.Domain == "" {
		switch p.ProviderType {
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

	return p.Domain
}

// IsGroup returns true if the configuration is for a group.
func (p ProviderConfig) IsGroup() bool {
	return p.Group != ""
}

// DebugLog logs the AppConfiguration details at debug level.
func (a AppConfiguration) DebugLog(logger *zerolog.Logger) {
	for name, config := range a.Configurations {
		config.SourceProvider.DebugLog(logger).Msg(name + ":SourceProvider")

		for key, target := range config.ProviderTargets {
			target.DebugLog(logger).Msg(key + ":TargetProvider")
		}
	}
}
