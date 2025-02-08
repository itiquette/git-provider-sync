// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package stringconvert provides utility functions for string manipulations
// such as removing non-alphanumeric characters, handling linebreaks,
// and manipulating file names.
package stringconvert

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"itiquette/git-provider-sync/internal/log"
)

var (
	// Regex to match non-alphanumeric chars, except hyphens between alphanumerics.

	doubleHyphenRegex    = regexp.MustCompile(`-{2,}`)
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9-]|^-|-$`)
	linebreakReplacer    = strings.NewReplacer(
		"\r\n", " ", "\r", " ", "\n", " ", "\v", " ",
		"\f", " ", "\u0085", " ", "\u2028", " ", "\u2029", " ",
	)
)

// RemoveNonAlphaNumericChars removes all non-alphanumeric characters from the input string,
// except for underscores and hyphens.
func RemoveNonAlphaNumericChars(ctx context.Context, input string) string {
	result := nonAlphanumericRegex.ReplaceAllString(input, "")

	normalized := doubleHyphenRegex.ReplaceAllString(result, "-")
	log.Logger(ctx).Debug().
		Str("input", input).
		Str("result", result).
		Msg("Removed non-alphanumeric characters")

	return normalized
}

// RemoveLinebreaks replaces all types of linebreak characters in the input string with a space.
func RemoveLinebreaks(input string) string {
	return linebreakReplacer.Replace(input)
}

// FileNameWithoutExt removes the file extension from the given file name.
// If the filename is the same as the extension it just returns the name i.e .myfile.
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
