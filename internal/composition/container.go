// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/adapters/config"
	"itiquette/git-provider-sync/internal/adapters/transport"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
	"itiquette/git-provider-sync/internal/log"
)

// Container holds all application dependencies wired together.
// It follows immutable design - once created, dependencies cannot be changed.
// Use cases are created on-demand with proper dependency injection.
type Container struct {
	config          ports.AppConfiguration
	configAdapter   ports.Configuration
	providerFactory *ProviderFactory
	gitFactory      *GitFactory
	httpFactory     *transport.HTTPFactory
	logger          ports.Logger
}

// ContainerConfig contains configuration for building the container.
type ContainerConfig struct {
	ConfigPath     string
	Environment    string
	LogLevel       string
	DryRun         bool
	SkipTLSVerify  bool
	MaxConcurrency int
}

// NewContainer creates a new dependency injection container.
// All dependencies are wired explicitly without service locator patterns.
func NewContainer(ctx context.Context, containerConfig ContainerConfig) (*Container, error) {
	// 1. Load configuration
	configAdapter := config.New()

	configSource := ports.NewConfigurationSource(ports.SourceTypeFile, containerConfig.ConfigPath)

	appConfig, err := configAdapter.Load(ctx, configSource)
	if err != nil {
		return nil, fmt.Errorf("failed to load application configuration: %w", err)
	}

	// 2. Create HTTP factory with configuration
	httpConfig := getHTTPConfig(appConfig, containerConfig)
	httpFactory := transport.NewHTTPFactory(httpConfig)

	// 3. Create logger
	logger := log.NewSimpleLoggerWithLevel(appConfig.GlobalSettings.LogLevel)

	// 4. Create provider factory
	providerFactory := NewProviderFactoryWithTransport(httpFactory, logger)

	// 5. Create git factory
	gitConfig := getGitConfig(appConfig, containerConfig)
	gitFactory := NewGitFactory(gitConfig)

	// Container with factories only - use cases created on demand
	return &Container{
		config:          appConfig,
		configAdapter:   configAdapter,
		providerFactory: providerFactory,
		gitFactory:      gitFactory,
		httpFactory:     httpFactory,
		logger:          logger,
	}, nil
}

// Configuration returns the application configuration.
func (c *Container) Configuration() ports.AppConfiguration {
	return c.config
}

// ConfigAdapter returns the configuration adapter.
func (c *Container) ConfigAdapter() ports.Configuration {
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

// HTTPFactory returns the HTTP factory.
func (c *Container) HTTPFactory() *transport.HTTPFactory {
	return c.httpFactory
}

// CreateSyncUseCase creates a sync use case with proper dependency injection.
func (c *Container) CreateSyncUseCase(
	repositoryProvider ports.RepositoryProvider,
	gitOperations ports.GitOperations,
) sync.SyncRepositoriesUseCase {
	return sync.NewSyncRepositoriesUseCase(c.configAdapter, repositoryProvider, gitOperations, c.logger)
}

// CreateValidateUseCase creates a validate use case with proper dependency injection.
func (c *Container) CreateValidateUseCase(
	repositoryProvider ports.RepositoryProvider,
) sync.ValidateSyncUseCase {
	return sync.NewValidateSyncUseCase(repositoryProvider, c.configAdapter)
}

// CreateFilterUseCase creates a filter use case.
func (c *Container) CreateFilterUseCase() sync.FilterRepositoriesUseCase {
	return sync.FilterRepositoriesUseCase{}
}

// CreateProvider creates a provider for the given configuration.
func (c *Container) CreateProvider(ctx context.Context, providerConfig ports.ProviderConfig) (ports.RepositoryProvider, error) {
	return c.providerFactory.CreateProviderFromConfig(ctx, providerConfig)
}

// CreateGitOperations creates git operations for the given configuration.
func (c *Container) CreateGitOperations(config ports.GitConfig) (ports.GitOperations, error) {
	return c.gitFactory.CreateOperations(config)
}

// Close performs cleanup of container resources.
func (c *Container) Close() error {
	// Stop configuration watching if active
	if c.configAdapter != nil {
		err := c.configAdapter.StopWatching()
		if err != nil {
			return fmt.Errorf("failed to stop configuration watching: %w", err)
		}
	}

	return nil
}

// Private helper functions for configuration

// getHTTPConfig extracts HTTP configuration from app configuration.
func getHTTPConfig(appConfig ports.AppConfiguration, containerConfig ContainerConfig) transport.HTTPConfig {
	httpConfig := transport.GetDefaultHTTPConfig()

	// Override with global settings
	if appConfig.GlobalSettings.LogLevel != "" {
		// Configure timeout based on log level (debug = longer timeout)
		if appConfig.GlobalSettings.LogLevel == ports.LogLevelDebug {
			httpConfig.Timeout = httpConfig.Timeout * 2
		}
	}

	// Apply container-specific overrides
	httpConfig.SkipTLSVerify = containerConfig.SkipTLSVerify

	return httpConfig
}

// getGitConfig extracts git configuration from app configuration.
func getGitConfig(appConfig ports.AppConfiguration, containerConfig ContainerConfig) ports.GitConfig {
	return ports.GitConfig{
		PreferredImplementation: "go-git", // Default to go-git
		UserName:                "git-provider-sync",
		UserEmail:               "sync@git-provider-sync.local",
		MaxConcurrent:           containerConfig.MaxConcurrency,
		VerifySSL:               !containerConfig.SkipTLSVerify,
		Debug:                   appConfig.GlobalSettings.LogLevel == ports.LogLevelDebug,
	}
}

// ContainerBuilder provides a fluent interface for building containers.
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
