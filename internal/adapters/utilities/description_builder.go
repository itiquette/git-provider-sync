// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package utilities

import (
	"strings"
)

// BuildDescription creates a repository description combining upstream URL and existing description.
// This restores the buildDescription functionality from main branch mirrorfacade.go.
//
// Parameters:
//   - userDescriptionPrefix: Custom description prefix (if provided)
//   - upstreamURL: The original source repository URL
//   - existingDescription: The current repository description
//
// Returns a formatted description string with line breaks removed.
func BuildDescription(userDescriptionPrefix, upstreamURL, existingDescription string) string {
	var description string

	// Use custom prefix if provided, otherwise use default format
	if userDescriptionPrefix != "" {
		description = userDescriptionPrefix
	} else {
		description = "Git Provider Sync cloned this from: " + upstreamURL + ": "
	}

	// Append existing description if available
	if existingDescription != "" {
		description += existingDescription
	}

	// Remove line breaks as in main branch
	return RemoveLinebreaksLocal(description)
}

// RemoveLinebreaksLocal removes line breaks from a string.
// This restores the stringconvert.RemoveLinebreaks functionality from main branch.
func RemoveLinebreaksLocal(text string) string {
	// Replace both \n and \r with spaces
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Clean up multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return strings.TrimSpace(text)
}

// BuildRepositoryDescription creates a description for repository creation.
// This is a comprehensive helper that handles all description building scenarios.
func BuildRepositoryDescription(options DescriptionOptions) string {
	if options.CustomPrefix != "" {
		// User provided custom prefix - use it directly
		description := options.CustomPrefix
		if options.ExistingDescription != "" {
			description += " " + options.ExistingDescription
		}
		return RemoveLinebreaksLocal(description)
	}

	// Build default description format
	var parts []string

	if options.SourceURL != "" {
		parts = append(parts, "Git Provider Sync cloned this from: "+options.SourceURL)
	}

	if options.ExistingDescription != "" {
		parts = append(parts, options.ExistingDescription)
	}

	// If no source URL and no existing description, provide minimal description
	if len(parts) == 0 {
		parts = append(parts, "Repository synchronized by Git Provider Sync")
	}

	description := strings.Join(parts, ": ")
	return RemoveLinebreaksLocal(description)
}

// DescriptionOptions contains all parameters for building repository descriptions.
type DescriptionOptions struct {
	CustomPrefix        string // User-provided description prefix
	SourceURL           string // Original repository URL
	ExistingDescription string // Current repository description
	RepositoryName      string // Repository name (for fallback)
}

// GetDescriptionFromGPSUpstream extracts description using GPSUPSTREAM remote URL.
// This restores the main branch pattern of using GPSUPSTREAM remote for description building.
func GetDescriptionFromGPSUpstream(gpsUpstreamURL, existingDescription, customPrefix string) string {
	if gpsUpstreamURL == "" {
		// Fallback if no GPSUPSTREAM remote
		return BuildRepositoryDescription(DescriptionOptions{
			CustomPrefix:        customPrefix,
			ExistingDescription: existingDescription,
		})
	}

	return BuildDescription(customPrefix, gpsUpstreamURL, existingDescription)
}
