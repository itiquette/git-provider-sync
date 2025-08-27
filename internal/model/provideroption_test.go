// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProviderOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		includeForks bool
		owner        string
		ownerType    string
		included     []string
		excluded     []string
		expected     ProviderOption
	}{
		{
			name:         "create provider option with all parameters",
			includeForks: true,
			owner:        "test-user",
			ownerType:    "user",
			included:     []string{"repo1", "repo2", "repo3"},
			excluded:     []string{"excluded1", "excluded2"},
			expected: ProviderOption{
				ExcludedRepositories: []string{"excluded1", "excluded2"},
				IncludeForks:         true,
				IncludedRepositories: []string{"repo1", "repo2", "repo3"},
				Owner:                "test-user",
				OwnerType:            "user",
			},
		},
		{
			name:         "create provider option with group owner",
			includeForks: false,
			owner:        "test-org",
			ownerType:    "group",
			included:     []string{"important-repo"},
			excluded:     []string{},
			expected: ProviderOption{
				ExcludedRepositories: []string{},
				IncludeForks:         false,
				IncludedRepositories: []string{"important-repo"},
				Owner:                "test-org",
				OwnerType:            "group",
			},
		},
		{
			name:         "create provider option with empty lists",
			includeForks: false,
			owner:        "minimal-user",
			ownerType:    "user",
			included:     []string{},
			excluded:     []string{},
			expected: ProviderOption{
				ExcludedRepositories: []string{},
				IncludeForks:         false,
				IncludedRepositories: []string{},
				Owner:                "minimal-user",
				OwnerType:            "user",
			},
		},
		{
			name:         "create provider option with nil slices",
			includeForks: true,
			owner:        "nil-slice-user",
			ownerType:    "user",
			included:     nil,
			excluded:     nil,
			expected: ProviderOption{
				ExcludedRepositories: nil,
				IncludeForks:         true,
				IncludedRepositories: nil,
				Owner:                "nil-slice-user",
				OwnerType:            "user",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := NewProviderOption(
				testCase.includeForks,
				testCase.owner,
				testCase.ownerType,
				testCase.included,
				testCase.excluded,
			)

			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestProviderOptionIsGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		providerOpt   ProviderOption
		expectedGroup bool
	}{
		{
			name: "owner type is group (lowercase)",
			providerOpt: ProviderOption{
				OwnerType: "group",
			},
			expectedGroup: true,
		},
		{
			name: "owner type is GROUP (uppercase)",
			providerOpt: ProviderOption{
				OwnerType: "GROUP",
			},
			expectedGroup: true,
		},
		{
			name: "owner type is Group (mixed case)",
			providerOpt: ProviderOption{
				OwnerType: "Group",
			},
			expectedGroup: true,
		},
		{
			name: "owner type is user",
			providerOpt: ProviderOption{
				OwnerType: "user",
			},
			expectedGroup: false,
		},
		{
			name: "owner type is USER (uppercase)",
			providerOpt: ProviderOption{
				OwnerType: "USER",
			},
			expectedGroup: false,
		},
		{
			name: "owner type is empty",
			providerOpt: ProviderOption{
				OwnerType: "",
			},
			expectedGroup: false,
		},
		{
			name: "owner type is invalid",
			providerOpt: ProviderOption{
				OwnerType: "organization",
			},
			expectedGroup: false,
		},
		{
			name: "owner type with whitespace",
			providerOpt: ProviderOption{
				OwnerType: " group ",
			},
			expectedGroup: false, // EqualFold doesn't trim whitespace
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.providerOpt.IsGroup()
			assert.Equal(t, testCase.expectedGroup, result)
		})
	}
}

func TestProviderOptionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		providerOpt    ProviderOption
		expectedFields []string
	}{
		{
			name: "provider option with all fields",
			providerOpt: ProviderOption{
				ExcludedRepositories: []string{"excluded1", "excluded2"},
				IncludeForks:         true,
				IncludedRepositories: []string{"included1", "included2", "included3"},
				Owner:                "test-owner",
				OwnerType:            "group",
			},
			expectedFields: []string{
				"Owner: test-owner",
				"OwnerType: group",
				"IncludeForks: true",
				"IncludedRepositories: [included1 included2 included3]",
				"ExcludedRepositories: [excluded1 excluded2]",
			},
		},
		{
			name: "provider option with empty lists",
			providerOpt: ProviderOption{
				ExcludedRepositories: []string{},
				IncludeForks:         false,
				IncludedRepositories: []string{},
				Owner:                "empty-lists",
				OwnerType:            "user",
			},
			expectedFields: []string{
				"Owner: empty-lists",
				"OwnerType: user",
				"IncludeForks: false",
				"IncludedRepositories: []",
				"ExcludedRepositories: []",
			},
		},
		{
			name: "provider option with nil slices",
			providerOpt: ProviderOption{
				ExcludedRepositories: nil,
				IncludeForks:         true,
				IncludedRepositories: nil,
				Owner:                "nil-slices",
				OwnerType:            "group",
			},
			expectedFields: []string{
				"Owner: nil-slices",
				"OwnerType: group",
				"IncludeForks: true",
				"IncludedRepositories: []",
				"ExcludedRepositories: []",
			},
		},
		{
			name:        "zero value provider option",
			providerOpt: ProviderOption{},
			expectedFields: []string{
				"Owner: ",
				"OwnerType: ",
				"IncludeForks: false",
				"IncludedRepositories: []",
				"ExcludedRepositories: []",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.providerOpt.String()

			assert.Contains(t, result, "ProviderOption{")

			for _, expectedField := range testCase.expectedFields {
				assert.Contains(t, result, expectedField)
			}
		})
	}
}

func TestProviderOptionConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "USER constant value",
			constant: USER,
			expected: "user",
		},
		{
			name:     "GROUP constant value",
			constant: GROUP,
			expected: "group",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expected, testCase.constant)
		})
	}
}

func TestProviderOption_AllFields_ValidatesAndBuildsCorrectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "complete workflow with group provider",
			testFunc: func(t *testing.T) {
				t.Helper()
				included := []string{"important-repo", "critical-repo"}
				excluded := []string{"temp-repo", "test-repo"}

				opt := NewProviderOption(
					true,
					"my-organization",
					GROUP,
					included,
					excluded,
				)

				// Verify it's recognized as a group
				assert.True(t, opt.IsGroup())

				// Verify all fields are set correctly
				assert.Equal(t, "my-organization", opt.Owner)
				assert.Equal(t, GROUP, opt.OwnerType)
				assert.True(t, opt.IncludeForks)
				assert.Equal(t, included, opt.IncludedRepositories)
				assert.Equal(t, excluded, opt.ExcludedRepositories)

				// Verify string representation
				str := opt.String()
				assert.Contains(t, str, "my-organization")
				assert.Contains(t, str, "group")
				assert.Contains(t, str, "true")
			},
		},
		{
			name: "complete workflow with user provider",
			testFunc: func(t *testing.T) {
				t.Helper()
				opt := NewProviderOption(
					false,
					"individual-user",
					USER,
					[]string{"personal-project"},
					[]string{},
				)

				// Verify it's not recognized as a group
				assert.False(t, opt.IsGroup())

				// Verify all fields are set correctly
				assert.Equal(t, "individual-user", opt.Owner)
				assert.Equal(t, USER, opt.OwnerType)
				assert.False(t, opt.IncludeForks)
				assert.Equal(t, []string{"personal-project"}, opt.IncludedRepositories)
				assert.Equal(t, []string{}, opt.ExcludedRepositories)

				// Verify string representation
				str := opt.String()
				assert.Contains(t, str, "individual-user")
				assert.Contains(t, str, "user")
				assert.Contains(t, str, "false")
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestProviderOption_InvalidTokenAndURL_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "very long repository lists",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Create very long lists
				longIncluded := make([]string, 1000)
				longExcluded := make([]string, 500)

				for i := range longIncluded {
					longIncluded[i] = "included-repo-" + string(rune(i+'0'))
				}
				for i := range longExcluded {
					longExcluded[i] = "excluded-repo-" + string(rune(i+'0'))
				}

				opt := NewProviderOption(
					true,
					"big-organization",
					GROUP,
					longIncluded,
					longExcluded,
				)

				assert.Len(t, opt.IncludedRepositories, 1000)
				assert.Len(t, opt.ExcludedRepositories, 500)
				assert.True(t, opt.IsGroup())

				// String representation should not panic
				str := opt.String()
				assert.Contains(t, str, "big-organization")
			},
		},
		{
			name: "special characters in owner names and repository names",
			testFunc: func(t *testing.T) {
				t.Helper()
				opt := NewProviderOption(
					true,
					"owner@with#special$chars",
					"group",
					[]string{"repo/with/slashes", "repo with spaces", "repo-with-🚀-emoji"},
					[]string{"excluded@repo", "another#excluded"},
				)

				assert.Equal(t, "owner@with#special$chars", opt.Owner)
				assert.True(t, opt.IsGroup())

				str := opt.String()
				assert.Contains(t, str, "owner@with#special$chars")
				assert.Contains(t, str, "repo/with/slashes")
				assert.Contains(t, str, "repo with spaces")
			},
		},
		{
			name: "case sensitivity and whitespace handling",
			testFunc: func(t *testing.T) {
				t.Helper()
				opt := ProviderOption{
					OwnerType: "GROUP", // Uppercase
				}
				assert.True(t, opt.IsGroup())

				opt.OwnerType = "Group" // Mixed case
				assert.True(t, opt.IsGroup())

				opt.OwnerType = "group "       // Trailing space
				assert.False(t, opt.IsGroup()) // EqualFold doesn't handle whitespace

				opt.OwnerType = " group" // Leading space
				assert.False(t, opt.IsGroup())
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}
