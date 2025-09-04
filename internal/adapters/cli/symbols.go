// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"os"

	"itiquette/git-provider-sync/internal/adapters/terminal"
)

// Symbols provides minimal visual indicators for output
// Be idiomatic: use ASCII by default, Unicode when appropriate.
type Symbols struct {
	Check   string // Success/valid
	Cross   string // Error/invalid
	Arrow   string // Next action
	Info    string // Information
	Warning string // Warning
}

// GetSymbols returns appropriate symbols based on terminal capabilities
// Don't overengineer: just check NO_COLOR and basic Unicode support.
func GetSymbols(colorMode terminal.ColorMode) Symbols {
	// ASCII fallback for NO_COLOR or when colors are disabled
	if os.Getenv("NO_COLOR") != "" || colorMode == terminal.ColorNever {
		return getASCIISymbols()
	}

	// Use Unicode if terminal supports it
	if isUnicodeSupported() {
		return getUnicodeSymbols()
	}

	return getASCIISymbols()
}

// GetASCIISymbols returns plain ASCII symbols for maximum compatibility.
func getASCIISymbols() Symbols {
	return Symbols{
		Check:   "[OK]",
		Cross:   "[!!]",
		Arrow:   "->",
		Info:    "[i]",
		Warning: "[!]",
	}
}

// GetUnicodeSymbols returns minimal Unicode symbols for better readability.
func getUnicodeSymbols() Symbols {
	return Symbols{
		Check:   "✓",
		Cross:   "✗",
		Arrow:   "→",
		Info:    "ℹ",
		Warning: "⚠",
	}
}

// IsUnicodeSupported checks if the terminal likely supports Unicode
// Be functional: simple check based on environment.
func isUnicodeSupported() bool {
	// Check TERM for dumb terminals
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	// Check for UTF-8 locale
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}

	// Simple check for UTF-8 support
	return lang != "" && (contains(lang, "UTF-8") || contains(lang, "utf8"))
}

// Contains is a simple string contains check to avoid importing strings.
func contains(text, substr string) bool {
	if len(substr) > len(text) {
		return false
	}

	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
