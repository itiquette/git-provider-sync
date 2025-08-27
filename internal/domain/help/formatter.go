// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package help provides pure functional help text formatting services.
package help

import (
	"fmt"
	"sort"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Service provides pure functional help formatting operations.
// All functions in this service are pure - they have no side effects
// and produce the same output for the same input.
type Service struct {
	formatter ports.HelpFormatter
}

// NewService creates a new help formatting service.
// This is a pure constructor function.
func NewService(formatter ports.HelpFormatter) Service {
	return Service{formatter: formatter}
}

// FormatHelp converts structured help content into formatted text.
// This is a pure function that performs no I/O or side effects.
func (s Service) FormatHelp(content ports.HelpContent) string {
	var sections []string

	// Title and description
	if content.Title != "" {
		sections = append(sections, content.Title)
	}

	if content.Description != "" {
		sections = append(sections, content.Description)
	}

	// Usage section
	if content.Usage != "" {
		sections = append(sections, s.formatUsageSection(content.Usage))
	}

	// Quick start examples
	if len(content.Examples) > 0 {
		sections = append(sections, s.formatExamplesSection(content.Examples))
	}

	// Commands section
	if len(content.Commands) > 0 {
		sections = append(sections, s.formatCommandsSection(content.Commands))
	}

	// Flags section
	if len(content.Flags) > 0 {
		sections = append(sections, s.formatFlagsSection(content.Flags))
	}

	// Support section
	if !isEmptySupport(content.Support) {
		sections = append(sections, s.formatSupportSection(content.Support))
	}

	return strings.Join(sections, "\n\n")
}

// formatUsageSection creates a formatted usage section.
// Pure function with no side effects.
func (s Service) formatUsageSection(usage string) string {
	header := s.formatter.Section("USAGE")

	return fmt.Sprintf("%s\n  %s", header, usage)
}

// formatExamplesSection creates a formatted examples section.
// Pure function that sorts examples consistently.
func (s Service) formatExamplesSection(examples []ports.HelpExample) string {
	header := s.formatter.Section("QUICK START")

	lines := make([]string, 0, len(examples))

	for _, example := range examples {
		line := fmt.Sprintf("  %s  # %s", example.Command, example.Description)
		lines = append(lines, line)
	}

	return fmt.Sprintf("%s\n%s", header, strings.Join(lines, "\n"))
}

// formatCommandsSection creates a formatted commands section.
// Pure function that groups commands logically.
func (s Service) formatCommandsSection(commands []ports.HelpCommand) string {
	header := s.formatter.Section("COMMANDS")

	lines := make([]string, 0, len(commands))

	// Sort commands by name for consistency
	sortedCommands := make([]ports.HelpCommand, len(commands))
	copy(sortedCommands, commands)
	sort.Slice(sortedCommands, func(i, j int) bool {
		return sortedCommands[i].Name < sortedCommands[j].Name
	})

	for _, cmd := range sortedCommands {
		line := fmt.Sprintf("  %-8s %s", cmd.Name, cmd.Description)
		lines = append(lines, line)
	}

	return fmt.Sprintf("%s\n%s", header, strings.Join(lines, "\n"))
}

// formatFlagsSection creates a formatted flags section with grouping.
// Pure function that separates common and advanced flags.
func (s Service) formatFlagsSection(flags []ports.HelpFlag) string {
	commonFlags := filterFlags(flags, true)
	advancedFlags := filterFlags(flags, false)

	var sections []string

	if len(commonFlags) > 0 {
		header := s.formatter.Section("COMMON OPTIONS")
		lines := s.formatFlagList(commonFlags)
		sections = append(sections, fmt.Sprintf("%s\n%s", header, strings.Join(lines, "\n")))
	}

	if len(advancedFlags) > 0 {
		header := s.formatter.Section("ADVANCED OPTIONS")
		lines := s.formatFlagList(advancedFlags)
		sections = append(sections, fmt.Sprintf("%s\n%s", header, strings.Join(lines, "\n")))
	}

	return strings.Join(sections, "\n\n")
}

// formatFlagList formats a list of flags consistently.
// Pure function with consistent formatting rules.
func (s Service) formatFlagList(flags []ports.HelpFlag) []string {
	lines := make([]string, 0, len(flags))

	for _, flag := range flags {
		line := fmt.Sprintf("  %-20s %s", flag.Name, flag.Description)
		lines = append(lines, line)
	}

	return lines
}

// formatSupportSection creates a formatted support section.
// Pure function that handles optional support information.
func (s Service) formatSupportSection(support ports.HelpSupport) string {
	header := s.formatter.Section("SUPPORT")

	var lines []string

	if support.Documentation != "" {
		lines = append(lines, "  Documentation: "+support.Documentation)
	}

	if support.Issues != "" {
		lines = append(lines, "  Issues:        "+support.Issues)
	}

	if support.Discussions != "" {
		lines = append(lines, "  Discussions:   "+support.Discussions)
	}

	return fmt.Sprintf("%s\n%s", header, strings.Join(lines, "\n"))
}

// Helper functions (all pure)

// filterFlags separates flags by their common/advanced status.
// Pure function that creates new slices without modifying input.
func filterFlags(flags []ports.HelpFlag, isCommon bool) []ports.HelpFlag {
	var filtered []ports.HelpFlag

	for _, flag := range flags {
		if flag.IsCommon == isCommon {
			filtered = append(filtered, flag)
		}
	}

	return filtered
}

// isEmptySupport checks if support information is empty.
// Pure predicate function.
func isEmptySupport(support ports.HelpSupport) bool {
	return support.Documentation == "" &&
		support.Issues == "" &&
		support.Discussions == ""
}
