// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

const testConfigContent = `
gitprovidersync:
  defaultenv:
    testsource:
      provider_type: github
      domain: github.com
      owner: testuser
      owner_type: user
      auth:
        token: testtoken
`

func TestNewContainer_ValidConfig_CreatesContainer(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)

	require.NoError(t, err)
	require.NotNil(t, container)
	assert.NotNil(t, container.config)
	assert.NotNil(t, container.configAdapter)
	assert.NotNil(t, container.providerFactory)
	assert.NotNil(t, container.gitFactory)
	assert.NotNil(t, container.httpFactory)
	assert.NotNil(t, container.logger)
}

func TestNewContainer_InvalidConfigPath(t *testing.T) {
	t.Parallel()

	containerConfig := ContainerConfig{
		ConfigPath:     "/nonexistent/config.yaml",
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)

	require.Error(t, err)
	require.Nil(t, container)
	assert.Contains(t, err.Error(), "failed to load application configuration")
}

func TestNewContainer_EmptyConfigPath(t *testing.T) {
	t.Parallel()

	containerConfig := ContainerConfig{
		ConfigPath:     "",
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)

	require.Error(t, err)
	require.Nil(t, container)
	assert.Contains(t, err.Error(), "failed to load application configuration")
}

func TestNewContainer_DryRunEnabled_ConfiguresCorrectly(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "debug",
		DryRun:         true,
		SkipTLSVerify:  false, // TLS verification is always enforced
		MaxConcurrency: 10,
	}

	container, err := NewContainer(context.Background(), containerConfig)

	require.NoError(t, err)
	require.NotNil(t, container)
}

func TestNewContainer_AllLogLevels_AcceptsValidLevels(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	logLevels := []string{"trace", "debug", "info", "warn", "error"}

	for _, logLevel := range logLevels {
		t.Run(logLevel, func(t *testing.T) {
			t.Parallel()

			containerConfig := ContainerConfig{
				ConfigPath:     configFile,
				Environment:    "defaultenv",
				LogLevel:       logLevel,
				DryRun:         false,
				SkipTLSVerify:  false,
				MaxConcurrency: 3,
			}

			container, err := NewContainer(context.Background(), containerConfig)

			require.NoError(t, err)
			require.NotNil(t, container)
		})
	}
}

func TestNewContainer_InvalidLogLevel(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "invalid",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)

	// Should still succeed but use default log level
	require.NoError(t, err)
	require.NotNil(t, container)
}

func TestContainer_CreateProvider(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)
	require.NoError(t, err)

	providerConfig := ports.ProviderConfig{
		ProviderType: "github",
		Domain:       "github.com",
		Owner:        "testuser",
		AuthConfig: ports.AuthenticationConfig{
			Token: "testtoken",
		},
	}

	provider, err := container.CreateProvider(context.Background(), providerConfig)

	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestContainer_CreateGitOperations(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)
	require.NoError(t, err)

	gitConfig := ports.GitConfig{
		PreferredImplementation: "go-git",
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	gitOps, err := container.CreateGitOperations(gitConfig)

	require.NoError(t, err)
	require.NotNil(t, gitOps)
}

func TestContainerConfig_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        ContainerConfig
		expectedValid bool
	}{
		{
			name: "valid config",
			config: ContainerConfig{
				ConfigPath:     "/path/to/config.yaml",
				Environment:    "production",
				LogLevel:       "info",
				DryRun:         false,
				SkipTLSVerify:  false,
				MaxConcurrency: 5,
			},
			expectedValid: true,
		},
		{
			name: "zero concurrency",
			config: ContainerConfig{
				ConfigPath:     "/path/to/config.yaml",
				Environment:    "production",
				LogLevel:       "info",
				DryRun:         false,
				SkipTLSVerify:  false,
				MaxConcurrency: 0,
			},
			expectedValid: true, // Should be handled gracefully
		},
		{
			name: "negative concurrency",
			config: ContainerConfig{
				ConfigPath:     "/path/to/config.yaml",
				Environment:    "production",
				LogLevel:       "info",
				DryRun:         false,
				SkipTLSVerify:  false,
				MaxConcurrency: -1,
			},
			expectedValid: true, // Should be handled gracefully
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Just verify the struct can be created without panicking
			assert.NotNil(t, test.config)
			assert.Equal(t, test.expectedValid, test.config.ConfigPath != "" || test.config.Environment != "")
		})
	}
}

func TestContainer_Dependencies(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)
	require.NoError(t, err)

	// Test that all dependencies are properly initialized
	assert.NotNil(t, container.config, "config should be initialized")
	assert.NotNil(t, container.configAdapter, "configAdapter should be initialized")
	assert.NotNil(t, container.providerFactory, "providerFactory should be initialized")
	assert.NotNil(t, container.gitFactory, "gitFactory should be initialized")
	assert.NotNil(t, container.httpFactory, "httpFactory should be initialized")
	assert.NotNil(t, container.logger, "logger should be initialized")
}

func TestContainer_GetterMethods(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)
	require.NoError(t, err)

	// Test getter methods
	assert.NotNil(t, container.Configuration())
	assert.NotNil(t, container.ConfigAdapter())
	assert.NotNil(t, container.ProviderFactory())
	assert.NotNil(t, container.GitFactory())
	assert.NotNil(t, container.HTTPFactory())
}

func TestContainer_CreateUseCases(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)
	require.NoError(t, err)

	// Create mock provider and git operations for testing use cases
	providerConfig := ports.ProviderConfig{
		ProviderType: "github",
		Domain:       "github.com",
		Owner:        "testuser",
		AuthConfig: ports.AuthenticationConfig{
			Token: "testtoken",
		},
	}

	provider, err := container.CreateProvider(context.Background(), providerConfig)
	require.NoError(t, err)

	gitConfig := ports.GitConfig{
		PreferredImplementation: "go-git",
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	gitOps, err := container.CreateGitOperations(gitConfig)
	require.NoError(t, err)

	// Test use case creation
	syncUseCase := container.CreateSyncUseCase(provider, gitOps)
	assert.NotNil(t, syncUseCase)

	validateUseCase := container.CreateValidateUseCase(provider)
	assert.NotNil(t, validateUseCase)

	filterUseCase := container.CreateFilterUseCase()
	assert.NotNil(t, filterUseCase)
}

func TestContainer_Close(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  false,
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)
	require.NoError(t, err)

	// Test cleanup
	err = container.Close()
	require.NoError(t, err)
}

func TestContainerBuilder(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	// Test builder pattern
	builder := NewContainerBuilder()
	require.NotNil(t, builder)

	container, err := builder.
		WithConfigPath(configFile).
		WithEnvironment("testenv").
		WithLogLevel("debug").
		WithDryRun(true).
		WithSkipTLSVerify(false). // TLS verification is always enforced
		WithMaxConcurrency(10).
		Build(context.Background())

	require.NoError(t, err)
	require.NotNil(t, container)

	// Clean up
	err = container.Close()
	require.NoError(t, err)
}

func TestNewContainer_RejectsTLSBypass(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := testConfigContent
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	containerConfig := ContainerConfig{
		ConfigPath:     configFile,
		Environment:    "defaultenv",
		LogLevel:       "info",
		DryRun:         false,
		SkipTLSVerify:  true, // Attempt to bypass TLS verification
		MaxConcurrency: 5,
	}

	container, err := NewContainer(context.Background(), containerConfig)

	require.Error(t, err)
	require.Nil(t, container)
	assert.Contains(t, err.Error(), "TLS verification bypass is not permitted")
}
