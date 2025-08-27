// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProjectOption(t *testing.T) {
	t.Parallel()

	option := NewProjectOption()

	assert.NotNil(t, option)
	assert.Empty(t, option.Description)
	assert.False(t, option.Disabled)
	assert.Empty(t, option.Visibility)
}

func TestProjectOption_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		option   ProjectOption
		contains []string
	}{
		{
			name: "complete configuration",
			option: ProjectOption{
				Description: "Test project description",
				Disabled:    true,
				Visibility:  "private",
			},
			contains: []string{
				"ProjectOption:",
				"Type: Test project description",
				"Disabled: true",
				"Visibility: private",
			},
		},
		{
			name: "minimal configuration",
			option: ProjectOption{
				Description: "Simple project",
				Disabled:    false,
				Visibility:  "public",
			},
			contains: []string{
				"ProjectOption:",
				"Type: Simple project",
				"Disabled: false",
				"Visibility: public",
			},
		},
		{
			name:   "default configuration",
			option: ProjectOption{},
			contains: []string{
				"ProjectOption:",
				"Type: ",
				"Disabled: false",
				"Visibility: ",
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

func TestProjectOption_Fields(t *testing.T) {
	t.Parallel()

	option := ProjectOption{
		Description: "Test description",
		Disabled:    true,
		Visibility:  "internal",
	}

	assert.Equal(t, "Test description", option.Description)
	assert.True(t, option.Disabled)
	assert.Equal(t, "internal", option.Visibility)
}

func TestProjectOption_BooleanValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		disabled bool
		expected string
	}{
		{
			name:     "disabled true",
			disabled: true,
			expected: "Disabled: true",
		},
		{
			name:     "disabled false",
			disabled: false,
			expected: "Disabled: false",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			option := ProjectOption{Disabled: test.disabled}
			result := option.String()

			assert.Contains(t, result, test.expected)
		})
	}
}
