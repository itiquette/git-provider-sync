// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package ports

// StringUtils defines the interface for string manipulation utilities.
// This port provides string operations needed by the domain without depending on shared packages.
type StringUtils interface {
	// RemoveNonAlphaNumericChars removes non-alphanumeric characters from a string.
	RemoveNonAlphaNumericChars(s string) string

	// AddBasicAuthToURL adds basic authentication to a URL.
	AddBasicAuthToURL(urlStr, username, password string) string

	// SanitizeURL removes sensitive information from URLs.
	SanitizeURL(url string) string
}
