// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/sync"
)

func TestFilterRepositoriesUseCase_ValidateFilterRequest(t *testing.T) {
	t.Parallel()

	useCase := sync.NewFilterRepositoriesUseCase(nil)
	require.NotNil(t, useCase)

	tests := []struct {
		name    string
		request sync.FilterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with repositories",
			request: sync.FilterRequest{
				Repositories: []entities.Repository{
					{}, // Empty repository is fine for validation
				},
			},
			wantErr: false,
		},
		{
			name:    "empty request is valid",
			request: sync.FilterRequest{},
			wantErr: false,
		},
		{
			name: "valid activity limit duration",
			request: sync.FilterRequest{
				ActiveFromLimit: "24h",
			},
			wantErr: false,
		},
		{
			name: "invalid activity limit duration",
			request: sync.FilterRequest{
				ActiveFromLimit: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid activity limit duration",
		},
		{
			name: "valid duration formats",
			request: sync.FilterRequest{
				ActiveFromLimit: "720h", // 30 days in hours
			},
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := useCase.ValidateFilterRequest(context.Background(), testCase.request)
			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errMsg != "" {
					assert.Contains(t, err.Error(), testCase.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
