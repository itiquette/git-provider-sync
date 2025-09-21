// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define sentinel errors for testing.
var errTokenEmpty = errors.New("token file is empty")

func TestValidateToken_InvalidFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		token       string
		wantErr     error
		errContains string
	}{
		{
			name:    "valid_token",
			token:   "ghp_xxxxxxxxxxxxxxxxxxxx",
			wantErr: nil,
		},
		{
			name:        "empty_token",
			token:       "",
			wantErr:     errTokenEmpty,
			errContains: "empty",
		},
		{
			name:        "unexpanded_variable",
			token:       "${GITHUB_TOKEN}",
			wantErr:     errors.New("unexpanded variable"),
			errContains: "unexpanded variable",
		},
		{
			name:        "token_with_spaces",
			token:       "token with spaces",
			wantErr:     errors.New("invalid token format"),
			errContains: "spaces",
		},
	}

	for _, testCase := range tests {
		// Capture range variable
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateToken(testCase.token)

			if testCase.wantErr != nil {
				require.Error(t, err)

				if testCase.errContains != "" {
					assert.Contains(t, err.Error(), testCase.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
