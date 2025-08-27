// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoriesOption_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		option   RepositoriesOption
		contains []string
	}{
		{
			name: "with exclude and include lists",
			option: RepositoriesOption{
				Exclude: []string{"repo1", "repo2", "test-*"},
				Include: []string{"important-*", "main", "develop"},
			},
			contains: []string{
				"RepositoryOption:",
				"Exclude [repo1 repo2 test-*]",
				"Include: [important-* main develop]",
			},
		},
		{
			name: "only exclude list",
			option: RepositoriesOption{
				Exclude: []string{"unwanted", "legacy"},
				Include: []string{},
			},
			contains: []string{
				"RepositoryOption:",
				"Exclude [unwanted legacy]",
				"Include: []",
			},
		},
		{
			name: "only include list",
			option: RepositoriesOption{
				Exclude: []string{},
				Include: []string{"wanted", "important"},
			},
			contains: []string{
				"RepositoryOption:",
				"Exclude []",
				"Include: [wanted important]",
			},
		},
		{
			name: "both lists empty",
			option: RepositoriesOption{
				Exclude: []string{},
				Include: []string{},
			},
			contains: []string{
				"RepositoryOption:",
				"Exclude []",
				"Include: []",
			},
		},
		{
			name: "nil slices",
			option: RepositoriesOption{
				Exclude: nil,
				Include: nil,
			},
			contains: []string{
				"RepositoryOption:",
				"Exclude []",
				"Include: []",
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

func TestRepositoriesOption_Fields(t *testing.T) {
	t.Parallel()

	option := RepositoriesOption{
		Exclude: []string{"exclude1", "exclude2"},
		Include: []string{"include1", "include2", "include3"},
	}

	assert.Equal(t, []string{"exclude1", "exclude2"}, option.Exclude)
	assert.Equal(t, []string{"include1", "include2", "include3"}, option.Include)
}

func TestRepositoriesOption_EmptyLists(t *testing.T) {
	t.Parallel()

	option := RepositoriesOption{}

	assert.Nil(t, option.Exclude)
	assert.Nil(t, option.Include)

	// Test that String() handles nil slices gracefully
	result := option.String()
	assert.Contains(t, result, "RepositoryOption:")
}

func TestRepositoriesOption_SingleElements(t *testing.T) {
	t.Parallel()

	option := RepositoriesOption{
		Exclude: []string{"single-exclude"},
		Include: []string{"single-include"},
	}

	result := option.String()

	assert.Contains(t, result, "Exclude [single-exclude]")
	assert.Contains(t, result, "Include: [single-include]")
}
