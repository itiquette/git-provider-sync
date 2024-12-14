// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"testing"

	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestGetVisibility(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name     string
		input    gitlab.VisibilityValue
		expected string
	}{
		{
			name:     "public visibility",
			input:    gitlab.PublicVisibility,
			expected: PUBLIC,
		},
		{
			name:     "private visibility",
			input:    gitlab.PrivateVisibility,
			expected: "private",
		},
		{
			name:     "internal visibility",
			input:    gitlab.InternalVisibility,
			expected: "internal",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(_ *testing.T) {
			require.Equal(tabletest.expected, getVisibility(tabletest.input))
		})
	}
}

func TestToVisibility(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name     string
		input    string
		expected gitlab.VisibilityValue
	}{
		{
			name:     "private visibility",
			input:    "private",
			expected: gitlab.PrivateVisibility,
		},
		{
			name:     "internal visibility",
			input:    "internal",
			expected: gitlab.InternalVisibility,
		},
		{
			name:     "public visibility",
			input:    PUBLIC,
			expected: gitlab.PublicVisibility,
		},
		{
			name:     "unknown visibility defaults to internal",
			input:    "unknown",
			expected: gitlab.InternalVisibility,
		},
		{
			name:     "empty string defaults to internal",
			input:    "",
			expected: gitlab.InternalVisibility,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(_ *testing.T) {
			require.Equal(tabletest.expected, toVisibility(tabletest.input))
		})
	}
}
