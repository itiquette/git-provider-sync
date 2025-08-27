// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"time"
)

// AppConfigOption represents a functional option for configuring an ImmutableAppConfiguration.
type AppConfigOption func(ImmutableAppConfiguration) ImmutableAppConfiguration

// EnvironmentOption represents a functional option for configuring an ImmutableEnvironmentConfiguration.
type EnvironmentOption func(ImmutableEnvironmentConfiguration) ImmutableEnvironmentConfiguration

// SourceOption represents a functional option for configuring an ImmutableSourceConfiguration.
type SourceOption func(ImmutableSourceConfiguration) ImmutableSourceConfiguration

// MirrorOption represents a functional option for configuring an ImmutableMirrorConfiguration.
type MirrorOption func(ImmutableMirrorConfiguration) ImmutableMirrorConfiguration

// AuthOption represents a functional option for configuring an ImmutableAuthenticationConfiguration.
type AuthOption func(ImmutableAuthenticationConfiguration) ImmutableAuthenticationConfiguration

// GlobalSettingsOption represents a functional option for configuring ImmutableGlobalSettings.
type GlobalSettingsOption func(ImmutableGlobalSettings) ImmutableGlobalSettings

// App Configuration Options

// WithEnvironment adds an environment to the app configuration.
func WithEnvironment(name string, envOptions ...EnvironmentOption) AppConfigOption {
	return func(config ImmutableAppConfiguration) ImmutableAppConfiguration {
		env := NewEnvironment(envOptions...)

		return config.WithEnvironment(name, env)
	}
}

// WithGitHubEnvironment adds a GitHub environment with sensible defaults.
func WithGitHubEnvironment(name, owner, token string, _ ...MirrorOption) AppConfigOption {
	return WithEnvironment(name,
		WithSource(
			WithSourceProvider("github"),
			WithSourceDomain("github.com"),
			WithSourceOwner(owner),
			WithSourceAuth(WithAuthToken(token)),
		),
	)
}

// WithGitLabEnvironment adds a GitLab environment with sensible defaults.
func WithGitLabEnvironment(name, domain, owner, token string, _ ...MirrorOption) AppConfigOption {
	return WithEnvironment(name,
		WithSource(
			WithSourceProvider("gitlab"),
			WithSourceDomain(domain),
			WithSourceOwner(owner),
			WithSourceAuth(WithAuthToken(token)),
		),
	)
}

// WithGiteaEnvironment adds a Gitea environment with sensible defaults.
func WithGiteaEnvironment(name, domain, owner, token string, _ ...MirrorOption) AppConfigOption {
	return WithEnvironment(name,
		WithSource(
			WithSourceProvider("gitea"),
			WithSourceDomain(domain),
			WithSourceOwner(owner),
			WithSourceAuth(WithAuthToken(token)),
		),
	)
}

// WithLogLevel sets the global log level.
func WithLogLevel(level LogLevel) AppConfigOption {
	return func(config ImmutableAppConfiguration) ImmutableAppConfiguration {
		settings := config.GlobalSettings().WithLogLevel(level)

		return config.WithGlobalSettings(settings)
	}
}

// WithLogFormat sets the global log format.
func WithLogFormat(format LogFormat) AppConfigOption {
	return func(config ImmutableAppConfiguration) ImmutableAppConfiguration {
		settings := config.GlobalSettings().WithLogFormat(format)

		return config.WithGlobalSettings(settings)
	}
}

// WithTempDirectory sets the temporary directory.
func WithTempDirectory(path string) AppConfigOption {
	return func(config ImmutableAppConfiguration) ImmutableAppConfiguration {
		settings := config.GlobalSettings().WithTempDirectory(path)

		return config.WithGlobalSettings(settings)
	}
}

// WithMetrics enables metrics collection.
func WithMetrics(enabled bool, port int) AppConfigOption {
	return func(config ImmutableAppConfiguration) ImmutableAppConfiguration {
		settings := config.GlobalSettings().WithMetricsEnabled(enabled).WithMetricsPort(port)

		return config.WithGlobalSettings(settings)
	}
}

// Environment Configuration Options

// WithSource sets the source configuration for an environment.
func WithSource(sourceOptions ...SourceOption) EnvironmentOption {
	return func(env ImmutableEnvironmentConfiguration) ImmutableEnvironmentConfiguration {
		source := NewSource(sourceOptions...)

		return env.WithSource(source)
	}
}

// WithMirror adds a mirror to an environment.
func WithMirror(name string, mirrorOptions ...MirrorOption) EnvironmentOption {
	return func(env ImmutableEnvironmentConfiguration) ImmutableEnvironmentConfiguration {
		mirror := NewMirror(mirrorOptions...)

		return env.WithMirror(name, mirror)
	}
}

// WithGitHubMirror adds a GitHub mirror with sensible defaults.
func WithGitHubMirror(name, owner, token string) EnvironmentOption {
	return WithMirror(name,
		WithMirrorProvider("github"),
		WithMirrorDomain("github.com"),
		WithMirrorOwner(owner),
		WithMirrorAuth(WithAuthToken(token)),
	)
}

// WithGitLabMirror adds a GitLab mirror with sensible defaults.
func WithGitLabMirror(name, domain, owner, token string) EnvironmentOption {
	return WithMirror(name,
		WithMirrorProvider("gitlab"),
		WithMirrorDomain(domain),
		WithMirrorOwner(owner),
		WithMirrorAuth(WithAuthToken(token)),
	)
}

// WithDirectoryMirror adds a directory mirror.
func WithDirectoryMirror(name, path string) EnvironmentOption {
	return WithMirror(name,
		WithMirrorProvider("directory"),
		WithMirrorPath(path),
	)
}

// WithArchiveMirror adds an archive mirror.
func WithArchiveMirror(name, path string) EnvironmentOption {
	return WithMirror(name,
		WithMirrorProvider("archive"),
		WithMirrorPath(path),
	)
}

// WithEnvironmentEnabled sets whether the environment is enabled.
func WithEnvironmentEnabled(enabled bool) EnvironmentOption {
	return func(env ImmutableEnvironmentConfiguration) ImmutableEnvironmentConfiguration {
		return env.WithEnabled(enabled)
	}
}

// WithDryRun enables or disables dry run mode.
func WithDryRun(enabled bool) EnvironmentOption {
	return func(env ImmutableEnvironmentConfiguration) ImmutableEnvironmentConfiguration {
		options := env.Options().WithDryRun(enabled)

		return env.WithOptions(options)
	}
}

// WithParallelExecution configures parallel execution.
func WithParallelExecution(enabled bool, maxConcurrency int) EnvironmentOption {
	return func(env ImmutableEnvironmentConfiguration) ImmutableEnvironmentConfiguration {
		options := env.Options().WithParallel(enabled).WithMaxConcurrency(maxConcurrency)

		return env.WithOptions(options)
	}
}

// WithTimeout sets the timeout for operations.
func WithTimeout(timeout time.Duration) EnvironmentOption {
	return func(env ImmutableEnvironmentConfiguration) ImmutableEnvironmentConfiguration {
		options := env.Options().WithTimeout(timeout)

		return env.WithOptions(options)
	}
}

// WithRetryPolicy configures retry behavior.
func WithRetryPolicy(attempts int, delay time.Duration) EnvironmentOption {
	return func(env ImmutableEnvironmentConfiguration) ImmutableEnvironmentConfiguration {
		options := env.Options().WithRetryAttempts(attempts).WithRetryDelay(delay)

		return env.WithOptions(options)
	}
}

// Source Configuration Options

// WithSourceProvider sets the source provider type.
func WithSourceProvider(providerType string) SourceOption {
	return func(source ImmutableSourceConfiguration) ImmutableSourceConfiguration {
		return source.WithProviderType(providerType)
	}
}

// WithSourceDomain sets the source domain.
func WithSourceDomain(domain string) SourceOption {
	return func(source ImmutableSourceConfiguration) ImmutableSourceConfiguration {
		return source.WithDomain(domain)
	}
}

// WithSourceOwner sets the source owner.
func WithSourceOwner(owner string) SourceOption {
	return func(source ImmutableSourceConfiguration) ImmutableSourceConfiguration {
		return source.WithOwner(owner)
	}
}

// WithSourceAuth sets the source authentication.
func WithSourceAuth(authOptions ...AuthOption) SourceOption {
	return func(source ImmutableSourceConfiguration) ImmutableSourceConfiguration {
		auth := NewAuth(authOptions...)

		return source.WithAuthentication(auth)
	}
}

// WithIncludePatterns sets repository include patterns.
func WithIncludePatterns(patterns []string) SourceOption {
	return func(source ImmutableSourceConfiguration) ImmutableSourceConfiguration {
		repo := source.Repository().WithIncludePatterns(patterns)

		return source.WithRepository(repo)
	}
}

// WithExcludePatterns sets repository exclude patterns.
func WithExcludePatterns(patterns []string) SourceOption {
	return func(source ImmutableSourceConfiguration) ImmutableSourceConfiguration {
		repo := source.Repository().WithExcludePatterns(patterns)

		return source.WithRepository(repo)
	}
}

// WithIncludeForks configures fork inclusion.
func WithIncludeForks(include bool) SourceOption {
	return func(source ImmutableSourceConfiguration) ImmutableSourceConfiguration {
		repo := source.Repository().WithIncludeForks(include)

		return source.WithRepository(repo)
	}
}

// WithIncludeArchived configures archived repository inclusion.
func WithIncludeArchived(include bool) SourceOption {
	return func(source ImmutableSourceConfiguration) ImmutableSourceConfiguration {
		repo := source.Repository().WithIncludeArchived(include)

		return source.WithRepository(repo)
	}
}

// Mirror Configuration Options

// WithMirrorProvider sets the mirror provider type.
func WithMirrorProvider(providerType string) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		return mirror.WithProviderType(providerType)
	}
}

// WithMirrorDomain sets the mirror domain.
func WithMirrorDomain(domain string) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		return mirror.WithDomain(domain)
	}
}

// WithMirrorOwner sets the mirror owner.
func WithMirrorOwner(owner string) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		return mirror.WithOwner(owner)
	}
}

// WithMirrorPath sets the mirror path (for directory/archive providers).
func WithMirrorPath(path string) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		return mirror.WithPath(path)
	}
}

// WithMirrorAuth sets the mirror authentication.
func WithMirrorAuth(authOptions ...AuthOption) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		auth := NewAuth(authOptions...)

		return mirror.WithAuthentication(auth)
	}
}

// WithMirrorEnabled sets whether the mirror is enabled.
func WithMirrorEnabled(enabled bool) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		return mirror.WithEnabled(enabled)
	}
}

// WithCreateIfNotExists configures whether to create repositories if they don't exist.
func WithCreateIfNotExists(create bool) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		options := mirror.Options().WithCreateIfNotExists(create)

		return mirror.WithOptions(options)
	}
}

// WithUpdateDescription configures whether to update repository descriptions.
func WithUpdateDescription(update bool) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		options := mirror.Options().WithUpdateDescription(update)

		return mirror.WithOptions(options)
	}
}

// WithSyncVisibility configures whether to sync repository visibility.
func WithSyncVisibility(sync bool) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		options := mirror.Options().WithSyncVisibility(sync)

		return mirror.WithOptions(options)
	}
}

// WithSyncTopics configures whether to sync repository topics.
func WithSyncTopics(sync bool) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		options := mirror.Options().WithSyncTopics(sync)

		return mirror.WithOptions(options)
	}
}

// WithBranchProtection configures whether to sync branch protection rules.
func WithBranchProtection(sync bool) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		options := mirror.Options().WithSyncBranchProtection(sync)

		return mirror.WithOptions(options)
	}
}

// WithGitLFS configures whether to enable Git LFS.
func WithGitLFS(enabled bool) MirrorOption {
	return func(mirror ImmutableMirrorConfiguration) ImmutableMirrorConfiguration {
		options := mirror.Options().WithEnableLFS(enabled)

		return mirror.WithOptions(options)
	}
}

// Authentication Options

// WithAuthToken sets token-based authentication.
func WithAuthToken(token string) AuthOption {
	return func(auth ImmutableAuthenticationConfiguration) ImmutableAuthenticationConfiguration {
		return auth.WithType(AuthenticationTypeToken).WithToken(token)
	}
}

// WithAuthBasic sets basic authentication (username/password).
func WithAuthBasic(username, password string) AuthOption {
	return func(auth ImmutableAuthenticationConfiguration) ImmutableAuthenticationConfiguration {
		return auth.WithType(AuthenticationTypeBasic).WithUsername(username).WithPassword(password)
	}
}

// WithAuthSSH sets SSH key authentication.
func WithAuthSSH(keyPath, username string) AuthOption {
	return func(auth ImmutableAuthenticationConfiguration) ImmutableAuthenticationConfiguration {
		return auth.WithType(AuthenticationTypeSSH).WithSSHKeyPath(keyPath).WithUsername(username)
	}
}

// WithAuthSSHKey sets SSH key authentication with key content.
func WithAuthSSHKey(keyContent, username string) AuthOption {
	return func(auth ImmutableAuthenticationConfiguration) ImmutableAuthenticationConfiguration {
		return auth.WithType(AuthenticationTypeSSH).WithSSHKey(keyContent).WithUsername(username)
	}
}

// Convenience constructors using functional options

// NewAppConfiguration creates a new app configuration with the given options.
// This is the recommended approach for building configurations functionally.
// Example:
//
//	config := NewAppConfiguration(
//	    WithGitHubEnvironment("prod", "myorg", "token123"),
//	    WithGlobalLogLevel(LogLevelInfo),
//	    WithGlobalMetrics(true, 8080),
//	)
func NewAppConfiguration(options ...AppConfigOption) ImmutableAppConfiguration {
	config := NewAppConfigurationBuilder().Build()
	for _, option := range options {
		config = option(config)
	}

	return config
}

// ComposeAppConfigOptions combines multiple options into a single option.
// Useful for creating reusable configuration presets.
func ComposeAppConfigOptions(options ...AppConfigOption) AppConfigOption {
	return func(config ImmutableAppConfiguration) ImmutableAppConfiguration {
		for _, option := range options {
			config = option(config)
		}

		return config
	}
}

// Configuration Presets - Examples of functional composition

// ProductionDefaults provides sensible defaults for production environments.
func ProductionDefaults() AppConfigOption {
	return ComposeAppConfigOptions(
		WithLogLevel(LogLevelInfo),
		WithMetrics(true, 8080),
		WithTempDirectory("/tmp/git-provider-sync"),
	)
}

// DevelopmentDefaults provides sensible defaults for development environments.
func DevelopmentDefaults() AppConfigOption {
	return ComposeAppConfigOptions(
		WithLogLevel(LogLevelDebug),
		WithMetrics(false, 0),
		WithTempDirectory("/tmp/git-provider-sync-dev"),
	)
}

// GitHubToGitLabMirror creates a common GitHub -> GitLab mirror configuration.
func GitHubToGitLabMirror(
	envName,
	githubOwner, githubToken,
	gitlabDomain, gitlabOwner, gitlabToken string,
) AppConfigOption {
	return WithEnvironment(envName,
		WithSource(
			WithSourceProvider("github"),
			WithSourceDomain("github.com"),
			WithSourceOwner(githubOwner),
			WithSourceAuth(WithAuthToken(githubToken)),
		),
		WithMirror("gitlab-mirror",
			WithMirrorProvider("gitlab"),
			WithMirrorDomain(gitlabDomain),
			WithMirrorOwner(gitlabOwner),
			WithMirrorAuth(WithAuthToken(gitlabToken)),
			WithMirrorEnabled(true),
		),
	)
}

// NewEnvironment creates a new environment configuration with the given options.
func NewEnvironment(options ...EnvironmentOption) ImmutableEnvironmentConfiguration {
	env := ImmutableEnvironmentConfiguration{
		enabled: true, // Default to enabled
		options: ImmutableEnvironmentOptions{
			dryRun:         false,
			parallel:       true,
			maxConcurrency: 5,
			timeout:        30 * time.Second,
			retryAttempts:  3,
			retryDelay:     time.Second,
			logLevel:       LogLevelInfo,
		},
		mirrors: make(map[string]ImmutableMirrorConfiguration),
	}

	for _, option := range options {
		env = option(env)
	}

	return env
}

// NewSource creates a new source configuration with the given options.
func NewSource(options ...SourceOption) ImmutableSourceConfiguration {
	source := ImmutableSourceConfiguration{
		repository: ImmutableRepositoryConfiguration{
			includeForks:    false, // Default to exclude forks
			includeArchived: false, // Default to exclude archived
			includePrivate:  true,  // Default to include private
		},
		rateLimit: ImmutableRateLimitConfiguration{
			requestsPerHour: 1000,
			burstLimit:      10,
			backoffStrategy: BackoffStrategyExponential,
		},
	}

	for _, option := range options {
		source = option(source)
	}

	return source
}

// NewMirror creates a new mirror configuration with the given options.
func NewMirror(options ...MirrorOption) ImmutableMirrorConfiguration {
	mirror := ImmutableMirrorConfiguration{
		enabled: true, // Default to enabled
		options: ImmutableMirrorOptionsConfiguration{
			createIfNotExists:    true,  // Default to create if not exists
			updateDescription:    true,  // Default to sync descriptions
			syncVisibility:       true,  // Default to sync visibility
			syncTopics:           true,  // Default to sync topics
			syncDefaultBranch:    true,  // Default to sync default branch
			syncBranchProtection: false, // Default to not sync branch protection (security)
			preservePullRequests: true,  // Default to preserve PRs
			preserveIssues:       true,  // Default to preserve issues
			enableLFS:            false, // Default to disable LFS
		},
	}

	for _, option := range options {
		mirror = option(mirror)
	}

	return mirror
}

// NewAuth creates a new authentication configuration with the given options.
func NewAuth(options ...AuthOption) ImmutableAuthenticationConfiguration {
	auth := ImmutableAuthenticationConfiguration{
		authType: AuthenticationTypeNone, // Default to no auth
	}

	for _, option := range options {
		auth = option(auth)
	}

	return auth
}

// Common configuration presets

// DefaultGitHubConfig creates a default GitHub configuration.
func DefaultGitHubConfig(owner, token string) ImmutableAppConfiguration {
	return NewAppConfiguration(
		WithGitHubEnvironment("github", owner, token),
		WithLogLevel(LogLevelInfo),
		WithLogFormat(LogFormatJSON),
	)
}

// DefaultGitLabConfig creates a default GitLab configuration.
func DefaultGitLabConfig(domain, owner, token string) ImmutableAppConfiguration {
	return NewAppConfiguration(
		WithGitLabEnvironment("gitlab", domain, owner, token),
		WithLogLevel(LogLevelInfo),
		WithLogFormat(LogFormatJSON),
	)
}

// DevelopmentConfig creates a development configuration with verbose logging.
func DevelopmentConfig() ImmutableAppConfiguration {
	return NewAppConfiguration(
		WithLogLevel(LogLevelDebug),
		WithLogFormat(LogFormatConsole),
		WithTempDirectory("/tmp/git-provider-sync-dev"),
	)
}

// ProductionConfig creates a production configuration with minimal logging.
func ProductionConfig() ImmutableAppConfiguration {
	return NewAppConfiguration(
		WithLogLevel(LogLevelInfo),
		WithLogFormat(LogFormatJSON),
		WithMetrics(true, 8080),
	)
}
