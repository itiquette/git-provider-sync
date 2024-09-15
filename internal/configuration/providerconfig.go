// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"fmt"
	"itiquette/git-provider-sync/internal/model"
	"strings"

	"github.com/rs/zerolog"
)

// ProviderConfig represents the configuration for a single provider.
type ProviderConfig struct {
	Provider         string            `koanf:"provider"`
	Domain           string            `koanf:"domain"`
	Token            string            `koanf:"token"`
	User             string            `koanf:"user"`
	Group            string            `koanf:"group"`
	GitInfo          model.GitInfo     `koanf:"gitoption"`
	Exclude          map[string]string `koanf:"exclude"`
	Include          map[string]string `koanf:"include"`
	Providerspecific map[string]string `koanf:"providerspecific"`
}

// String returns a string representation of ProviderConfig, masking the token.
func (p ProviderConfig) String() string {
	return fmt.Sprintf("ProviderConfig: Provider: %s, Domain: %s, Token: <****>, User: %s, GitInfo: %v,  Group: %s, Exclude: %v, Include: %v, ProviderSpecific: %v",
		p.Provider, p.Domain, p.User, p.GitInfo, p.Group, p.Exclude, p.Include, p.Providerspecific)
}

// DebugLog logs the ProviderConfig details at debug level.
func (p ProviderConfig) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	event := logger.Debug(). //nolint:zerologlint
					Str("provider", p.Provider).
					Fields(p.Exclude).
					Fields(p.Include)

	switch strings.ToLower(p.Provider) {
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
	return p.Providerspecific["archivetargetdir"]
}

// DirectoryTargetDir returns the directory target directory.
func (p ProviderConfig) DirectoryTargetDir() string {
	return p.Providerspecific["directorytargetdir"]
}

// IsGroup returns true if the configuration is for a group.
func (p ProviderConfig) IsGroup() bool {
	return p.Group != ""
}

// IncludedRepositories returns a slice of included repository names.
func (p ProviderConfig) IncludedRepositories() []string {
	return splitAndTrim(p.Include["repositories"])
}

// ExcludedRepositories returns a slice of excluded repository names.
func (p ProviderConfig) ExcludedRepositories() []string {
	return splitAndTrim(p.Exclude["repositories"])
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
