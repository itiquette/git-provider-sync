// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package terminal

import "os"

// ColorMode represents when to use colors in output
// Be idiomatic: use standard naming from tools like ls, grep.
type ColorMode string

const (
	// ColorAuto enables colors only when output is to a TTY (default).
	ColorAuto ColorMode = "auto"
	// ColorAlways forces colors even when piping/redirecting.
	ColorAlways ColorMode = "always"
	// ColorNever disables colors completely.
	ColorNever ColorMode = "never"
)

// Color provides minimal ANSI color codes for terminal output
// Be functional and idiomatic: immutable color codes determined at startup.
type Color struct {
	Red   string
	Green string
	Bold  string
	Reset string
}

// NewColor creates color codes based on color mode and TTY detection
// Returns empty strings if colors should be disabled
// Don't overengineer: just 3 colors for essential highlighting.
func NewColor(mode ColorMode, isTTY bool) Color {
	if !shouldUseColor(mode, isTTY) {
		return Color{} // No colors
	}

	return Color{
		Red:   "\033[31m",
		Green: "\033[32m",
		Bold:  "\033[1m",
		Reset: "\033[0m",
	}
}

// ShouldUseColor determines if colors should be used
// Be functional: pure function with explicit inputs
// Respects NO_COLOR environment variable per https://no-color.org
func shouldUseColor(mode ColorMode, isTTY bool) bool {
	// NO_COLOR takes precedence over everything
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// TERM=dumb means no color support
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto, "": // Empty string defaults to auto
		return isTTY
	default:
		return isTTY // Unknown modes default to auto behavior
	}
}

// Success formats a string in green (if TTY).
func (c Color) Success(s string) string {
	if c.Green == "" {
		return s
	}

	return c.Green + s + c.Reset
}

// Error formats a string in red (if TTY).
func (c Color) Error(s string) string {
	if c.Red == "" {
		return s
	}

	return c.Red + s + c.Reset
}

// Header formats a string in bold (if TTY).
func (c Color) Header(s string) string {
	if c.Bold == "" {
		return s
	}

	return c.Bold + s + c.Reset
}
