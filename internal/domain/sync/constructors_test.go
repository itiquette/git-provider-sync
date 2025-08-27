// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFilterRepositoriesUseCase(t *testing.T) {
	t.Parallel()

	// Use nil for simplicity since we're only testing constructor doesn't panic
	useCase := NewFilterRepositoriesUseCase(nil)

	// We can't access private fields, but we can verify the constructor
	// doesn't panic and returns a valid instance
	assert.NotNil(t, useCase)
}

func TestNewToMirrorsUseCase(t *testing.T) {
	t.Parallel()

	// Use nil for simplicity since we're only testing constructor doesn't panic
	useCase := NewToMirrorsUseCase(nil, nil, nil, nil)

	// We can't access private fields, but we can verify the constructor
	// doesn't panic and returns a valid instance
	assert.NotNil(t, useCase)
}

func TestNewBranchProtectionUseCase(t *testing.T) {
	t.Parallel()

	// Use nil for simplicity since we're only testing constructor doesn't panic
	useCase := NewBranchProtectionUseCase(nil, nil)

	// We can't access private fields, but we can verify the constructor
	// doesn't panic and returns a valid instance
	assert.NotNil(t, useCase)
}

func TestNewValidateSyncUseCase(t *testing.T) {
	t.Parallel()

	// Use nil for simplicity since we're only testing constructor doesn't panic
	useCase := NewValidateSyncUseCase(nil, nil)

	// We can't access private fields, but we can verify the constructor
	// doesn't panic and returns a valid instance
	assert.NotNil(t, useCase)
}
