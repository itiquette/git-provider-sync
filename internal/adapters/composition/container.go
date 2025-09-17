// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"

	"itiquette/git-provider-sync/internal/adapters/config"
	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/adapters/logging"
	"itiquette/git-provider-sync/internal/adapters/repository/archive"
	"itiquette/git-provider-sync/internal/adapters/shared"
	"itiquette/git-provider-sync/internal/adapters/transport"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
)

// Container holds all application dependencies wired together
// follows immutable design - once created, dependencies cannot be changed
// Use cases are created on-demand with proper dependency injection.
type Container struct {
	config          ports.AppConfiguration
	configAdapter   ports.Configuration
	providerFactory *ProviderFactory
	gitFactory      *GitFactory
	httpFactory     *transport.HTTPFactory
	fileSystem      ports.FileSystem
	logger          ports.Logger
	stringUtils     ports.StringUtils
}

// ContainerConfig contains configuration for building the container.
type ContainerConfig struct {
	ConfigPath     string
	Environment    string
	LogLevel       string
	OutputFormat   string // Output format: console, json, plain
	DryRun         bool
	SkipTLSVerify  bool
	MaxConcurrency int
}

// NewContainer creates a new dependency injection container
// All dependencies are wired explicitly without service locator patterns.
func NewContainer(ctx context.Context, containerConfig ContainerConfig) (*Container, error) {
	// 1. Load configuration
	configAdapter := config.New()

	configSource := ports.NewConfigurationSource(ports.SourceTypeFile, containerConfig.ConfigPath)
	configSource.Required = true

	appConfig, err := configAdapter.Load(ctx, configSource)
	if err != nil {
		return nil, fmt.Errorf("failed to load application configuration: %w", err)
	}

	// 2. Create HTTP factory with configuration
	httpConfig := getHTTPConfig(appConfig, containerConfig)

	httpFactory, err := transport.NewHTTPFactory(httpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP factory: %w", err)
	}

	// 3. Create logger using proper hexagonal adapter
	zerologLevel := convertLogLevel(appConfig.GlobalSettings.LogLevel)

	// Determine if we should suppress logger output
	// Logger only writes to stderr in debug/trace/verbose modes
	// In normal/quiet modes, formatters handle all user-facing output
	suppressLogs := containerConfig.LogLevel != "debug" &&
		containerConfig.LogLevel != "trace" &&
		containerConfig.LogLevel != "verbose" &&
		string(appConfig.GlobalSettings.LogLevel) != "debug" &&
		string(appConfig.GlobalSettings.LogLevel) != "trace"

	var zerologInstance zerolog.Logger

	switch {
	case suppressLogs:
		// Suppress INFO and DEBUG logs in normal operation - formatters handle output
		// Only show warnings and errors for actual issues
		zerologInstance = zerolog.New(os.Stderr).Level(zerolog.WarnLevel).With().Timestamp().Logger()
	case containerConfig.OutputFormat == "json":
		// JSON format for structured output (when logging is enabled)
		zerologInstance = zerolog.New(os.Stderr).Level(zerologLevel).With().Timestamp().Logger()
	default:
		// Console format for human-readable output (when logging is enabled)
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
			NoColor:    containerConfig.OutputFormat == "plain", // Plain format = no colors
		}
		zerologInstance = zerolog.New(consoleWriter).Level(zerologLevel).With().Timestamp().Logger()
	}

	logger := logging.NewZerologAdapter(&zerologInstance)

	// 4. Create provider factory
	providerFactory := NewProviderFactory(httpFactory, logger)

	// 5. Create file system adapter (needed by git factory)
	fileSystem := filesystem.NewOSFileSystem()

	// 6. Create git factory with file system dependency
	gitConfig := getGitConfig(appConfig, containerConfig)
	gitFactory := NewGitFactory(gitConfig, fileSystem)

	// 7. Create string utils adapter
	stringUtils := shared.NewStringUtilsAdapter()

	// Container with factories only - use cases created on demand
	return &Container{
		config:          appConfig,
		configAdapter:   configAdapter,
		providerFactory: providerFactory,
		gitFactory:      gitFactory,
		httpFactory:     httpFactory,
		fileSystem:      fileSystem,
		logger:          logger,
		stringUtils:     stringUtils,
	}, nil
}

// Configuration returns the application configuration.
func (c *Container) Configuration() ports.AppConfiguration {
	return c.config
}

// ConfigAdapter returns the configuration adapter.
func (c *Container) ConfigAdapter() ports.Configuration { //nolint:ireturn // Getter returning interface
	return c.configAdapter
}

// ProviderFactory returns the provider factory.
func (c *Container) ProviderFactory() *ProviderFactory {
	return c.providerFactory
}

// GitFactory returns the git operations factory.
func (c *Container) GitFactory() *GitFactory {
	return c.gitFactory
}

// StringUtils returns the string utilities adapter.
func (c *Container) StringUtils() ports.StringUtils {
	return c.stringUtils
}

// HTTPFactory returns the HTTP factory.
func (c *Container) HTTPFactory() *transport.HTTPFactory {
	return c.httpFactory
}

// FileSystem returns the file system adapter.
func (c *Container) FileSystem() ports.FileSystem { //nolint:ireturn // Getter returning interface
	return c.fileSystem
}

// Logger returns the logger adapter.
func (c *Container) Logger() ports.Logger { //nolint:ireturn // Getter returning interface
	return c.logger
}

// CreateSyncUseCase creates a sync use case with dependency injection.
func (c *Container) CreateSyncUseCase(
	repositoryProvider ports.RepositoryProvider,
	gitOperations ports.GitOperations,
) sync.RepositoriesUseCase {
	// Create archive operations with configured temp directories
	tempDir := c.config.GlobalSettings.TempDirectory
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	archiveDir := c.config.GlobalSettings.CacheDirectory
	if archiveDir == "" {
		archiveDir = os.TempDir()
	}

	archiveOps := archive.NewOperations(c.logger, tempDir, archiveDir)

	return sync.NewRepositoriesUseCase(c.configAdapter, repositoryProvider, gitOperations, archiveOps, c.fileSystem, c.logger, c.stringUtils)
}

// CreateValidateUseCase creates a validation use case with dependency injection.
func (c *Container) CreateValidateUseCase(
	repositoryProvider ports.RepositoryProvider,
) sync.ValidateSyncUseCase {
	return sync.NewValidateSyncUseCase(repositoryProvider, c.configAdapter)
}

// CreateFilterUseCase creates a filter use case.
func (c *Container) CreateFilterUseCase() sync.FilterRepositoriesUseCase {
	return sync.FilterRepositoriesUseCase{}
}

// CreateProvider creates a repository provider from configuration.
func (c *Container) CreateProvider(ctx context.Context, providerConfig ports.ProviderConfig) (ports.RepositoryProvider, error) { //nolint:ireturn // Factory method returning interface
	return c.providerFactory.CreateProviderFromConfig(ctx, providerConfig)
}

// CreateGitOperations creates git operations from configuration.
func (c *Container) CreateGitOperations(config ports.GitConfig) (ports.GitOperations, error) { //nolint:ireturn // Factory method returning interface
	return c.gitFactory.CreateOperations(config)
}

// Close performs cleanup of container resources.
func (c *Container) Close() error {
	return nil
}

// Private helper functions for configuration

// GetHTTPConfig extracts HTTP configuration from app configuration.
func getHTTPConfig(appConfig ports.AppConfiguration, containerConfig ContainerConfig) transport.HTTPConfig {
	httpConfig := transport.GetDefaultHTTPConfig()

	// Override with global settings
	if appConfig.GlobalSettings.LogLevel != "" {
		// Configure timeout based on log level (debug = longer timeout)
		if appConfig.GlobalSettings.LogLevel == ports.LogLevelDebug {
			httpConfig.Timeout *= 2
		}
	}

	// Apply container-specific overrides
	httpConfig.SkipTLSVerify = containerConfig.SkipTLSVerify

	return httpConfig
}

// GetGitConfig extracts git configuration from app configuration.
func getGitConfig(appConfig ports.AppConfiguration, containerConfig ContainerConfig) ports.GitConfig {
	// Get git timeout from first environment's configuration if available
	gitTimeout := 5 * time.Minute // Default 5 minutes

	for _, env := range appConfig.Environments {
		if env.Options.Timeout > 0 {
			gitTimeout = env.Options.Timeout

			break
		}
	}

	return ports.GitConfig{
		PreferredImplementation: "go-git", // Default to go-git
		UserName:                "git-provider-sync",
		UserEmail:               "sync@git-provider-sync.local",
		MaxConcurrent:           containerConfig.MaxConcurrency,
		VerifySSL:               !containerConfig.SkipTLSVerify,
		Debug:                   appConfig.GlobalSettings.LogLevel == ports.LogLevelDebug,
		Timeout:                 gitTimeout,
	}
}

// ConvertLogLevel converts domain LogLevel to zerolog.Level.
func convertLogLevel(level ports.LogLevel) zerolog.Level {
	switch level {
	case ports.LogLevelTrace:
		return zerolog.TraceLevel
	case ports.LogLevelDebug:
		return zerolog.DebugLevel
	case ports.LogLevelInfo:
		return zerolog.InfoLevel
	case ports.LogLevelWarn:
		return zerolog.WarnLevel
	case ports.LogLevelError:
		return zerolog.ErrorLevel
	case ports.LogLevelFatal:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// ContainerBuilder builds Container instances.
type ContainerBuilder struct {
	config ContainerConfig
}

// NewContainerBuilder creates a new container builder.
func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{
		config: ContainerConfig{
			ConfigPath:     "config.yaml",
			Environment:    "development",
			LogLevel:       "info",
			DryRun:         false,
			SkipTLSVerify:  false,
			MaxConcurrency: 5,
		},
	}
}

// WithConfigPath sets the configuration file path.
func (b *ContainerBuilder) WithConfigPath(path string) *ContainerBuilder {
	b.config.ConfigPath = path

	return b
}

// WithEnvironment sets the environment.
func (b *ContainerBuilder) WithEnvironment(env string) *ContainerBuilder {
	b.config.Environment = env

	return b
}

// WithLogLevel sets the log level.
func (b *ContainerBuilder) WithLogLevel(level string) *ContainerBuilder {
	b.config.LogLevel = level

	return b
}

// WithDryRun enables or disables dry run mode.
func (b *ContainerBuilder) WithDryRun(dryRun bool) *ContainerBuilder {
	b.config.DryRun = dryRun

	return b
}

// WithSkipTLSVerify enables or disables TLS verification.
func (b *ContainerBuilder) WithSkipTLSVerify(skip bool) *ContainerBuilder {
	b.config.SkipTLSVerify = skip

	return b
}

// WithMaxConcurrency sets the maximum concurrency.
func (b *ContainerBuilder) WithMaxConcurrency(maxConcurrency int) *ContainerBuilder {
	b.config.MaxConcurrency = maxConcurrency

	return b
}

// Build creates the container with the configured options.
func (b *ContainerBuilder) Build(ctx context.Context) (*Container, error) {
	return NewContainer(ctx, b.config)
}
