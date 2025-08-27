// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// InterruptHandler provides user feedback for interruptible operations.
// Simple, functional, idiomatic - no overengineering.
type InterruptHandler struct {
	writer       io.Writer
	mu           sync.Mutex
	lastProgress time.Time
}

// NewInterruptHandler creates a new interrupt handler.
func NewInterruptHandler(writer io.Writer) ports.InterruptHandler {
	return &InterruptHandler{
		writer: writer,
	}
}

// ShowInterruptible indicates an operation can be interrupted with Ctrl-C.
func (h *InterruptHandler) ShowInterruptible(ctx context.Context, operation string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	select {
	case <-ctx.Done():
		// Already cancelled, don't show the message
		return
	default:
		_, _ = fmt.Fprintf(h.writer, "→ %s (Press Ctrl-C to cancel)\n", operation)
	}
}

// ShowProgress displays progress for long operations.
func (h *InterruptHandler) ShowProgress(ctx context.Context, current, total int, description string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Rate limit progress updates to avoid spam
	now := time.Now()
	if now.Sub(h.lastProgress) < 100*time.Millisecond {
		return
	}

	h.lastProgress = now

	select {
	case <-ctx.Done():
		// Operation cancelled, don't update progress
		return
	default:
		if total > 0 {
			percentage := (current * 100) / total
			_, _ = fmt.Fprintf(h.writer, "\r[%3d%%] %s", percentage, description)
		} else {
			_, _ = fmt.Fprintf(h.writer, "\r[...] %s", description)
		}
	}
}

// ShowCancelled indicates an operation was cancelled by the user.
func (h *InterruptHandler) ShowCancelled(_ context.Context, operation string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, _ = fmt.Fprintf(h.writer, "\n× %s cancelled\n", operation)
}
