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

	_ = NewGitFactory(config)
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

	_, err := factory.CreateOperations(operationConfig)

	require.NoError(t, err)
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

	_, err := factory.CreateOperations(operationConfig)

	require.NoError(t, err)
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

	_, err := factory.CreateOperations(operationConfig)

	require.NoError(t, err)
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

	_, err := factory.CreateOperations(operationConfig)

	require.NoError(t, err)
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

	_, err := factory.CreateOperations(config)

	require.NoError(t, err)
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
