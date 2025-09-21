// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define sentinel errors for clone options validation.
var (
	ErrMissingRepositoryName = errors.New("repository name is required")
	ErrMissingSourceURL      = errors.New("source URL is required")
	ErrMissingTargetPath     = errors.New("target path is required")
)

func TestCloneOptionsBuilder_Build(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		builderSetup  func() CloneOptionsBuilder
		wantErr       error
		expectedError string
	}{
		{
			name: "builds_successfully_with_all_required_fields",
			builderSetup: func() CloneOptionsBuilder {
				return NewCloneOptionsBuilder().
					WithRepositoryName("test-repo").
					WithSourceURL("https://github.com/user/repo.git").
					WithTargetPath("/tmp/test")
			},
			wantErr: nil,
		},
		{
			name: "returns_error_when_repository_name_missing",
			builderSetup: func() CloneOptionsBuilder {
				return NewCloneOptionsBuilder().
					WithSourceURL("https://github.com/user/repo.git").
					WithTargetPath("/tmp/test")
			},
			wantErr:       ErrMissingRepositoryName,
			expectedError: "repository name is required",
		},
		{
			name: "returns_error_when_source_url_missing",
			builderSetup: func() CloneOptionsBuilder {
				return NewCloneOptionsBuilder().
					WithRepositoryName("test-repo").
					WithTargetPath("/tmp/test")
			},
			wantErr:       ErrMissingSourceURL,
			expectedError: "source URL is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			builder := test.builderSetup()
			result, err := builder.Build()

			if test.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
				assert.Equal(t, CloneOptions{}, result)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, CloneOptions{}, result)
			}
		})
	}
}

func TestNewCloneOptionsFromRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		repo        Repository
		auth        AuthConfig
		mirror      bool
		expectError bool
		expected    CloneOptions
	}{
		{
			name: "creates_options_from_repository_with_https_url",
			repo: Repository{
				name:     "test-repo",
				httpsURL: "https://github.com/user/repo.git",
				sshURL:   "git@github.com:user/repo.git",
			},
			auth:        AuthConfig{},
			mirror:      false,
			expectError: false,
			expected: CloneOptions{
				repositoryName: "test-repo",
				sourceURL:      "https://github.com/user/repo.git",
				protocol:       "https",
				isMirror:       false,
				isNonBare:      false,
				useASCIIName:   false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := NewCloneOptionsFromRepository(test.repo, test.auth, test.mirror)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.repositoryName, result.repositoryName)
				assert.Equal(t, test.expected.sourceURL, result.sourceURL)
				assert.Equal(t, test.expected.protocol, result.protocol)
				assert.Equal(t, test.expected.isMirror, result.isMirror)
				assert.Equal(t, test.expected.isNonBare, result.isNonBare)
				assert.Equal(t, test.expected.useASCIIName, result.useASCIIName)
			}
		})
	}
}

func TestNewMirrorCloneOptions(t *testing.T) {
	t.Parallel()

	repoName := "test-repo"
	sourceURL := "https://github.com/user/repo.git"
	auth := AuthConfig{}

	result, err := NewMirrorCloneOptions(repoName, sourceURL, auth)

	require.NoError(t, err)
	assert.Equal(t, repoName, result.repositoryName)
	assert.Equal(t, sourceURL, result.sourceURL)
	assert.True(t, result.isMirror)
	assert.False(t, result.isNonBare)
}

func TestNewRegularCloneOptions(t *testing.T) {
	t.Parallel()

	repoName := "test-repo"
	sourceURL := "https://github.com/user/repo.git"
	targetPath := "/tmp/test"
	auth := AuthConfig{}

	result, err := NewRegularCloneOptions(repoName, sourceURL, targetPath, auth)

	require.NoError(t, err)
	assert.Equal(t, repoName, result.repositoryName)
	assert.Equal(t, sourceURL, result.sourceURL)
	assert.Equal(t, targetPath, result.targetPath)
	assert.False(t, result.isMirror)
	assert.True(t, result.isNonBare)
}

func TestCloneOptions_EffectiveName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  CloneOptions
		expected string
	}{
		{
			name: "normal repository name",
			options: CloneOptions{
				repositoryName: "test-repo",
				useASCIIName:   false,
			},
			expected: "test-repo",
		},
		{
			name: "repository name with ASCII conversion",
			options: CloneOptions{
				repositoryName: "tëst-repö",
				useASCIIName:   true,
			},
			expected: "t-st-rep",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := test.options.EffectiveName()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCloneOptions_IsSecureProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		protocol string
		expected bool
	}{
		{
			name:     "https is secure",
			protocol: "https",
			expected: true,
		},
		{
			name:     "ssh is secure",
			protocol: "ssh",
			expected: true,
		},
		{
			name:     "http is not secure",
			protocol: "http",
			expected: false,
		},
		{
			name:     "empty protocol is not secure",
			protocol: "",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			options := CloneOptions{protocol: test.protocol}
			assert.Equal(t, test.expected, options.IsSecureProtocol())
		})
	}
}

func TestCloneOptions_RequiresAuthentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  CloneOptions
		expected bool
	}{
		{
			name: "has token authentication",
			options: CloneOptions{
				authConfig: AuthConfig{token: "test-token"},
			},
			expected: true,
		},
		{
			name: "no authentication",
			options: CloneOptions{
				authConfig: AuthConfig{authType: AuthTypeNone},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := test.options.RequiresAuthentication()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCloneOptions_GetCloneURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  CloneOptions
		httpsURL string
		sshURL   string
		expected string
	}{
		{
			name: "prefer https when protocol is https",
			options: CloneOptions{
				protocol: "https",
			},
			httpsURL: "https://github.com/user/repo.git",
			sshURL:   "git@github.com:user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name: "prefer ssh when protocol is ssh",
			options: CloneOptions{
				protocol: "ssh",
			},
			httpsURL: "https://github.com/user/repo.git",
			sshURL:   "git@github.com:user/repo.git",
			expected: "git@github.com:user/repo.git",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := test.options.GetCloneURL(test.httpsURL, test.sshURL)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCloneOptions_WithUpdatedAuth(t *testing.T) {
	t.Parallel()

	originalOptions := CloneOptions{
		repositoryName: "test-repo",
		authConfig:     AuthConfig{},
	}

	newAuth := AuthConfig{token: "new-token"}
	updatedOptions := originalOptions.WithUpdatedAuth(newAuth)

	assert.Equal(t, newAuth, updatedOptions.authConfig)
	assert.Equal(t, originalOptions.repositoryName, updatedOptions.repositoryName)
	assert.NotEqual(t, originalOptions.authConfig, updatedOptions.authConfig)
}

func TestCloneOptions_WithUpdatedProtocol(t *testing.T) {
	t.Parallel()

	originalOptions := CloneOptions{
		repositoryName: "test-repo",
		protocol:       "https",
	}

	updatedOptions := originalOptions.WithUpdatedProtocol("ssh")

	assert.Equal(t, "ssh", updatedOptions.protocol)
	assert.Equal(t, originalOptions.repositoryName, updatedOptions.repositoryName)
	assert.NotEqual(t, originalOptions.protocol, updatedOptions.protocol)
}

func TestCloneOptionsBuilder_ChainedCalls(t *testing.T) {
	t.Parallel()

	options, err := NewCloneOptionsBuilder().
		WithRepositoryName("test-repo").
		WithSourceURL("https://github.com/user/repo.git").
		WithTargetPath("/tmp/test").
		WithMirror(true).
		WithNonBare(false).
		WithASCIIName(true).
		WithProtocol("https").
		Build()

	require.NoError(t, err)
	assert.Equal(t, "test-repo", options.RepositoryName())
	assert.Equal(t, "https://github.com/user/repo.git", options.SourceURL())
	assert.Equal(t, "/tmp/test", options.TargetPath())
	assert.True(t, options.IsMirror())
	assert.False(t, options.IsNonBare())
	assert.True(t, options.UseASCIIName())
	assert.Equal(t, "https", options.Protocol())
}
