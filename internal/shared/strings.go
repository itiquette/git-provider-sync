// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package shared provides common utilities used across the application.
package shared

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	// Regex patterns for string cleaning.
	doubleHyphenRegex    = regexp.MustCompile(`-{2,}`)
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9-]|^-|-$`)

	// linebreak replacer for all Unicode linebreak types.
	linebreakReplacer = strings.NewReplacer( //nolint:gochecknoglobals // Shared string processing utility
		"\r\n", " ", "\r", " ", "\n", " ", "\v", " ",
		"\f", " ", "\u0085", " ", "\u2028", " ", "\u2029", " ",
	)
)

// RemoveNonAlphaNumericChars sanitizes strings for safe repository naming operations.
func RemoveNonAlphaNumericChars(input string) string {
	result := nonAlphanumericRegex.ReplaceAllString(input, "")

	return doubleHyphenRegex.ReplaceAllString(result, "-")
}

// RemoveLinebreaks normalizes text by converting all linebreak types to spaces.
func RemoveLinebreaks(input string) string {
	return linebreakReplacer.Replace(input)
}

// AddBasicAuthToURL embeds credentials into URL for authenticated repository operations.
func AddBasicAuthToURL(urlStr, username, password string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String()
}

// RemoveBasicAuthFromURL sanitizes URLs for safe logging and display purposes.
func RemoveBasicAuthFromURL(urlStr string, stripInsteadMask bool) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	if parsedURL.User == nil {
		return parsedURL.String()
	}

	if stripInsteadMask {
		parsedURL.User = nil
	} else {
		username := parsedURL.User.Username()
		if _, hasPassword := parsedURL.User.Password(); hasPassword {
			parsedURL.User = url.UserPassword(username, "SECRET")
		}
	}

	return parsedURL.String()
}
