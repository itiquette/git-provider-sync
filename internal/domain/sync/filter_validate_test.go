// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/sync"
)

// Define sentinel errors for testing.
var (
	errInvalidActivityLimit = errors.New("invalid activity limit duration")
)

func TestValidateFilterRequest_InputFormats(t *testing.T) {
	t.Parallel()

	useCase := sync.NewFilterRepositoriesUseCase(nil)
	require.NotNil(t, useCase)

	tests := []struct {
		name           string
		request        sync.FilterRequest
		wantErr        error  // Use error type instead of bool
		errMsgContains string // For partial error message matching
	}{
		{
			name: "valid_request_with_repositories",
			request: sync.FilterRequest{
				Repositories: []entities.Repository{
					{}, // Empty repository is fine for validation
				},
			},
			wantErr: nil,
		},
		{
			name:    "empty_request_is_valid",
			request: sync.FilterRequest{},
			wantErr: nil,
		},
		{
			name: "valid_activity_limit_duration_hours",
			request: sync.FilterRequest{
				ActiveFromLimit: "24h",
			},
			wantErr: nil,
		},
		{
			name: "invalid_activity_limit_duration_format",
			request: sync.FilterRequest{
				ActiveFromLimit: "invalid",
			},
			wantErr:        errInvalidActivityLimit, // Using sentinel error for clarity
			errMsgContains: "invalid activity limit duration",
		},
		{
			name: "valid_duration_format_30_days",
			request: sync.FilterRequest{
				ActiveFromLimit: "720h", // 30 days in hours
			},
			wantErr: nil,
		},
		{
			name: "valid_duration_format_minutes",
			request: sync.FilterRequest{
				ActiveFromLimit: "30m",
			},
			wantErr: nil,
		},
		{
			name: "valid_duration_format_seconds",
			request: sync.FilterRequest{
				ActiveFromLimit: "3600s",
			},
			wantErr: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.ValidateFilterRequest(context.Background(), testCase.request)

			if testCase.wantErr != nil {
				require.Error(t, err, "Expected error but got nil")

				// Check for specific error message content
				if testCase.errMsgContains != "" {
					assert.Contains(t, err.Error(), testCase.errMsgContains,
						"Error message should contain expected text")
				}
			} else {
				require.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}
