// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/adapters/terminal"
)

// PanicHandler handles unexpected panics and provides user-friendly error messages.
type PanicHandler struct {
	writer   io.Writer
	version  string
	crashDir string
	color    terminal.Color
}

// NewPanicHandler creates a new panic handler.
func NewPanicHandler(writer io.Writer, version string) *PanicHandler {
	homeDir, _ := os.UserHomeDir()
	crashDir := filepath.Join(homeDir, ".local", "share", "git-provider-sync", "crashes")

	// Check if writer is a TTY for color detection
	isTTY := false

	if f, ok := writer.(*os.File); ok {
		stat, _ := f.Stat()
		isTTY = (stat.Mode() & os.ModeCharDevice) != 0
	}

	return &PanicHandler{
		writer:   writer,
		version:  version,
		crashDir: crashDir,
		color:    terminal.NewColor(terminal.ColorAuto, isTTY),
	}
}

// HandlePanic recovers from a panic and provides user-friendly output.
func (h *PanicHandler) HandlePanic() {
	if r := recover(); r != nil {
		h.handlePanicRecovery(r)
		os.Exit(99) // Exit code 99 for unexpected errors
	}
}

func (h *PanicHandler) handlePanicRecovery(recovered any) {
	// Get stack trace
	stack := debug.Stack()

	// Try to save crash report
	crashFile := h.saveCrashReport(recovered, stack)

	// Display user-friendly message
	h.displayPanicMessage(recovered, crashFile)
}

func (h *PanicHandler) saveCrashReport(recovered any, stack []byte) string {
	// Create crash directory if it doesn't exist
	if err := os.MkdirAll(h.crashDir, 0750); err != nil {
		// Can't save crash report, return empty
		return ""
	}

	// Generate crash filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("crash-%s.txt", timestamp)
	crashPath := filepath.Join(h.crashDir, filename)

	// Write crash report
	content := h.formatCrashReport(recovered, stack)
	if err := os.WriteFile(crashPath, []byte(content), 0600); err != nil {
		return ""
	}

	return crashPath
}

func (h *PanicHandler) formatCrashReport(recovered any, stack []byte) string {
	var builder strings.Builder

	builder.WriteString("Git Provider Sync Crash Report\n")
	builder.WriteString(strings.Repeat("=", 60) + "\n\n")

	builder.WriteString("Version: " + h.version + "\n")
	builder.WriteString("Time: " + time.Now().Format(time.RFC3339) + "\n")
	builder.WriteString("OS: " + runtime.GOOS + "/" + runtime.GOARCH + "\n")
	builder.WriteString("Go Version: " + runtime.Version() + "\n\n")

	builder.WriteString("Panic: ")
	builder.WriteString(fmt.Sprintf("%v", recovered))
	builder.WriteString("\n\n")

	builder.WriteString("Stack Trace:\n")
	builder.WriteString(strings.Repeat("-", 60) + "\n")
	builder.Write(stack)

	return builder.String()
}

func (h *PanicHandler) displayPanicMessage(recovered any, crashFile string) {
	symbols := GetSymbols(terminal.ColorAuto)

	// Title with error symbol
	_, _ = fmt.Fprintf(h.writer, "\n%s %s\n",
		h.color.Error(symbols.Cross),
		h.color.Header("Unexpected Error"))

	// Brief explanation
	_, _ = fmt.Fprintf(h.writer, "\nThe application encountered an unexpected error and had to stop.\n")

	// Error details (simplified)
	errMsg := fmt.Sprintf("%v", recovered)
	if len(errMsg) > 100 {
		errMsg = errMsg[:97] + "..."
	}

	_, _ = fmt.Fprintf(h.writer, "\n%s %s\n",
		h.color.Header("Error:"),
		errMsg)

	// Crash report location if saved
	if crashFile != "" {
		_, _ = fmt.Fprintf(h.writer, "\n%s Crash report saved to:\n   %s\n",
			symbols.Info,
			crashFile)
	}

	// Next steps
	_, _ = fmt.Fprintf(h.writer, "\n%s\n",
		h.color.Header("What to do next:"))

	_, _ = fmt.Fprintf(h.writer, "\n1. Try running the command again\n")
	_, _ = fmt.Fprintf(h.writer, "2. Check your configuration file for errors\n")
	_, _ = fmt.Fprintf(h.writer, "3. Run with --log-level=debug for more details\n")

	// Report this issue
	_, _ = fmt.Fprintf(h.writer, "\n%s\n",
		h.color.Header("Report this issue:"))

	bugURL := h.generateBugReportURL(recovered, crashFile)
	_, _ = fmt.Fprintf(h.writer, "\n   %s\n", bugURL)

	if crashFile != "" {
		_, _ = fmt.Fprintf(h.writer, "\n   Please attach the crash report when filing the issue.\n")
	}

	_, _ = fmt.Fprintln(h.writer)
}

func (h *PanicHandler) generateBugReportURL(recovered any, crashFile string) string {
	baseURL := "https://github.com/itiquette/git-provider-sync/issues/new"

	// Build proper URL with query parameters
	params := url.Values{}
	params.Add("labels", "bug")
	params.Add("template", "bug_report.yml")

	// Generate a concise title
	title := fmt.Sprintf("[BUG] Unexpected error: %v", recovered)
	if len(title) > 80 {
		title = title[:77] + "..."
	}

	params.Add("title", title)

	// For YAML templates, we can't pre-fill the body fields
	// So we'll format a body that users can copy-paste if needed
	body := h.formatIssueBody(recovered, crashFile)
	params.Add("body", body)

	// Parse base URL and add query parameters
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// Fallback to simple URL if parsing fails
		return baseURL
	}

	parsedURL.RawQuery = params.Encode()

	return parsedURL.String()
}

func (h *PanicHandler) formatIssueBody(recovered any, crashFile string) string {
	var builder strings.Builder

	builder.WriteString("## Description\n")
	builder.WriteString(fmt.Sprintf("Unexpected error occurred: %v\n\n", recovered))

	builder.WriteString("## Error Details\n")
	builder.WriteString("```\n")
	builder.WriteString(fmt.Sprintf("%v\n", recovered))
	builder.WriteString("```\n\n")

	builder.WriteString("## System Information\n")
	builder.WriteString(fmt.Sprintf("- Version: %s\n", h.version))
	builder.WriteString(fmt.Sprintf("- OS: %s\n", runtime.GOOS))
	builder.WriteString(fmt.Sprintf("- Architecture: %s\n", runtime.GOARCH))
	builder.WriteString(fmt.Sprintf("- Go Version: %s\n\n", runtime.Version()))

	if crashFile != "" {
		builder.WriteString("## Crash Report\n")
		builder.WriteString(fmt.Sprintf("A crash report was saved to: `%s`\n", crashFile))
		builder.WriteString("Please attach this file to the issue.\n\n")
	}

	builder.WriteString("## Additional Context\n")
	builder.WriteString("Please add any additional context about the problem here.\n")

	return builder.String()
}
