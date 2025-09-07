// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"os"
	"sort"
	"strings"

	"itiquette/git-provider-sync/internal/adapters/configuration/dto"
)

// ConfigSource represents a configuration from a specific source with priority.
type ConfigSource struct {
	Name     string                // Source identifier (e.g., "user", "project", "env")
	Config   *dto.AppConfiguration // The configuration data
	Priority int                   // Higher number = higher precedence
}

// MergeConfigs merges multiple configuration sources according to priority
// Pure function - no side effects, no mutations.
func MergeConfigs(sources []ConfigSource) *dto.AppConfiguration {
	if len(sources) == 0 {
		return &dto.AppConfiguration{
			GitProviderSyncConfs: make(map[string]dto.Environment),
		}
	}

	// Sort by priority (lowest to highest)
	sorted := make([]ConfigSource, len(sources))
	copy(sorted, sources)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	// Start with empty config
	result := &dto.AppConfiguration{
		GitProviderSyncConfs: make(map[string]dto.Environment),
	}

	// Apply each source in order
	for _, source := range sorted {
		if source.Config != nil {
			result = mergeAppConfig(result, source.Config)
		}
	}

	return result
}

// MergeAppConfig merges two AppConfiguration instances
// returns a new instance without modifying inputs.
func mergeAppConfig(base, override *dto.AppConfiguration) *dto.AppConfiguration {
	if base == nil && override == nil {
		return &dto.AppConfiguration{
			GitProviderSyncConfs: make(map[string]dto.Environment),
		}
	}

	if base == nil {
		return copyAppConfig(override)
	}

	if override == nil {
		return copyAppConfig(base)
	}

	result := &dto.AppConfiguration{
		GitProviderSyncConfs: make(map[string]dto.Environment),
	}

	// Copy all base environments
	for envName, env := range base.GitProviderSyncConfs {
		result.GitProviderSyncConfs[envName] = copyEnvironment(env)
	}

	// Merge or add override environments
	for envName, overrideEnv := range override.GitProviderSyncConfs {
		if baseEnv, exists := result.GitProviderSyncConfs[envName]; exists {
			result.GitProviderSyncConfs[envName] = mergeEnvironments(baseEnv, overrideEnv)
		} else {
			result.GitProviderSyncConfs[envName] = copyEnvironment(overrideEnv)
		}
	}

	return result
}

// MergeEnvironments merges two Environment instances.
func mergeEnvironments(base, override dto.Environment) dto.Environment {
	result := make(dto.Environment)

	// Copy all base sources
	for sourceName, source := range base {
		result[sourceName] = copySyncConfig(source)
	}

	// Merge or add override sources
	for sourceName, overrideSource := range override {
		if baseSource, exists := result[sourceName]; exists {
			result[sourceName] = mergeSyncConfigs(baseSource, overrideSource)
		} else {
			result[sourceName] = copySyncConfig(overrideSource)
		}
	}

	return result
}

// MergeSyncConfigs merges two SyncConfig instances.
func mergeSyncConfigs(base, override dto.SyncConfig) dto.SyncConfig {
	result := copySyncConfig(base)

	// Merge BaseConfig fields
	result.BaseConfig = mergeBaseConfigs(base.BaseConfig, override.BaseConfig)

	// Override string fields if not empty
	if override.ActiveFromLimit != "" {
		result.ActiveFromLimit = override.ActiveFromLimit
	}

	// Override boolean fields (always take override value)
	result.IncludeForks = override.IncludeForks

	// Merge Repositories
	result.Repositories = mergeRepositoryOptions(base.Repositories, override.Repositories)

	// Merge Mirrors
	result.Mirrors = mergeMirrors(base.Mirrors, override.Mirrors)

	return result
}

// MergeBaseConfigs merges two BaseConfig instances.
func mergeBaseConfigs(base, override dto.BaseConfig) dto.BaseConfig {
	result := base

	// Merge Auth
	result.Auth = mergeAuthConfigs(base.Auth, override.Auth)

	// Override string fields if not empty
	if override.Domain != "" {
		result.Domain = override.Domain
	}

	if override.Owner != "" {
		result.Owner = override.Owner
	}

	if override.OwnerType != "" {
		result.OwnerType = override.OwnerType
	}

	if override.ProviderType != "" {
		result.ProviderType = override.ProviderType
	}

	// Override boolean (always take override value)
	result.UseGitBinary = override.UseGitBinary

	return result
}

// MergeAuthConfigs merges two AuthConfig instances.
func mergeAuthConfigs(base, override dto.AuthConfig) dto.AuthConfig {
	result := base

	// Merge string fields
	result.CertDirPath = mergeString(base.CertDirPath, override.CertDirPath)
	result.HTTPScheme = mergeString(base.HTTPScheme, override.HTTPScheme)
	result.Token = mergeString(base.Token, override.Token)
	result.TokenFile = mergeString(base.TokenFile, override.TokenFile)
	result.Protocol = mergeString(base.Protocol, override.Protocol)
	result.ProxyURL = mergeString(base.ProxyURL, override.ProxyURL)
	result.SSHCommand = mergeString(base.SSHCommand, override.SSHCommand)
	result.SSHURLRewriteFrom = mergeString(base.SSHURLRewriteFrom, override.SSHURLRewriteFrom)
	result.SSHURLRewriteTo = mergeString(base.SSHURLRewriteTo, override.SSHURLRewriteTo)

	// Merge int fields
	result.RequestTimeout = mergeInt(base.RequestTimeout, override.RequestTimeout)
	result.GitTimeout = mergeInt(base.GitTimeout, override.GitTimeout)
	result.HTTPTimeout = mergeInt(base.HTTPTimeout, override.HTTPTimeout)

	return result
}

// MergeString returns override if not empty, otherwise base.
func mergeString(base, override string) string {
	if override != "" {
		return override
	}

	return base
}

// MergeInt returns override if not zero, otherwise base.
func mergeInt(base, override int) int {
	if override != 0 {
		return override
	}

	return base
}

// MergeRepositoryOptions merges two RepositoriesOption instances.
func mergeRepositoryOptions(base, override dto.RepositoriesOption) dto.RepositoriesOption {
	result := dto.RepositoriesOption{
		Include: mergeStringSlices(base.Include, override.Include),
		Exclude: mergeStringSlices(base.Exclude, override.Exclude),
	}

	return result
}

// MergeStringSlices merges two string slices
// Override completely replaces base if not empty.
func mergeStringSlices(base, override []string) []string {
	if len(override) > 0 {
		result := make([]string, len(override))
		copy(result, override)

		return result
	}

	if len(base) > 0 {
		result := make([]string, len(base))
		copy(result, base)

		return result
	}

	return []string{}
}

// MergeMirrors merges two mirror maps.
func mergeMirrors(base, override map[string]dto.MirrorConfig) map[string]dto.MirrorConfig {
	if base == nil && override == nil {
		return make(map[string]dto.MirrorConfig)
	}

	result := make(map[string]dto.MirrorConfig)

	// Copy all base mirrors
	for name, mirror := range base {
		result[name] = copyMirrorConfig(mirror)
	}

	// Merge or add override mirrors
	for name, overrideMirror := range override {
		if baseMirror, exists := result[name]; exists {
			result[name] = mergeMirrorConfigs(baseMirror, overrideMirror)
		} else {
			result[name] = copyMirrorConfig(overrideMirror)
		}
	}

	return result
}

// MergeMirrorConfigs merges two MirrorConfig instances.
func mergeMirrorConfigs(base, override dto.MirrorConfig) dto.MirrorConfig {
	result := copyMirrorConfig(base)

	// Merge BaseConfig
	result.BaseConfig = mergeBaseConfigs(base.BaseConfig, override.BaseConfig)

	// Override Path if not empty
	if override.Path != "" {
		result.Path = override.Path
	}

	// Merge Settings
	result.Settings = mergeMirrorSettings(base.Settings, override.Settings)

	return result
}

// MergeMirrorSettings merges two MirrorSettings instances
// Note: For booleans, we always take the override value since Go doesn't distinguish
// Between unset and false. In practice, this is handled at the config loading level
// Where only explicitly set values are included in the override dto.
func mergeMirrorSettings(base, override dto.MirrorSettings) dto.MirrorSettings {
	result := base

	// Limitation: Go's zero values can't distinguish unset from false
	// Production should use pointers or IsSet fields
	result.AlphaNumHyphName = override.AlphaNumHyphName
	result.Disabled = override.Disabled
	result.ForcePush = override.ForcePush
	result.IgnoreInvalidName = override.IgnoreInvalidName

	// Override string fields if not empty
	if override.DescriptionPrefix != "" {
		result.DescriptionPrefix = override.DescriptionPrefix
	}

	if override.GitHubUploadURL != "" {
		result.GitHubUploadURL = override.GitHubUploadURL
	}

	if override.Visibility != "" {
		result.Visibility = override.Visibility
	}

	return result
}

// Copy functions to ensure immutability

func copyAppConfig(cfg *dto.AppConfiguration) *dto.AppConfiguration {
	if cfg == nil {
		return &dto.AppConfiguration{
			GitProviderSyncConfs: make(map[string]dto.Environment),
		}
	}

	result := &dto.AppConfiguration{
		GitProviderSyncConfs: make(map[string]dto.Environment),
	}

	for envName, env := range cfg.GitProviderSyncConfs {
		result.GitProviderSyncConfs[envName] = copyEnvironment(env)
	}

	return result
}

func copyEnvironment(env dto.Environment) dto.Environment {
	result := make(dto.Environment)
	for name, sync := range env {
		result[name] = copySyncConfig(sync)
	}

	return result
}

func copySyncConfig(cfg dto.SyncConfig) dto.SyncConfig {
	result := dto.SyncConfig{
		BaseConfig:      cfg.BaseConfig,
		ActiveFromLimit: cfg.ActiveFromLimit,
		IncludeForks:    cfg.IncludeForks,
		Repositories: dto.RepositoriesOption{
			Include: copyStringSlice(cfg.Repositories.Include),
			Exclude: copyStringSlice(cfg.Repositories.Exclude),
		},
		Mirrors: make(map[string]dto.MirrorConfig),
	}

	for name, mirror := range cfg.Mirrors {
		result.Mirrors[name] = copyMirrorConfig(mirror)
	}

	return result
}

func copyMirrorConfig(cfg dto.MirrorConfig) dto.MirrorConfig {
	return dto.MirrorConfig{
		BaseConfig: cfg.BaseConfig,
		Path:       cfg.Path,
		Settings:   cfg.Settings,
	}
}

func copyStringSlice(slice []string) []string {
	if slice == nil {
		return nil
	}

	result := make([]string, len(slice))
	copy(result, slice)

	return result
}

// ProcessRepositoryLists converts comma-separated strings to slices
// is a pure function that returns a new configuration.
func ProcessRepositoryLists(cfg *dto.AppConfiguration) *dto.AppConfiguration {
	result := copyAppConfig(cfg)

	for envName, env := range result.GitProviderSyncConfs {
		for sourceName, source := range env {
			if len(source.Repositories.Include) == 1 && strings.Contains(source.Repositories.Include[0], ",") {
				source.Repositories.Include = splitCommaSeparated(source.Repositories.Include[0])
			}

			if len(source.Repositories.Exclude) == 1 && strings.Contains(source.Repositories.Exclude[0], ",") {
				source.Repositories.Exclude = splitCommaSeparated(source.Repositories.Exclude[0])
			}

			env[sourceName] = source

			for mirrorName, mirror := range source.Mirrors {
				// Note: MirrorConfig doesn't have Repositories in the current struct,
				// But keeping this for completeness
				source.Mirrors[mirrorName] = mirror
			}
		}

		result.GitProviderSyncConfs[envName] = env
	}

	return result
}

// SplitCommaSeparated splits a comma-separated string and trims whitespace.
func splitCommaSeparated(s string) []string {
	parts := strings.Split(s, ",")

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// ExpandVariables expands environment variables in configuration
// is a pure function that takes an environment map for expansion.
func ExpandVariables(cfg *dto.AppConfiguration, envVars map[string]string) *dto.AppConfiguration {
	result := copyAppConfig(cfg)

	for envName, env := range result.GitProviderSyncConfs {
		for sourceName, source := range env {
			// Expand in auth
			source.Auth.Token = expandVar(source.Auth.Token, envVars)
			source.Auth.TokenFile = expandVar(source.Auth.TokenFile, envVars)
			source.Auth.ProxyURL = expandVar(source.Auth.ProxyURL, envVars)
			source.Auth.CertDirPath = expandVar(source.Auth.CertDirPath, envVars)
			source.Auth.SSHCommand = expandVar(source.Auth.SSHCommand, envVars)

			env[sourceName] = source

			// Expand in mirrors
			for mirrorName, mirror := range source.Mirrors {
				mirror.Auth.Token = expandVar(mirror.Auth.Token, envVars)
				mirror.Auth.TokenFile = expandVar(mirror.Auth.TokenFile, envVars)
				mirror.Auth.ProxyURL = expandVar(mirror.Auth.ProxyURL, envVars)
				mirror.Auth.CertDirPath = expandVar(mirror.Auth.CertDirPath, envVars)
				mirror.Auth.SSHCommand = expandVar(mirror.Auth.SSHCommand, envVars)
				mirror.Path = expandVar(mirror.Path, envVars)

				source.Mirrors[mirrorName] = mirror
			}
		}

		result.GitProviderSyncConfs[envName] = env
	}

	return result
}

// ExpandVar expands environment variables in a string
// Pure function using provided environment map.
func expandVar(input string, envVars map[string]string) string {
	if input == "" || !strings.Contains(input, "$") {
		return input
	}

	// Use os.Expand for proper variable expansion
	return os.Expand(input, func(key string) string {
		if val, ok := envVars[key]; ok {
			return val
		}

		return "" // Return empty for undefined vars
	})
}

// ApplyProviderTokens applies provider-specific token overrides
// is a pure function that returns a new configuration.
func ApplyProviderTokens(cfg *dto.AppConfiguration, providerTokens map[string]string) *dto.AppConfiguration {
	result := copyAppConfig(cfg)

	for envName, env := range result.GitProviderSyncConfs {
		for sourceName, source := range env {
			// Check for provider-specific token
			tokenKey := "GPS_" + strings.ToUpper(source.ProviderType) + "_TOKEN"
			if token, exists := providerTokens[tokenKey]; exists && token != "" {
				source.Auth.Token = token
				source.Auth.TokenFile = "" // Clear token_file when env var is set
			}

			env[sourceName] = source

			// Apply to mirrors
			for mirrorName, mirror := range source.Mirrors {
				tokenKey := "GPS_" + strings.ToUpper(mirror.ProviderType) + "_TOKEN"
				if token, exists := providerTokens[tokenKey]; exists && token != "" {
					mirror.Auth.Token = token
					mirror.Auth.TokenFile = ""
				}

				source.Mirrors[mirrorName] = mirror
			}
		}

		result.GitProviderSyncConfs[envName] = env
	}

	return result
}
