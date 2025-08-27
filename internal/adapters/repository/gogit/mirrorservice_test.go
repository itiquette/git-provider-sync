// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package gogit

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// testGoGitLogger is a simple no-op logger for testing MirrorService.
type testGoGitLogger struct{}

func (l testGoGitLogger) Trace(_ context.Context, _ string, _ map[string]interface{}) {}
func (l testGoGitLogger) Debug(_ context.Context, _ string, _ map[string]interface{}) {}
func (l testGoGitLogger) Info(_ context.Context, _ string, _ map[string]interface{})  {}
func (l testGoGitLogger) Warn(_ context.Context, _ string, _ map[string]interface{})  {}
func (l testGoGitLogger) Error(_ context.Context, _ string, _ map[string]interface{}) {}
func (l testGoGitLogger) Fatal(_ context.Context, _ string, _ map[string]interface{}) {}
func (l testGoGitLogger) IsLevelEnabled(_ ports.LogLevel) bool                        { return true }

func TestNewMirrorService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logger   ports.Logger
		tempDir  string
		expected bool
	}{
		{
			name:     "valid logger and temp dir",
			logger:   testGoGitLogger{},
			tempDir:  "/tmp/test",
			expected: true,
		},
		{
			name:     "nil logger",
			logger:   nil,
			tempDir:  "/tmp/test",
			expected: true,
		},
		{
			name:     "empty temp dir",
			logger:   testGoGitLogger{},
			tempDir:  "",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := NewMirrorService(test.logger, test.tempDir)

			assert.NotNil(t, service)
			assert.Equal(t, test.logger, service.logger)
			assert.Equal(t, test.tempDir, service.tempDir)
			assert.Nil(t, service.progressWriter)
		})
	}
}

func TestMirrorService_SetProgressWriter(t *testing.T) {
	t.Parallel()

	logger := testGoGitLogger{}

	tests := []struct {
		name   string
		writer io.Writer
	}{
		{
			name:   "set buffer writer",
			writer: &bytes.Buffer{},
		},
		{
			name:   "set nil writer",
			writer: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			testService := NewMirrorService(logger, "/tmp/test")
			testService.SetProgressWriter(test.writer)
			assert.Equal(t, test.writer, testService.progressWriter)
		})
	}
}

func TestMirrorService_shouldIncludeBranch(t *testing.T) {
	t.Parallel()

	logger := testGoGitLogger{}
	service := NewMirrorService(logger, "/tmp/test")

	tests := []struct {
		name            string
		branchName      string
		includePatterns []string
		excludePatterns []string
		expected        bool
	}{
		{
			name:            "no patterns - include all",
			branchName:      "main",
			includePatterns: []string{},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "exact match include",
			branchName:      "main",
			includePatterns: []string{"main", "develop"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "exact match exclude",
			branchName:      "temp",
			includePatterns: []string{},
			excludePatterns: []string{"temp", "test"},
			expected:        false,
		},
		{
			name:            "wildcard include match",
			branchName:      "feature/new-functionality",
			includePatterns: []string{"feature/*"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "wildcard exclude match",
			branchName:      "temp/experimental",
			includePatterns: []string{},
			excludePatterns: []string{"temp/*"},
			expected:        false,
		},
		{
			name:            "exclude takes precedence over include",
			branchName:      "feature/temp",
			includePatterns: []string{"feature/*"},
			excludePatterns: []string{"*/temp"},
			expected:        false,
		},
		{
			name:            "include pattern no match",
			branchName:      "hotfix/critical",
			includePatterns: []string{"feature/*", "develop"},
			excludePatterns: []string{},
			expected:        false,
		},
		{
			name:            "multiple include patterns - first matches",
			branchName:      "feature/auth",
			includePatterns: []string{"feature/*", "bugfix/*"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "multiple include patterns - second matches",
			branchName:      "bugfix/login-issue",
			includePatterns: []string{"feature/*", "bugfix/*"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "star pattern includes everything",
			branchName:      "any-branch-name",
			includePatterns: []string{"*"},
			excludePatterns: []string{},
			expected:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.shouldIncludeBranch(test.branchName, test.includePatterns, test.excludePatterns)
			assert.Equal(t, test.expected, result, "Branch '%s' inclusion should be %v", test.branchName, test.expected)
		})
	}
}

func TestMirrorService_shouldIncludeTag(t *testing.T) {
	t.Parallel()

	logger := testGoGitLogger{}
	service := NewMirrorService(logger, "/tmp/test")

	tests := []struct {
		name            string
		tagName         string
		includePatterns []string
		excludePatterns []string
		expected        bool
	}{
		{
			name:            "no patterns - include all",
			tagName:         "v1.0.0",
			includePatterns: []string{},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "exact match include",
			tagName:         "v1.0.0",
			includePatterns: []string{"v1.0.0", "v2.0.0"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "exact match exclude",
			tagName:         "beta",
			includePatterns: []string{},
			excludePatterns: []string{"beta", "alpha"},
			expected:        false,
		},
		{
			name:            "wildcard include match",
			tagName:         "v1.2.3",
			includePatterns: []string{"v*"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "wildcard exclude match",
			tagName:         "beta-1.0",
			includePatterns: []string{},
			excludePatterns: []string{"beta-*"},
			expected:        false,
		},
		{
			name:            "exclude takes precedence over include",
			tagName:         "v1.0.0-beta",
			includePatterns: []string{"v*"},
			excludePatterns: []string{"*-beta"},
			expected:        false,
		},
		{
			name:            "include pattern no match",
			tagName:         "release-1.0",
			includePatterns: []string{"v*", "build-*"},
			excludePatterns: []string{},
			expected:        false,
		},
		{
			name:            "multiple include patterns - matches one",
			tagName:         "build-123",
			includePatterns: []string{"v*", "build-*"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "star pattern includes everything",
			tagName:         "any-tag-name",
			includePatterns: []string{"*"},
			excludePatterns: []string{},
			expected:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.shouldIncludeTag(test.tagName, test.includePatterns, test.excludePatterns)
			assert.Equal(t, test.expected, result, "Tag '%s' inclusion should be %v", test.tagName, test.expected)
		})
	}
}

func TestMirrorService_matchesPattern(t *testing.T) {
	t.Parallel()

	logger := testGoGitLogger{}
	service := NewMirrorService(logger, "/tmp/test")

	tests := []struct {
		name     string
		input    string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			input:    "main",
			pattern:  "main",
			expected: true,
		},
		{
			name:     "exact match - no match",
			input:    "main",
			pattern:  "develop",
			expected: false,
		},
		{
			name:     "star pattern matches everything",
			input:    "anything",
			pattern:  "*",
			expected: true,
		},
		{
			name:     "prefix wildcard",
			input:    "feature/new-auth",
			pattern:  "feature/*",
			expected: true,
		},
		{
			name:     "prefix wildcard - no match",
			input:    "bugfix/login",
			pattern:  "feature/*",
			expected: false,
		},
		{
			name:     "suffix wildcard",
			input:    "test-branch",
			pattern:  "*-branch",
			expected: true,
		},
		{
			name:     "suffix wildcard - no match",
			input:    "branch-test",
			pattern:  "*-branch",
			expected: false,
		},
		{
			name:     "middle wildcard",
			input:    "v1.2.3",
			pattern:  "v*3",
			expected: true,
		},
		{
			name:     "middle wildcard - no match",
			input:    "v1.2.4",
			pattern:  "v*3",
			expected: false,
		},
		{
			name:     "multiple wildcards - not supported",
			input:    "test",
			pattern:  "*test*",
			expected: false, // Implementation only supports single wildcard split
		},
		{
			name:     "double wildcard - not supported",
			input:    "anything",
			pattern:  "**",
			expected: false, // Implementation only supports single wildcard split
		},
		{
			name:     "case sensitive match",
			input:    "Main",
			pattern:  "main",
			expected: false,
		},
		{
			name:     "case sensitive wildcard",
			input:    "Feature/test",
			pattern:  "feature/*",
			expected: false,
		},
		{
			name:     "complex branch name with wildcard",
			input:    "feature/JIRA-123/user-authentication",
			pattern:  "feature/*",
			expected: true,
		},
		{
			name:     "version tag with simple wildcard",
			input:    "v1.2.3-alpha.1",
			pattern:  "v*",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.matchesPattern(test.input, test.pattern)
			assert.Equal(t, test.expected, result, "Pattern '%s' matching '%s' should be %v", test.pattern, test.input, test.expected)
		})
	}
}

func TestGenerateRandomID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "generates non-empty ID",
			description: "The function should return a non-empty string",
		},
		{
			name:        "generates consistent ID for same process",
			description: "Multiple calls in same process should return same PID-based ID",
		},
		{
			name:        "generates numeric ID",
			description: "The ID should be a valid integer (PID)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			generatedID := generateRandomID()

			// Test non-empty
			assert.NotEmpty(t, generatedID, "Generated ID should not be empty")

			// Test that it's a valid integer (since it's based on PID)
			parsedID, err := strconv.Atoi(generatedID)
			require.NoError(t, err, "Generated ID should be a valid integer")
			assert.Positive(t, parsedID, "Generated ID should be positive")

			// Test consistency - multiple calls should return same value in same process
			secondID := generateRandomID()
			assert.Equal(t, generatedID, secondID, "Multiple calls should return same PID-based ID")

			t.Logf("Generated ID: %s (as integer: %d)", generatedID, parsedID)
		})
	}
}

func TestGenerateRandomID_Properties(t *testing.T) {
	t.Parallel()

	// Test properties of the generated ID
	randomID := generateRandomID()

	// Verify it's suitable for directory names (no special characters)
	assert.NotContains(t, randomID, "/", "ID should not contain path separators")
	assert.NotContains(t, randomID, "\\", "ID should not contain backslashes")
	assert.NotContains(t, randomID, " ", "ID should not contain spaces")
	assert.NotContains(t, randomID, "\n", "ID should not contain newlines")
	assert.NotContains(t, randomID, "\t", "ID should not contain tabs")

	// Verify it's not too long for filesystem limits
	assert.LessOrEqual(t, len(randomID), 255, "ID should not exceed typical filename length limits")

	// Verify it starts with a digit (since it's a PID)
	assert.True(t, randomID[0] >= '0' && randomID[0] <= '9', "ID should start with a digit")
}

func TestMirrorService_EdgeCases(t *testing.T) {
	t.Parallel()

	logger := testGoGitLogger{}
	service := NewMirrorService(logger, "/tmp/test")

	t.Run("shouldIncludeBranch with empty branch name", func(t *testing.T) {
		t.Parallel()

		result := service.shouldIncludeBranch("", []string{"*"}, []string{})
		assert.True(t, result, "Empty branch name should match star pattern")
	})

	t.Run("shouldIncludeTag with empty tag name", func(t *testing.T) {
		t.Parallel()

		result := service.shouldIncludeTag("", []string{"*"}, []string{})
		assert.True(t, result, "Empty tag name should match star pattern")
	})

	t.Run("matchesPattern with empty strings", func(t *testing.T) {
		t.Parallel()

		result1 := service.matchesPattern("", "")
		assert.True(t, result1, "Empty string should match empty pattern")

		result2 := service.matchesPattern("", "*")
		assert.True(t, result2, "Empty string should match star pattern")

		result3 := service.matchesPattern("test", "")
		assert.False(t, result3, "Non-empty string should not match empty pattern")
	})

	t.Run("matchesPattern with complex wildcard patterns", func(t *testing.T) {
		t.Parallel()

		// Test pattern with multiple asterisks (only first two parts are used)
		result := service.matchesPattern("prefix-middle-suffix", "prefix-*-suffix")
		assert.True(t, result, "Should match prefix and suffix pattern")

		// Test pattern that splits into more than 2 parts (only uses first split)
		result2 := service.matchesPattern("a-b-c-d", "a-*-c-*")
		assert.False(t, result2, "Complex patterns with multiple wildcards are not supported")
	})
}
