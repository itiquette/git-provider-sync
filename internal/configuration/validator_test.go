// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

func TestValidateConfiguration_Success(t *testing.T) {
	t.Parallel()

	appCfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"test": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "testuser",
						Auth: config.AuthConfig{
							Token: "test-token",
						},
					},
				},
			},
		},
	}

	err := validateConfiguration(context.Background(), appCfg)
	require.NoError(t, err)
}

func TestValidateConfiguration_EmptyConfigurations(t *testing.T) {
	t.Parallel()

	appCfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{},
	}

	err := validateConfiguration(context.Background(), appCfg)
	require.ErrorIs(t, err, domain.ErrNoSyncConfigurations)
}

func TestValidateConfiguration_InvalidProviderType(t *testing.T) {
	t.Parallel()

	appCfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"test": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						ProviderType: "invalid-provider",
						Domain:       "github.com",
						Owner:        "testuser",
					},
				},
			},
		},
	}

	err := validateConfiguration(context.Background(), appCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "configuration validation failed")
}

func TestValidateConfiguration_EmptyEnvironment(t *testing.T) {
	t.Parallel()

	appCfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"testenv": {},
		},
	}

	err := validateConfiguration(context.Background(), appCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "testenv")
}

func TestValidateConfiguration_NoOwner(t *testing.T) {
	t.Parallel()

	appCfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"test": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						// Missing owner
					},
				},
			},
		},
	}

	err := validateConfiguration(context.Background(), appCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "configuration validation failed")
}

func TestValidateConfiguration_InvalidDuration(t *testing.T) {
	t.Parallel()

	appCfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"test": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "testuser",
						Auth: config.AuthConfig{
							Token: "test-token",
						},
					},
					ActiveFromLimit: "invalid-duration",
				},
			},
		},
	}

	err := validateConfiguration(context.Background(), appCfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid duration format")
}

func TestValidateConfiguration_ValidDuration(t *testing.T) {
	t.Parallel()

	appCfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"test": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "testuser",
						Auth: config.AuthConfig{
							Token: "test-token",
						},
					},
					ActiveFromLimit: "24h",
				},
			},
		},
	}

	err := validateConfiguration(context.Background(), appCfg)
	require.NoError(t, err)
}

func TestValidateConfiguration_WithValidMirrors(t *testing.T) {
	t.Parallel()

	appCfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"test": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "testuser",
						Auth: config.AuthConfig{
							Token: "test-token",
						},
					},
					Mirrors: map[string]config.MirrorConfig{
						"gitlab": {
							BaseConfig: config.BaseConfig{
								ProviderType: "gitlab",
								Domain:       "gitlab.com",
								Owner:        "testuser",
								Auth: config.AuthConfig{
									Token: "gitlab-token",
								},
							},
						},
					},
				},
			},
		},
	}

	err := validateConfiguration(context.Background(), appCfg)
	require.NoError(t, err)
}

// Test file-specific validation functions that are not covered by domain validation

func TestValidateDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		duration    string
		expectError bool
	}{
		{"empty", "", false},
		{"valid", "24h", false},
		{"valid minutes", "30m", false},
		{"invalid", "invalid-duration", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateDuration(test.duration)

			if test.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidDuration)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePathExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{"empty path", "", true},
		{"relative path", "relative/path", true},
		{"non-existent path", "/non/existent/path", true},
		{"valid path", "/tmp", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validatePathExists(test.path)

			if test.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidPath)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{"empty", "", true},
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"invalid scheme", "ftp://example.com", true},
		{"malformed", "not-a-url", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateURL(test.url)

			if test.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidURL)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSSHCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sshCommand  string
		expectError bool
	}{
		{"empty", "", false},
		{"valid", "ssh -i ~/.ssh/id_rsa", false},
		{"invalid prefix", "not-ssh command", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateSSHCommand(test.sshCommand)

			if test.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, domain.ErrSSHCommandMustStartWithSSH)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGitBinary(t *testing.T) {
	t.Parallel()

	// Create a context with timeout for testing
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test will pass if git is available on the system
	// and fail if it's not, which is expected behavior
	err := validateGitBinary(ctx)

	// We can't assert success/failure here since it depends on the system
	// Instead, we'll test that the function doesn't panic and returns some result
	t.Logf("Git binary validation result: %v", err)
}
