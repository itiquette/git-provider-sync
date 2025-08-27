// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package composition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestNewGitFactory(t *testing.T) {
	t.Parallel()

	config := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "go-git",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	factory := NewGitFactory(config)

	require.NotNil(t, factory)
	// GitFactory is now stateless - no config field
}

func TestGitFactory_CreateOperations_GoGit(t *testing.T) {
	t.Parallel()

	factoryConfig := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "go-git",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	factory := NewGitFactory(factoryConfig)

	operationConfig := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "go-git",
		MaxConcurrent:           3,
		VerifySSL:               true,
		Debug:                   false,
	}

	operations, err := factory.CreateOperations(operationConfig)

	require.NoError(t, err)
	require.NotNil(t, operations)
}

func TestGitFactory_CreateOperations_GitBinary(t *testing.T) {
	t.Parallel()

	factoryConfig := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "git-binary",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	factory := NewGitFactory(factoryConfig)

	operationConfig := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "git-binary",
		MaxConcurrent:           3,
		VerifySSL:               true,
		Debug:                   false,
	}

	operations, err := factory.CreateOperations(operationConfig)

	require.NoError(t, err)
	require.NotNil(t, operations)
}

func TestGitFactory_CreateOperations_Directory(t *testing.T) {
	t.Parallel()

	factoryConfig := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "directory",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	factory := NewGitFactory(factoryConfig)

	operationConfig := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "directory",
		MaxConcurrent:           3,
		VerifySSL:               true,
		Debug:                   false,
	}

	operations, err := factory.CreateOperations(operationConfig)

	require.NoError(t, err)
	require.NotNil(t, operations)
}

func TestGitFactory_CreateOperations_Archive(t *testing.T) {
	t.Parallel()

	factoryConfig := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "archive",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	factory := NewGitFactory(factoryConfig)

	operationConfig := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "archive",
		MaxConcurrent:           3,
		VerifySSL:               true,
		Debug:                   false,
	}

	operations, err := factory.CreateOperations(operationConfig)

	require.NoError(t, err)
	require.NotNil(t, operations)
}

func TestGitFactory_CreateOperations_UnsupportedImplementation(t *testing.T) {
	t.Parallel()

	factory := NewGitFactory(ports.GitConfig{})

	config := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "unsupported",
		MaxConcurrent:           5,
	}

	operations, err := factory.CreateOperations(config)

	require.Error(t, err)
	require.Nil(t, operations)
	assert.Contains(t, err.Error(), "unsupported git implementation")
}

func TestGitFactory_CreateOperations_EmptyImplementation(t *testing.T) {
	t.Parallel()

	factory := NewGitFactory(ports.GitConfig{})

	config := ports.GitConfig{
		UserName:                "testuser",
		UserEmail:               "test@example.com",
		PreferredImplementation: "", // Empty defaults to go-git
		MaxConcurrent:           5,
	}

	operations, err := factory.CreateOperations(config)

	require.NoError(t, err)
	require.NotNil(t, operations)
}

func TestGitFactory_AvailableImplementations(t *testing.T) {
	t.Parallel()

	implementations := AvailableImplementations()

	require.NotEmpty(t, implementations)
	assert.Contains(t, implementations, ProviderTypeGoGit)
	assert.Contains(t, implementations, ProviderTypeGitBinary)
	assert.Contains(t, implementations, ProviderTypeDirectory)
	assert.Contains(t, implementations, ProviderTypeArchive)
}

func TestGitFactory_IsImplementationAvailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		implementation string
		expected       bool
	}{
		{"go-git available", ProviderTypeGoGit, true},
		{"git-binary available", ProviderTypeGitBinary, true},
		{"directory available", ProviderTypeDirectory, true},
		{"archive available", ProviderTypeArchive, true},
		{"unknown unavailable", "unknown", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := IsImplementationAvailable(test.implementation)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGitFactory_GetDefaultConfig(t *testing.T) {
	t.Parallel()

	defaultConfig := GetDefaultConfig()

	assert.Equal(t, ProviderTypeGoGit, defaultConfig.PreferredImplementation)
	assert.Equal(t, "git-provider-sync", defaultConfig.UserName)
	assert.Equal(t, "sync@git-provider-sync.local", defaultConfig.UserEmail)
	assert.Equal(t, 5, defaultConfig.MaxConcurrent)
	assert.True(t, defaultConfig.VerifySSL)
	assert.False(t, defaultConfig.Debug)
}
