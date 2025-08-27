// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfirmOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation string
		expected  bool
	}{
		{
			name:      "non_interactive_returns_false",
			operation: "test operation",
			expected:  false, // In tests, IsInput() returns false
		},
	}

	for _, testCase := range tests {
		// Capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// In test environment, IsInput() and IsError() return false
			// so ConfirmOperation should always return false
			result := ConfirmOperation(testCase.operation)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
