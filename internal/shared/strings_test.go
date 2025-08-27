// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package shared

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:dupl // Table-driven tests for different functions have similar structure
func TestRemoveNonAlphaNumericChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "alphanumeric string unchanged",
			input:    "hello123",
			expected: "hello123",
		},
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: "helloworld",
		},
		{
			name:     "string with special characters",
			input:    "hello@world#123!",
			expected: "helloworld123",
		},
		{
			name:     "string with hyphens preserved",
			input:    "hello-world-123",
			expected: "hello-world-123",
		},
		{
			name:     "string with multiple consecutive hyphens",
			input:    "hello----world",
			expected: "hello-world",
		},
		{
			name:     "string starting with hyphen",
			input:    "-hello-world",
			expected: "hello-world",
		},
		{
			name:     "string ending with hyphen",
			input:    "hello-world-",
			expected: "hello-world",
		},
		{
			name:     "string with underscores removed",
			input:    "hello_world_123",
			expected: "helloworld123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "@#$%^&*()",
			expected: "",
		},
		{
			name:     "mixed case with numbers and symbols",
			input:    "Hello123-World@456!",
			expected: "Hello123-World456",
		},
		{
			name:     "unicode characters",
			input:    "héllo-wörld",
			expected: "hllo-wrld",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := RemoveNonAlphaNumericChars(testCase.input)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

//nolint:dupl // Table-driven tests for different functions have similar structure
func TestRemoveLinebreaks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no linebreaks",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "single newline",
			input:    "hello\nworld",
			expected: "hello world",
		},
		{
			name:     "carriage return newline",
			input:    "hello\r\nworld",
			expected: "hello world",
		},
		{
			name:     "multiple types of linebreaks",
			input:    "hello\rworld\ntest\vvalue",
			expected: "hello world test value",
		},
		{
			name:     "form feed character",
			input:    "hello\fworld",
			expected: "hello world",
		},
		{
			name:     "unicode line separator",
			input:    "hello\u2028world",
			expected: "hello world",
		},
		{
			name:     "unicode paragraph separator",
			input:    "hello\u2029world",
			expected: "hello world",
		},
		{
			name:     "next line character",
			input:    "hello\u0085world",
			expected: "hello world",
		},
		{
			name:     "multiple consecutive linebreaks",
			input:    "hello\n\n\nworld",
			expected: "hello   world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only linebreaks",
			input:    "\n\r\v\f",
			expected: "    ",
		},
		{
			name:     "mixed content with all linebreak types",
			input:    "line1\nline2\rline3\r\nline4\vline5\fline6\u0085line7\u2028line8\u2029line9",
			expected: "line1 line2 line3 line4 line5 line6 line7 line8 line9",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := RemoveLinebreaks(testCase.input)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestAddBasicAuthToURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		url      string
		username string
		password string
		expected string
	}{
		{
			name:     "simple HTTP URL",
			url:      "http://example.com",
			username: "user",
			password: "pass",
			expected: "http://user:pass@example.com",
		},
		{
			name:     "HTTPS URL with path",
			url:      "https://api.github.com/repos",
			username: "myuser",
			password: "mytoken",
			expected: "https://myuser:mytoken@api.github.com/repos",
		},
		{
			name:     "URL with port",
			url:      "http://localhost:8080/api",
			username: "admin",
			password: "secret",
			expected: "http://admin:secret@localhost:8080/api",
		},
		{
			name:     "URL with query parameters",
			url:      "https://example.com/search?q=test&limit=10",
			username: "user",
			password: "pass",
			expected: "https://user:pass@example.com/search?q=test&limit=10",
		},
		{
			name:     "URL with fragment",
			url:      "https://example.com/page#section",
			username: "user",
			password: "pass",
			expected: "https://user:pass@example.com/page#section",
		},
		{
			name:     "invalid URL gets modified anyway",
			url:      "not-a-valid-url",
			username: "user",
			password: "pass",
			expected: "//user:pass@not-a-valid-url",
		},
		{
			name:     "empty username and password",
			url:      "https://example.com",
			username: "",
			password: "",
			expected: "https://:@example.com",
		},
		{
			name:     "special characters in credentials",
			url:      "https://example.com",
			username: "user@domain",
			password: "p@ss:w0rd",
			expected: "https://user%40domain:p%40ss%3Aw0rd@example.com",
		},
		{
			name:     "URL already has auth (gets replaced)",
			url:      "https://olduser:oldpass@example.com",
			username: "newuser",
			password: "newpass",
			expected: "https://newuser:newpass@example.com",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := AddBasicAuthToURL(testCase.url, testCase.username, testCase.password)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestRemoveBasicAuthFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		url              string
		stripInsteadMask bool
		expected         string
	}{
		{
			name:             "URL with auth - strip mode",
			url:              "https://user:pass@example.com",
			stripInsteadMask: true,
			expected:         "https://example.com",
		},
		{
			name:             "URL with auth - mask mode",
			url:              "https://user:pass@example.com",
			stripInsteadMask: false,
			expected:         "https://user:SECRET@example.com",
		},
		{
			name:             "URL without auth - strip mode",
			url:              "https://example.com",
			stripInsteadMask: true,
			expected:         "https://example.com",
		},
		{
			name:             "URL without auth - mask mode",
			url:              "https://example.com",
			stripInsteadMask: false,
			expected:         "https://example.com",
		},
		{
			name:             "URL with username only - strip mode",
			url:              "https://user@example.com",
			stripInsteadMask: true,
			expected:         "https://example.com",
		},
		{
			name:             "URL with username only - mask mode",
			url:              "https://user@example.com",
			stripInsteadMask: false,
			expected:         "https://user@example.com",
		},
		{
			name:             "URL with auth and path - strip mode",
			url:              "https://user:token@api.github.com/repos/owner/repo",
			stripInsteadMask: true,
			expected:         "https://api.github.com/repos/owner/repo",
		},
		{
			name:             "URL with auth and path - mask mode",
			url:              "https://user:token@api.github.com/repos/owner/repo",
			stripInsteadMask: false,
			expected:         "https://user:SECRET@api.github.com/repos/owner/repo",
		},
		{
			name:             "URL with auth and query params - strip mode",
			url:              "https://user:pass@example.com/api?version=v1&format=json",
			stripInsteadMask: true,
			expected:         "https://example.com/api?version=v1&format=json",
		},
		{
			name:             "URL with auth and query params - mask mode",
			url:              "https://user:pass@example.com/api?version=v1&format=json",
			stripInsteadMask: false,
			expected:         "https://user:SECRET@example.com/api?version=v1&format=json",
		},
		{
			name:             "invalid URL returns original",
			url:              "not-a-valid-url",
			stripInsteadMask: true,
			expected:         "not-a-valid-url",
		},
		{
			name:             "URL with port and auth - strip mode",
			url:              "http://admin:secret@localhost:8080/dashboard",
			stripInsteadMask: true,
			expected:         "http://localhost:8080/dashboard",
		},
		{
			name:             "URL with port and auth - mask mode",
			url:              "http://admin:secret@localhost:8080/dashboard",
			stripInsteadMask: false,
			expected:         "http://admin:SECRET@localhost:8080/dashboard",
		},
		{
			name:             "URL with special characters in auth - strip mode",
			url:              "https://user%40domain:p%40ss%3Aw0rd@example.com",
			stripInsteadMask: true,
			expected:         "https://example.com",
		},
		{
			name:             "URL with special characters in auth - mask mode",
			url:              "https://user%40domain:p%40ss%3Aw0rd@example.com",
			stripInsteadMask: false,
			expected:         "https://user%40domain:SECRET@example.com",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := RemoveBasicAuthFromURL(testCase.url, testCase.stripInsteadMask)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestRegexPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "doubleHyphenRegex matches multiple hyphens",
			testFunc: func(t *testing.T) {
				t.Helper()
				testCases := []struct {
					input    string
					expected bool
				}{
					{"hello--world", true},
					{"hello---world", true},
					{"hello----world", true},
					{"hello-world", false},
					{"helloworld", false},
					{"--hello", true},
					{"world--", true},
				}

				for _, tc := range testCases {
					matches := doubleHyphenRegex.MatchString(tc.input)
					assert.Equal(t, tc.expected, matches, "input: %s", tc.input)
				}
			},
		},
		{
			name: "nonAlphanumericRegex matches correctly",
			testFunc: func(t *testing.T) {
				t.Helper()
				testCases := []struct {
					input    string
					expected bool
				}{
					{"hello@world", true},
					{"hello123", false},
					{"hello-world", false},
					{"-hello", true},      // leading hyphen
					{"hello-", true},      // trailing hyphen
					{"hello_world", true}, // underscore
					{"héllo", true},       // unicode
				}

				for _, tc := range testCases {
					matches := nonAlphanumericRegex.MatchString(tc.input)
					assert.Equal(t, tc.expected, matches, "input: %s", tc.input)
				}
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

func TestLinebreakReplacer(t *testing.T) {
	t.Parallel()

	// Test that the linebreak replacer handles all the expected characters
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "carriage return newline",
			input:    "hello\r\nworld",
			expected: "hello world",
		},
		{
			name:     "carriage return only",
			input:    "hello\rworld",
			expected: "hello world",
		},
		{
			name:     "newline only",
			input:    "hello\nworld",
			expected: "hello world",
		},
		{
			name:     "vertical tab",
			input:    "hello\vworld",
			expected: "hello world",
		},
		{
			name:     "form feed",
			input:    "hello\fworld",
			expected: "hello world",
		},
		{
			name:     "next line (NEL)",
			input:    "hello\u0085world",
			expected: "hello world",
		},
		{
			name:     "line separator",
			input:    "hello\u2028world",
			expected: "hello world",
		},
		{
			name:     "paragraph separator",
			input:    "hello\u2029world",
			expected: "hello world",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := linebreakReplacer.Replace(testCase.input)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestStringFunctions_LargeInputs(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping benchmarks in short mode")
	}

	// These are not actual benchmarks but test that functions work with large inputs
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "performance with large string input",
			testFunc: func(t *testing.T) {
				t.Helper()
				// Create a large string with various problematic characters
				largeInput := strings.Repeat("hello\nworld@123!test----", 1000)
				// Test that it doesn't panic or take too long
				result := RemoveLinebreaks(RemoveNonAlphaNumericChars(largeInput))
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "hello")
				assert.Contains(t, result, "world")
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
