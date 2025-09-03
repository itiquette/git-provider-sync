// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package printcmd provides functionality to print Git Provider Sync configuration.
// It allows users to view their current configuration settings in a readable format.
package printcmd

import (
	"context"
	"fmt"
	"io"
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
	"itiquette/git-provider-sync/internal/domain/validation"
	"itiquette/git-provider-sync/internal/log"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

// NewPrintCommand creates and returns a new cli.Command for the 'print' subcommand.
// It displays the current Git Provider Sync configuration using the default formatter.
//
// Example usage:
//
//	git-provider-sync print
//	git-provider-sync print --connectivity-check
//
// The command will output the full configuration including all sources
// and their respective settings, optionally with connectivity status.
func NewPrintCommand() *cli.Command {
	return NewPrintCommandWithWriter(os.Stdout)
}

// NewPrintCommandWithWriter creates a print command with a custom writer for testing.
func NewPrintCommandWithWriter(writer io.Writer) *cli.Command {
	cmd := &cli.Command{
		Name:  "print",
		Usage: "Print the current configuration",
		Description: `The 'print' command outputs the current, aggregated Git Provider Sync configuration to stdout.
It loads the configuration from available sources and displays it in a formatted manner.
Optionally tests connectivity to configured providers using the --connectivity-check flag.`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runPrintWithWriter(ctx, cmd, writer)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "connectivity-check",
				Usage:    "Test connectivity to configured providers",
				Category: "Validation",
			},
		},
	}

	return cmd
}

// runPrintWithWriter executes the logic for the 'print' command with custom writer.
// Uses proper error handling with exit codes and dependency injection.
func runPrintWithWriter(ctx context.Context, cmd *cli.Command, writer io.Writer) error {
	logger := log.Logger(ctx)

	// Create CLI configuration functionally
	cliConfig, err := baseoption.ExtractRootInputOptions(cmd)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to extract CLI options")

		// Use consistent error formatting
		// Default to auto mode - honors NO_COLOR environment variable
		symbols := cliAdapters.GetSymbols(terminal.ColorAuto)

		errorMsg := cliAdapters.FormatErrorWithSuggestion(err,
			"Failed to extract CLI options", symbols)
		fmt.Fprint(os.Stderr, errorMsg)

		if !testing.Testing() {
			os.Exit(2) // Configuration error
		}

		return fmt.Errorf("failed to extract CLI options: %w", err)
	}

	ctx = cliAdapters.WithCLIConfig(ctx, cliConfig)

	// CLI config already added to context by baseOpt.AddRootInputOptionsToContext
	// Initialize logger using proven approach
	logLevel := cmd.String("log-level")
	quiet := cliConfig.Quiet()
	outputFormat := cliConfig.OutputFormat()
	ctx = log.InitLogger(ctx, logLevel, quiet, false, outputFormat)

	configLoaderInstance := configuration.DefaultConfigLoader{}

	conf, err := configLoaderInstance.LoadConfiguration(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load configuration")

		// Use consistent error formatting
		colorMode := terminal.ColorMode(cliConfig.ColorMode())
		symbols := cliAdapters.GetSymbols(colorMode)

		errorMsg := cliAdapters.FormatErrorWithSuggestion(err,
			"Failed to load configuration", symbols)
		fmt.Fprint(os.Stderr, errorMsg)

		if !testing.Testing() {
			os.Exit(2) // Configuration error
		}

		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if connectivity check is requested
	connectivityCheck := cmd.Bool("connectivity-check")

	// Output format already extracted above

	// Create output formatter with color mode from CLI config
	colorMode := terminal.ColorMode(cliConfig.ColorMode())

	formatter := cliAdapters.NewOutputFormatterWithColorMode(colorMode)
	if err := formatter.FormatConfiguration(*conf, outputFormat, writer); err != nil {
		logger.Error().Err(err).Msg("Failed to format configuration output")

		// Use consistent error formatting
		symbols := cliAdapters.GetSymbols(colorMode)
		errorMsg := cliAdapters.FormatErrorWithSuggestion(err,
			"Failed to format configuration output", symbols)
		fmt.Fprint(os.Stderr, errorMsg)

		if !testing.Testing() {
			os.Exit(2) // Configuration error
		}

		return fmt.Errorf("failed to format configuration output: %w", err)
	}

	// Optionally test connectivity
	if connectivityCheck {
		logger.Info().Msg("Testing connectivity to configured providers...")

		if err := testAndDisplayConnectivity(ctx, *conf, outputFormat, writer); err != nil {
			logger.Error().Err(err).Msg("Failed to test connectivity")
		}
	}

	// Add simple command suggestion
	if outputFormat == "console" {
		fmt.Fprintf(os.Stderr, "\nNext: gitprovidersync sync --dry-run\n")
	}

	return nil
}

// handleConfigurationError returns specific error messages based on the configuration error type.
// Deprecated: Use FormatErrorWithSuggestion for consistency.
func handleConfigurationError(err error) {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "failed to find a configuration"):
		// Keep for backward compatibility but prefer FormatErrorWithSuggestion
		fmt.Fprintf(os.Stderr, "\n✗ No configuration found\n\n")
		fmt.Fprintf(os.Stderr, "Quick fix:\n")
		fmt.Fprintf(os.Stderr, "  echo 'gitprovidersync:\n    default:\n      source:\n        provider_type: github\n        owner: YOUR_USERNAME\n        owner_type: user\n        mirrors:\n          backup:\n            provider_type: gitlab\n            owner: BACKUP_USERNAME' > gitprovidersync.yaml\n\n")
		fmt.Fprintf(os.Stderr, "Or specify existing config: --config-file path/to/config.yaml\n")
		fmt.Fprintf(os.Stderr, "Then run: gitprovidersync print\n\n")
		fmt.Fprintf(os.Stderr, "Configuration guide: https://github.com/itiquette/git-provider-sync/blob/main/docs/configuration.md\n")
		fmt.Fprintf(os.Stderr, "Full example: https://github.com/itiquette/git-provider-sync/blob/main/examples/gitprovidersync-complete-example.yaml\n\n")

	case strings.Contains(errMsg, "error loading config"):
		fmt.Fprintf(os.Stderr, "\nYAML syntax error in configuration file:\n")
		fmt.Fprintf(os.Stderr, "%s\n\n", errMsg)
		fmt.Fprintf(os.Stderr, "Fix the YAML syntax and try again.\n")
		fmt.Fprintf(os.Stderr, "Tip: Check for missing quotes, incorrect indentation, or extra characters.\n\n")

	case strings.Contains(errMsg, "error unmarshalling yaml config"):
		fmt.Fprintf(os.Stderr, "\nConfiguration structure error:\n")
		fmt.Fprintf(os.Stderr, "%s\n\n", errMsg)
		fmt.Fprintf(os.Stderr, "The YAML structure doesn't match the expected configuration format.\n")
		fmt.Fprintf(os.Stderr, "Check the documentation for the correct configuration structure.\n\n")

	case strings.Contains(errMsg, "failed to validate configuration"):
		handleValidationError(errMsg)

	default:
		fmt.Fprintf(os.Stderr, "\nConfiguration error:\n")
		fmt.Fprintf(os.Stderr, "%s\n\n", errMsg)
		fmt.Fprintf(os.Stderr, "Run 'gitprovidersync --help' for more information.\n\n")
	}
}

// handleValidationError returns specific guidance for configuration validation errors.
func handleValidationError(errMsg string) {
	fmt.Fprintf(os.Stderr, "\nConfiguration validation failed:\n")

	// Extract the specific validation error details
	if strings.Contains(errMsg, "invalid environment") {
		// Extract environment name and specific error
		lines := strings.Split(errMsg, ":")
		if len(lines) >= 3 {
			envPart := strings.TrimSpace(lines[1])
			errorPart := strings.TrimSpace(strings.Join(lines[2:], ":"))

			fmt.Fprintf(os.Stderr, "Environment '%s': %s\n\n", envPart, errorPart)
		} else {
			fmt.Fprintf(os.Stderr, "%s\n\n", errMsg)
		}

		// Provide specific guidance based on common validation errors
		switch {
		case strings.Contains(errMsg, "no owner configured"):
			fmt.Fprintf(os.Stderr, "✗ Missing owner field\n\n")
			fmt.Fprintf(os.Stderr, "Quick fix: Add owner to your config:\n")
			fmt.Fprintf(os.Stderr, "  owner: YOUR_USERNAME\n")
			fmt.Fprintf(os.Stderr, "  owner_type: user\n\n")

		case strings.Contains(errMsg, "unsupported provider"):
			fmt.Fprintf(os.Stderr, "✗ Invalid provider type\n\n")
			fmt.Fprintf(os.Stderr, "Quick fix: Use valid provider_type:\n")
			fmt.Fprintf(os.Stderr, "  provider_type: github  # or gitlab, gitea\n\n")

		case strings.Contains(errMsg, "no domain configured"):
			fmt.Fprintf(os.Stderr, "✗ Missing domain field\n\n")
			fmt.Fprintf(os.Stderr, "Quick fix: Add domain to your config:\n")
			fmt.Fprintf(os.Stderr, "  domain: github.com  # or gitlab.com, your-gitea.com\n\n")

		default:
			fmt.Fprintf(os.Stderr, "✗ Configuration validation failed\n\n")
			fmt.Fprintf(os.Stderr, "Check the configuration file for missing or invalid fields.\n\n")
		}
	} else {
		fmt.Fprintf(os.Stderr, "%s\n\n", errMsg)
	}

	fmt.Fprintf(os.Stderr, "Run 'gitprovidersync --help' for more information about configuration.\n\n")
}

// testAndDisplayConnectivity tests connectivity to configured providers and displays results.
func testAndDisplayConnectivity(ctx context.Context, config gpsconfig.AppConfiguration, outputFormat string, writer io.Writer) error {
	// Create connectivity adapter
	connectivityAdapter := validationAdapters.NewConnectivityAdapter(30 * time.Second)

	// Test connectivity for each environment and sync config
	var allResults []validation.ConnectivityResult

	for envName, env := range config.GitProviderSyncConfs {
		for configName, syncConfig := range env {
			if syncConfig.ProviderType == "" || syncConfig.Domain == "" {
				continue
			}

			// Create basic HTTP connectivity test
			connectivityValidation := validation.ConnectivityValidation{
				Type:        validation.ConnectivityTypeHTTP,
				Target:      buildHTTPSURL(syncConfig.Domain),
				Timeout:     10 * time.Second,
				Description: fmt.Sprintf("HTTP connectivity to %s (env: %s, config: %s)", syncConfig.Domain, envName, configName),
				Required:    true,
			}

			result := connectivityAdapter.ValidateConnectivity(ctx, connectivityValidation)
			allResults = append(allResults, result)
		}
	}

	// Display results
	if err := displayConnectivityResults(allResults, outputFormat, writer); err != nil {
		return fmt.Errorf("failed to display connectivity results: %w", err)
	}

	return nil
}

// buildHTTPSURL creates an HTTPS URL from a domain.
func buildHTTPSURL(domain string) string {
	if domain == "" {
		return ""
	}

	return "https://" + domain
}

// displayConnectivityResults formats and displays connectivity test results.
func displayConnectivityResults(results []validation.ConnectivityResult, outputFormat string, writer io.Writer) error {
	if len(results) == 0 {
		if _, err := fmt.Fprintln(writer, "\nNo connectivity tests performed."); err != nil {
			return fmt.Errorf("failed to write connectivity results: %w", err)
		}

		return nil
	}

	switch outputFormat {
	case "json":
		return displayConnectivityResultsJSON(results, writer)
	case "plain":
		return displayConnectivityResultsPlain(results, writer)
	default:
		return displayConnectivityResultsConsole(results, writer)
	}
}

// displayConnectivityResultsConsole formats connectivity results for console output.
func displayConnectivityResultsConsole(results []validation.ConnectivityResult, writer io.Writer) error {
	if err := writeConnectivityHeader(writer); err != nil {
		return err
	}

	if err := writeConnectivityResults(results, writer); err != nil {
		return err
	}

	if err := writeConnectivitySummary(results, writer); err != nil {
		return err
	}

	return nil
}

// writeConnectivityHeader writes the header for connectivity results.
func writeConnectivityHeader(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "\nConnectivity Test Results:"); err != nil {
		return fmt.Errorf("failed to write connectivity results header: %w", err)
	}

	if _, err := fmt.Fprintln(writer, "="+strings.Repeat("=", 50)); err != nil {
		return fmt.Errorf("failed to write connectivity results separator: %w", err)
	}

	return nil
}

// writeConnectivityResults writes individual connectivity test results.
func writeConnectivityResults(results []validation.ConnectivityResult, writer io.Writer) error {
	for _, result := range results {
		if err := writeConnectivityResult(result, writer); err != nil {
			return err
		}
	}

	return nil
}

// writeConnectivityResult writes a single connectivity test result.
func writeConnectivityResult(result validation.ConnectivityResult, writer io.Writer) error {
	status := "✓"
	if !result.Success {
		status = "✗"
	}

	if _, err := fmt.Fprintf(writer, "%s %s", status, result.Validation.Description); err != nil {
		return fmt.Errorf("failed to write connectivity result: %w", err)
	}

	if result.Duration > 0 {
		if _, err := fmt.Fprintf(writer, " (%.2fms)", float64(result.Duration.Nanoseconds())/1e6); err != nil {
			return fmt.Errorf("failed to write connectivity result duration: %w", err)
		}
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return fmt.Errorf("failed to write connectivity result newline: %w", err)
	}

	if !result.Success && result.Error != nil {
		if _, err := fmt.Fprintf(writer, "  Error: %s\n", result.Error.Error()); err != nil {
			return fmt.Errorf("failed to write connectivity result error: %w", err)
		}
	}

	return nil
}

// writeConnectivitySummary writes the summary of connectivity test results.
func writeConnectivitySummary(results []validation.ConnectivityResult, writer io.Writer) error {
	total := len(results)
	successful := countSuccessfulResults(results)

	if _, err := fmt.Fprintf(writer, "\nSummary: %d/%d tests passed\n", successful, total); err != nil {
		return fmt.Errorf("failed to write connectivity results summary: %w", err)
	}

	return nil
}

// countSuccessfulResults counts the number of successful connectivity results.
func countSuccessfulResults(results []validation.ConnectivityResult) int {
	successful := 0

	for _, result := range results {
		if result.Success {
			successful++
		}
	}

	return successful
}

// displayConnectivityResultsPlain formats connectivity results for plain output.
func displayConnectivityResultsPlain(results []validation.ConnectivityResult, writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "\nCONNECTIVITY_RESULTS"); err != nil {
		return fmt.Errorf("failed to write connectivity results header: %w", err)
	}

	for index, result := range results {
		status := "PASS"
		if !result.Success {
			status = "FAIL"
		}

		if _, err := fmt.Fprintf(writer, "TEST_%d\t%s\t%s\t%.2f\n",
			index+1,
			status,
			result.Validation.Description,
			float64(result.Duration.Nanoseconds())/1e6); err != nil {
			return fmt.Errorf("failed to write connectivity result: %w", err)
		}
	}

	return nil
}

// displayConnectivityResultsJSON formats connectivity results as JSON.
func displayConnectivityResultsJSON(results []validation.ConnectivityResult, writer io.Writer) error {
	if err := writeConnectivityJSONHeader(writer); err != nil {
		return err
	}

	if err := writeConnectivityJSONResults(results, writer); err != nil {
		return err
	}

	if err := writeConnectivityJSONFooter(writer); err != nil {
		return err
	}

	return nil
}

// writeConnectivityJSONHeader writes the JSON header for connectivity results.
func writeConnectivityJSONHeader(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "\n\"connectivity_results\": ["); err != nil {
		return fmt.Errorf("failed to write connectivity results JSON header: %w", err)
	}

	return nil
}

// writeConnectivityJSONResults writes connectivity results in JSON format.
func writeConnectivityJSONResults(results []validation.ConnectivityResult, writer io.Writer) error {
	for i, result := range results {
		if i > 0 {
			if _, err := fmt.Fprintln(writer, ","); err != nil {
				return fmt.Errorf("failed to write connectivity result separator: %w", err)
			}
		}

		if err := writeConnectivityJSONResult(result, writer); err != nil {
			return err
		}
	}

	return nil
}

// writeConnectivityJSONResult writes a single connectivity result in JSON format.
func writeConnectivityJSONResult(result validation.ConnectivityResult, writer io.Writer) error {
	errorMsg := ""
	if result.Error != nil {
		errorMsg = result.Error.Error()
	}

	if _, err := fmt.Fprintf(writer, "  {\n"); err != nil {
		return fmt.Errorf("failed to write connectivity result JSON: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "    \"description\": \"%s\",\n", result.Validation.Description); err != nil {
		return fmt.Errorf("failed to write connectivity result description: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "    \"success\": %v,\n", result.Success); err != nil {
		return fmt.Errorf("failed to write connectivity result success: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "    \"duration_ms\": %.2f,\n", float64(result.Duration.Nanoseconds())/1e6); err != nil {
		return fmt.Errorf("failed to write connectivity result duration: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "    \"error\": \"%s\"\n", errorMsg); err != nil {
		return fmt.Errorf("failed to write connectivity result error: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "  }"); err != nil {
		return fmt.Errorf("failed to write connectivity result JSON close: %w", err)
	}

	return nil
}

// writeConnectivityJSONFooter writes the JSON footer for connectivity results.
func writeConnectivityJSONFooter(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "\n]"); err != nil {
		return fmt.Errorf("failed to write connectivity results JSON close: %w", err)
	}

	return nil
}
