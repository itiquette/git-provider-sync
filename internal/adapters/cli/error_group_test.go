// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorGroup_SingleError(t *testing.T) {
	t.Parallel()

	errorGroup := NewErrorGroup("Clone")
	errorGroup.Add("repo1", errors.New("authentication failed"))

	symbols := getASCIISymbols()
	output := errorGroup.Format(symbols)

	assert.Contains(t, output, "[!!] Clone failed: repo1")
	assert.Contains(t, output, "Reason: authentication failed")
}

func TestErrorGroup_MultipleAuthErrors(t *testing.T) {
	t.Parallel()

	errorGroup := NewErrorGroup("Clone")
	errorGroup.Add("repo1", errors.New("401 unauthorized"))
	errorGroup.Add("repo2", errors.New("authentication failed"))
	errorGroup.Add("repo3", errors.New("403 forbidden"))
	errorGroup.Add("repo4", errors.New("401 unauthorized"))
	errorGroup.Add("repo5", errors.New("authentication failed"))

	symbols := getASCIISymbols()
	output := errorGroup.Format(symbols)

	// Should group as authentication problems
	assert.Contains(t, output, "[!!] Clone failed for 5 items")
	assert.Contains(t, output, "Authentication problem (5 failed)")
	assert.Contains(t, output, "repo1")
	assert.Contains(t, output, "repo2")
	assert.Contains(t, output, "repo3")
	assert.Contains(t, output, "... and 2 more")
	assert.Contains(t, output, "-> Fix authentication:")
}

func TestErrorGroup_MixedErrors(t *testing.T) {
	t.Parallel()

	errorGroup := NewErrorGroup("Sync")
	errorGroup.Add("repo1", errors.New("401 unauthorized"))
	errorGroup.Add("repo2", errors.New("connection timeout"))
	errorGroup.Add("repo3", errors.New("404 not found"))
	errorGroup.Add("repo4", errors.New("401 unauthorized"))

	symbols := getASCIISymbols()
	output := errorGroup.Format(symbols)

	// Should show different error categories
	assert.Contains(t, output, "[!!] Sync failed for 4 items")
	assert.Contains(t, output, "Authentication problem (2 failed)")
	assert.Contains(t, output, "Network problem (1 failed)")
	assert.Contains(t, output, "Not found (1 failed)")
}

func TestErrorGroup_NoErrors(t *testing.T) {
	t.Parallel()

	errorGroup := NewErrorGroup("Push")
	symbols := getASCIISymbols()
	output := errorGroup.Format(symbols)

	assert.Empty(t, output)
	assert.False(t, errorGroup.HasErrors())
	assert.Equal(t, 0, errorGroup.Count())
}

func TestErrorGroup_ErrorInterface(t *testing.T) {
	t.Parallel()

	errorGroup := NewErrorGroup("Fetch")
	errorGroup.Add("repo1", errors.New("test error"))
	errorGroup.Add("repo2", errors.New("another error"))

	// Should implement error interface
	var err error = errorGroup
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Fetch failed for 2 items")
}

func TestErrorGroup_Classification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		errorMsg      string
		expectedClass string
	}{
		{"auth 401", "401 unauthorized", "Authentication problem"},
		{"auth 403", "403 forbidden", "Authentication problem"},
		{"auth explicit", "authentication failed", "Authentication problem"},
		{"network timeout", "connection timeout", "Network problem"},
		{"network refused", "connection refused", "Network problem"},
		{"network host", "no such host", "Network problem"},
		{"not found 404", "404 not found", "Not found"},
		{"not found explicit", "repository does not exist", "Not found"},
		{"permission", "permission denied", "Permission denied"},
		{"rate limit", "rate limit exceeded", "Rate limited"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := classifyError(errors.New(tt.errorMsg))
			assert.Equal(t, tt.expectedClass, result)
		})
	}
}

func TestErrorGroup_GetErrors(t *testing.T) {
	t.Parallel()

	errorGroup := NewErrorGroup("Test")
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	errorGroup.Add("item1", err1)
	errorGroup.Add("item2", err2)

	errs := errorGroup.GetErrors()
	require.Len(t, errs, 2)
	assert.Contains(t, errs[0].Error(), "error 1")
	assert.Contains(t, errs[1].Error(), "error 2")
}

func TestErrorGroup_Suggestions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errors   []error
		wantHelp string
	}{
		{
			name: "auth errors get auth help",
			errors: []error{
				errors.New("401 unauthorized"),
				errors.New("authentication failed"),
			},
			wantHelp: "Check token:",
		},
		{
			name: "network errors get network help",
			errors: []error{
				errors.New("connection timeout"),
				errors.New("network unreachable"),
			},
			wantHelp: "Check internet connection",
		},
		{
			name: "rate limit gets specific help",
			errors: []error{
				errors.New("rate limit exceeded"),
				errors.New("API rate limit"),
			},
			wantHelp: "Wait a few minutes",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			errorGroup := NewErrorGroup("Operation")
			for i, err := range testCase.errors {
				errorGroup.Add(strings.ReplaceAll(testCase.name, " ", "-")+string(rune(i)), err)
			}

			symbols := getASCIISymbols()
			output := errorGroup.Format(symbols)

			assert.Contains(t, output, testCase.wantHelp)
		})
	}
}
