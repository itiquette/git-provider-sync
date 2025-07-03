// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package shared provides common utilities used across the application.
package shared

import (
	"context"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Regex patterns for sophisticated string cleaning.
	doubleHyphenRegex    = regexp.MustCompile(`-{2,}`)
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9-]|^-|-$`)

	// Comprehensive linebreak replacer for all Unicode linebreak types.
	linebreakReplacer = strings.NewReplacer(
		"\r\n", " ", "\r", " ", "\n", " ", "\v", " ",
		"\f", " ", "\u0085", " ", "\u2028", " ", "\u2029", " ",
	)
)

// RemoveNonAlphaNumericChars removes all non-alphanumeric characters from the input string,
// except for underscores and hyphens.
func RemoveNonAlphaNumericChars(ctx context.Context, input string) string {
	result := nonAlphanumericRegex.ReplaceAllString(input, "")
	return doubleHyphenRegex.ReplaceAllString(result, "-")
}

// RemoveLinebreaks replaces all types of linebreak characters in the input string with a space.
func RemoveLinebreaks(input string) string {
	return linebreakReplacer.Replace(input)
}

// FileNameWithoutExt removes the file extension from the given file name.
func FileNameWithoutExt(fileName string) string {
	if fileName == filepath.Ext(fileName) {
		return fileName
	}

	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// CleanString removes non-alphanumeric characters and linebreaks from the input string.
func CleanString(ctx context.Context, input string) string {
	return RemoveLinebreaks(RemoveNonAlphaNumericChars(ctx, input))
}

// AddBasicAuthToURL adds basic authentication to a URL.
func AddBasicAuthToURL(ctx context.Context, urlStr, username, password string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	parsedURL.User = url.UserPassword(username, password)
	return parsedURL.String()
}

// RemoveBasicAuthFromURL removes or masks basic authentication from a URL.
func RemoveBasicAuthFromURL(ctx context.Context, urlStr string, stripInsteadMask bool) string {
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
