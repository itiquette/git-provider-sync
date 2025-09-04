// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DebugWriter tees debug output to both console and a debug log file.
type DebugWriter struct {
	console  io.Writer
	file     *os.File
	filePath string
	mu       sync.Mutex
}

// NewDebugWriter creates a new debug writer that writes to both console and file
// Returns the original writer if debug mode is not enabled or file creation fails.
func NewDebugWriter(console io.Writer, enabled bool) (io.Writer, string, error) {
	if !enabled {
		return console, "", nil
	}

	// Create debug log directory
	tmpDir := os.TempDir()

	debugDir := filepath.Join(tmpDir, "git-provider-sync-debug")
	if err := os.MkdirAll(debugDir, 0750); err != nil {
		// Fall back to console only - ignore directory creation error
		//nolint:nilerr // Intentionally returning nil - debug logging is optional
		return console, "", nil
	}

	// Create unique debug log file
	timestamp := time.Now().Format("20060102-150405")
	pid := os.Getpid()
	filename := fmt.Sprintf("gps-debug-%s-%d.log", timestamp, pid)
	filePath := filepath.Join(debugDir, filename)

	file, err := os.Create(filePath) //nolint:gosec // Path is constructed internally with timestamp //nolint:gosec // Path is controlled, not user input
	if err != nil {
		// Fall back to console only - ignore file creation error
		//nolint:nilerr // Intentionally returning nil - debug logging is optional
		return console, "", nil
	}

	// Write header to debug file
	header := fmt.Sprintf("Git Provider Sync Debug Log\n"+
		"Started: %s\n"+
		"PID: %d\n"+
		"%s\n\n",
		time.Now().Format(time.RFC3339),
		pid,
		"============================================================")
	_, _ = file.WriteString(header)

	return &DebugWriter{
		console:  console,
		file:     file,
		filePath: filePath,
	}, filePath, nil
}

// Write writes data to both console and debug file.
func (w *DebugWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Write to console first (user sees this)
	bytesWritten, err := w.console.Write(data)
	if err != nil {
		return bytesWritten, fmt.Errorf("console write failed: %w", err)
	}

	// Write to debug file (best effort)
	if w.file != nil {
		_, _ = w.file.Write(data) // Ignore errors for debug file
	}

	return bytesWritten, nil
}

// Path returns the debug file path.
func (w *DebugWriter) Path() string {
	return w.filePath
}

// Close closes the debug file if open.
func (w *DebugWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("failed to close debug file: %w", err)
		}

		w.file = nil
	}

	return nil
}

// TeeWriter creates a writer that tees output to a debug file when enabled
// is a convenience function for setting up debug logging.
func TeeWriter(console io.Writer, logLevel string) (io.Writer, string) {
	// Enable debug file for debug and trace levels
	enabled := logLevel == "debug" || logLevel == "trace"

	writer, debugPath, err := NewDebugWriter(console, enabled)
	if err != nil {
		// Fall back to console only
		return console, ""
	}

	return writer, debugPath
}
