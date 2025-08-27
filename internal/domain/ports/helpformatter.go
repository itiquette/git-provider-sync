// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package ports defines interfaces for hexagonal architecture adapters.
package ports

// HelpFormatter provides terminal-independent text formatting capabilities.
// This port allows the domain layer to format help text without knowing
// about specific terminal implementations.
type HelpFormatter interface {
	// Bold formats text as bold/emphasized, falling back gracefully
	Bold(text string) string

	// Section formats a section header with consistent styling
	Section(title string) string

	// IsColorSupported returns true if the terminal supports color output
	IsColorSupported() bool
}

// HelpContent represents structured help information in the domain layer.
// This is a pure data structure with no formatting logic.
type HelpContent struct {
	Title       string
	Description string
	Usage       string
	Examples    []HelpExample
	Commands    []HelpCommand
	Flags       []HelpFlag
	Support     HelpSupport
}

// HelpExample represents a usage example with description.
type HelpExample struct {
	Description string
	Command     string
}

// HelpCommand represents a subcommand in help output.
type HelpCommand struct {
	Name        string
	Description string
}

// HelpFlag represents a command-line flag in help output.
type HelpFlag struct {
	Name        string
	Description string
	IsCommon    bool // true for frequently used flags
}

// HelpSupport represents support and documentation links.
type HelpSupport struct {
	Documentation string
	Issues        string
	Discussions   string
}
