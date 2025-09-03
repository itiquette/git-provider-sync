// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package statuscmd provides functionality to display Git Provider Sync system status.
// It shows configuration validity, provider connectivity, and suggests next actions.
package statuscmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"

	"itiquette/git-provider-sync/cmd/baseoption"
	cliAdapters "itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/adapters/terminal"
	validationAdapters "itiquette/git-provider-sync/internal/adapters/validation"
	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/validation"
	"itiquette/git-provider-sync/internal/log"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

// NewStatusCommand creates and returns a new cli.Command for the 'status' subcommand.
// It displays the current system status including configuration validity,
// provider connectivity, and suggested next actions.
//
// Example usage:
//
//	gitprovidersync status
//	gitprovidersync status --connectivity-check
//
// The command shows configuration status, provider reachability, and actionable suggestions.
func NewStatusCommand() *cli.Command {
	cmd := &cli.Command{
		Name:  "status",
		Usage: "Show system status and suggest next actions",
		Description: `The 'status' command displays the current state of the Git Provider Sync system.
It validates configuration, optionally tests connectivity, and provides actionable suggestions.

This helps you understand what needs attention and what commands to run next.`,
		Action: runStatus,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "connectivity-check",
				Usage:    "Test connectivity to configured providers",
				Category: "Validation",
			},
			&cli.BoolFlag{
				Name:     "skip-suggestions",
				Usage:    "Don't show suggested next actions",
				Category: "Output Control",
			},
		},
	}

	return cmd
}

// runStatus executes the status command logic.
func runStatus(ctx context.Context, cmd *cli.Command) error {
	logger := log.Logger(ctx)

	// Create CLI configuration functionally
	cliConfig, err := baseoption.ExtractRootInputOptions(cmd)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create CLI configuration")

		if !testing.Testing() {
			os.Exit(2) // Configuration error
		}

		return fmt.Errorf("failed to extract CLI options: %w", err)
	}

	ctx = cliAdapters.WithCLIConfig(ctx, cliConfig)

	// Initialize logger
	logLevel := cmd.String("log-level")
	quiet := cliConfig.Quiet()
	outputFormat := cliConfig.OutputFormat()
	ctx = log.InitLogger(ctx, logLevel, quiet, false, outputFormat)

	// → Reading configuration...
	logger.Info().Msg("→ Reading configuration...")

	configLoaderInstance := configuration.DefaultConfigLoader{}

	conf, err := configLoaderInstance.LoadConfiguration(ctx)
	if err != nil {
		// Get symbols for error formatting
		colorMode := terminal.ColorMode(cliConfig.ColorMode())
		symbols := cliAdapters.GetSymbols(colorMode)

		// Format error with suggestions
		errorMsg := cliAdapters.FormatErrorWithSuggestion(err,
			"Failed to load configuration", symbols)
		fmt.Fprint(os.Stderr, errorMsg)

		if !testing.Testing() {
			os.Exit(2) // Configuration error
		}

		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Extract command flags
	connectivityCheck := cmd.Bool("connectivity-check")
	skipSuggestions := cmd.Bool("skip-suggestions")

	// Perform basic validation
	logger.Info().Msg("→ Validating system status...")

	// Simple status check - configuration is valid if it loaded successfully
	status := createSystemStatus(*conf, connectivityCheck)

	// Optionally test connectivity
	if connectivityCheck {
		logger.Info().Msg("→ Testing connectivity to providers...")

		connectivityAdapter := validationAdapters.NewConnectivityAdapter(30 * time.Second)
		status.ConnectivityResults = testProviderConnectivity(ctx, *conf, connectivityAdapter)
	}

	// Format and display results
	log.Logger(ctx).Debug().Str("outputFormat", outputFormat).Msg("Formatting status output")
	statusOutput := formatSystemStatus(status, connectivityCheck, skipSuggestions, outputFormat)

	// Output to stdout (data) vs stderr (progress messages)
	fmt.Print(statusOutput)

	// Exit with error code if there are critical issues
	if status.HasCriticalIssues {
		if !testing.Testing() {
			os.Exit(1) // System has critical issues
		}

		return domain.ErrSystemCriticalIssues
	}

	return nil
}

// SystemStatus represents a simplified status of the system.
type SystemStatus struct {
	ConfigurationValid  bool
	EnvironmentCount    int
	ConnectivityResults []validation.ConnectivityResult
	ConnectivityChecked bool
	HasCriticalIssues   bool
	Issues              []string
	Warnings            []string
	Suggestions         []string
}

// createSystemStatus creates a system status from the configuration.
func createSystemStatus(config config.AppConfiguration, connectivityCheck bool) SystemStatus {
	// Count environments
	envCount := 0
	for _, env := range config.GitProviderSyncConfs {
		envCount += len(env)
	}

	status := SystemStatus{
		ConfigurationValid:  true, // If config loaded, it's valid
		EnvironmentCount:    envCount,
		ConnectivityChecked: connectivityCheck,
		HasCriticalIssues:   false,
		Issues:              []string{},
		Warnings:            []string{},
		Suggestions:         []string{},
	}

	// Basic validation - check if we have any environments
	if envCount == 0 {
		status.HasCriticalIssues = true
		status.Issues = append(status.Issues, "No environments configured")
		status.Suggestions = append(status.Suggestions, "Add environment configuration to gitprovidersync.yaml")
	}

	return status
}

// testProviderConnectivity tests connectivity to providers in the configuration.
func testProviderConnectivity(ctx context.Context, config config.AppConfiguration, adapter *validationAdapters.ConnectivityAdapter) []validation.ConnectivityResult {
	var results []validation.ConnectivityResult

	// For each environment and sync config, test basic HTTP connectivity
	for envName, env := range config.GitProviderSyncConfs {
		for configName, syncConfig := range env {
			if syncConfig.ProviderType == "" || syncConfig.Domain == "" {
				continue
			}

			// Create basic HTTP connectivity test
			validation := validation.ConnectivityValidation{
				Type:        validation.ConnectivityTypeHTTP,
				Target:      buildHTTPSURL(syncConfig.Domain),
				Timeout:     10 * time.Second,
				Description: fmt.Sprintf("HTTP connectivity to %s (env: %s, config: %s)", syncConfig.Domain, envName, configName),
				Required:    true,
			}

			result := adapter.ValidateConnectivity(ctx, validation)
			results = append(results, result)
		}
	}

	return results
}

// buildHTTPSURL creates an HTTPS URL from a domain.
func buildHTTPSURL(domain string) string {
	if domain == "" {
		return ""
	}

	return "https://" + domain
}

// formatSystemStatus creates formatted output for the system status.
func formatSystemStatus(status SystemStatus, _ /* connectivityChecked */, skipSuggestions bool, outputFormat string) string {
	switch outputFormat {
	case "json":
		return formatStatusJSON(status)
	case "plain":
		return formatStatusPlain(status, skipSuggestions)
	case "console", "":
		return formatStatusConsole(status, skipSuggestions)
	default:
		// Default to plain for unknown formats
		return formatStatusPlain(status, skipSuggestions)
	}
}

// formatStatusConsole creates human-readable console output.
func formatStatusConsole(status SystemStatus, skipSuggestions bool) string {
	// Get appropriate symbols based on terminal capabilities
	// Default to auto mode - honors NO_COLOR environment variable
	symbols := cliAdapters.GetSymbols(terminal.ColorAuto)

	var output strings.Builder

	// Add header
	output.WriteString("Git Provider Sync Status\n")
	output.WriteString("========================\n\n")

	// Configuration status with symbol
	if status.ConfigurationValid {
		fmt.Fprintf(&output, "%s Configuration: Valid (%d environment",
			symbols.Check, status.EnvironmentCount)
	} else {
		fmt.Fprintf(&output, "%s Configuration: Invalid (%d environment",
			symbols.Cross, status.EnvironmentCount)
	}

	if status.EnvironmentCount != 1 {
		output.WriteString("s")
	}

	output.WriteString(")\n")

	// Connectivity if checked
	if status.ConnectivityChecked {
		output.WriteString(formatConnectivityWithSymbols(status, symbols))
	}

	// Overall status
	output.WriteString("\n")

	if status.HasCriticalIssues {
		fmt.Fprintf(&output, "%s Issues need attention\n", symbols.Cross)
	} else {
		fmt.Fprintf(&output, "%s Ready for sync operations\n", symbols.Check)
	}

	// Issues if any
	if len(status.Issues) > 0 || len(status.Warnings) > 0 {
		output.WriteString(formatIssuesWithSymbols(status, symbols))
	}

	// Always provide next steps
	if !skipSuggestions {
		output.WriteString(formatNextSteps(status, symbols))
	}

	return output.String()
}

// formatConnectivityWithSymbols formats connectivity status with appropriate symbols.
func formatConnectivityWithSymbols(status SystemStatus, symbols cliAdapters.Symbols) string {
	_, requiredFailed, optionalFailed := validation.CountConnectivityFailures(status.ConnectivityResults)

	var output strings.Builder

	if requiredFailed == 0 {
		fmt.Fprintf(&output, "%s Provider Connectivity: All providers reachable\n", symbols.Check)
	} else {
		fmt.Fprintf(&output, "%s Connectivity: %d provider(s) unreachable\n",
			symbols.Cross, requiredFailed)
	}

	if optionalFailed > 0 {
		fmt.Fprintf(&output, "%s Warning: %d optional test(s) failed\n",
			symbols.Warning, optionalFailed)
	}

	return output.String()
}

// formatIssuesWithSymbols formats issues with appropriate symbols.
func formatIssuesWithSymbols(status SystemStatus, symbols cliAdapters.Symbols) string {
	var output strings.Builder

	fmt.Fprintln(&output, "\nIssues Found:")

	for _, issue := range status.Issues {
		fmt.Fprintf(&output, "  %s %s\n", symbols.Cross, issue)
	}

	for _, warning := range status.Warnings {
		fmt.Fprintf(&output, "  ! %s\n", warning)
	}

	return output.String()
}

// formatNextSteps provides actionable next steps based on status.
func formatNextSteps(status SystemStatus, symbols cliAdapters.Symbols) string {
	var output strings.Builder

	fmt.Fprintf(&output, "\n%s Next steps:\n", symbols.Arrow)

	switch {
	case !status.ConfigurationValid:
		fmt.Fprintln(&output, "  Fix configuration errors above")
		fmt.Fprintln(&output, "  Run 'gitprovidersync status --verbose' for details")
	case !status.ConnectivityChecked:
		fmt.Fprintln(&output, "  gitprovidersync status --connectivity-check  # Test connections")
		fmt.Fprintln(&output, "  gitprovidersync sync --dry-run              # Preview sync")
	case !status.HasCriticalIssues:
		fmt.Fprintln(&output, "  gitprovidersync sync --dry-run  # Preview changes")
		fmt.Fprintln(&output, "  gitprovidersync sync            # Execute sync")
	default:
		fmt.Fprintln(&output, "  Fix connectivity issues shown above")
		fmt.Fprintln(&output, "  Verify authentication tokens are set correctly")
	}

	// Add custom suggestions
	for _, suggestion := range status.Suggestions {
		fmt.Fprintf(&output, "  %s\n", suggestion)
	}

	return output.String()
}

// formatStatusPlain creates plain text output for pipelines.
func formatStatusPlain(status SystemStatus, _ /* skipSuggestions */ bool) string {
	statusText := "OK"
	if status.HasCriticalIssues {
		statusText = "ERROR"
	} else if len(status.Warnings) > 0 {
		statusText = "WARNING"
	}

	output := fmt.Sprintf("STATUS\t%s\n", statusText)
	output += fmt.Sprintf("CONFIG_VALID\t%v\n", status.ConfigurationValid)
	output += fmt.Sprintf("ENVIRONMENTS\t%d\n", status.EnvironmentCount)

	if status.ConnectivityChecked {
		_, requiredFailed, optionalFailed := validation.CountConnectivityFailures(status.ConnectivityResults)
		output += fmt.Sprintf("CONNECTIVITY_REQUIRED_FAILED\t%d\n", requiredFailed)
		output += fmt.Sprintf("CONNECTIVITY_OPTIONAL_FAILED\t%d\n", optionalFailed)
	}

	output += fmt.Sprintf("TOTAL_ERRORS\t%d\n", len(status.Issues))
	output += fmt.Sprintf("TOTAL_WARNINGS\t%d\n", len(status.Warnings))

	return output
}

// formatStatusJSON creates JSON output for programmatic use.
func formatStatusJSON(status SystemStatus) string {
	// Create a simplified JSON structure
	statusMap := map[string]any{
		"overall_success":      !status.HasCriticalIssues,
		"configuration_valid":  status.ConfigurationValid,
		"environment_count":    status.EnvironmentCount,
		"total_errors":         len(status.Issues),
		"total_warnings":       len(status.Warnings),
		"connectivity_checked": status.ConnectivityChecked,
		"issues":               status.Issues,
		"warnings":             status.Warnings,
		"suggestions":          status.Suggestions,
	}

	if status.ConnectivityChecked {
		_, requiredFailed, optionalFailed := validation.CountConnectivityFailures(status.ConnectivityResults)
		statusMap["connectivity_required_failed"] = requiredFailed
		statusMap["connectivity_optional_failed"] = optionalFailed
	}

	// Simple JSON formatting (avoiding external dependencies)
	return formatJSONOutput(statusMap)
}

// formatJSONOutput creates basic JSON output without external dependencies.
func formatJSONOutput(data map[string]any) string {
	output := "{\n"

	for key, value := range data {
		output += formatJSONValue(key, value, "")
	}

	// Remove trailing comma
	output = removeTrailingComma(output)
	output += "}\n"

	return output
}

// formatJSONValue formats a single JSON key-value pair with proper indentation.
func formatJSONValue(key string, value any, indent string) string {
	baseIndent := "  " + indent

	switch typedValue := value.(type) {
	case string:
		return fmt.Sprintf("%s\"%s\": \"%s\",\n", baseIndent, key, typedValue)
	case int:
		return fmt.Sprintf("%s\"%s\": %d,\n", baseIndent, key, typedValue)
	case int64:
		return fmt.Sprintf("%s\"%s\": %d,\n", baseIndent, key, typedValue)
	case bool:
		return fmt.Sprintf("%s\"%s\": %v,\n", baseIndent, key, typedValue)
	case []string:
		return formatJSONStringArray(key, typedValue, baseIndent)
	case map[string]any:
		return formatJSONObject(key, typedValue, baseIndent)
	default:
		return ""
	}
}

// formatJSONStringArray formats a string array as JSON.
func formatJSONStringArray(key string, values []string, indent string) string {
	output := fmt.Sprintf("%s\"%s\": [", indent, key)

	for index, item := range values {
		if index > 0 {
			output += ", "
		}

		output += fmt.Sprintf("\"%s\"", item)
	}

	output += "],\n"

	return output
}

// formatJSONObject formats a nested object as JSON.
func formatJSONObject(key string, objValue map[string]any, indent string) string {
	output := fmt.Sprintf("%s\"%s\": {\n", indent, key)

	for subKey, subValue := range objValue {
		output += formatJSONObjectValue(subKey, subValue, indent+"  ")
	}

	// Remove trailing comma
	output = removeTrailingComma(output)
	output += indent + "},\n"

	return output
}

// formatJSONObjectValue formats a value within a nested object.
func formatJSONObjectValue(key string, value any, indent string) string {
	switch typedValue := value.(type) {
	case string:
		return fmt.Sprintf("%s\"%s\": \"%s\",\n", indent, key, typedValue)
	case bool:
		return fmt.Sprintf("%s\"%s\": %v,\n", indent, key, typedValue)
	case []string:
		return formatJSONStringArray(key, typedValue, indent)
	default:
		return ""
	}
}

// removeTrailingComma removes trailing comma and newline from JSON output.
func removeTrailingComma(output string) string {
	if len(output) > 2 && output[len(output)-2:] == ",\n" {
		return output[:len(output)-2] + "\n"
	}

	return output
}

// handleStatusError provides specific error messages for status command failures.
func handleStatusError(err error, outputFormat string) {
	errMsg := err.Error()

	if outputFormat == "json" {
		fmt.Printf("{\"error\": \"%s\", \"status\": \"ERROR\"}\n", errMsg)

		return
	}

	if outputFormat == "plain" {
		fmt.Printf("STATUS\tERROR\nERROR_MESSAGE\t%s\n", errMsg)

		return
	}

	// Console format
	fmt.Printf("✗ Status Check Failed: %s\n\n", errMsg)
	fmt.Printf("Suggestions:\n")
	fmt.Printf("  • Check that gitprovidersync.yaml exists in the current directory\n")
	fmt.Printf("  • Or specify config file: --config-file path/to/config.yaml\n")
	fmt.Printf("  • Run 'gitprovidersync --help' for more information\n\n")
}

// getLastSyncInfoFromPath reads simple last sync info from a specific file path.
// Testable with custom file paths.
func getLastSyncInfoFromPath(filePath string) string {
	content, err := os.ReadFile(filePath) //nolint:gosec // File path is controlled and validated
	if err != nil {
		return ""
	}

	// Simple parsing
	var timestamp int64

	var repos, successful, failed, skipped int

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "timestamp="):
			_, _ = fmt.Sscanf(line, "timestamp=%d", &timestamp)
		case strings.HasPrefix(line, "repos="):
			_, _ = fmt.Sscanf(line, "repos=%d", &repos)
		case strings.HasPrefix(line, "successful="):
			_, _ = fmt.Sscanf(line, "successful=%d", &successful)
		case strings.HasPrefix(line, "failed="):
			_, _ = fmt.Sscanf(line, "failed=%d", &failed)
		case strings.HasPrefix(line, "skipped="):
			_, _ = fmt.Sscanf(line, "skipped=%d", &skipped)
		}
	}

	if timestamp > 0 {
		syncTime := time.Unix(timestamp, 0)

		return fmt.Sprintf("✓ Last sync: %s (%d repos, %d successful)\n",
			syncTime.Format("2006-01-02 15:04:05"), repos, successful)
	}

	return ""
}

// Legacy functions kept for tests - to be refactored.
//
//nolint:unused // Used in tests
func formatConfigurationStatus(_ SystemStatus) string {
	return "" // Stub for tests
}

func formatConnectivityStatus(status SystemStatus) string {
	if !status.ConnectivityChecked {
		return "- Provider Connectivity: Not checked (use --connectivity-check)"
	}

	// Count successes and failures
	requiredFailed := 0
	optionalFailed := 0

	for _, result := range status.ConnectivityResults {
		if !result.Success {
			if result.Validation.Required {
				requiredFailed++
			} else {
				optionalFailed++
			}
		}
	}

	var result strings.Builder

	if requiredFailed > 0 {
		fmt.Fprintf(&result, "✗ Provider Connectivity: %d required provider(s) unreachable", requiredFailed)
	} else {
		result.WriteString("✓ Provider Connectivity: All required providers reachable")
	}

	if optionalFailed > 0 {
		fmt.Fprintf(&result, "\n! Warning: %d optional connectivity test(s) failed", optionalFailed)
	}

	return result.String()
}

func formatOverallStatus(status SystemStatus) string {
	if status.HasCriticalIssues {
		return "✗ Issues need attention"
	}

	return "✓ Ready for sync operations"
}

func formatIssuesSection(status SystemStatus) string {
	if len(status.Issues) == 0 && len(status.Warnings) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("\nIssues Found:\n")

	for _, issue := range status.Issues {
		fmt.Fprintf(&result, "  ✗ %s\n", issue)
	}

	for _, warning := range status.Warnings {
		fmt.Fprintf(&result, "  ! %s\n", warning)
	}

	return result.String()
}

func formatSuggestionsSection(status SystemStatus) string {
	var suggestions []string

	switch {
	case !status.ConfigurationValid:
		suggestions = append(suggestions, "Fix configuration errors and run 'gitprovidersync status' again")
	case !status.ConnectivityChecked:
		suggestions = append(suggestions, "Run 'gitprovidersync status --connectivity-check' to test provider connections")
	case status.HasCriticalIssues:
		suggestions = append(suggestions, "Fix connectivity issues shown above")
		suggestions = append(suggestions, "Check authentication tokens and network connectivity")
	default:
		suggestions = append(suggestions, "Run 'gitprovidersync sync --dry-run' to preview changes")
		suggestions = append(suggestions, "Run 'gitprovidersync sync' to perform the sync")
	}

	// Add custom suggestions
	suggestions = append(suggestions, status.Suggestions...)

	if len(suggestions) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("\n→ Next steps:\n")

	for _, suggestion := range suggestions {
		fmt.Fprintf(&result, "  %s\n", suggestion)
	}

	return result.String()
}
