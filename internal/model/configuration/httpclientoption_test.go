// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClientOption_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		option   HTTPClientOption
		contains []string
	}{
		{
			name: "complete configuration",
			option: HTTPClientOption{
				Scheme:      "https",
				ProxyURL:    "http://proxy.example.com:8080",
				Token:       "secret-token-123",
				CertDirPath: "/path/to/certs",
			},
			contains: []string{
				"HTTPClientOption:",
				"ProxyURL http://proxy.example.com:8080",
				"Token: <****>",
			},
		},
		{
			name: "minimal configuration",
			option: HTTPClientOption{
				Scheme: "http",
			},
			contains: []string{
				"HTTPClientOption:",
				"ProxyURL ",
				"Token: <****>",
			},
		},
		{
			name:   "empty configuration",
			option: HTTPClientOption{},
			contains: []string{
				"HTTPClientOption:",
				"Token: <****>",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := test.option.String()

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestMaskToken(t *testing.T) {
	t.Parallel()

	result := maskToken()
	assert.Equal(t, "<****>", result)
}

func TestHTTPClientOption_Fields(t *testing.T) {
	t.Parallel()

	option := HTTPClientOption{
		Scheme:      "https",
		ProxyURL:    "http://proxy.example.com:8080",
		Token:       "secret-token-123",
		CertDirPath: "/path/to/certs",
	}

	assert.Equal(t, "https", option.Scheme)
	assert.Equal(t, "http://proxy.example.com:8080", option.ProxyURL)
	assert.Equal(t, "secret-token-123", option.Token)
	assert.Equal(t, "/path/to/certs", option.CertDirPath)
}
