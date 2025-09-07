// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"strings"

	"github.com/urfave/cli/v3"

	"itiquette/git-provider-sync/internal/adapters/terminal"
	"itiquette/git-provider-sync/internal/domain/help"
	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	commandSync = "sync"
)

// HelpRenderer adapts urfave/cli commands to our hexagonal help formatting system
// adapter bridges the gap between urfave/cli's command structure and our domain config.
type HelpRenderer struct {
	helpService help.Service
}

// NewHelpRenderer creates a new help renderer with terminal-aware formatting
// constructor sets up the complete hexagonal architecture for help formatting.
func NewHelpRenderer() *HelpRenderer {
	formatter := terminal.NewHelpFormatter()
	helpService := help.NewService(formatter)

	return &HelpRenderer{
		helpService: helpService,
	}
}

// NewHelpRendererWithFormatter creates a renderer with a custom formatter
// constructor allows dependency injection for testing purposes.
func NewHelpRendererWithFormatter(formatter ports.HelpFormatter) *HelpRenderer {
	helpService := help.NewService(formatter)

	return &HelpRenderer{
		helpService: helpService,
	}
}

// RenderRootHelp converts a urfave/cli root command to formatted help text.
func (r *HelpRenderer) RenderRootHelp(cmd *cli.Command) string {
	content := r.extractRootHelpContent(cmd)

	return r.helpService.FormatHelp(content)
}

// RenderSubcommandHelp converts a urfave/cli subcommand to formatted help text.
func (r *HelpRenderer) RenderSubcommandHelp(cmd *cli.Command) string {
	content := r.extractSubcommandHelpContent(cmd)

	return r.helpService.FormatHelp(content)
}

// ExtractRootHelpContent extracts structured help content from a root command
// Pure function that transforms urfave/cli structures to domain models.
func (r *HelpRenderer) extractRootHelpContent(cmd *cli.Command) ports.HelpContent {
	return ports.HelpContent{
		Title:       cmd.Usage,
		Description: r.extractDescription(cmd),
		Usage:       r.extractUsage(cmd),
		Examples:    r.extractRootExamples(),
		Commands:    r.extractCommands(cmd),
		Flags:       r.extractRootFlags(cmd),
		Support:     r.extractSupportInfo(),
	}
}

// ExtractSubcommandHelpContent extracts help content from a subcommand
// Pure function specialized for subcommand formatting.
func (r *HelpRenderer) extractSubcommandHelpContent(cmd *cli.Command) ports.HelpContent {
	return ports.HelpContent{
		Title:       cmd.Usage,
		Description: r.extractDescription(cmd),
		Usage:       r.extractUsage(cmd),
		Examples:    r.extractSubcommandExamples(cmd),
		Commands:    []ports.HelpCommand{}, // Subcommands typically don't have nested commands
		Flags:       r.extractSubcommandFlags(cmd),
		Support:     r.extractDocumentationLink(cmd),
	}
}

// ExtractDescription extracts the long description, falling back to usage
// Pure function with consistent fallback behavior.
func (r *HelpRenderer) extractDescription(cmd *cli.Command) string {
	if cmd.Description != "" {
		// Extract only the first paragraph for conciseness
		lines := splitIntoLines(cmd.Description)
		if len(lines) > 0 {
			return lines[0]
		}
	}

	return cmd.Usage
}

// ExtractUsage creates a clean usage line
// Pure function that formats usage consistently.
func (r *HelpRenderer) extractUsage(cmd *cli.Command) string {
	if cmd.Action != nil {
		return cmd.Name + " [options]"
	}

	return cmd.Name + " [command]"
}

// ExtractRootExamples returns quick start examples for the root command.
func (r *HelpRenderer) extractRootExamples() []ports.HelpExample {
	return []ports.HelpExample{
		{Description: "verify configuration", Command: "gitprovidersync print"},
		{Description: "preview changes", Command: "gitprovidersync sync --dry-run"},
		{Description: "execute sync", Command: "gitprovidersync sync"},
	}
}

// ExtractSubcommandExamples extracts examples from subcommand descriptions
// Pure function that parses examples from command descriptions.
func (r *HelpRenderer) extractSubcommandExamples(cmd *cli.Command) []ports.HelpExample {
	// For sync command
	if cmd.Name == commandSync {
		return []ports.HelpExample{
			{Description: "run synchronization", Command: "gitprovidersync sync"},
			{Description: "preview changes only", Command: "gitprovidersync sync --dry-run"},
		}
	}

	// For print command
	if cmd.Name == "print" {
		return []ports.HelpExample{
			{Description: "show current configuration", Command: "gitprovidersync print"},
			{Description: "machine-readable output", Command: "gitprovidersync print --format json"},
		}
	}

	return []ports.HelpExample{}
}

// ExtractCommands extracts available subcommands
// Pure function that filters and sorts commands.
func (r *HelpRenderer) extractCommands(cmd *cli.Command) []ports.HelpCommand {
	var commands []ports.HelpCommand

	for _, subCmd := range cmd.Commands {
		if !subCmd.Hidden {
			commands = append(commands, ports.HelpCommand{
				Name:        subCmd.Name,
				Description: subCmd.Usage,
			})
		}
	}

	return commands
}

// ExtractRootFlags extracts and categorizes root command flags
// Pure function that determines common vs advanced flags.
func (r *HelpRenderer) extractRootFlags(cmd *cli.Command) []ports.HelpFlag {
	flags := make([]ports.HelpFlag, 0, len(cmd.Flags))

	// Common flags (frequently used)
	commonFlagNames := map[string]bool{
		"config-file": true,
		"quiet":       true,
		"help":        true,
	}

	for _, flag := range cmd.Flags {
		// Skip hidden flags by checking if it's a concrete type with Hidden field
		if isHiddenFlag(flag) {
			continue
		}

		flagName := "--" + flag.Names()[0]

		if len(flag.Names()) > 1 {
			// Add aliases
			for _, alias := range flag.Names()[1:] {
				if len(alias) == 1 {
					flagName = "-" + alias + ", " + flagName

					break
				}
			}
		}

		flags = append(flags, ports.HelpFlag{
			Name:        flagName,
			Description: getFlagUsage(flag),
			IsCommon:    commonFlagNames[flag.Names()[0]],
		})
	}

	return flags
}

// ExtractSubcommandFlags extracts flags from a subcommand
// Pure function with subcommand-specific flag categorization.
func (r *HelpRenderer) extractSubcommandFlags(cmd *cli.Command) []ports.HelpFlag {
	flags := make([]ports.HelpFlag, 0, len(cmd.Flags))

	// For sync command, dry-run and force-push are common
	commonSyncFlags := map[string]bool{
		"dry-run":    true,
		"force-push": true,
		"help":       true,
	}

	for _, flag := range cmd.Flags {
		// Skip hidden flags by checking if it's a concrete type with Hidden field
		if isHiddenFlag(flag) {
			continue
		}

		flagName := "--" + flag.Names()[0]

		if len(flag.Names()) > 1 {
			// Add aliases
			for _, alias := range flag.Names()[1:] {
				if len(alias) == 1 {
					flagName = "-" + alias + ", " + flagName

					break
				}
			}
		}

		var isCommon bool
		if cmd.Name == commandSync {
			isCommon = commonSyncFlags[flag.Names()[0]]
		} else {
			isCommon = flag.Names()[0] == "help" // help is always common
		}

		flags = append(flags, ports.HelpFlag{
			Name:        flagName,
			Description: getFlagUsage(flag),
			IsCommon:    isCommon,
		})
	}

	return flags
}

// ExtractSupportInfo returns support and documentation links.
func (r *HelpRenderer) extractSupportInfo() ports.HelpSupport {
	return ports.HelpSupport{
		Documentation: "https://github.com/itiquette/git-provider-sync/blob/main/README.adoc",
		Issues:        "https://github.com/itiquette/git-provider-sync/issues",
		Discussions:   "https://github.com/itiquette/git-provider-sync/discussions",
	}
}

// ExtractDocumentationLink returns subcommand-specific documentation links.
func (r *HelpRenderer) extractDocumentationLink(cmd *cli.Command) ports.HelpSupport {
	baseURL := "https://github.com/itiquette/git-provider-sync/blob/main/docs/usage.adoc"

	switch cmd.Name {
	case commandSync:
		return ports.HelpSupport{
			Documentation: baseURL + "#sync-command",
		}
	case "print":
		return ports.HelpSupport{
			Documentation: baseURL + "#print-command",
		}
	default:
		return ports.HelpSupport{
			Documentation: "https://github.com/itiquette/git-provider-sync/blob/main/README.adoc",
		}
	}
}

// Helper functions

// IsHiddenFlag checks if a flag is hidden by examining its concrete type.
func isHiddenFlag(flag cli.Flag) bool {
	// Use type assertion to check common flag types for Hidden field
	switch typedFlag := flag.(type) {
	case *cli.StringFlag:
		return typedFlag.Hidden
	case *cli.BoolFlag:
		return typedFlag.Hidden
	case *cli.IntFlag:
		return typedFlag.Hidden
	case *cli.Float64Flag:
		return typedFlag.Hidden
	case *cli.DurationFlag:
		return typedFlag.Hidden
	case *cli.StringSliceFlag:
		return typedFlag.Hidden
	case *cli.IntSliceFlag:
		return typedFlag.Hidden
	default:
		return false // Default to not hidden if we can't determine
	}
}

// GetFlagUsage extracts usage string from flag.
func getFlagUsage(flag cli.Flag) string {
	// Use type assertion to get usage from common flag types
	switch typedFlag := flag.(type) {
	case *cli.StringFlag:
		return typedFlag.Usage
	case *cli.BoolFlag:
		return typedFlag.Usage
	case *cli.IntFlag:
		return typedFlag.Usage
	case *cli.Float64Flag:
		return typedFlag.Usage
	case *cli.DurationFlag:
		return typedFlag.Usage
	case *cli.StringSliceFlag:
		return typedFlag.Usage
	case *cli.IntSliceFlag:
		return typedFlag.Usage
	default:
		return ""
	}
}

// SplitIntoLines splits text into lines, handling different line endings
// Pure function for consistent text processing.
func splitIntoLines(text string) []string {
	// Split on common line endings and filter empty lines
	lines := []string{}

	// Split the text on newlines
	splitLines := strings.Split(text, "\n")

	for _, line := range splitLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}
