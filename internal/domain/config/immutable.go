// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides immutable configuration utilities.
package config

import (
	"time"
)

// ImmutableAppConfiguration represents immutable application configuration.
type ImmutableAppConfiguration struct {
	environments   map[string]ImmutableEnvironmentConfiguration
	globalSettings ImmutableGlobalSettings
	metadata       ImmutableConfigurationMetadata
}

// ImmutableEnvironmentConfiguration represents immutable environment configuration.
type ImmutableEnvironmentConfiguration struct {
	name    string
	source  ImmutableSourceConfiguration
	mirrors map[string]ImmutableMirrorConfiguration
	options ImmutableEnvironmentOptions
	enabled bool
}

// ImmutableSourceConfiguration represents immutable source provider configuration.
type ImmutableSourceConfiguration struct {
	providerType   string
	domain         string
	owner          string
	authentication ImmutableAuthenticationConfiguration
	repository     ImmutableRepositoryConfiguration
	filtering      ImmutableFilterConfiguration
	rateLimit      ImmutableRateLimitConfiguration
}

// ImmutableMirrorConfiguration represents immutable mirror target configuration.
type ImmutableMirrorConfiguration struct {
	name           string
	providerType   string
	domain         string
	owner          string
	path           string
	authentication ImmutableAuthenticationConfiguration
	options        ImmutableMirrorOptionsConfiguration
	enabled        bool
}

// ImmutableAuthenticationConfiguration represents immutable authentication settings.
type ImmutableAuthenticationConfiguration struct {
	authType   AuthenticationType
	token      string
	username   string
	password   string
	sshKeyPath string
	sshKey     string
	passphrase string
}

// ImmutableRepositoryConfiguration represents immutable repository-specific settings.
type ImmutableRepositoryConfiguration struct {
	includePatterns []string
	excludePatterns []string
	includeForks    bool
	includeArchived bool
	includePrivate  bool
	defaultBranch   string
	topics          []string
}

// ImmutableFilterConfiguration represents immutable filtering options.
type ImmutableFilterConfiguration struct {
}

// ImmutableRateLimitConfiguration represents immutable rate limiting settings.
type ImmutableRateLimitConfiguration struct {
	requestsPerHour int
	burstLimit      int
	backoffStrategy BackoffStrategy
}

// ImmutableMirrorOptionsConfiguration represents immutable mirror-specific options.
type ImmutableMirrorOptionsConfiguration struct {
	useGitBinary         bool
	createIfNotExists    bool
	updateDescription    bool
	syncVisibility       bool
	syncTopics           bool
	syncDefaultBranch    bool
	syncBranchProtection bool
	preservePullRequests bool
	preserveIssues       bool
	enableLFS            bool
}

// ImmutableEnvironmentOptions represents immutable environment-specific options.
type ImmutableEnvironmentOptions struct {
	dryRun            bool
	parallel          bool
	maxConcurrency    int
	timeout           time.Duration
	retryAttempts     int
	retryDelay        time.Duration
	progressReporting bool
	logLevel          LogLevel
}

// ImmutableGlobalSettings represents immutable global application settings.
type ImmutableGlobalSettings struct {
	logLevel        LogLevel
	logFormat       LogFormat
	logFile         string
	tempDirectory   string
	cacheDirectory  string
	maxCacheSize    int64
	cacheTTL        time.Duration
	metricsEnabled  bool
	metricsPort     int
	healthCheckPort int
}

// ImmutableConfigurationMetadata contains immutable metadata about the configuration.
type ImmutableConfigurationMetadata struct {
	loadTime time.Time
}

// Supporting enums (these remain the same as they're already immutable).

// AuthenticationType represents the type of authentication used for git providers.
type AuthenticationType string

// BackoffStrategy represents retry backoff strategies.
type BackoffStrategy string

// LogLevel represents logging levels.
type LogLevel string

// LogFormat represents log output formats.
type LogFormat string

const (
	// AuthenticationTypeNone represents no authentication.
	AuthenticationTypeNone AuthenticationType = "none"
	// AuthenticationTypeToken represents token-based authentication.
	AuthenticationTypeToken AuthenticationType = "token"
	// AuthenticationTypeBasic represents basic authentication.
	AuthenticationTypeBasic AuthenticationType = "basic"
	// AuthenticationTypeSSH represents SSH key authentication.
	AuthenticationTypeSSH AuthenticationType = "ssh"
	// AuthenticationTypeOAuth represents OAuth authentication.
	AuthenticationTypeOAuth AuthenticationType = "oauth"
)

const (
	// BackoffStrategyLinear represents linear backoff strategy.
	BackoffStrategyLinear BackoffStrategy = "linear"
	// BackoffStrategyExponential represents exponential backoff strategy.
	BackoffStrategyExponential BackoffStrategy = "exponential"
	// BackoffStrategyFixed represents fixed backoff strategy.
	BackoffStrategyFixed BackoffStrategy = "fixed"
)

const (
	// LogLevelTrace represents trace log level.
	LogLevelTrace LogLevel = "trace"
	// LogLevelDebug represents debug log level.
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo represents info log level.
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn represents warn log level.
	LogLevelWarn LogLevel = "warn"
	// LogLevelError represents error log level.
	LogLevelError LogLevel = "error"
	// LogLevelFatal represents fatal log level.
	LogLevelFatal LogLevel = "fatal"
)

const (
	// LogFormatJSON represents JSON log format.
	LogFormatJSON LogFormat = "json"
	// LogFormatText represents text log format.
	LogFormatText LogFormat = "text"
	// LogFormatConsole represents console log format.
	LogFormatConsole LogFormat = "console"
)

// ConfigurationSource represents a source of configuration.
type ConfigurationSource struct {
	Type     SourceType
	Location string
	Priority int
	Required bool
	Format   ConfigurationFormat
}

// SourceType represents configuration source types.
type SourceType string

// ConfigurationFormat represents configuration file formats.
type ConfigurationFormat string

const (
	// SourceTypeFile represents file source type.
	SourceTypeFile SourceType = "file"
	// SourceTypeEnvironment represents environment source type.
	SourceTypeEnvironment SourceType = "environment"
	// SourceTypeEtcd represents etcd source type.
	SourceTypeEtcd SourceType = "etcd"
	// SourceTypeConsul represents consul source type.
	SourceTypeConsul SourceType = "consul"
	// SourceTypeVault represents vault source type.
	SourceTypeVault SourceType = "vault"
	// SourceTypeHTTP represents HTTP source type.
	SourceTypeHTTP SourceType = "http"
	// SourceTypeDefaults represents defaults source type.
	SourceTypeDefaults SourceType = "defaults"
)

const (
	// ConfigurationFormatYAML represents YAML configuration format.
	ConfigurationFormatYAML ConfigurationFormat = "yaml"
	// ConfigurationFormatJSON represents JSON configuration format.
	ConfigurationFormatJSON ConfigurationFormat = "json"
	// ConfigurationFormatTOML represents TOML configuration format.
	ConfigurationFormatTOML ConfigurationFormat = "toml"
	// ConfigurationFormatINI represents INI configuration format.
	ConfigurationFormatINI ConfigurationFormat = "ini"
	// ConfigurationFormatENV represents ENV configuration format.
	ConfigurationFormatENV ConfigurationFormat = "env"
)

// Accessor methods for ImmutableAppConfiguration

// Environments returns a copy of the environments map.
func (c ImmutableAppConfiguration) Environments() map[string]ImmutableEnvironmentConfiguration {
	result := make(map[string]ImmutableEnvironmentConfiguration, len(c.environments))
	for k, v := range c.environments {
		result[k] = v
	}

	return result
}

// GlobalSettings returns the global settings.
func (c ImmutableAppConfiguration) GlobalSettings() ImmutableGlobalSettings {
	return c.globalSettings
}

// Metadata returns the configuration metadata.
func (c ImmutableAppConfiguration) Metadata() ImmutableConfigurationMetadata {
	return c.metadata
}

// GetEnvironment returns a specific environment configuration.
func (c ImmutableAppConfiguration) GetEnvironment(name string) (ImmutableEnvironmentConfiguration, bool) {
	env, exists := c.environments[name]

	return env, exists
}

// Functional update methods for ImmutableAppConfiguration

// WithEnvironment returns a new configuration with the given environment added/updated.
func (c ImmutableAppConfiguration) WithEnvironment(name string, env ImmutableEnvironmentConfiguration) ImmutableAppConfiguration {
	newEnvs := make(map[string]ImmutableEnvironmentConfiguration, len(c.environments)+1)
	for k, v := range c.environments {
		newEnvs[k] = v
	}

	newEnvs[name] = env

	return ImmutableAppConfiguration{
		environments:   newEnvs,
		globalSettings: c.globalSettings,
		metadata:       c.metadata,
	}
}

// WithoutEnvironment returns a new configuration with the given environment removed.
func (c ImmutableAppConfiguration) WithoutEnvironment(name string) ImmutableAppConfiguration {
	newEnvs := make(map[string]ImmutableEnvironmentConfiguration, len(c.environments))

	for k, v := range c.environments {
		if k != name {
			newEnvs[k] = v
		}
	}

	return ImmutableAppConfiguration{
		environments:   newEnvs,
		globalSettings: c.globalSettings,
		metadata:       c.metadata,
	}
}

// WithGlobalSettings returns a new configuration with updated global settings.
func (c ImmutableAppConfiguration) WithGlobalSettings(settings ImmutableGlobalSettings) ImmutableAppConfiguration {
	return ImmutableAppConfiguration{
		environments:   c.environments,
		globalSettings: settings,
		metadata:       c.metadata,
	}
}

// WithMetadata returns a new configuration with updated metadata.
func (c ImmutableAppConfiguration) WithMetadata(metadata ImmutableConfigurationMetadata) ImmutableAppConfiguration {
	return ImmutableAppConfiguration{
		environments:   c.environments,
		globalSettings: c.globalSettings,
		metadata:       metadata,
	}
}

// Accessor methods for ImmutableEnvironmentConfiguration

// Name returns the environment name.
func (e ImmutableEnvironmentConfiguration) Name() string {
	return e.name
}

// Source returns the source configuration.
func (e ImmutableEnvironmentConfiguration) Source() ImmutableSourceConfiguration {
	return e.source
}

// Mirrors returns a copy of the mirrors map.
func (e ImmutableEnvironmentConfiguration) Mirrors() map[string]ImmutableMirrorConfiguration {
	result := make(map[string]ImmutableMirrorConfiguration, len(e.mirrors))
	for k, v := range e.mirrors {
		result[k] = v
	}

	return result
}

// Options returns the environment options.
func (e ImmutableEnvironmentConfiguration) Options() ImmutableEnvironmentOptions {
	return e.options
}

// Enabled returns whether the environment is enabled.
func (e ImmutableEnvironmentConfiguration) Enabled() bool {
	return e.enabled
}

// Functional update methods for ImmutableEnvironmentConfiguration

// WithName returns a new environment configuration with updated name.
func (e ImmutableEnvironmentConfiguration) WithName(name string) ImmutableEnvironmentConfiguration {
	return ImmutableEnvironmentConfiguration{
		name:    name,
		source:  e.source,
		mirrors: e.mirrors,
		options: e.options,
		enabled: e.enabled,
	}
}

// WithSource returns a new environment configuration with updated source.
func (e ImmutableEnvironmentConfiguration) WithSource(source ImmutableSourceConfiguration) ImmutableEnvironmentConfiguration {
	return ImmutableEnvironmentConfiguration{
		name:    e.name,
		source:  source,
		mirrors: e.mirrors,
		options: e.options,
		enabled: e.enabled,
	}
}

// WithMirror returns a new environment configuration with the given mirror added/updated.
func (e ImmutableEnvironmentConfiguration) WithMirror(name string, mirror ImmutableMirrorConfiguration) ImmutableEnvironmentConfiguration {
	newMirrors := make(map[string]ImmutableMirrorConfiguration, len(e.mirrors)+1)
	for k, v := range e.mirrors {
		newMirrors[k] = v
	}

	newMirrors[name] = mirror

	return ImmutableEnvironmentConfiguration{
		name:    e.name,
		source:  e.source,
		mirrors: newMirrors,
		options: e.options,
		enabled: e.enabled,
	}
}

// WithoutMirror returns a new environment configuration with the given mirror removed.
func (e ImmutableEnvironmentConfiguration) WithoutMirror(name string) ImmutableEnvironmentConfiguration {
	newMirrors := make(map[string]ImmutableMirrorConfiguration, len(e.mirrors))

	for k, v := range e.mirrors {
		if k != name {
			newMirrors[k] = v
		}
	}

	return ImmutableEnvironmentConfiguration{
		name:    e.name,
		source:  e.source,
		mirrors: newMirrors,
		options: e.options,
		enabled: e.enabled,
	}
}

// WithOptions returns a new environment configuration with updated options.
func (e ImmutableEnvironmentConfiguration) WithOptions(options ImmutableEnvironmentOptions) ImmutableEnvironmentConfiguration {
	return ImmutableEnvironmentConfiguration{
		name:    e.name,
		source:  e.source,
		mirrors: e.mirrors,
		options: options,
		enabled: e.enabled,
	}
}

// WithEnabled returns a new environment configuration with updated enabled status.
func (e ImmutableEnvironmentConfiguration) WithEnabled(enabled bool) ImmutableEnvironmentConfiguration {
	return ImmutableEnvironmentConfiguration{
		name:    e.name,
		source:  e.source,
		mirrors: e.mirrors,
		options: e.options,
		enabled: enabled,
	}
}

// Accessor methods for ImmutableSourceConfiguration

// ProviderType returns the provider type.
func (s ImmutableSourceConfiguration) ProviderType() string {
	return s.providerType
}

// Domain returns the domain.
func (s ImmutableSourceConfiguration) Domain() string {
	return s.domain
}

// Owner returns the owner.
func (s ImmutableSourceConfiguration) Owner() string {
	return s.owner
}

// Authentication returns the authentication configuration.
func (s ImmutableSourceConfiguration) Authentication() ImmutableAuthenticationConfiguration {
	return s.authentication
}

// Repository returns the repository configuration.
func (s ImmutableSourceConfiguration) Repository() ImmutableRepositoryConfiguration {
	return s.repository
}

// Filtering returns the filtering configuration.
func (s ImmutableSourceConfiguration) Filtering() ImmutableFilterConfiguration {
	return s.filtering
}

// RateLimit returns the rate limit configuration.
func (s ImmutableSourceConfiguration) RateLimit() ImmutableRateLimitConfiguration {
	return s.rateLimit
}

// Functional update methods for ImmutableSourceConfiguration

// WithProviderType returns a new source configuration with updated provider type.
func (s ImmutableSourceConfiguration) WithProviderType(providerType string) ImmutableSourceConfiguration {
	return ImmutableSourceConfiguration{
		providerType:   providerType,
		domain:         s.domain,
		owner:          s.owner,
		authentication: s.authentication,
		repository:     s.repository,
		filtering:      s.filtering,
		rateLimit:      s.rateLimit,
	}
}

// WithDomain returns a new source configuration with updated domain.
func (s ImmutableSourceConfiguration) WithDomain(domain string) ImmutableSourceConfiguration {
	return ImmutableSourceConfiguration{
		providerType:   s.providerType,
		domain:         domain,
		owner:          s.owner,
		authentication: s.authentication,
		repository:     s.repository,
		filtering:      s.filtering,
		rateLimit:      s.rateLimit,
	}
}

// WithOwner returns a new source configuration with updated owner.
func (s ImmutableSourceConfiguration) WithOwner(owner string) ImmutableSourceConfiguration {
	return ImmutableSourceConfiguration{
		providerType:   s.providerType,
		domain:         s.domain,
		owner:          owner,
		authentication: s.authentication,
		repository:     s.repository,
		filtering:      s.filtering,
		rateLimit:      s.rateLimit,
	}
}

// WithAuthentication returns a new source configuration with updated authentication.
func (s ImmutableSourceConfiguration) WithAuthentication(auth ImmutableAuthenticationConfiguration) ImmutableSourceConfiguration {
	return ImmutableSourceConfiguration{
		providerType:   s.providerType,
		domain:         s.domain,
		owner:          s.owner,
		authentication: auth,
		repository:     s.repository,
		filtering:      s.filtering,
		rateLimit:      s.rateLimit,
	}
}

// Accessor methods for ImmutableAuthenticationConfiguration

// Type returns the authentication type.
func (a ImmutableAuthenticationConfiguration) Type() AuthenticationType {
	return a.authType
}

// Token returns the token.
func (a ImmutableAuthenticationConfiguration) Token() string {
	return a.token
}

// Username returns the username.
func (a ImmutableAuthenticationConfiguration) Username() string {
	return a.username
}

// Password returns the password.
func (a ImmutableAuthenticationConfiguration) Password() string {
	return a.password
}

// SSHKeyPath returns the SSH key path.
func (a ImmutableAuthenticationConfiguration) SSHKeyPath() string {
	return a.sshKeyPath
}

// SSHKey returns the SSH key.
func (a ImmutableAuthenticationConfiguration) SSHKey() string {
	return a.sshKey
}

// Passphrase returns the passphrase.
func (a ImmutableAuthenticationConfiguration) Passphrase() string {
	return a.passphrase
}

// Functional update methods for ImmutableAuthenticationConfiguration

// WithToken returns a new authentication configuration with updated token.
func (a ImmutableAuthenticationConfiguration) WithToken(token string) ImmutableAuthenticationConfiguration {
	return ImmutableAuthenticationConfiguration{
		authType:   a.authType,
		token:      token,
		username:   a.username,
		password:   a.password,
		sshKeyPath: a.sshKeyPath,
		sshKey:     a.sshKey,
		passphrase: a.passphrase,
	}
}

// WithUsername returns a new authentication configuration with updated username.
func (a ImmutableAuthenticationConfiguration) WithUsername(username string) ImmutableAuthenticationConfiguration {
	return ImmutableAuthenticationConfiguration{
		authType:   a.authType,
		token:      a.token,
		username:   username,
		password:   a.password,
		sshKeyPath: a.sshKeyPath,
		sshKey:     a.sshKey,
		passphrase: a.passphrase,
	}
}

// Builder pattern for constructing immutable configurations

// AppConfigurationBuilder provides a fluent interface for building immutable configurations.
type AppConfigurationBuilder struct {
	config ImmutableAppConfiguration
}

// NewAppConfigurationBuilder creates a new configuration builder.
func NewAppConfigurationBuilder() *AppConfigurationBuilder {
	return &AppConfigurationBuilder{
		config: ImmutableAppConfiguration{
			environments: make(map[string]ImmutableEnvironmentConfiguration),
			globalSettings: ImmutableGlobalSettings{
				logLevel:  LogLevelInfo,
				logFormat: LogFormatJSON,
			},
			metadata: ImmutableConfigurationMetadata{
				loadTime: time.Now(),
			},
		},
	}
}

// WithEnvironment adds an environment to the configuration.
func (b *AppConfigurationBuilder) WithEnvironment(name string, env ImmutableEnvironmentConfiguration) *AppConfigurationBuilder {
	b.config = b.config.WithEnvironment(name, env)

	return b
}

// WithGlobalSettings sets the global settings.
func (b *AppConfigurationBuilder) WithGlobalSettings(settings ImmutableGlobalSettings) *AppConfigurationBuilder {
	b.config = b.config.WithGlobalSettings(settings)

	return b
}

// Build returns the built configuration.
func (b *AppConfigurationBuilder) Build() ImmutableAppConfiguration {
	return b.config
}

// Additional accessor and update methods needed by functional options

// ImmutableGlobalSettings accessor and update methods

// LogLevel returns the log level.
func (g ImmutableGlobalSettings) LogLevel() LogLevel {
	return g.logLevel
}

// LogFormat returns the log format.
func (g ImmutableGlobalSettings) LogFormat() LogFormat {
	return g.logFormat
}

// LogFile returns the log file path.
func (g ImmutableGlobalSettings) LogFile() string {
	return g.logFile
}

// TempDirectory returns the temporary directory.
func (g ImmutableGlobalSettings) TempDirectory() string {
	return g.tempDirectory
}

// MetricsEnabled returns whether metrics are enabled.
func (g ImmutableGlobalSettings) MetricsEnabled() bool {
	return g.metricsEnabled
}

// MetricsPort returns the metrics port.
func (g ImmutableGlobalSettings) MetricsPort() int {
	return g.metricsPort
}

// WithLogLevel returns a new global settings with updated log level.
func (g ImmutableGlobalSettings) WithLogLevel(level LogLevel) ImmutableGlobalSettings {
	return ImmutableGlobalSettings{
		logLevel:        level,
		logFormat:       g.logFormat,
		logFile:         g.logFile,
		tempDirectory:   g.tempDirectory,
		cacheDirectory:  g.cacheDirectory,
		maxCacheSize:    g.maxCacheSize,
		cacheTTL:        g.cacheTTL,
		metricsEnabled:  g.metricsEnabled,
		metricsPort:     g.metricsPort,
		healthCheckPort: g.healthCheckPort,
	}
}

// WithLogFormat returns a new global settings with updated log format.
func (g ImmutableGlobalSettings) WithLogFormat(format LogFormat) ImmutableGlobalSettings {
	return ImmutableGlobalSettings{
		logLevel:        g.logLevel,
		logFormat:       format,
		logFile:         g.logFile,
		tempDirectory:   g.tempDirectory,
		cacheDirectory:  g.cacheDirectory,
		maxCacheSize:    g.maxCacheSize,
		cacheTTL:        g.cacheTTL,
		metricsEnabled:  g.metricsEnabled,
		metricsPort:     g.metricsPort,
		healthCheckPort: g.healthCheckPort,
	}
}

// WithTempDirectory returns a new global settings with updated temp directory.
func (g ImmutableGlobalSettings) WithTempDirectory(path string) ImmutableGlobalSettings {
	return ImmutableGlobalSettings{
		logLevel:        g.logLevel,
		logFormat:       g.logFormat,
		logFile:         g.logFile,
		tempDirectory:   path,
		cacheDirectory:  g.cacheDirectory,
		maxCacheSize:    g.maxCacheSize,
		cacheTTL:        g.cacheTTL,
		metricsEnabled:  g.metricsEnabled,
		metricsPort:     g.metricsPort,
		healthCheckPort: g.healthCheckPort,
	}
}

// WithMetricsEnabled returns a new global settings with updated metrics enabled.
func (g ImmutableGlobalSettings) WithMetricsEnabled(enabled bool) ImmutableGlobalSettings {
	return ImmutableGlobalSettings{
		logLevel:        g.logLevel,
		logFormat:       g.logFormat,
		logFile:         g.logFile,
		tempDirectory:   g.tempDirectory,
		cacheDirectory:  g.cacheDirectory,
		maxCacheSize:    g.maxCacheSize,
		cacheTTL:        g.cacheTTL,
		metricsEnabled:  enabled,
		metricsPort:     g.metricsPort,
		healthCheckPort: g.healthCheckPort,
	}
}

// WithMetricsPort returns a new global settings with updated metrics port.
func (g ImmutableGlobalSettings) WithMetricsPort(port int) ImmutableGlobalSettings {
	return ImmutableGlobalSettings{
		logLevel:        g.logLevel,
		logFormat:       g.logFormat,
		logFile:         g.logFile,
		tempDirectory:   g.tempDirectory,
		cacheDirectory:  g.cacheDirectory,
		maxCacheSize:    g.maxCacheSize,
		cacheTTL:        g.cacheTTL,
		metricsEnabled:  g.metricsEnabled,
		metricsPort:     port,
		healthCheckPort: g.healthCheckPort,
	}
}

// ImmutableMirrorConfiguration accessor and update methods

// Name returns the mirror name.
func (m ImmutableMirrorConfiguration) Name() string {
	return m.name
}

// ProviderType returns the provider type.
func (m ImmutableMirrorConfiguration) ProviderType() string {
	return m.providerType
}

// Domain returns the domain.
func (m ImmutableMirrorConfiguration) Domain() string {
	return m.domain
}

// Owner returns the owner.
func (m ImmutableMirrorConfiguration) Owner() string {
	return m.owner
}

// Path returns the path.
func (m ImmutableMirrorConfiguration) Path() string {
	return m.path
}

// Authentication returns the authentication configuration.
func (m ImmutableMirrorConfiguration) Authentication() ImmutableAuthenticationConfiguration {
	return m.authentication
}

// Options returns the mirror options.
func (m ImmutableMirrorConfiguration) Options() ImmutableMirrorOptionsConfiguration {
	return m.options
}

// Enabled returns whether the mirror is enabled.
func (m ImmutableMirrorConfiguration) Enabled() bool {
	return m.enabled
}

// WithProviderType returns a new mirror configuration with updated provider type.
func (m ImmutableMirrorConfiguration) WithProviderType(providerType string) ImmutableMirrorConfiguration {
	return ImmutableMirrorConfiguration{
		name:           m.name,
		providerType:   providerType,
		domain:         m.domain,
		owner:          m.owner,
		path:           m.path,
		authentication: m.authentication,
		options:        m.options,
		enabled:        m.enabled,
	}
}

// WithDomain returns a new mirror configuration with updated domain.
func (m ImmutableMirrorConfiguration) WithDomain(domain string) ImmutableMirrorConfiguration {
	return ImmutableMirrorConfiguration{
		name:           m.name,
		providerType:   m.providerType,
		domain:         domain,
		owner:          m.owner,
		path:           m.path,
		authentication: m.authentication,
		options:        m.options,
		enabled:        m.enabled,
	}
}

// WithOwner returns a new mirror configuration with updated owner.
func (m ImmutableMirrorConfiguration) WithOwner(owner string) ImmutableMirrorConfiguration {
	return ImmutableMirrorConfiguration{
		name:           m.name,
		providerType:   m.providerType,
		domain:         m.domain,
		owner:          owner,
		path:           m.path,
		authentication: m.authentication,
		options:        m.options,
		enabled:        m.enabled,
	}
}

// WithPath returns a new mirror configuration with updated path.
func (m ImmutableMirrorConfiguration) WithPath(path string) ImmutableMirrorConfiguration {
	return ImmutableMirrorConfiguration{
		name:           m.name,
		providerType:   m.providerType,
		domain:         m.domain,
		owner:          m.owner,
		path:           path,
		authentication: m.authentication,
		options:        m.options,
		enabled:        m.enabled,
	}
}

// WithAuthentication returns a new mirror configuration with updated authentication.
func (m ImmutableMirrorConfiguration) WithAuthentication(auth ImmutableAuthenticationConfiguration) ImmutableMirrorConfiguration {
	return ImmutableMirrorConfiguration{
		name:           m.name,
		providerType:   m.providerType,
		domain:         m.domain,
		owner:          m.owner,
		path:           m.path,
		authentication: auth,
		options:        m.options,
		enabled:        m.enabled,
	}
}

// WithOptions returns a new mirror configuration with updated options.
func (m ImmutableMirrorConfiguration) WithOptions(options ImmutableMirrorOptionsConfiguration) ImmutableMirrorConfiguration {
	return ImmutableMirrorConfiguration{
		name:           m.name,
		providerType:   m.providerType,
		domain:         m.domain,
		owner:          m.owner,
		path:           m.path,
		authentication: m.authentication,
		options:        options,
		enabled:        m.enabled,
	}
}

// WithEnabled returns a new mirror configuration with updated enabled status.
func (m ImmutableMirrorConfiguration) WithEnabled(enabled bool) ImmutableMirrorConfiguration {
	return ImmutableMirrorConfiguration{
		name:           m.name,
		providerType:   m.providerType,
		domain:         m.domain,
		owner:          m.owner,
		path:           m.path,
		authentication: m.authentication,
		options:        m.options,
		enabled:        enabled,
	}
}

// ImmutableMirrorOptionsConfiguration methods

// WithCreateIfNotExists returns new options with updated createIfNotExists.
func (o ImmutableMirrorOptionsConfiguration) WithCreateIfNotExists(create bool) ImmutableMirrorOptionsConfiguration {
	return ImmutableMirrorOptionsConfiguration{
		useGitBinary:         o.useGitBinary,
		createIfNotExists:    create,
		updateDescription:    o.updateDescription,
		syncVisibility:       o.syncVisibility,
		syncTopics:           o.syncTopics,
		syncDefaultBranch:    o.syncDefaultBranch,
		syncBranchProtection: o.syncBranchProtection,
		preservePullRequests: o.preservePullRequests,
		preserveIssues:       o.preserveIssues,
		enableLFS:            o.enableLFS,
	}
}

// WithUpdateDescription returns new options with updated updateDescription.
func (o ImmutableMirrorOptionsConfiguration) WithUpdateDescription(update bool) ImmutableMirrorOptionsConfiguration {
	return ImmutableMirrorOptionsConfiguration{
		useGitBinary:         o.useGitBinary,
		createIfNotExists:    o.createIfNotExists,
		updateDescription:    update,
		syncVisibility:       o.syncVisibility,
		syncTopics:           o.syncTopics,
		syncDefaultBranch:    o.syncDefaultBranch,
		syncBranchProtection: o.syncBranchProtection,
		preservePullRequests: o.preservePullRequests,
		preserveIssues:       o.preserveIssues,
		enableLFS:            o.enableLFS,
	}
}

// WithSyncVisibility returns new options with updated syncVisibility.
func (o ImmutableMirrorOptionsConfiguration) WithSyncVisibility(sync bool) ImmutableMirrorOptionsConfiguration {
	return ImmutableMirrorOptionsConfiguration{
		useGitBinary:         o.useGitBinary,
		createIfNotExists:    o.createIfNotExists,
		updateDescription:    o.updateDescription,
		syncVisibility:       sync,
		syncTopics:           o.syncTopics,
		syncDefaultBranch:    o.syncDefaultBranch,
		syncBranchProtection: o.syncBranchProtection,
		preservePullRequests: o.preservePullRequests,
		preserveIssues:       o.preserveIssues,
		enableLFS:            o.enableLFS,
	}
}

// WithSyncTopics returns new options with updated syncTopics.
func (o ImmutableMirrorOptionsConfiguration) WithSyncTopics(sync bool) ImmutableMirrorOptionsConfiguration {
	return ImmutableMirrorOptionsConfiguration{
		useGitBinary:         o.useGitBinary,
		createIfNotExists:    o.createIfNotExists,
		updateDescription:    o.updateDescription,
		syncVisibility:       o.syncVisibility,
		syncTopics:           sync,
		syncDefaultBranch:    o.syncDefaultBranch,
		syncBranchProtection: o.syncBranchProtection,
		preservePullRequests: o.preservePullRequests,
		preserveIssues:       o.preserveIssues,
		enableLFS:            o.enableLFS,
	}
}

// WithSyncBranchProtection returns new options with updated syncBranchProtection.
func (o ImmutableMirrorOptionsConfiguration) WithSyncBranchProtection(sync bool) ImmutableMirrorOptionsConfiguration {
	return ImmutableMirrorOptionsConfiguration{
		useGitBinary:         o.useGitBinary,
		createIfNotExists:    o.createIfNotExists,
		updateDescription:    o.updateDescription,
		syncVisibility:       o.syncVisibility,
		syncTopics:           o.syncTopics,
		syncDefaultBranch:    o.syncDefaultBranch,
		syncBranchProtection: sync,
		preservePullRequests: o.preservePullRequests,
		preserveIssues:       o.preserveIssues,
		enableLFS:            o.enableLFS,
	}
}

// WithEnableLFS returns new options with updated enableLFS.
func (o ImmutableMirrorOptionsConfiguration) WithEnableLFS(enabled bool) ImmutableMirrorOptionsConfiguration {
	return ImmutableMirrorOptionsConfiguration{
		useGitBinary:         o.useGitBinary,
		createIfNotExists:    o.createIfNotExists,
		updateDescription:    o.updateDescription,
		syncVisibility:       o.syncVisibility,
		syncTopics:           o.syncTopics,
		syncDefaultBranch:    o.syncDefaultBranch,
		syncBranchProtection: o.syncBranchProtection,
		preservePullRequests: o.preservePullRequests,
		preserveIssues:       o.preserveIssues,
		enableLFS:            enabled,
	}
}

// ImmutableAuthenticationConfiguration update methods

// WithType returns a new auth configuration with updated type.
func (a ImmutableAuthenticationConfiguration) WithType(authType AuthenticationType) ImmutableAuthenticationConfiguration {
	return ImmutableAuthenticationConfiguration{
		authType:   authType,
		token:      a.token,
		username:   a.username,
		password:   a.password,
		sshKeyPath: a.sshKeyPath,
		sshKey:     a.sshKey,
		passphrase: a.passphrase,
	}
}

// WithPassword returns a new auth configuration with updated password.
func (a ImmutableAuthenticationConfiguration) WithPassword(password string) ImmutableAuthenticationConfiguration {
	return ImmutableAuthenticationConfiguration{
		authType:   a.authType,
		token:      a.token,
		username:   a.username,
		password:   password,
		sshKeyPath: a.sshKeyPath,
		sshKey:     a.sshKey,
		passphrase: a.passphrase,
	}
}

// WithSSHKeyPath returns a new auth configuration with updated SSH key path.
func (a ImmutableAuthenticationConfiguration) WithSSHKeyPath(path string) ImmutableAuthenticationConfiguration {
	return ImmutableAuthenticationConfiguration{
		authType:   a.authType,
		token:      a.token,
		username:   a.username,
		password:   a.password,
		sshKeyPath: path,
		sshKey:     a.sshKey,
		passphrase: a.passphrase,
	}
}

// WithSSHKey returns a new auth configuration with updated SSH key.
func (a ImmutableAuthenticationConfiguration) WithSSHKey(key string) ImmutableAuthenticationConfiguration {
	return ImmutableAuthenticationConfiguration{
		authType:   a.authType,
		token:      a.token,
		username:   a.username,
		password:   a.password,
		sshKeyPath: a.sshKeyPath,
		sshKey:     key,
		passphrase: a.passphrase,
	}
}

// ImmutableRepositoryConfiguration methods

// IncludePatterns returns the include patterns.
func (r ImmutableRepositoryConfiguration) IncludePatterns() []string {
	return append([]string(nil), r.includePatterns...)
}

// ExcludePatterns returns the exclude patterns.
func (r ImmutableRepositoryConfiguration) ExcludePatterns() []string {
	return append([]string(nil), r.excludePatterns...)
}

// IncludeForks returns whether to include forks.
func (r ImmutableRepositoryConfiguration) IncludeForks() bool {
	return r.includeForks
}

// IncludeArchived returns whether to include archived repositories.
func (r ImmutableRepositoryConfiguration) IncludeArchived() bool {
	return r.includeArchived
}

// WithIncludePatterns returns new repository config with updated include patterns.
func (r ImmutableRepositoryConfiguration) WithIncludePatterns(patterns []string) ImmutableRepositoryConfiguration {
	newPatterns := make([]string, len(patterns))
	copy(newPatterns, patterns)

	return ImmutableRepositoryConfiguration{
		includePatterns: newPatterns,
		excludePatterns: r.excludePatterns,
		includeForks:    r.includeForks,
		includeArchived: r.includeArchived,
		includePrivate:  r.includePrivate,
		defaultBranch:   r.defaultBranch,
		topics:          r.topics,
	}
}

// WithExcludePatterns returns new repository config with updated exclude patterns.
func (r ImmutableRepositoryConfiguration) WithExcludePatterns(patterns []string) ImmutableRepositoryConfiguration {
	newPatterns := make([]string, len(patterns))
	copy(newPatterns, patterns)

	return ImmutableRepositoryConfiguration{
		includePatterns: r.includePatterns,
		excludePatterns: newPatterns,
		includeForks:    r.includeForks,
		includeArchived: r.includeArchived,
		includePrivate:  r.includePrivate,
		defaultBranch:   r.defaultBranch,
		topics:          r.topics,
	}
}

// WithIncludeForks returns new repository config with updated includeForks.
func (r ImmutableRepositoryConfiguration) WithIncludeForks(include bool) ImmutableRepositoryConfiguration {
	return ImmutableRepositoryConfiguration{
		includePatterns: r.includePatterns,
		excludePatterns: r.excludePatterns,
		includeForks:    include,
		includeArchived: r.includeArchived,
		includePrivate:  r.includePrivate,
		defaultBranch:   r.defaultBranch,
		topics:          r.topics,
	}
}

// WithIncludeArchived returns new repository config with updated includeArchived.
func (r ImmutableRepositoryConfiguration) WithIncludeArchived(include bool) ImmutableRepositoryConfiguration {
	return ImmutableRepositoryConfiguration{
		includePatterns: r.includePatterns,
		excludePatterns: r.excludePatterns,
		includeForks:    r.includeForks,
		includeArchived: include,
		includePrivate:  r.includePrivate,
		defaultBranch:   r.defaultBranch,
		topics:          r.topics,
	}
}

// WithRepository returns a new source configuration with updated repository config.
func (s ImmutableSourceConfiguration) WithRepository(repo ImmutableRepositoryConfiguration) ImmutableSourceConfiguration {
	return ImmutableSourceConfiguration{
		providerType:   s.providerType,
		domain:         s.domain,
		owner:          s.owner,
		authentication: s.authentication,
		repository:     repo,
		filtering:      s.filtering,
		rateLimit:      s.rateLimit,
	}
}

// ImmutableEnvironmentOptions methods

// DryRun returns whether dry run is enabled.
func (e ImmutableEnvironmentOptions) DryRun() bool {
	return e.dryRun
}

// Parallel returns whether parallel execution is enabled.
func (e ImmutableEnvironmentOptions) Parallel() bool {
	return e.parallel
}

// MaxConcurrency returns the maximum concurrency.
func (e ImmutableEnvironmentOptions) MaxConcurrency() int {
	return e.maxConcurrency
}

// Timeout returns the timeout.
func (e ImmutableEnvironmentOptions) Timeout() time.Duration {
	return e.timeout
}

// RetryAttempts returns the retry attempts.
func (e ImmutableEnvironmentOptions) RetryAttempts() int {
	return e.retryAttempts
}

// RetryDelay returns the retry delay.
func (e ImmutableEnvironmentOptions) RetryDelay() time.Duration {
	return e.retryDelay
}

// WithDryRun returns new options with updated dryRun.
func (e ImmutableEnvironmentOptions) WithDryRun(dryRun bool) ImmutableEnvironmentOptions {
	return ImmutableEnvironmentOptions{
		dryRun:            dryRun,
		parallel:          e.parallel,
		maxConcurrency:    e.maxConcurrency,
		timeout:           e.timeout,
		retryAttempts:     e.retryAttempts,
		retryDelay:        e.retryDelay,
		progressReporting: e.progressReporting,
		logLevel:          e.logLevel,
	}
}

// WithParallel returns new options with updated parallel.
func (e ImmutableEnvironmentOptions) WithParallel(parallel bool) ImmutableEnvironmentOptions {
	return ImmutableEnvironmentOptions{
		dryRun:            e.dryRun,
		parallel:          parallel,
		maxConcurrency:    e.maxConcurrency,
		timeout:           e.timeout,
		retryAttempts:     e.retryAttempts,
		retryDelay:        e.retryDelay,
		progressReporting: e.progressReporting,
		logLevel:          e.logLevel,
	}
}

// WithMaxConcurrency returns new options with updated maxConcurrency.
func (e ImmutableEnvironmentOptions) WithMaxConcurrency(maxConcurrency int) ImmutableEnvironmentOptions {
	return ImmutableEnvironmentOptions{
		dryRun:            e.dryRun,
		parallel:          e.parallel,
		maxConcurrency:    maxConcurrency,
		timeout:           e.timeout,
		retryAttempts:     e.retryAttempts,
		retryDelay:        e.retryDelay,
		progressReporting: e.progressReporting,
		logLevel:          e.logLevel,
	}
}

// WithTimeout returns new options with updated timeout.
func (e ImmutableEnvironmentOptions) WithTimeout(timeout time.Duration) ImmutableEnvironmentOptions {
	return ImmutableEnvironmentOptions{
		dryRun:            e.dryRun,
		parallel:          e.parallel,
		maxConcurrency:    e.maxConcurrency,
		timeout:           timeout,
		retryAttempts:     e.retryAttempts,
		retryDelay:        e.retryDelay,
		progressReporting: e.progressReporting,
		logLevel:          e.logLevel,
	}
}

// WithRetryAttempts returns new options with updated retryAttempts.
func (e ImmutableEnvironmentOptions) WithRetryAttempts(attempts int) ImmutableEnvironmentOptions {
	return ImmutableEnvironmentOptions{
		dryRun:            e.dryRun,
		parallel:          e.parallel,
		maxConcurrency:    e.maxConcurrency,
		timeout:           e.timeout,
		retryAttempts:     attempts,
		retryDelay:        e.retryDelay,
		progressReporting: e.progressReporting,
		logLevel:          e.logLevel,
	}
}

// WithRetryDelay returns new options with updated retryDelay.
func (e ImmutableEnvironmentOptions) WithRetryDelay(delay time.Duration) ImmutableEnvironmentOptions {
	return ImmutableEnvironmentOptions{
		dryRun:            e.dryRun,
		parallel:          e.parallel,
		maxConcurrency:    e.maxConcurrency,
		timeout:           e.timeout,
		retryAttempts:     e.retryAttempts,
		retryDelay:        delay,
		progressReporting: e.progressReporting,
		logLevel:          e.logLevel,
	}
}
