// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"context"
	"fmt"
	"time"

	"itiquette/git-provider-sync/internal/domain"
)

// Configuration defines the interface for configuration management (secondary port).
// This port is implemented by adapters that handle configuration loading from files, environment, etc.
type Configuration interface {
	// Configuration loading
	Load(ctx context.Context, source ConfigurationSource) (AppConfiguration, error)
	LoadMultiple(ctx context.Context, sources []ConfigurationSource) (AppConfiguration, error)
	Reload(ctx context.Context) (AppConfiguration, error)

	// Configuration validation
	Validate(config AppConfiguration) ([]ConfigurationError, error)
	ValidateEnvironment(env EnvironmentConfiguration) error

	// Configuration metadata
	GetSources() []ConfigurationSource
	GetLastModified() time.Time
	GetVersion() string
}

// AppConfiguration represents the complete application configuration.
type AppConfiguration struct {
	Environments   map[string]EnvironmentConfiguration
	GlobalSettings GlobalSettings
	Metadata       ConfigurationMetadata
}

// EnvironmentConfiguration represents configuration for a specific environment.
type EnvironmentConfiguration struct {
	Name    string
	Source  SourceConfiguration
	Mirrors map[string]MirrorConfiguration
	Options EnvironmentOptions
	Enabled bool
}

// SourceConfiguration represents the source provider configuration.
type SourceConfiguration struct {
	ProviderType   string
	Domain         string
	Owner          string
	Authentication AuthenticationConfiguration
	Repository     RepositoryConfiguration
	Filtering      FilterConfiguration
	RateLimit      RateLimitConfiguration
}

// MirrorConfiguration represents a mirror target configuration.
type MirrorConfiguration struct {
	Name           string
	ProviderType   string
	Domain         string
	Owner          string
	Path           string
	Authentication AuthenticationConfiguration
	Options        MirrorOptionsConfiguration
	Enabled        bool
}

// AuthenticationConfiguration represents authentication settings.
type AuthenticationConfiguration struct {
	Type       AuthenticationType
	Protocol   string // "ssh", "https", "tls" - restores main branch protocol selection
	Token      string
	Username   string
	Password   string
	SSHKeyPath string
	SSHKey     string
	Passphrase string
}

// RepositoryConfiguration represents repository-specific settings.
type RepositoryConfiguration struct {
	IncludePatterns []string
	ExcludePatterns []string
	IncludeForks    bool
	IncludeArchived bool
	IncludePrivate  bool
	DefaultBranch   string
	Topics          []string
}

// FilterConfiguration represents filtering options.
type FilterConfiguration struct {
	Languages     []string
	MinSize       int64
	MaxSize       int64
	ActiveSince   *time.Time
	InactiveSince *time.Time
	HasIssues     *bool
	HasWiki       *bool
	HasProjects   *bool
}

// RateLimitConfiguration represents rate limiting settings.
type RateLimitConfiguration struct {
	RequestsPerHour int
	BurstLimit      int
	BackoffStrategy BackoffStrategy
}

// MirrorOptionsConfiguration represents mirror-specific options.
type MirrorOptionsConfiguration struct {
	UseGitBinary         bool
	CreateIfNotExists    bool
	UpdateDescription    bool
	SyncVisibility       bool
	SyncTopics           bool
	SyncDefaultBranch    bool
	SyncBranchProtection bool
	PreservePullRequests bool
	PreserveIssues       bool
	EnableLFS            bool
}

// EnvironmentOptions represents environment-specific options.
type EnvironmentOptions struct {
	DryRun            bool
	Parallel          bool
	MaxConcurrency    int
	Timeout           time.Duration
	RetryAttempts     int
	RetryDelay        time.Duration
	ProgressReporting bool
	LogLevel          LogLevel
}

// GlobalSettings represents global application settings.
type GlobalSettings struct {
	LogLevel        LogLevel
	LogFormat       LogFormat
	LogFile         string
	TempDirectory   string
	CacheDirectory  string
	MaxCacheSize    int64
	CacheTTL        time.Duration
	MetricsEnabled  bool
	MetricsPort     int
	HealthCheckPort int
}

// ConfigurationMetadata contains metadata about the configuration.
type ConfigurationMetadata struct {
	Version     string
	LoadTime    time.Time
	Sources     []ConfigurationSource
	Environment string
	Validated   bool
	Checksum    string
}

// Enums and constants

// AuthenticationType represents the type of authentication.
type AuthenticationType string

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

// BackoffStrategy represents retry backoff strategies.
type BackoffStrategy string

const (
	// BackoffStrategyLinear represents linear backoff strategy.
	BackoffStrategyLinear BackoffStrategy = "linear"
	// BackoffStrategyExponential represents exponential backoff strategy.
	BackoffStrategyExponential BackoffStrategy = "exponential"
	// BackoffStrategyFixed represents fixed backoff strategy.
	BackoffStrategyFixed BackoffStrategy = "fixed"
)

// LogLevel represents logging levels.
type LogLevel string

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

// LogFormat represents logging formats.
type LogFormat string

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

// SourceType represents the type of configuration source.
type SourceType string

const (
	// SourceTypeFile represents file-based configuration source.
	SourceTypeFile SourceType = "file"
	// SourceTypeEnvironment represents environment variable configuration source.
	SourceTypeEnvironment SourceType = "environment"
	// SourceTypeEtcd represents etcd configuration source.
	SourceTypeEtcd SourceType = "etcd"
	// SourceTypeConsul represents consul configuration source.
	SourceTypeConsul SourceType = "consul"
	// SourceTypeVault represents vault configuration source.
	SourceTypeVault SourceType = "vault"
	// SourceTypeHTTP represents HTTP configuration source.
	SourceTypeHTTP SourceType = "http"
	// SourceTypeDefaults represents default configuration source.
	SourceTypeDefaults SourceType = "defaults"
)

// ConfigurationFormat represents the format of configuration data.
type ConfigurationFormat string

const (
	// ConfigurationFormatYAML represents YAML configuration format.
	ConfigurationFormatYAML ConfigurationFormat = "yaml"
	// ConfigurationFormatJSON represents JSON configuration format.
	ConfigurationFormatJSON ConfigurationFormat = "json"
	// ConfigurationFormatTOML represents TOML configuration format.
	ConfigurationFormatTOML ConfigurationFormat = "toml"
	// ConfigurationFormatINI represents INI configuration format.
	ConfigurationFormatINI ConfigurationFormat = "ini"
	// ConfigurationFormatENV represents environment variable configuration format.
	ConfigurationFormatENV ConfigurationFormat = "env"
)

// Supporting types

// ConfigurationError represents an error in configuration.
type ConfigurationError struct {
	Field    string
	Value    interface{}
	Err      error
	Severity ErrorSeverity
	Source   string
}

// ErrorSeverity represents the severity of a configuration error.
type ErrorSeverity string

const (
	// ErrorSeverityError represents error severity level.
	ErrorSeverityError ErrorSeverity = "error"
	// ErrorSeverityWarning represents warning severity level.
	ErrorSeverityWarning ErrorSeverity = "warning"
	// ErrorSeverityInfo represents info severity level.
	ErrorSeverityInfo ErrorSeverity = "info"
)

// ConfigurationValidator validates configuration values.
type ConfigurationValidator interface {
	ValidateEnvironment(env EnvironmentConfiguration) error
	ValidateSource(source SourceConfiguration) error
	ValidateMirror(mirror MirrorConfiguration) error
	ValidateAuthentication(auth AuthenticationConfiguration) error
	ValidateGlobal(global GlobalSettings) error
}

// ConfigurationBuilder interface removed: was unused (22 methods with zero implementations).
// The codebase uses concrete value types with functional options instead.
// See internal/domain/config/immutable.go for the actual implementation pattern.

// ConfigurationProvider provides configurations for different scenarios.
type ConfigurationProvider interface {
	// Preset configurations
	GetDefaultConfiguration() AppConfiguration
	GetDevelopmentConfiguration() AppConfiguration
	GetProductionConfiguration() AppConfiguration
	GetTestConfiguration() AppConfiguration

	// Provider-specific configurations
	GetGitHubConfiguration(owner, token string) SourceConfiguration
	GetGitLabConfiguration(owner, domain, token string) SourceConfiguration
	GetGiteaConfiguration(owner, domain, token string) SourceConfiguration

	// Validation presets
	GetValidationRules() map[string]ValidationRule
	GetSecurityPolicy() SecurityPolicy
}

// ValidationRule represents a configuration validation rule.
type ValidationRule struct {
	Field         string
	Required      bool
	Type          string
	Pattern       string
	MinValue      interface{}
	MaxValue      interface{}
	AllowedValues []interface{}
	Dependencies  []string
}

// SecurityPolicy represents security policies for configuration.
type SecurityPolicy struct {
	RequireAuthentication   bool
	AllowPlaintextPasswords bool
	RequireSSL              bool
	AllowedDomains          []string
	BlockedDomains          []string
	MaxTokenLength          int
	TokenPatterns           []string
}

// ConfigurationCache provides caching for configuration.
type ConfigurationCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	Size() int
}

// Advanced configuration interfaces

// DynamicConfiguration allows dynamic configuration updates.
type DynamicConfiguration interface {
	Configuration

	// Dynamic updates
	UpdateEnvironment(name string, env EnvironmentConfiguration) error
	UpdateMirror(envName, mirrorName string, mirror MirrorConfiguration) error
	UpdateGlobalSettings(settings GlobalSettings) error

	// Hot reload
	EnableHotReload(interval time.Duration) error
	DisableHotReload() error

	// Configuration drift detection
	DetectDrift() ([]ConfigurationDrift, error)
	ResolveDrift(drifts []ConfigurationDrift) error
}

// ConfigurationDrift represents a drift in configuration.
type ConfigurationDrift struct {
	Field          string
	Expected       interface{}
	Actual         interface{}
	Source         string
	Severity       DriftSeverity
	AutoResolvable bool
}

// DriftSeverity represents the severity of configuration drift.
type DriftSeverity string

const (
	// DriftSeverityCritical represents critical drift severity.
	DriftSeverityCritical DriftSeverity = "critical"
	// DriftSeverityMajor represents major drift severity.
	DriftSeverityMajor DriftSeverity = "major"
	// DriftSeverityMinor represents minor drift severity.
	DriftSeverityMinor DriftSeverity = "minor"
	// DriftSeverityInfo represents info drift severity.
	DriftSeverityInfo DriftSeverity = "info"
)

// Helper functions

// NewConfigurationSource creates a new configuration source.
func NewConfigurationSource(sourceType SourceType, location string) ConfigurationSource {
	return ConfigurationSource{
		Type:     sourceType,
		Location: location,
		Priority: 0,
		Required: false,
		Format:   ConfigurationFormatYAML,
	}
}

// NewAppConfiguration creates a new application configuration.
func NewAppConfiguration() AppConfiguration {
	return AppConfiguration{
		Environments: make(map[string]EnvironmentConfiguration),
		GlobalSettings: GlobalSettings{
			LogLevel:     LogLevelInfo,
			LogFormat:    LogFormatJSON,
			MaxCacheSize: 100 * 1024 * 1024, // 100MB
			CacheTTL:     time.Hour,
		},
		Metadata: ConfigurationMetadata{
			LoadTime: time.Now(),
			Sources:  []ConfigurationSource{},
		},
	}
}

// Validation helper functions

// ValidateRequired checks if required fields are present.
func ValidateRequired(value interface{}, fieldName string) error {
	if value == nil || value == "" {
		return &ConfigurationError{
			Field:    fieldName,
			Err:      domain.ErrFieldRequired,
			Severity: ErrorSeverityError,
		}
	}

	return nil
}

// ValidateEnum checks if value is in allowed list.
func ValidateEnum(value string, allowed []string, fieldName string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}

	return &ConfigurationError{
		Field:    fieldName,
		Value:    value,
		Err:      fmt.Errorf("%w: %v", domain.ErrValueNotInAllowedList, allowed),
		Severity: ErrorSeverityError,
	}
}

// Error implements the error interface for ConfigurationError.
func (ce ConfigurationError) Error() string {
	return fmt.Sprintf("configuration %s in field %s: %v", ce.Severity, ce.Field, ce.Err)
}
