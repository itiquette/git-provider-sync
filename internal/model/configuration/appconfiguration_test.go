// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaseConfig_GetDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   BaseConfig
		expected string
	}{
		{
			name: "Custom domain overrides default",
			config: BaseConfig{
				ProviderType: "gitlab",
				Domain:       "custom.gitlab.com",
			},
			expected: "custom.gitlab.com",
		},
		{
			name: "Unknown provider returns empty",
			config: BaseConfig{
				ProviderType: "unknown",
			},
			expected: "",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			t.Parallel()

			result := tabletest.config.GetDomain()
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestBaseConfig_FillDefaults(t *testing.T) {
	t.Parallel()

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
					GitTimeout:     300,
					HTTPTimeout:    30,
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
					GitTimeout:     300, // Default is set
					HTTPTimeout:    30,  // Default is set
				},
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			t.Parallel()

			tabletest.config.FillDefaults()
			require.Equal(t, tabletest.expected, tabletest.config)
		})
	}
}
