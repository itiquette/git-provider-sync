// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

//nolint:funcorder // Complex formatter with many helper methods
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"itiquette/git-provider-sync/internal/adapters/terminal"
	"itiquette/git-provider-sync/internal/domain/ports"
	sync "itiquette/git-provider-sync/internal/domain/sync"
	model "itiquette/git-provider-sync/internal/model/configuration"
)

// Constants for output formats.
const (
	FormatJSON    = "json"
	formatConsole = "console"
	formatPlain   = "plain"
	statusFailed  = "failed"
	statusSkipped = "skipped"
)

// Static errors for err113 compliance.
var (
	ErrUnsupportedOutputFormat = errors.New("unsupported output format")
	ErrInvalidSyncResultsType  = errors.New("invalid sync results type")
)

// OutputFormatter implements ports.OutputFormatter for CLI applications.
// Supports console (human-readable), json (structured), and plain (tabular) output formats.
type OutputFormatter struct {
	color terminal.Color
}

// NewOutputFormatter creates a new CLI output formatter with TTY-aware color support.
// Be functional: detect TTY once at creation, immutable thereafter.
func NewOutputFormatter() ports.OutputFormatter { //nolint:ireturn // Factory function returning interface
	return NewOutputFormatterWithColorMode(terminal.ColorAuto)
}

// NewOutputFormatterWithColorMode creates formatter with explicit color mode.
// Be hexagonal: color mode is injected, not read from global state.
func NewOutputFormatterWithColorMode(mode terminal.ColorMode) ports.OutputFormatter { //nolint:ireturn // Factory function returning interface
	isTTY := terminal.IsOutput()

	return &OutputFormatter{
		color: terminal.NewColor(mode, isTTY),
	}
}

// SupportedFormats returns the list of supported output formats.
func (f *OutputFormatter) SupportedFormats() []string {
	return []string{formatConsole, FormatJSON, formatPlain}
}

// FormatConfiguration formats application configuration for output.
func (f *OutputFormatter) FormatConfiguration(appCfg model.AppConfiguration, format string, writer io.Writer) error {
	switch format {
	case formatConsole:
		return f.formatConfigurationConsole(appCfg, writer)
	case FormatJSON:
		return f.formatConfigurationJSON(appCfg, writer)
	case formatPlain:
		return f.formatConfigurationPlain(appCfg, writer)
	default:
		return fmt.Errorf("%w: %s (supported: %s)", ErrUnsupportedOutputFormat, format, strings.Join(f.SupportedFormats(), ", "))
	}
}

// formatConfigurationConsole renders configuration in structured console format with proper indentation.
func (f *OutputFormatter) formatConfigurationConsole(appCfg model.AppConfiguration, writer io.Writer) error {
	const indentSize = 2

	// Use color for header if TTY
	header := f.color.Header("Git Provider Sync Configuration")
	if _, err := fmt.Fprintf(writer, "\n%s\n", header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := fmt.Fprintln(writer, strings.Repeat("=", 30)); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	if _, err := fmt.Fprintln(writer, header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := fmt.Fprintln(writer, strings.Repeat("=", 30)); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	for envName, env := range appCfg.GitProviderSyncConfs {
		if err := f.printEnvironment(envName, env, writer, 0, indentSize); err != nil {
			return fmt.Errorf("failed to print environment %s: %w", envName, err)
		}
	}

	return nil
}

// printEnvironment renders a single environment section with hierarchical indentation.
func (f *OutputFormatter) printEnvironment(name string, env model.Environment, writer io.Writer, level, indentSize int) error {
	indent := strings.Repeat(" ", level*indentSize)

	envHeader := f.color.Header("Environment: " + name)
	if _, err := fmt.Fprintf(writer, "\n%s%s\n", indent, envHeader); err != nil {
		return fmt.Errorf("failed to write environment name: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "%s%s\n", indent, strings.Repeat("-", 20)); err != nil {
		return fmt.Errorf("failed to write environment separator: %w", err)
	}

	for sourceName, syncConfig := range env {
		if err := f.printSyncConfig(sourceName, syncConfig, writer, level+1, indentSize); err != nil {
			return fmt.Errorf("failed to print sync config %s: %w", sourceName, err)
		}
	}

	return nil
}

// printSyncConfig renders SyncConfig details with structured indentation and validation.
func (f *OutputFormatter) printSyncConfig(name string, syncCfg model.SyncConfig, writer io.Writer, level, indentSize int) error {
	indent := strings.Repeat(" ", level*indentSize)

	if err := f.writeSyncConfigHeader(writer, indent, name); err != nil {
		return err
	}

	if err := f.writeMandatoryFields(writer, indent, syncCfg); err != nil {
		return err
	}

	f.writeOptionalFields(writer, indent, syncCfg) // Don't fail on optional fields

	if !f.isEmptyAuthConfig(syncCfg.Auth) {
		if err := f.printAuthConfig(syncCfg.Auth, writer, level+1, indentSize); err != nil {
			return err
		}
	}

	if !f.isEmptyRepositoriesOption(syncCfg.Repositories) {
		if err := f.printRepositoriesOption(syncCfg.Repositories, writer, level+1, indentSize); err != nil {
			return err
		}
	}

	if len(syncCfg.Mirrors) > 0 {
		return f.printMirrorsSection(writer, syncCfg, level, indentSize)
	}

	return nil
}

func (f *OutputFormatter) writeSyncConfigHeader(writer io.Writer, indent, name string) error {
	if _, err := fmt.Fprintf(writer, "\n%sSync Configuration: %s\n", indent, name); err != nil {
		return fmt.Errorf("failed to write sync config name: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeMandatoryFields(writer io.Writer, indent string, syncCfg model.SyncConfig) error {
	fields := []struct{ label, value string }{
		{"Provider Type", syncCfg.ProviderType},
		{"Domain", syncCfg.GetDomain()},
		{"Owner", syncCfg.Owner},
		{"Owner Type", syncCfg.OwnerType},
	}

	for _, field := range fields {
		if _, err := fmt.Fprintf(writer, "%s%s: %s\n", indent, field.label, field.value); err != nil {
			return fmt.Errorf("failed to write %s: %w", field.label, err)
		}
	}

	return nil
}

func (f *OutputFormatter) writeOptionalFields(writer io.Writer, indent string, syncCfg model.SyncConfig) {
	if syncCfg.IncludeForks {
		fmt.Fprintf(writer, "%sInclude Forks: %t\n", indent, syncCfg.IncludeForks) //nolint:errcheck // Best effort output
	}

	if syncCfg.UseGitBinary {
		fmt.Fprintf(writer, "%sUse Git Binary: %t\n", indent, syncCfg.UseGitBinary) //nolint:errcheck // Best effort output
	}

	if syncCfg.ActiveFromLimit != "" {
		fmt.Fprintf(writer, "%sActive From Limit: %s\n", indent, syncCfg.ActiveFromLimit) //nolint:errcheck // Best effort output
	}
}

func (f *OutputFormatter) printMirrorsSection(writer io.Writer, syncCfg model.SyncConfig, level, indentSize int) error {
	indentSub := strings.Repeat(" ", level*indentSize)
	if _, err := fmt.Fprintf(writer, "\n%sMirror Configurations:\n", indentSub); err != nil {
		return fmt.Errorf("failed to write mirror configurations header: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "%s%s\n", "  ", strings.Repeat("-", 20)); err != nil {
		return fmt.Errorf("failed to write mirror configurations separator: %w", err)
	}

	for mirrorName, mirror := range syncCfg.Mirrors {
		if err := f.printMirrorConfig(mirrorName, mirror, writer, level+1, indentSize); err != nil {
			return fmt.Errorf("failed to print mirror config %s: %w", mirrorName, err)
		}
	}

	return nil
}

// printAuthConfig renders authentication configuration with security token masking.
func (f *OutputFormatter) printAuthConfig(authCfg model.AuthConfig, writer io.Writer, level, indentSize int) error {
	indent := strings.Repeat(" ", level*indentSize)

	if err := f.writeAuthHeader(writer, indent); err != nil {
		return err
	}

	if err := f.writeAuthMandatoryFields(writer, indent, authCfg); err != nil {
		return err
	}

	if err := f.writeAuthOptionalFields(writer, indent, authCfg); err != nil {
		return err
	}

	if err := f.writeAuthSSHConfiguration(writer, indent, authCfg); err != nil {
		return err
	}

	return nil
}

// printMirrorConfig renders mirror configuration with hierarchical structure validation.
func (f *OutputFormatter) printMirrorConfig(name string, mirrorCfg model.MirrorConfig, writer io.Writer, level, indentSize int) error {
	indent := strings.Repeat(" ", level*indentSize)

	if err := f.writeMirrorConfigHeader(writer, indent, name); err != nil {
		return err
	}

	if err := f.writeMirrorConfigMandatoryFields(writer, indent, mirrorCfg); err != nil {
		return err
	}

	if err := f.writeMirrorConfigOptionalFields(writer, indent, mirrorCfg); err != nil {
		return err
	}

	if err := f.writeMirrorConfigSections(writer, mirrorCfg, level, indentSize); err != nil {
		return err
	}

	return nil
}

func (f *OutputFormatter) writeMirrorConfigHeader(writer io.Writer, indent, name string) error {
	if _, err := fmt.Fprintf(writer, "\n%sMirror: %s\n", indent, name); err != nil {
		return fmt.Errorf("failed to write mirror name: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeMirrorConfigMandatoryFields(writer io.Writer, indent string, mirrorCfg model.MirrorConfig) error {
	if _, err := fmt.Fprintf(writer, "%sType: %s\n", indent, mirrorCfg.ProviderType); err != nil {
		return fmt.Errorf("failed to write mirror type: %w", err)
	}

	if mirrorCfg.Domain != "" {
		if _, err := fmt.Fprintf(writer, "%sDomain: %s\n", indent, mirrorCfg.GetDomain()); err != nil {
			return fmt.Errorf("failed to write mirror domain: %w", err)
		}
	}

	if mirrorCfg.Owner != "" {
		if _, err := fmt.Fprintf(writer, "%sOwner: %s\n", indent, mirrorCfg.Owner); err != nil {
			return fmt.Errorf("failed to write mirror owner: %w", err)
		}
	}

	if _, err := fmt.Fprintf(writer, "%sOwner Type: %s\n", indent, mirrorCfg.OwnerType); err != nil {
		return fmt.Errorf("failed to write mirror owner type: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeMirrorConfigOptionalFields(writer io.Writer, indent string, mirrorCfg model.MirrorConfig) error {
	if mirrorCfg.UseGitBinary {
		if _, err := fmt.Fprintf(writer, "%sUse Git Binary: %t\n", indent, mirrorCfg.UseGitBinary); err != nil {
			return fmt.Errorf("failed to write mirror use git binary: %w", err)
		}
	}

	if mirrorCfg.Path != "" {
		if _, err := fmt.Fprintf(writer, "%sPath: %s\n", indent, mirrorCfg.Path); err != nil {
			return fmt.Errorf("failed to write mirror path: %w", err)
		}
	}

	return nil
}

func (f *OutputFormatter) writeMirrorConfigSections(writer io.Writer, mirrorCfg model.MirrorConfig, level, indentSize int) error {
	// Print Mirror Settings if they're not empty
	if !f.isEmptyMirrorSettings(mirrorCfg.Settings) {
		if err := f.printMirrorSettings(mirrorCfg.Settings, writer, level+1, indentSize); err != nil {
			return fmt.Errorf("failed to print mirror settings: %w", err)
		}
	}

	// Print Mirror Auth Configuration if it's not empty
	if !f.isEmptyAuthConfig(mirrorCfg.Auth) {
		if err := f.printAuthConfig(mirrorCfg.Auth, writer, level+1, indentSize); err != nil {
			return fmt.Errorf("failed to print mirror auth config: %w", err)
		}
	}

	return nil
}

// printMirrorSettings renders mirror-specific settings with conditional field display.
func (f *OutputFormatter) printMirrorSettings(settings model.MirrorSettings, writer io.Writer, level, indentSize int) error {
	indent := strings.Repeat(" ", level*indentSize)

	if err := f.writeMirrorSettingsHeader(writer, indent); err != nil {
		return err
	}

	if err := f.writeMirrorSettingsFields(writer, indent, settings); err != nil {
		return err
	}

	return nil
}

func (f *OutputFormatter) writeMirrorSettingsHeader(writer io.Writer, indent string) error {
	if _, err := fmt.Fprintf(writer, "\n%sSettings:\n", indent); err != nil {
		return fmt.Errorf("failed to write settings header: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeMirrorSettingsFields(writer io.Writer, indent string, settings model.MirrorSettings) error {
	// Define settings fields to write (only non-default values)
	settingsFields := []struct {
		condition bool
		label     string
		value     any
	}{
		{settings.AlphaNumHyphName, "ASCII Name", settings.AlphaNumHyphName},
		{settings.DescriptionPrefix != "", "Description Prefix", settings.DescriptionPrefix},
		{settings.Disabled, "Disabled", settings.Disabled},
		{settings.ForcePush, "Force Push", settings.ForcePush},
		{settings.GitHubUploadURL != "", "GitHub Upload URL", settings.GitHubUploadURL},
		{settings.IgnoreInvalidName, "Ignore Invalid Name", settings.IgnoreInvalidName},
		{settings.Visibility != "", "Visibility", settings.Visibility},
	}

	for _, field := range settingsFields {
		if field.condition {
			if _, err := fmt.Fprintf(writer, "%s%s: %v\n", indent, field.label, field.value); err != nil {
				return fmt.Errorf("failed to write %s: %w", field.label, err)
			}
		}
	}

	return nil
}

// printRepositoriesOption renders repository filter patterns with include/exclude separation.
func (f *OutputFormatter) printRepositoriesOption(opt model.RepositoriesOption, writer io.Writer, level, indentSize int) error {
	indent := strings.Repeat(" ", level*indentSize)
	if _, err := fmt.Fprintf(writer, "\n%sRepositories:\n", indent); err != nil {
		return fmt.Errorf("failed to write repositories header: %w", err)
	}

	if len(opt.Include) > 0 {
		if _, err := fmt.Fprintf(writer, "%sInclude:\n", indent); err != nil {
			return fmt.Errorf("failed to write include header: %w", err)
		}

		for _, pattern := range opt.Include {
			if _, err := fmt.Fprintf(writer, "%s  %s\n", indent, pattern); err != nil {
				return fmt.Errorf("failed to write include pattern: %w", err)
			}
		}
	}

	if len(opt.Exclude) > 0 {
		if _, err := fmt.Fprintf(writer, "%sExclude:\n", indent); err != nil {
			return fmt.Errorf("failed to write exclude header: %w", err)
		}

		for _, pattern := range opt.Exclude {
			if _, err := fmt.Fprintf(writer, "%s  %s\n", indent, pattern); err != nil {
				return fmt.Errorf("failed to write exclude pattern: %w", err)
			}
		}
	}

	return nil
}

// Helper functions to check if configurations are empty.
func (f *OutputFormatter) isEmptyAuthConfig(authCfg model.AuthConfig) bool {
	return authCfg.Protocol == "" &&
		authCfg.HTTPScheme == "" &&
		authCfg.Token == "" &&
		authCfg.ProxyURL == "" &&
		authCfg.CertDirPath == "" &&
		authCfg.SSHCommand == "" &&
		authCfg.SSHURLRewriteFrom == "" &&
		authCfg.SSHURLRewriteTo == ""
}

func (f *OutputFormatter) isEmptyRepositoriesOption(opt model.RepositoriesOption) bool {
	return len(opt.Include) == 0 && len(opt.Exclude) == 0
}

func (f *OutputFormatter) isEmptyMirrorSettings(settings model.MirrorSettings) bool {
	return !settings.AlphaNumHyphName &&
		settings.DescriptionPrefix == "" &&
		!settings.Disabled &&
		!settings.ForcePush &&
		settings.GitHubUploadURL == "" &&
		!settings.IgnoreInvalidName &&
		settings.Visibility == ""
}

// formatConfigurationJSON outputs configuration as structured JSON.
func (f *OutputFormatter) formatConfigurationJSON(appCfg model.AppConfiguration, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(appCfg); err != nil {
		return fmt.Errorf("failed to encode configuration as JSON: %w", err)
	}

	return nil
}

// formatConfigurationPlain outputs configuration as tabular text for pipeline compatibility.
// One record per line, tab-separated for easy parsing with grep, awk, cut.
// Human-second, machine-first format.
func (f *OutputFormatter) formatConfigurationPlain(appCfg model.AppConfiguration, writer io.Writer) error {
	// Tab-separated header for column identification
	if _, err := fmt.Fprintln(writer, "ENVIRONMENT\tSOURCE\tPROVIDER\tDOMAIN\tOWNER\tOWNER_TYPE\tMIRRORS"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// One line per source configuration
	for envName, env := range appCfg.GitProviderSyncConfs {
		for sourceName, syncConfig := range env {
			// Collect mirror names for this source
			mirrorNames := make([]string, 0, len(syncConfig.Mirrors))
			for mirrorName := range syncConfig.Mirrors {
				mirrorNames = append(mirrorNames, mirrorName)
			}

			mirrorList := "-"
			if len(mirrorNames) > 0 {
				mirrorList = strings.Join(mirrorNames, ",")
			}

			// Tab-separated values, no multi-line cells
			line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s",
				envName,
				sourceName,
				syncConfig.ProviderType,
				syncConfig.GetDomain(),
				syncConfig.Owner,
				syncConfig.OwnerType,
				mirrorList,
			)
			if _, err := fmt.Fprintln(writer, line); err != nil {
				return fmt.Errorf("failed to write configuration line: %w", err)
			}
		}
	}

	return nil
}

// FormatSyncResults formats sync operation results for output.
// Progress and status information should go to stderr, data to stdout.
func (f *OutputFormatter) FormatSyncResults(results any, format string, dataWriter, progressWriter io.Writer) error {
	switch format {
	case formatConsole:
		return f.formatSyncResultsConsole(results, dataWriter, progressWriter)
	case FormatJSON:
		return f.formatSyncResultsJSON(results, dataWriter, progressWriter)
	case formatPlain:
		return f.formatSyncResultsPlain(results, dataWriter, progressWriter)
	default:
		return fmt.Errorf("%w: %s (supported: %s)", ErrUnsupportedOutputFormat, format, strings.Join(f.SupportedFormats(), ", "))
	}
}

// formatSyncResultsConsole outputs human-readable sync results.
func (f *OutputFormatter) formatSyncResultsConsole(results any, dataWriter, progressWriter io.Writer) error {
	syncResults, ok := results.(*sync.Results)
	if !ok {
		return ErrInvalidSyncResultsType
	}

	if err := f.writeSyncProgress(syncResults, progressWriter); err != nil {
		return err
	}

	if err := f.writeSyncSummary(syncResults, dataWriter); err != nil {
		return err
	}

	if err := f.writeSyncDetailedResults(syncResults, dataWriter); err != nil {
		return err
	}

	return nil
}

func (f *OutputFormatter) writeSyncProgress(syncResults *sync.Results, progressWriter io.Writer) error {
	// Progress to stderr, like git
	// Be idiomatic: inform about state changes
	if syncResults.DryRun {
		if _, err := fmt.Fprintf(progressWriter, "Dry run completed in %.2f seconds\n", syncResults.DurationSeconds); err != nil {
			return fmt.Errorf("failed to write progress: %w", err)
		}
	} else {
		if _, err := fmt.Fprintf(progressWriter, "Synchronization completed in %.2f seconds\n", syncResults.DurationSeconds); err != nil {
			return fmt.Errorf("failed to write progress: %w", err)
		}
	}

	return nil
}

func (f *OutputFormatter) writeSyncSummary(syncResults *sync.Results, dataWriter io.Writer) error {
	if err := f.writeSyncSummaryHeader(dataWriter); err != nil {
		return err
	}

	if err := f.writeSyncSummaryStats(syncResults, dataWriter); err != nil {
		return err
	}

	if err := f.writeSyncSummaryMode(syncResults, dataWriter); err != nil {
		return err
	}

	return nil
}

func (f *OutputFormatter) writeSyncSummaryHeader(dataWriter io.Writer) error {
	// Clear indication that state has changed
	header := f.color.Header("Repository State Changes")
	if _, err := fmt.Fprintf(dataWriter, "\n%s\n", header); err != nil {
		return fmt.Errorf("failed to write summary header: %w", err)
	}

	if _, err := fmt.Fprintf(dataWriter, "========================\n"); err != nil {
		return fmt.Errorf("failed to write summary separator: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeSyncSummaryStats(syncResults *sync.Results, dataWriter io.Writer) error {
	stats := []struct {
		label string
		value any
	}{
		{"Total Sources", syncResults.TotalSources},
		{"Total Mirrors", syncResults.TotalMirrors},
		{"Total Repositories", syncResults.TotalRepositories},
		{"Successful Syncs", syncResults.SuccessfulSyncs},
		{"Failed Syncs", syncResults.FailedSyncs},
		{"Skipped Syncs", syncResults.SkippedSyncs},
		{"Duration", fmt.Sprintf("%.2f seconds", syncResults.DurationSeconds)},
	}

	for _, stat := range stats {
		if _, err := fmt.Fprintf(dataWriter, "%s: %v\n", stat.label, stat.value); err != nil {
			return fmt.Errorf("failed to write %s: %w", stat.label, err)
		}
	}

	return nil
}

func (f *OutputFormatter) writeSyncSummaryMode(syncResults *sync.Results, dataWriter io.Writer) error {
	if syncResults.DryRun {
		if _, err := fmt.Fprintf(dataWriter, "Mode: DRY RUN\n"); err != nil {
			return fmt.Errorf("failed to write dry run mode: %w", err)
		}
	}

	return nil
}

func (f *OutputFormatter) writeSyncDetailedResults(syncResults *sync.Results, dataWriter io.Writer) error {
	if len(syncResults.Results) == 0 {
		return nil
	}

	if err := f.writeSyncDetailedResultsHeader(dataWriter); err != nil {
		return err
	}

	for _, result := range syncResults.Results {
		if err := f.writeSyncDetailedResult(result, dataWriter); err != nil {
			return err
		}
	}

	return nil
}

func (f *OutputFormatter) writeSyncDetailedResultsHeader(dataWriter io.Writer) error {
	// Like git push, show what happened to each repository
	if _, err := fmt.Fprintf(dataWriter, "\nTo mirrors:\n"); err != nil {
		return fmt.Errorf("failed to write detailed results header: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeSyncDetailedResult(result sync.Result, dataWriter io.Writer) error {
	// Like git push, show clear state changes
	// Format: source/repo -> target: action status
	var (
		statusSymbol string
		statusMsg    string
	)

	switch result.Status {
	case "success", "synced":
		statusSymbol = f.color.Success("✓")
		statusMsg = "done"
	case statusFailed, "error":
		statusSymbol = f.color.Error("✗")

		statusMsg = statusFailed
		if result.Error != "" {
			statusMsg = "failed: " + result.Error
		}
	case statusSkipped:
		statusSymbol = "-"
		statusMsg = statusSkipped
	default:
		statusSymbol = "?"
		statusMsg = result.Status
	}

	// Show what changed, git-push style
	// Be idiomatic: clear and concise state changes

	// Format like: " ✓ repo-name -> mirror-name: synced (1.2s)"
	if _, err := fmt.Fprintf(dataWriter, " %s %s -> %s: %s (%.1fs)\n",
		statusSymbol,
		result.Repository,
		result.Mirror,
		statusMsg,
		result.DurationSeconds); err != nil {
		return fmt.Errorf("failed to write detailed result: %w", err)
	}

	return nil
}

// formatSyncResultsJSON outputs structured JSON sync results.
func (f *OutputFormatter) formatSyncResultsJSON(results any, dataWriter, progressWriter io.Writer) error {
	syncResults, ok := results.(*sync.Results)
	if !ok {
		return ErrInvalidSyncResultsType
	}

	// Progress to stderr
	if _, err := fmt.Fprintf(progressWriter, "Sync completed in %.2f seconds\n", syncResults.DurationSeconds); err != nil {
		return fmt.Errorf("failed to write progress: %w", err)
	}

	// Structured data to stdout
	encoder := json.NewEncoder(dataWriter)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(syncResults); err != nil {
		return fmt.Errorf("failed to encode JSON results: %w", err)
	}

	return nil
}

// formatSyncResultsPlain outputs tabular sync results for pipeline compatibility.
// formatSyncResultsPlain outputs sync results as tab-separated values for pipeline processing.
// Machine-readable format: one record per line, no formatting, suitable for grep/awk.
func (f *OutputFormatter) formatSyncResultsPlain(results any, dataWriter, _ io.Writer) error {
	syncResults, ok := results.(*sync.Results)
	if !ok {
		return ErrInvalidSyncResultsType
	}

	// Simple tab-separated header for column identification
	if _, err := fmt.Fprintf(dataWriter, "STATUS\tREPOSITORY\tSOURCE\tMIRROR\tDURATION\tERROR\n"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// One line per repository sync result
	for _, result := range syncResults.Results {
		// Use simple status codes for easy parsing
		status := "OK"

		switch result.Status {
		case "failed":
			status = "FAIL"
		case "skipped":
			status = "SKIP"
		}

		// Empty string for no error (easier to parse than "-")
		errorField := ""
		if result.Error != "" {
			// Replace tabs and newlines in error messages to maintain one line per record
			errorField = strings.ReplaceAll(strings.ReplaceAll(result.Error, "\t", " "), "\n", " ")
		}

		// Tab-separated values, guaranteed one line per record
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%.2f\t%s",
			status,
			result.Repository,
			result.Source,
			result.Mirror,
			result.DurationSeconds,
			errorField,
		)
		if _, err := fmt.Fprintln(dataWriter, line); err != nil {
			return fmt.Errorf("failed to write sync result: %w", err)
		}
	}

	return nil
}

func (f *OutputFormatter) writeAuthHeader(writer io.Writer, indent string) error {
	if _, err := fmt.Fprintf(writer, "\n%sAuthentication:\n", indent); err != nil {
		return fmt.Errorf("failed to write authentication header: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeAuthMandatoryFields(writer io.Writer, indent string, authCfg model.AuthConfig) error {
	if _, err := fmt.Fprintf(writer, "%sProtocol: %s\n", indent, authCfg.Protocol); err != nil {
		return fmt.Errorf("failed to write protocol: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeAuthOptionalFields(writer io.Writer, indent string, authCfg model.AuthConfig) error {
	if authCfg.HTTPScheme != "" {
		if _, err := fmt.Fprintf(writer, "%sHTTP Scheme: %s\n", indent, authCfg.HTTPScheme); err != nil {
			return fmt.Errorf("failed to write HTTP scheme: %w", err)
		}
	}

	if authCfg.Token != "" {
		if _, err := fmt.Fprintf(writer, "%sToken: <*****>\n", indent); err != nil {
			return fmt.Errorf("failed to write token: %w", err)
		}
	}

	if authCfg.ProxyURL != "" {
		if _, err := fmt.Fprintf(writer, "%sProxy URL: %s\n", indent, authCfg.ProxyURL); err != nil {
			return fmt.Errorf("failed to write proxy URL: %w", err)
		}
	}

	if authCfg.CertDirPath != "" {
		if _, err := fmt.Fprintf(writer, "%sCertificate Directory: %s\n", indent, authCfg.CertDirPath); err != nil {
			return fmt.Errorf("failed to write certificate directory: %w", err)
		}
	}

	return nil
}

func (f *OutputFormatter) writeAuthSSHConfiguration(writer io.Writer, indent string, authCfg model.AuthConfig) error {
	if !f.hasSSHConfiguration(authCfg) {
		return nil
	}

	if err := f.writeSSHHeader(writer, indent); err != nil {
		return err
	}

	return f.writeSSHFields(writer, indent, authCfg)
}

// hasSSHConfiguration checks if any SSH configuration fields are set.
func (f *OutputFormatter) hasSSHConfiguration(authCfg model.AuthConfig) bool {
	return authCfg.SSHCommand != "" || authCfg.SSHURLRewriteFrom != "" || authCfg.SSHURLRewriteTo != ""
}

func (f *OutputFormatter) writeSSHHeader(writer io.Writer, indent string) error {
	if _, err := fmt.Fprintf(writer, "\n%sSSH Configuration:\n", indent); err != nil {
		return fmt.Errorf("failed to write SSH configuration header: %w", err)
	}

	return nil
}

func (f *OutputFormatter) writeSSHFields(writer io.Writer, indent string, authCfg model.AuthConfig) error {
	if authCfg.SSHCommand != "" {
		if _, err := fmt.Fprintf(writer, "%sCommand: %s\n", indent, authCfg.SSHCommand); err != nil {
			return fmt.Errorf("failed to write SSH command: %w", err)
		}
	}

	if authCfg.SSHURLRewriteFrom != "" {
		if _, err := fmt.Fprintf(writer, "%sURL Rewrite From: %s\n", indent, authCfg.SSHURLRewriteFrom); err != nil {
			return fmt.Errorf("failed to write SSH URL rewrite from: %w", err)
		}
	}

	if authCfg.SSHURLRewriteTo != "" {
		if _, err := fmt.Fprintf(writer, "%sURL Rewrite To: %s\n", indent, authCfg.SSHURLRewriteTo); err != nil {
			return fmt.Errorf("failed to write SSH URL rewrite to: %w", err)
		}
	}

	return nil
}
