// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package help

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// MockHelpFormatter implements ports.HelpFormatter for testing.
type mockHelpFormatter struct {
	colorSupported bool
}

func (m mockHelpFormatter) Bold(text string) string {
	if m.colorSupported {
		return "[BOLD]" + text + "[/BOLD]"
	}

	return text
}

func (m mockHelpFormatter) Section(title string) string {
	if m.colorSupported {
		return "[SECTION]" + title + "[/SECTION]"
	}

	return title
}

func (m mockHelpFormatter) IsColorSupported() bool {
	return m.colorSupported
}

func TestNewService(t *testing.T) {
	t.Parallel()

	formatter := mockHelpFormatter{colorSupported: true}
	service := NewService(formatter)

	assert.Equal(t, formatter, service.formatter)
}

func TestFormatHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  ports.HelpContent
		contains []string
	}{
		{
			name: "complete help content",
			content: ports.HelpContent{
				Title:       "Test Application",
				Description: "A test application for unit testing",
				Usage:       "testapp [options] command",
				Examples: []ports.HelpExample{
					{Command: "testapp start", Description: "Start the application"},
					{Command: "testapp stop", Description: "Stop the application"},
				},
				Commands: []ports.HelpCommand{
					{Name: "start", Description: "Start the service"},
					{Name: "stop", Description: "Stop the service"},
				},
				Flags: []ports.HelpFlag{
					{Name: "--verbose", Description: "Enable verbose output", IsCommon: true},
					{Name: "--debug", Description: "Enable debug mode", IsCommon: false},
				},
				Support: ports.HelpSupport{
					Documentation: "https://example.com/docs",
					Issues:        "https://example.com/issues",
					Discussions:   "https://example.com/discussions",
				},
			},
			contains: []string{
				"Test Application",
				"A test application for unit testing",
				"USAGE",
				"testapp [options] command",
				"QUICK START",
				"testapp start  # Start the application",
				"testapp stop  # Stop the application",
				"COMMANDS",
				"start",
				"stop",
				"COMMON OPTIONS",
				"--verbose",
				"ADVANCED OPTIONS",
				"--debug",
				"SUPPORT",
				"Documentation: https://example.com/docs",
				"Issues:        https://example.com/issues",
				"Discussions:   https://example.com/discussions",
			},
		},
		{
			name: "minimal content",
			content: ports.HelpContent{
				Title: "Minimal App",
			},
			contains: []string{"Minimal App"},
		},
		{
			name:     "empty content",
			content:  ports.HelpContent{},
			contains: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			formatter := mockHelpFormatter{colorSupported: false}
			service := NewService(formatter)

			result := service.FormatHelp(test.content)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFormatUsageSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		usage          string
		colorSupported bool
		contains       []string
	}{
		{
			name:           "simple usage without color",
			usage:          "app [options]",
			colorSupported: false,
			contains:       []string{"USAGE", "  app [options]"},
		},
		{
			name:           "simple usage with color",
			usage:          "app [options]",
			colorSupported: true,
			contains:       []string{"[SECTION]USAGE[/SECTION]", "  app [options]"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			formatter := mockHelpFormatter{colorSupported: test.colorSupported}
			service := NewService(formatter)

			result := service.formatUsageSection(test.usage)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFormatExamplesSection(t *testing.T) {
	t.Parallel()

	examples := []ports.HelpExample{
		{Command: "app start", Description: "Start the app"},
		{Command: "app stop", Description: "Stop the app"},
	}

	formatter := mockHelpFormatter{colorSupported: false}
	service := NewService(formatter)

	result := service.formatExamplesSection(examples)

	assert.Contains(t, result, "QUICK START")
	assert.Contains(t, result, "  app start  # Start the app")
	assert.Contains(t, result, "  app stop  # Stop the app")
}

func TestFormatCommandsSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		commands []ports.HelpCommand
		contains []string
	}{
		{
			name: "multiple commands sorted alphabetically",
			commands: []ports.HelpCommand{
				{Name: "stop", Description: "Stop the service"},
				{Name: "start", Description: "Start the service"},
				{Name: "restart", Description: "Restart the service"},
			},
			contains: []string{
				"COMMANDS",
				"restart",
				"start",
				"stop",
			},
		},
		{
			name: "single command",
			commands: []ports.HelpCommand{
				{Name: "run", Description: "Run the application"},
			},
			contains: []string{
				"COMMANDS",
				"run      Run the application",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			formatter := mockHelpFormatter{colorSupported: false}
			service := NewService(formatter)

			result := service.formatCommandsSection(test.commands)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}

			// Verify alphabetical ordering for multiple commands
			if len(test.commands) > 1 {
				lines := strings.Split(result, "\n")

				commandLines := lines[1:] // Skip header
				for i := 1; i < len(commandLines); i++ {
					prevCmd := strings.Fields(commandLines[i-1])[0]
					currCmd := strings.Fields(commandLines[i])[0]
					assert.Less(t, prevCmd, currCmd, "Commands should be sorted alphabetically")
				}
			}
		})
	}
}

func TestFormatFlagsSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    []ports.HelpFlag
		contains []string
	}{
		{
			name: "mixed common and advanced flags",
			flags: []ports.HelpFlag{
				{Name: "--verbose", Description: "Enable verbose output", IsCommon: true},
				{Name: "--help", Description: "Show help", IsCommon: true},
				{Name: "--debug-level", Description: "Set debug level", IsCommon: false},
				{Name: "--trace", Description: "Enable tracing", IsCommon: false},
			},
			contains: []string{
				"COMMON OPTIONS",
				"--verbose",
				"--help",
				"ADVANCED OPTIONS",
				"--debug-level",
				"--trace",
			},
		},
		{
			name: "only common flags",
			flags: []ports.HelpFlag{
				{Name: "--version", Description: "Show version", IsCommon: true},
			},
			contains: []string{
				"COMMON OPTIONS",
				"--version",
			},
		},
		{
			name: "only advanced flags",
			flags: []ports.HelpFlag{
				{Name: "--internal-debug", Description: "Internal debug mode", IsCommon: false},
			},
			contains: []string{
				"ADVANCED OPTIONS",
				"--internal-debug",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			formatter := mockHelpFormatter{colorSupported: false}
			service := NewService(formatter)

			result := service.formatFlagsSection(test.flags)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFormatFlagList(t *testing.T) {
	t.Parallel()

	flags := []ports.HelpFlag{
		{Name: "--verbose", Description: "Enable verbose output", IsCommon: true},
		{Name: "--help", Description: "Show help message", IsCommon: true},
	}

	formatter := mockHelpFormatter{colorSupported: false}
	service := NewService(formatter)

	result := service.formatFlagList(flags)

	assert.Len(t, result, 2)
	assert.Contains(t, result[0], "--verbose")
	assert.Contains(t, result[0], "Enable verbose output")
	assert.Contains(t, result[1], "--help")
	assert.Contains(t, result[1], "Show help message")
}

func TestFormatSupportSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		support  ports.HelpSupport
		contains []string
	}{
		{
			name: "complete support info",
			support: ports.HelpSupport{
				Documentation: "https://example.com/docs",
				Issues:        "https://example.com/issues",
				Discussions:   "https://example.com/discussions",
			},
			contains: []string{
				"SUPPORT",
				"Documentation: https://example.com/docs",
				"Issues:        https://example.com/issues",
				"Discussions:   https://example.com/discussions",
			},
		},
		{
			name: "partial support info",
			support: ports.HelpSupport{
				Documentation: "https://example.com/docs",
				Issues:        "https://example.com/issues",
			},
			contains: []string{
				"SUPPORT",
				"Documentation: https://example.com/docs",
				"Issues:        https://example.com/issues",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			formatter := mockHelpFormatter{colorSupported: false}
			service := NewService(formatter)

			result := service.formatSupportSection(test.support)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFilterFlags(t *testing.T) {
	t.Parallel()

	flags := []ports.HelpFlag{
		{Name: "--verbose", Description: "Verbose output", IsCommon: true},
		{Name: "--help", Description: "Show help", IsCommon: true},
		{Name: "--debug-level", Description: "Debug level", IsCommon: false},
		{Name: "--trace", Description: "Enable tracing", IsCommon: false},
	}

	commonFlags := filterFlags(flags, true)
	advancedFlags := filterFlags(flags, false)

	assert.Len(t, commonFlags, 2)
	assert.Len(t, advancedFlags, 2)

	// Verify common flags
	assert.Equal(t, "--verbose", commonFlags[0].Name)
	assert.Equal(t, "--help", commonFlags[1].Name)

	// Verify advanced flags
	assert.Equal(t, "--debug-level", advancedFlags[0].Name)
	assert.Equal(t, "--trace", advancedFlags[1].Name)
}

func TestIsEmptySupport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		support  ports.HelpSupport
		expected bool
	}{
		{
			name:     "completely empty",
			support:  ports.HelpSupport{},
			expected: true,
		},
		{
			name: "has documentation only",
			support: ports.HelpSupport{
				Documentation: "https://example.com/docs",
			},
			expected: false,
		},
		{
			name: "has issues only",
			support: ports.HelpSupport{
				Issues: "https://example.com/issues",
			},
			expected: false,
		},
		{
			name: "has discussions only",
			support: ports.HelpSupport{
				Discussions: "https://example.com/discussions",
			},
			expected: false,
		},
		{
			name: "has all fields",
			support: ports.HelpSupport{
				Documentation: "https://example.com/docs",
				Issues:        "https://example.com/issues",
				Discussions:   "https://example.com/discussions",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isEmptySupport(test.support)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestServiceWithColorSupport(t *testing.T) {
	t.Parallel()

	content := ports.HelpContent{
		Title: "Test App",
		Usage: "testapp [options]",
	}

	formatter := mockHelpFormatter{colorSupported: true}
	service := NewService(formatter)

	result := service.FormatHelp(content)

	assert.Contains(t, result, "[SECTION]USAGE[/SECTION]")
}

func TestServiceWithoutColorSupport(t *testing.T) {
	t.Parallel()

	content := ports.HelpContent{
		Title: "Test App",
		Usage: "testapp [options]",
	}

	formatter := mockHelpFormatter{colorSupported: false}
	service := NewService(formatter)

	result := service.FormatHelp(content)

	assert.Contains(t, result, "USAGE")
	assert.NotContains(t, result, "[SECTION]")
}
