// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestBaseConfig_GetDomain(t *testing.T) {
	tests := []struct {
		name     string
		config   BaseConfig
		expected string
	}{
		{
			name: "GitHub default domain",
			config: BaseConfig{
				ProviderType: "github",
			},
			expected: "github.com",
		},
		{
			name: "GitLab default domain",
			config: BaseConfig{
				ProviderType: "gitlab",
			},
			expected: "gitlab.com",
		},
		{
			name: "Gitea default domain",
			config: BaseConfig{
				ProviderType: "gitea",
			},
			expected: "gitea.com",
		},
		{
			name: "Custom domain",
			config: BaseConfig{
				ProviderType: "gitlab",
				Domain:       "custom.gitlab.com",
			},
			expected: "custom.gitlab.com",
		},
		{
			name: "Unknown provider type",
			config: BaseConfig{
				ProviderType: "unknown",
			},
			expected: "",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result := tabletest.config.GetDomain()
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestBaseConfig_FillDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   BaseConfig
		expected BaseConfig
	}{
		{
			name: "Empty config gets defaults",
			config: BaseConfig{
				ProviderType: "gitlab",
			},
			expected: BaseConfig{
				ProviderType: "gitlab",
				Domain:       "gitlab.com",
				OwnerType:    "group",
				Auth: AuthConfig{
					HTTPScheme:     HTTPS,
					Protocol:       TLS,
					RequestTimeout: 30,
				},
			},
		},
		{
			name: "Custom values are preserved",
			config: BaseConfig{
				ProviderType: "github",
				Domain:       "custom.github.com",
				OwnerType:    "user",
				Auth: AuthConfig{
					HTTPScheme:     HTTP,
					Protocol:       SSH,
					RequestTimeout: 31,
				},
			},
			expected: BaseConfig{
				ProviderType: "github",
				Domain:       "custom.github.com",
				OwnerType:    "user",
				Auth: AuthConfig{
					HTTPScheme:     HTTP,
					Protocol:       SSH,
					RequestTimeout: 31,
				},
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			tabletest.config.FillDefaults()
			require.Equal(t, tabletest.expected, tabletest.config)
		})
	}
}

func TestSyncConfig_String(t *testing.T) {
	tests := []struct {
		name     string
		config   SyncConfig
		expected string
	}{
		{
			name: "Complete config",
			config: SyncConfig{
				BaseConfig: BaseConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "owner1",
					OwnerType:    "group",
					Auth:         AuthConfig{Token: "secret"},
				},
				Repositories: RepositoriesOption{},
				Mirrors: map[string]MirrorConfig{
					"mirror1": {
						BaseConfig: BaseConfig{
							ProviderType: "gitlab",
							Domain:       "gitlab.com",
						},
					},
				},
			},
			expected: "SyncConfig: ProviderType: github, Domain: github.com, Owner: owner1, OwnerType: group",
		},
		{
			name: "Minimal config",
			config: SyncConfig{
				BaseConfig: BaseConfig{
					ProviderType: "gitlab",
					Domain:       "gitlab.com",
				},
			},
			expected: "SyncConfig: ProviderType: gitlab, Domain: gitlab.com, Owner: , OwnerType: ",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result := tabletest.config.String()
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestSyncConfig_DebugLog(t *testing.T) {
	tests := []struct {
		name     string
		config   SyncConfig
		expected []string
	}{
		{
			name: "GitHub config",
			config: SyncConfig{
				BaseConfig: BaseConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "owner1",
					OwnerType:    "group",
				},
			},
			expected: []string{"type", "github", "domain", "github.com", "owner", "owner1", "ownerType", "group"},
		},
		// {
		// 	name: "Config with mirrors",
		// 	config: SyncConfig{
		// 		BaseConfig: BaseConfig{
		// 			ProviderType: "gitlab",
		// 			Domain:       "gitlab.com",
		// 		},
		// 		Mirrors: map[string]MirrorConfig{
		// 			"mirror1": {
		// 				BaseConfig: BaseConfig{
		// 					ProviderType: "archive",
		// 				},
		// 			},
		// 		},
		// 	},
		// 	expected: []string{"type", "gitlab", "domain", "gitlab.com", "mirror_0"},
		// },
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := zerolog.New(&buf)
			tabletest.config.DebugLog(&logger).Msg(tabletest.name)

			logOutput := buf.String()
			for _, exp := range tabletest.expected {
				require.Contains(t, logOutput, exp)
			}
		})
	}
}

func TestAppConfiguration_DebugLog(t *testing.T) {
	config := AppConfiguration{
		GitProviderSyncConfs: map[string]Environment{
			"test": {
				"source1": SyncConfig{
					BaseConfig: BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
					},
				},
				"source2": SyncConfig{
					BaseConfig: BaseConfig{
						ProviderType: "gitlab",
						Domain:       "gitlab.com",
					},
					Mirrors: map[string]MirrorConfig{
						"mirror1": {
							BaseConfig: BaseConfig{
								ProviderType: "archive",
							},
							Path: "/path/to/archive",
						},
						"mirror2": {
							BaseConfig: BaseConfig{
								ProviderType: "directory",
							},
							Path: "/path/to/directory",
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	config.DebugLog(&logger)

	logOutput := buf.String()
	require.Contains(t, logOutput, "Environment: test")
	require.Contains(t, logOutput, "Source: source1")
	require.Contains(t, logOutput, "Source: source2")
	require.Contains(t, logOutput, "github.com")
	require.Contains(t, logOutput, "gitlab.com")
	require.Contains(t, logOutput, "/path/to/archive")
	require.Contains(t, logOutput, "/path/to/directory")
}

func TestAuthConfig_String(t *testing.T) {
	tests := []struct {
		name     string
		config   AuthConfig
		expected string
	}{
		{
			name: "Complete auth config",
			config: AuthConfig{
				Protocol:    TLS,
				HTTPScheme:  HTTPS,
				ProxyURL:    "http://proxy",
				CertDirPath: "/certs",
				SSHCommand:  "ssh -i key",
			},
			expected: "AuthConfig: Protocol: tls, HTTPScheme: https, ProxyURL: http://proxy, CertDirPath: /certs, SSHCommand: ssh -i key",
		},
		{
			name: "Minimal auth config",
			config: AuthConfig{
				Protocol: SSH,
			},
			expected: "AuthConfig: Protocol: ssh, HTTPScheme: , ProxyURL: , CertDirPath: , SSHCommand: ",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result := tabletest.config.String()
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestMirrorConfig_ProviderTypes(t *testing.T) {
	mirrors := map[string]MirrorConfig{
		"archive": {
			BaseConfig: BaseConfig{
				ProviderType: "archive",
			},
		},
		"directory": {
			BaseConfig: BaseConfig{
				ProviderType: "directory",
			},
		},
		"git": {
			BaseConfig: BaseConfig{
				ProviderType: "github",
			},
		},
	}

	require.True(t, mirrors["archive"].IsArchive())
	require.False(t, mirrors["archive"].IsDirectory())

	require.True(t, mirrors["directory"].IsDirectory())
	require.False(t, mirrors["directory"].IsArchive())

	require.False(t, mirrors["git"].IsArchive())
	require.False(t, mirrors["git"].IsDirectory())
}
