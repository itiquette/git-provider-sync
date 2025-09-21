// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package printcmd prints Git Provider Sync configuration
// in various readable formats
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
	"itiquette/git-provider-sync/internal/adapters/configuration"
	gps "itiquette/git-provider-sync/internal/adapters/configuration/dto"
	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/adapters/log"
	"itiquette/git-provider-sync/internal/adapters/terminal"
	validationAdapters "itiquette/git-provider-sync/internal/adapters/validation"
	"itiquette/git-provider-sync/internal/domain/validation"
)

// NewPrintCommand creates and returns a new cli.Command for the 'print' subcommand
// displays the current Git Provider Sync configuration using the default formatter
// Example usage:
//
//	git-provider-sync print
//	git-provider-sync print --connectivity-check
//
// Command will output the full configuration including all sources
// And their respective settings, optionally with connectivity status.
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

// RunPrintWithWriter executes the logic for the 'print' command with custom writer
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

// TestAndDisplayConnectivity tests connectivity to configured providers and displays results.
func testAndDisplayConnectivity(ctx context.Context, config gps.AppConfiguration, outputFormat string, writer io.Writer) error {
	// Create connectivity adapter with filesystem
	fs := filesystem.NewOSFileSystem()
	connectivityAdapter := validationAdapters.NewConnectivityAdapter(30*time.Second, fs)

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

// BuildHTTPSURL creates an HTTPS URL from a domain.
func buildHTTPSURL(domain string) string {
	if domain == "" {
		return ""
	}

	return "https://" + domain
}

// DisplayConnectivityResults formats and displays connectivity test results.
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

// DisplayConnectivityResultsConsole formats connectivity results for console output.
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

// WriteConnectivityHeader writes the header for connectivity results.
func writeConnectivityHeader(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "\nConnectivity Test Results:"); err != nil {
		return fmt.Errorf("failed to write connectivity results header: %w", err)
	}

	if _, err := fmt.Fprintln(writer, "="+strings.Repeat("=", 50)); err != nil {
		return fmt.Errorf("failed to write connectivity results separator: %w", err)
	}

	return nil
}

// WriteConnectivityResults writes individual connectivity test results.
func writeConnectivityResults(results []validation.ConnectivityResult, writer io.Writer) error {
	for _, result := range results {
		if err := writeConnectivityResult(result, writer); err != nil {
			return err
		}
	}

	return nil
}

// WriteConnectivityResult writes a single connectivity test result.
func writeConnectivityResult(result validation.ConnectivityResult, writer io.Writer) error {
	status := "✓"
	if !result.Success {
		status = "✗"
	}

	if _, err := fmt.Fprintf(writer, "%s %s", status, result.Validation.Description); err != nil {
		return fmt.Errorf("failed to write connectivity result: %w", err)
	}

	if result.Duration > 0 {
		if _, err := fmt.Fprintf(writer, " (%dms)", result.Duration.Milliseconds()); err != nil {
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

// WriteConnectivitySummary writes the summary of connectivity test results.
func writeConnectivitySummary(results []validation.ConnectivityResult, writer io.Writer) error {
	total := len(results)
	successful := countSuccessfulResults(results)

	if _, err := fmt.Fprintf(writer, "\nSummary: %d/%d tests passed\n", successful, total); err != nil {
		return fmt.Errorf("failed to write connectivity results summary: %w", err)
	}

	return nil
}

// CountSuccessfulResults counts the number of successful connectivity results.
func countSuccessfulResults(results []validation.ConnectivityResult) int {
	successful := 0

	for _, result := range results {
		if result.Success {
			successful++
		}
	}

	return successful
}

// DisplayConnectivityResultsPlain formats connectivity results for plain output.
func displayConnectivityResultsPlain(results []validation.ConnectivityResult, writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "\nCONNECTIVITY_RESULTS"); err != nil {
		return fmt.Errorf("failed to write connectivity results header: %w", err)
	}

	for index, result := range results {
		status := "PASS"
		if !result.Success {
			status = "FAIL"
		}

		if _, err := fmt.Fprintf(writer, "TEST_%d\t%s\t%s\t%d\n",
			index+1,
			status,
			result.Validation.Description,
			result.Duration.Milliseconds()); err != nil {
			return fmt.Errorf("failed to write connectivity result: %w", err)
		}
	}

	return nil
}

// DisplayConnectivityResultsJSON formats connectivity results as JSON.
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

// WriteConnectivityJSONHeader writes the JSON header for connectivity results.
func writeConnectivityJSONHeader(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "\n\"connectivity_results\": ["); err != nil {
		return fmt.Errorf("failed to write connectivity results JSON header: %w", err)
	}

	return nil
}

// WriteConnectivityJSONResults writes connectivity results in JSON format.
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

// WriteConnectivityJSONResult writes a single connectivity result in JSON format.
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

	if _, err := fmt.Fprintf(writer, "    \"duration_ms\": %d,\n", result.Duration.Milliseconds()); err != nil {
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

// WriteConnectivityJSONFooter writes the JSON footer for connectivity results.
func writeConnectivityJSONFooter(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "\n]"); err != nil {
		return fmt.Errorf("failed to write connectivity results JSON close: %w", err)
	}

	return nil
}
