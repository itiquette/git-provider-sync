// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// errorhandling.go - Error handling utilities restored from main branch using hexagonal architecture
package utilities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ErrorHandler provides sophisticated error handling capabilities.
// This restores the advanced error handling functionality from main branch.
type ErrorHandler struct {
	logger ports.Logger
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(logger ports.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError provides user-friendly error handling with contextual messages.
// This restores the main branch error.HandleError functionality.
func (eh *ErrorHandler) HandleError(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	// Log the original error for debugging
	eh.logger.Error(ctx, "Operation failed", map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
	})

	// Provide user-friendly message based on error type
	userFriendlyMsg := eh.provideUserFriendlyMessage(err, operation)

	// Create new error with context
	contextualError := fmt.Errorf("%s failed: %s", operation, userFriendlyMsg)

	eh.logger.Debug(ctx, "Generated user-friendly error", map[string]interface{}{
		"original":     err.Error(),
		"userFriendly": userFriendlyMsg,
	})

	return contextualError
}

// provideUserFriendlyMessage converts technical errors to user-friendly messages.
// This restores the main branch error handling patterns.
func (eh *ErrorHandler) provideUserFriendlyMessage(err error, operation string) string {
	errStr := strings.ToLower(err.Error())

	// Network-related errors
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "network unreachable") {
		return "network connection failed - check your internet connection and provider domain"
	}

	// Authentication errors
	if strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "token") {
		return "authentication failed - check your access token or credentials"
	}

	// Permission errors
	if strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "access denied") {
		return "insufficient permissions - check repository access rights"
	}

	// Repository not found
	if strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "404") {
		return "repository or resource not found - verify repository name and access"
	}

	// Git operation errors
	if strings.Contains(errStr, "git") && strings.Contains(errStr, "clone") {
		return "git clone operation failed - check repository URL and access permissions"
	}

	if strings.Contains(errStr, "git") && strings.Contains(errStr, "push") {
		return "git push operation failed - check target repository permissions and conflicts"
	}

	// File system errors
	if strings.Contains(errStr, "no space left") {
		return "insufficient disk space - free up disk space and try again"
	}

	if strings.Contains(errStr, "permission denied") && strings.Contains(operation, "directory") {
		return "directory access denied - check file system permissions"
	}

	// Rate limiting
	if strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests") {
		return "API rate limit exceeded - wait before retrying or check rate limits"
	}

	// Timeout errors
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") {
		return "operation timed out - try again or increase timeout settings"
	}

	// Repository name validation
	if strings.Contains(errStr, "invalid") && strings.Contains(errStr, "name") {
		return "invalid repository name - check naming requirements for target provider"
	}

	// Generic fallback
	return fmt.Sprintf("operation failed - %s", err.Error())
}

// WrapError wraps an error with additional context.
func (eh *ErrorHandler) WrapError(err error, context string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if len(args) > 0 {
		context = fmt.Sprintf(context, args...)
	}

	return fmt.Errorf("%s: %w", context, err)
}

// IsRetryableError determines if an error is retryable.
func (eh *ErrorHandler) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Network errors are usually retryable
	retryablePatterns := []string{
		"connection refused",
		"timeout",
		"deadline exceeded",
		"temporary failure",
		"rate limit",
		"too many requests",
		"503", // Service unavailable
		"502", // Bad gateway
		"504", // Gateway timeout
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// CategorizeError categorizes errors into different types.
func (eh *ErrorHandler) CategorizeError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") {
		return "network"
	}

	if strings.Contains(errStr, "auth") || strings.Contains(errStr, "unauthorized") {
		return "authentication"
	}

	if strings.Contains(errStr, "permission") || strings.Contains(errStr, "forbidden") {
		return "permission"
	}

	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "404") {
		return "not_found"
	}

	if strings.Contains(errStr, "invalid") && strings.Contains(errStr, "name") {
		return "invalid_name"
	}

	if strings.Contains(errStr, "rate limit") {
		return "rate_limit"
	}

	if strings.Contains(errStr, "timeout") {
		return "timeout"
	}

	return "general"
}

// CollectErrors collects multiple errors into a single error.
func (eh *ErrorHandler) CollectErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	var errorMessages []string

	for i, err := range errors {
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("%d: %s", i+1, err.Error()))
		}
	}

	if len(errorMessages) == 0 {
		return nil
	}

	return fmt.Errorf("multiple errors occurred:\n%s", strings.Join(errorMessages, "\n"))
}

// LogError logs an error with appropriate level and context.
func (eh *ErrorHandler) LogError(ctx context.Context, err error, operation string, additionalContext map[string]interface{}) {
	if err == nil {
		return
	}

	logContext := map[string]interface{}{
		"operation":    operation,
		"error":        err.Error(),
		"error_type":   eh.CategorizeError(err),
		"is_retryable": eh.IsRetryableError(err),
	}

	// Merge additional context
	for k, v := range additionalContext {
		logContext[k] = v
	}

	if eh.IsRetryableError(err) {
		eh.logger.Warn(ctx, "Retryable error occurred", logContext)
	} else {
		eh.logger.Error(ctx, "Non-retryable error occurred", logContext)
	}
}

// CreateErrorWithCategory creates an error with a specific category for tracking.
func (eh *ErrorHandler) CreateErrorWithCategory(category, message string, cause error) error {
	if cause != nil {
		return fmt.Errorf("[%s] %s: %w", category, message, cause)
	}

	return fmt.Errorf("[%s] %s", category, message)
}

// Common error types used throughout the application

var (
	// ErrInvalidRepositoryName indicates an invalid repository name
	ErrInvalidRepositoryName = errors.New("invalid repository name")

	// ErrRepositoryNotFound indicates a repository was not found
	ErrRepositoryNotFound = errors.New("repository not found")

	// ErrAuthenticationFailed indicates authentication failure
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrNetworkFailure indicates network connectivity issues
	ErrNetworkFailure = errors.New("network failure")

	// ErrPermissionDenied indicates insufficient permissions
	ErrPermissionDenied = errors.New("permission denied")

	// ErrRateLimitExceeded indicates API rate limit exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrOperationTimeout indicates operation timeout
	ErrOperationTimeout = errors.New("operation timeout")
)
