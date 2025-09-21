// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"errors"
	"fmt"
	"os"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// AdapterConfig holds configuration for building an Adapter.
// This is an immutable configuration object.
type AdapterConfig struct {
	Config     ports.GitConfig
	Logger     ports.Logger
	FileSystem ports.FileSystem
	TempDir    string
	MirrorSvc  *MirrorService
}

// AdapterOption is a functional option for building AdapterConfig.
// Unlike the old Option, this builds configuration, doesn't mutate objects.
type AdapterOption func(AdapterConfig) AdapterConfig

// WithConfigFileSystem sets a custom filesystem implementation.
func WithConfigFileSystem(fs ports.FileSystem) AdapterOption {
	return func(cfg AdapterConfig) AdapterConfig {
		cfg.FileSystem = fs

		return cfg
	}
}

// WithConfigTempDir sets a custom temporary directory.
func WithConfigTempDir(tempDir string) AdapterOption {
	return func(cfg AdapterConfig) AdapterConfig {
		cfg.TempDir = tempDir

		return cfg
	}
}

// WithConfigMirrorService sets a custom mirror service (for testing).
func WithConfigMirrorService(svc *MirrorService) AdapterOption {
	return func(cfg AdapterConfig) AdapterConfig {
		cfg.MirrorSvc = svc

		return cfg
	}
}

// BuildAdapterConfig creates an AdapterConfig using functional options.
// This is a pure function - no side effects, no mutations.
func BuildAdapterConfig(config ports.GitConfig, logger ports.Logger, opts ...AdapterOption) AdapterConfig {
	// Start with default configuration
	cfg := AdapterConfig{
		Config:     config,
		Logger:     logger,
		FileSystem: filesystem.NewOSFileSystem(),
	}

	// Apply options functionally (each returns new config)
	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return cfg
}

// NewFunctional creates a new Adapter from configuration.
// This is the truly functional constructor - no mutations after creation.
//
//nolint:cyclop // Builder pattern requires multiple configuration checks
func NewFunctional(ctx context.Context, config AdapterConfig) (*Adapter, error) {
	// Validate required fields
	if config.Logger == nil {
		return nil, errors.New("logger is required")
	}

	if config.FileSystem == nil {
		config.FileSystem = filesystem.NewOSFileSystem()
	}

	// Create temp directory if not provided
	tempDir := config.TempDir
	if tempDir == "" {
		var err error

		tempDir, err = os.MkdirTemp("", TempDirPrefix+"*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
	}

	// Create mirror service if not provided
	mirrorSvc := config.MirrorSvc
	//nolint:nestif // Initialization logic requires nested checks
	if mirrorSvc == nil {
		timeout := config.Config.Timeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}

		var err error

		mirrorSvc, err = NewMirrorServiceWithTimeout(ctx, config.Logger, tempDir, timeout)
		if err != nil {
			originalErr := err
			// Clean up temp dir if we created it
			if config.TempDir == "" && tempDir != "" {
				// Best effort cleanup - we're already returning an error
				if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
					// Combine errors explicitly
					return nil, fmt.Errorf("failed to create git binary mirror service: %w (cleanup also failed: %w)", originalErr, cleanupErr)
				}
			}

			return nil, fmt.Errorf("failed to create git binary mirror service: %w", originalErr)
		}
	}

	// Create adapter with all dependencies - no mutations after this
	return &Adapter{
		config:     config.Config,
		fileSystem: config.FileSystem,
		logger:     config.Logger,
		tempDir:    tempDir,
		mirrorSvc:  mirrorSvc,
	}, nil
}
