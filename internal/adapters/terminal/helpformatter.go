// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package terminal provides terminal-specific adapters for the hexagonal architecture.
package terminal

import (
	"strings"

	"github.com/fatih/color"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// HelpFormatter implements terminal-specific help text formatting.
// This adapter handles color support detection and graceful fallbacks.
type HelpFormatter struct {
	boldFormatter *color.Color
	colorEnabled  bool
}

// NewHelpFormatter creates a new terminal-aware help formatter.
func NewHelpFormatter() *HelpFormatter {
	return &HelpFormatter{
		boldFormatter: color.New(color.Bold),
		colorEnabled:  !color.NoColor,
	}
}

// NewHelpFormatterWithColorSupport creates a formatter with explicit color control.
func NewHelpFormatterWithColorSupport(colorEnabled bool) *HelpFormatter {
	formatter := &HelpFormatter{
		boldFormatter: color.New(color.Bold),
		colorEnabled:  colorEnabled,
	}

	// Configure the formatter based on color support
	// We don't modify global color.NoColor here to avoid test interference
	if !colorEnabled {
		formatter.boldFormatter.DisableColor()
	} else {
		formatter.boldFormatter.EnableColor()
	}

	return formatter
}

// Bold formats text as bold/emphasized with graceful fallback.
// When color is disabled, it returns the text in uppercase for emphasis.
// This is a pure function with consistent behavior.
func (f *HelpFormatter) Bold(text string) string {
	// Handle empty strings to avoid unnecessary ANSI sequences
	if text == "" {
		return ""
	}

	if f.colorEnabled {
		return f.boldFormatter.Sprint(text)
	}
	// Graceful fallback: uppercase for emphasis in no-color terminals
	return strings.ToUpper(text)
}

// Section formats a section header with consistent styling.
// Section creates visually distinct section headers that work across different terminal types.
func (f *HelpFormatter) Section(title string) string {
	return f.Bold(title)
}

// IsColorSupported returns true if the terminal supports color output.
// This allows callers to make formatting decisions based on capabilities.
func (f *HelpFormatter) IsColorSupported() bool {
	return f.colorEnabled
}

// Compile-time interface compliance check.
var _ ports.HelpFormatter = (*HelpFormatter)(nil)
