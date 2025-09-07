// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package dto

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProjectInfoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		projectInfo  ProjectInfo
		expectedName string
	}{
		{
			name: "returns original name when ASCIIName is false",
			projectInfo: ProjectInfo{
				OriginalName: "my-original-repo",
				CleanName:    "my-clean-repo",
				ASCIIName:    false,
			},
			expectedName: "my-original-repo",
		},
		{
			name: "returns clean name when ASCIIName is true",
			projectInfo: ProjectInfo{
				OriginalName: "my-original-repo",
				CleanName:    "my-clean-repo",
				ASCIIName:    true,
			},
			expectedName: "my-clean-repo",
		},
		{
			name: "handles empty names",
			projectInfo: ProjectInfo{
				OriginalName: "",
				CleanName:    "",
				ASCIIName:    true,
			},
			expectedName: "",
		},
		{
			name: "handles special characters",
			projectInfo: ProjectInfo{
				OriginalName: "repo@with#special$chars",
				CleanName:    "repo-with-special-chars",
				ASCIIName:    true,
			},
			expectedName: "repo-with-special-chars",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := testCase.projectInfo.Name(ctx)
			assert.Equal(t, testCase.expectedName, result)
		})
	}
}

func TestProjectInfoTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		projectInfo        ProjectInfo
		expectedIsZeroTime bool
	}{
		{
			name: "returns zero time when LastActivityAt is nil",
			projectInfo: ProjectInfo{
				LastActivityAt: nil,
			},
			expectedIsZeroTime: true,
		},
		{
			name: "returns actual time when LastActivityAt is set",
			projectInfo: ProjectInfo{
				LastActivityAt: func() *time.Time {
					t := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)

					return &t
				}(),
			},
			expectedIsZeroTime: false,
		},
		{
			name: "returns zero time for empty time pointer",
			projectInfo: ProjectInfo{
				LastActivityAt: func() *time.Time {
					var t time.Time

					return &t
				}(),
			},
			expectedIsZeroTime: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.projectInfo.Time()

			if testCase.expectedIsZeroTime {
				assert.True(t, result.IsZero())
			} else {
				assert.False(t, result.IsZero())

				if testCase.projectInfo.LastActivityAt != nil {
					assert.Equal(t, *testCase.projectInfo.LastActivityAt, result)
				}
			}
		})
	}
}
