// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package auth

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// TokenReader handles reading authentication tokens from various sources.
type TokenReader struct {
	fileSystem ports.FileSystem
}

// NewTokenReader creates a new TokenReader with the given filesystem.
func NewTokenReader(fileSystem ports.FileSystem) *TokenReader {
	return &TokenReader{
		fileSystem: fileSystem,
	}
}

// ReadTokenFromStdin reads a token from stdin
// Used with --with-token flag following GitHub CLI pattern
// Supports both piped input and interactive terminal input.
func ReadTokenFromStdin() (string, error) {
	// Check if stdin is a terminal
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// Interactive mode - hide input for security
		fmt.Fprint(os.Stderr, "Enter token: ")
		byteToken, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr) // New line after hidden input

		if err != nil {
			return "", fmt.Errorf("failed to read token from terminal: %w", err)
		}

		return strings.TrimSpace(string(byteToken)), nil
	}

	// Non-interactive mode - read from pipe
	reader := bufio.NewReader(os.Stdin)

	token, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read token from stdin: %w", err)
	}

	return strings.TrimSpace(token), nil
}

// ReadTokenFromFile reads a token from a file
// File should contain only the token, with optional whitespace
// Use "-" to read from stdin (same as ReadTokenFromStdin).
func (tr *TokenReader) ReadTokenFromFile(path string) (string, error) {
	if path == "-" {
		return ReadTokenFromStdin()
	}

	data, err := tr.fileSystem.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read token file %s: %w", path, err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file %s is empty", path)
	}

	return token, nil
}

// ReadTokenFromFile is a legacy function that creates a TokenReader with OS filesystem.
// Deprecated: Use TokenReader.ReadTokenFromFile instead for better testability.
func ReadTokenFromFile(path string) (string, error) {
	// Import cycle prevents using filesystem.NewOSFileSystem() here
	// This is a compatibility shim - new code should use TokenReader
	if path == "-" {
		return ReadTokenFromStdin()
	}

	data, err := os.ReadFile(path) //nolint:gosec // User provides the path explicitly
	if err != nil {
		return "", fmt.Errorf("failed to read token file %s: %w", path, err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file %s is empty", path)
	}

	return token, nil
}

// ValidateToken performs basic validation on a token
// Returns an error if the token appears invalid.
func ValidateToken(token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	// Check for common placeholder values
	if strings.HasPrefix(token, "${") && strings.HasSuffix(token, "}") {
		return fmt.Errorf("token appears to be an unexpanded variable: %s", token)
	}

	// Warn about tokens that look like they might be exposed
	if strings.Contains(token, " ") {
		return errors.New("token contains spaces, which is unusual")
	}

	return nil
}
