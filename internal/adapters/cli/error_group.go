// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"strings"
)

// ErrorGroup collects and groups similar errors for better display
// Be functional: immutable after Format() is called
// Be hexagonal: this is an adapter for presentation, not domain logic
//
//nolint:errname // This is not an error type, it's a container for errors
type ErrorGroup struct {
	operation string
	errors    []errorItem
}

type errorItem struct {
	resource string // repo name, file path, etc.
	err      error
}

// NewErrorGroup creates a new error group for an operation
// Be idiomatic: return pointer for mutable collection phase.
func NewErrorGroup(operation string) *ErrorGroup {
	return &ErrorGroup{
		operation: operation,
		errors:    make([]errorItem, 0),
	}
}

// Add adds an error for a specific resource.
func (eg *ErrorGroup) Add(resource string, err error) {
	if err == nil {
		return
	}

	eg.errors = append(eg.errors, errorItem{resource: resource, err: err})
}

// HasErrors returns true if any errors were collected.
func (eg *ErrorGroup) HasErrors() bool {
	return len(eg.errors) > 0
}

// Count returns the number of errors.
func (eg *ErrorGroup) Count() int {
	return len(eg.errors)
}

// GetErrors returns the collected errors as a slice
// Be functional: return copy to prevent external mutation.
func (eg *ErrorGroup) GetErrors() []error {
	result := make([]error, 0, len(eg.errors))
	for _, item := range eg.errors {
		result = append(result, item.err)
	}

	return result
}

// Format groups and formats errors for display
// Don't overengineer: simple grouping by error pattern.
func (eg *ErrorGroup) Format(symbols Symbols) string {
	if len(eg.errors) == 0 {
		return ""
	}

	var output strings.Builder

	// Single error - show normally
	if len(eg.errors) == 1 {
		fmt.Fprintf(&output, "%s %s failed: %s\n",
			symbols.Cross, eg.operation, eg.errors[0].resource)
		fmt.Fprintf(&output, "  Reason: %s\n", eg.errors[0].err)

		return output.String()
	}

	// Multiple errors - group by type
	groups := eg.groupByErrorType()

	// Show summary
	fmt.Fprintf(&output, "%s %s failed for %d items\n",
		symbols.Cross, eg.operation, len(eg.errors))

	// Show grouped errors (max 3 examples per group)
	for errorType, items := range groups {
		fmt.Fprintf(&output, "\n  %s (%d failed):\n", errorType, len(items))

		// Show first 3 examples
		for i, item := range items {
			if i >= 3 {
				fmt.Fprintf(&output, "    ... and %d more\n", len(items)-3)

				break
			}

			fmt.Fprintf(&output, "    • %s\n", item.resource)
		}
	}

	// Add targeted help based on most common error
	help := eg.getSuggestionForMostCommon(symbols)
	if help != "" {
		fmt.Fprintln(&output, help)
	}

	return output.String()
}

// GroupByErrorType groups errors by their type
// Be functional: pure function, no side effects.
func (eg *ErrorGroup) groupByErrorType() map[string][]errorItem {
	groups := make(map[string][]errorItem)

	for _, item := range eg.errors {
		errType := classifyError(item.err)
		groups[errType] = append(groups[errType], item)
	}

	return groups
}

// ClassifyError determines the error type from the error message
// Be idiomatic: simple string matching, not complex parsing
//
//nolint:cyclop // Multiple error types need checking
func classifyError(err error) string {
	if err == nil {
		return "Unknown error"
	}

	msg := strings.ToLower(err.Error())

	// Check common patterns (most frequent first for performance)
	switch {
	case strings.Contains(msg, "401") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "authentication") ||
		strings.Contains(msg, "unauthorized"):
		return "Authentication problem"

	case strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "network"):
		return "Network problem"

	case strings.Contains(msg, "404") ||
		strings.Contains(msg, "not found") ||
		strings.Contains(msg, "does not exist"):
		return "Not found"

	case strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "access denied"):
		return "Permission denied"

	case strings.Contains(msg, "rate limit"):
		return "Rate limited"

	default:
		// Extract first meaningful part of error
		if idx := strings.Index(msg, ":"); idx > 0 && idx < 50 {
			return strings.TrimSpace(err.Error()[:idx])
		}

		return "Other errors"
	}
}

// GetSuggestionForMostCommon returns help for the most common error type
// Be functional: pure function based on input state.
func (eg *ErrorGroup) getSuggestionForMostCommon(symbols Symbols) string {
	groups := eg.groupByErrorType()

	// Find most common error type

	var mostCommon string

	maxCount := 0
	for errType, items := range groups {
		if len(items) > maxCount {
			maxCount = len(items)
			mostCommon = errType
		}
	}

	// Return specific help
	switch mostCommon {
	case "Authentication problem":
		return fmt.Sprintf("\n%s Fix authentication:\n  1. Check token: echo $GITHUB_TOKEN (or relevant provider)\n  2. Verify token has 'repo' scope\n  3. Test: gitprovidersync status --connectivity-check",
			symbols.Arrow)

	case "Network problem":
		return fmt.Sprintf("\n%s Fix connection:\n  1. Check internet connection\n  2. Verify domain is correct\n  3. Check proxy settings if behind firewall",
			symbols.Arrow)

	case "Not found":
		return fmt.Sprintf("\n%s Items not found:\n  1. Check names are correct\n  2. Verify you have access\n  3. Some may have been deleted",
			symbols.Arrow)

	case "Permission denied":
		return fmt.Sprintf("\n%s Fix permissions:\n  1. Check file/directory permissions\n  2. Ensure write access to target\n  3. Verify git credentials",
			symbols.Arrow)

	case "Rate limited":
		return fmt.Sprintf("\n%s Rate limit hit:\n  1. Wait a few minutes\n  2. Use authentication for higher limits\n  3. Reduce concurrent operations",
			symbols.Arrow)

	default:
		return ""
	}
}

// Error implements the error interface
// allows ErrorGroup to be used as an error type.
func (eg *ErrorGroup) Error() string {
	if len(eg.errors) == 0 {
		return ""
	}

	if len(eg.errors) == 1 {
		return fmt.Sprintf("%s failed for %s: %v", eg.operation, eg.errors[0].resource, eg.errors[0].err)
	}

	return fmt.Sprintf("%s failed for %d items", eg.operation, len(eg.errors))
}
