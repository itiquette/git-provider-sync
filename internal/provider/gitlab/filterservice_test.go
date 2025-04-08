// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2
package gitlab

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFilterService_FilterProjectinfos(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	oldDate := now.Add(-48 * time.Hour)

	tests := []struct {
		name         string
		projectInfos []model.ProjectInfo
		opt          model.ProviderOption
		expected     []model.ProjectInfo
		expectedErr  string
	}{
		{
			name: "filter projects within time range",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: &now},
				{ProjectID: "2", LastActivityAt: &oldDate},
			},
			expected: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: &now},
			},
		},
		{
			name: "skip nil dates",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: nil},
				{ProjectID: "2", LastActivityAt: &now},
			},
			expected: []model.ProjectInfo{
				{ProjectID: "2", LastActivityAt: &now},
			},
		},
		{
			name:         "empty project list",
			projectInfos: []model.ProjectInfo{},
			expected:     []model.ProjectInfo{},
		},
		{
			name: "all projects filtered out",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: &oldDate},
				{ProjectID: "2", LastActivityAt: &oldDate},
			},
			expected: []model.ProjectInfo{},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			filterIncludeExclude := func(_ context.Context, _ model.ProviderOption, projects []model.ProjectInfo) ([]model.ProjectInfo, error) {
				return projects, nil
			}

			isInInterval := func(_ context.Context, t time.Time) (bool, error) {
				return t.After(yesterday), nil
			}

			service := filterService{isInInterval: isInInterval}
			result, err := service.FilterProjectinfos(
				context.Background(),
				tabletest.opt,
				tabletest.projectInfos,
				filterIncludeExclude,
				isInInterval,
			)

			if tabletest.expectedErr != "" {
				require.EqualError(t, err, tabletest.expectedErr)
				require.Nil(t, result)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestFilterByDate(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	oldDate := now.Add(-48 * time.Hour)

	tests := []struct {
		name         string
		projectInfos []model.ProjectInfo
		expected     []model.ProjectInfo
		expectedErr  string
	}{
		{
			name: "filter by date success",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: &now},
				{ProjectID: "2", LastActivityAt: &oldDate},
			},
			expected: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: &now},
			},
		},
		{
			name: "handle nil dates",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: nil},
				{ProjectID: "2", LastActivityAt: &now},
			},
			expected: []model.ProjectInfo{
				{ProjectID: "2", LastActivityAt: &now},
			},
		},
		{
			name:         "empty list",
			projectInfos: []model.ProjectInfo{},
			expected:     []model.ProjectInfo{},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			isInInterval := func(_ context.Context, t time.Time) (bool, error) {
				return t.After(yesterday), nil
			}

			result, err := filterByDate(context.Background(), tabletest.projectInfos, isInInterval)

			if tabletest.expectedErr != "" {
				require.EqualError(t, err, tabletest.expectedErr)
				require.Nil(t, result)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tabletest.expected, result)
		})
	}
}
