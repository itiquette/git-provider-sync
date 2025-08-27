// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncRunOption_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		option   SyncRunOption
		contains []string
	}{
		{
			name: "complete configuration with all fields",
			option: SyncRunOption{
				ForcePush:         true,
				IgnoreInvalidName: true,
				AlphaNumHyphName:  true,
				ActiveFromLimit:   "2024-01-01",
			},
			contains: []string{
				"SyncRunOption{",
				"ForcePush: true",
				"IgnoreInvalidName: true",
				"ASCIIName: true",
				"ActiveFromLimit: 2024-01-01",
				"}",
			},
		},
		{
			name: "minimal configuration with all false booleans",
			option: SyncRunOption{
				ForcePush:         false,
				IgnoreInvalidName: false,
				AlphaNumHyphName:  false,
				ActiveFromLimit:   "",
			},
			contains: []string{
				"SyncRunOption{",
				"ForcePush: false",
				"IgnoreInvalidName: false",
				"ASCIIName: false",
				"}",
			},
		},
		{
			name: "mixed boolean values with active limit",
			option: SyncRunOption{
				ForcePush:         true,
				IgnoreInvalidName: false,
				AlphaNumHyphName:  true,
				ActiveFromLimit:   "2023-12-01T00:00:00Z",
			},
			contains: []string{
				"SyncRunOption{",
				"ForcePush: true",
				"IgnoreInvalidName: false",
				"ASCIIName: true",
				"ActiveFromLimit: 2023-12-01T00:00:00Z",
				"}",
			},
		},
		{
			name: "only active from limit set",
			option: SyncRunOption{
				ForcePush:         false,
				IgnoreInvalidName: false,
				AlphaNumHyphName:  false,
				ActiveFromLimit:   "1 week ago",
			},
			contains: []string{
				"SyncRunOption{",
				"ForcePush: false",
				"IgnoreInvalidName: false",
				"ASCIIName: false",
				"ActiveFromLimit: 1 week ago",
				"}",
			},
		},
		{
			name:   "default configuration (zero values)",
			option: SyncRunOption{},
			contains: []string{
				"SyncRunOption{",
				"ForcePush: false",
				"IgnoreInvalidName: false",
				"ASCIIName: false",
				"}",
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

func TestSyncRunOption_Fields(t *testing.T) {
	t.Parallel()

	option := SyncRunOption{
		ForcePush:         true,
		IgnoreInvalidName: false,
		AlphaNumHyphName:  true,
		ActiveFromLimit:   "2024-06-01",
	}

	assert.True(t, option.ForcePush)
	assert.False(t, option.IgnoreInvalidName)
	assert.True(t, option.AlphaNumHyphName)
	assert.Equal(t, "2024-06-01", option.ActiveFromLimit)
}

func TestSyncRunOption_BooleanCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		forcePush        bool
		ignoreInvalid    bool
		alphaNumHyph     bool
		expectedContains []string
	}{
		{
			name:          "all true",
			forcePush:     true,
			ignoreInvalid: true,
			alphaNumHyph:  true,
			expectedContains: []string{
				"ForcePush: true",
				"IgnoreInvalidName: true",
				"ASCIIName: true",
			},
		},
		{
			name:          "all false",
			forcePush:     false,
			ignoreInvalid: false,
			alphaNumHyph:  false,
			expectedContains: []string{
				"ForcePush: false",
				"IgnoreInvalidName: false",
				"ASCIIName: false",
			},
		},
		{
			name:          "mixed values - force and ascii true",
			forcePush:     true,
			ignoreInvalid: false,
			alphaNumHyph:  true,
			expectedContains: []string{
				"ForcePush: true",
				"IgnoreInvalidName: false",
				"ASCIIName: true",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			option := SyncRunOption{
				ForcePush:         test.forcePush,
				IgnoreInvalidName: test.ignoreInvalid,
				AlphaNumHyphName:  test.alphaNumHyph,
			}

			result := option.String()

			for _, expected := range test.expectedContains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestSyncRunOption_ActiveFromLimitHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		activeLimit   string
		shouldContain bool
	}{
		{
			name:          "empty string should not appear",
			activeLimit:   "",
			shouldContain: false,
		},
		{
			name:          "non-empty string should appear",
			activeLimit:   "2024-01-01",
			shouldContain: true,
		},
		{
			name:          "whitespace only should appear",
			activeLimit:   "   ",
			shouldContain: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			option := SyncRunOption{
				ActiveFromLimit: test.activeLimit,
			}

			result := option.String()

			if test.shouldContain {
				assert.Contains(t, result, "ActiveFromLimit: "+test.activeLimit)
			} else {
				assert.NotContains(t, result, "ActiveFromLimit:")
			}
		})
	}
}

func TestSyncRunOption_StringFormatting(t *testing.T) {
	t.Parallel()

	option := SyncRunOption{
		ForcePush:         true,
		IgnoreInvalidName: false,
		AlphaNumHyphName:  true,
		ActiveFromLimit:   "test",
	}

	result := option.String()

	// Verify proper spacing and structure
	assert.Contains(t, result, "SyncRunOption{ ForcePush: true IgnoreInvalidName: false ASCIIName: true ActiveFromLimit: test }")
}
