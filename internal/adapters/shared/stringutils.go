// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package shared provides shared utilities for adapters.
package shared

import (
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/shared"
)

// StringUtilsAdapter implements the ports.StringUtils interface.
type StringUtilsAdapter struct{}

// NewStringUtilsAdapter creates a new string utils adapter.
func NewStringUtilsAdapter() ports.StringUtils {
	return &StringUtilsAdapter{}
}

// RemoveNonAlphaNumericChars removes non-alphanumeric characters from a string.
func (s *StringUtilsAdapter) RemoveNonAlphaNumericChars(str string) string {
	return shared.RemoveNonAlphaNumericChars(str)
}

// AddBasicAuthToURL adds basic authentication to a URL.
func (s *StringUtilsAdapter) AddBasicAuthToURL(urlStr, username, password string) string {
	return shared.AddBasicAuthToURL(urlStr, username, password)
}

// SanitizeURL removes sensitive information from URLs.
func (s *StringUtilsAdapter) SanitizeURL(urlStr string) string {
	return shared.SanitizeURL(urlStr)
}
