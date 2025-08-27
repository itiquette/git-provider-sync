// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package ports

import "context"

// InterruptHandler provides user feedback during interruptible operations.
// This port follows hexagonal architecture - domain defines the interface,
// adapters implement the behavior. Keep it simple, functional, idiomatic.
type InterruptHandler interface {
	// ShowInterruptible displays a message indicating the operation can be interrupted
	ShowInterruptible(ctx context.Context, operation string)

	// ShowProgress displays progress for long-running operations
	ShowProgress(ctx context.Context, current, total int, description string)

	// ShowCancelled displays a message when an operation is cancelled
	ShowCancelled(ctx context.Context, operation string)
}
