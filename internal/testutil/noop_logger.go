// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"context"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// NoOpLogger is a logger that does nothing, useful for tests.
type NoOpLogger struct{}

// Ensure NoOpLogger implements ports.Logger.
var _ ports.Logger = (*NoOpLogger)(nil)

// NewNoOpLogger creates a new no-op logger.
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// Trace does nothing.
func (l *NoOpLogger) Trace(_ context.Context, _ string, _ map[string]any) {}

// Debug does nothing.
func (l *NoOpLogger) Debug(_ context.Context, _ string, _ map[string]any) {}

// Info does nothing.
func (l *NoOpLogger) Info(_ context.Context, _ string, _ map[string]any) {}

// Warn does nothing.
func (l *NoOpLogger) Warn(_ context.Context, _ string, _ map[string]any) {}

// Error does nothing.
func (l *NoOpLogger) Error(_ context.Context, _ string, _ map[string]any) {}

// Fatal does nothing.
func (l *NoOpLogger) Fatal(_ context.Context, _ string, _ map[string]any) {}

// IsLevelEnabled always returns false.
func (l *NoOpLogger) IsLevelEnabled(_ ports.LogLevel) bool { return false }

// GetLevel returns info level.
func (l *NoOpLogger) GetLevel() ports.LogLevel { return ports.LogLevelInfo }
