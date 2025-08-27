// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSHClientOption_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		option      SSHClientOption
		contains    []string
		notContains []string
	}{
		{
			name: "complete configuration",
			option: SSHClientOption{
				SSHCommand:        "ssh -i ~/.ssh/id_rsa",
				RewriteSSHURLFrom: "git@github.com:",
				RewriteSSHURLTo:   "https://github.com/",
			},
			contains: []string{
				"SSHClientOption{",
				"SSHCommand: ssh -i ~/.ssh/id_rsa",
				"RewriteSSHURLFrom: git@github.com:",
				"RewriteSSHURLTo: https://github.com/",
				"}",
			},
		},
		{
			name: "only SSH command",
			option: SSHClientOption{
				SSHCommand: "ssh -o StrictHostKeyChecking=no",
			},
			contains: []string{
				"SSHClientOption{",
				"SSHCommand: ssh -o StrictHostKeyChecking=no",
				"}",
			},
			notContains: []string{
				"RewriteSSHURLFrom:",
				"RewriteSSHURLTo:",
			},
		},
		{
			name: "only URL rewrite configuration",
			option: SSHClientOption{
				RewriteSSHURLFrom: "git@gitlab.com:",
				RewriteSSHURLTo:   "https://gitlab.com/",
			},
			contains: []string{
				"SSHClientOption{",
				"RewriteSSHURLFrom: git@gitlab.com:",
				"RewriteSSHURLTo: https://gitlab.com/",
				"}",
			},
			notContains: []string{
				"SSHCommand:",
			},
		},
		{
			name:   "empty configuration",
			option: SSHClientOption{},
			contains: []string{
				"SSHClientOption{",
				"}",
			},
			notContains: []string{
				"SSHCommand:",
				"RewriteSSHURLFrom:",
				"RewriteSSHURLTo:",
			},
		},
		{
			name: "partial configuration - only rewrite from",
			option: SSHClientOption{
				RewriteSSHURLFrom: "git@bitbucket.org:",
			},
			contains: []string{
				"SSHClientOption{",
				"RewriteSSHURLFrom: git@bitbucket.org:",
				"}",
			},
			notContains: []string{
				"SSHCommand:",
				"RewriteSSHURLTo:",
			},
		},
		{
			name: "partial configuration - only rewrite to",
			option: SSHClientOption{
				RewriteSSHURLTo: "https://bitbucket.org/",
			},
			contains: []string{
				"SSHClientOption{",
				"RewriteSSHURLTo: https://bitbucket.org/",
				"}",
			},
			notContains: []string{
				"SSHCommand:",
				"RewriteSSHURLFrom:",
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

			for _, notExpected := range test.notContains {
				assert.NotContains(t, result, notExpected)
			}
		})
	}
}

func TestSSHClientOption_Fields(t *testing.T) {
	t.Parallel()

	option := SSHClientOption{
		SSHCommand:        "custom-ssh",
		RewriteSSHURLFrom: "from-url",
		RewriteSSHURLTo:   "to-url",
	}

	assert.Equal(t, "custom-ssh", option.SSHCommand)
	assert.Equal(t, "from-url", option.RewriteSSHURLFrom)
	assert.Equal(t, "to-url", option.RewriteSSHURLTo)
}

func TestSSHClientOption_StringFormatting(t *testing.T) {
	t.Parallel()

	// Test that the string formatting uses proper spacing
	option := SSHClientOption{
		SSHCommand:        "ssh",
		RewriteSSHURLFrom: "from",
		RewriteSSHURLTo:   "to",
	}

	result := option.String()

	// Should have proper spacing between parts
	assert.Contains(t, result, "SSHClientOption{ SSHCommand: ssh RewriteSSHURLFrom: from RewriteSSHURLTo: to }")
}

func TestSSHClientOption_EmptyStrings(t *testing.T) {
	t.Parallel()

	// Test with empty strings (not nil)
	option := SSHClientOption{
		SSHCommand:        "",
		RewriteSSHURLFrom: "",
		RewriteSSHURLTo:   "",
	}

	result := option.String()

	// Empty strings should not be included in output
	assert.Equal(t, "SSHClientOption{ }", result)
}
