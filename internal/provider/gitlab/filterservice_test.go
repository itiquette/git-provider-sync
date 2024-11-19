// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlab

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

func TestFilterService_FilterProjectinfos(t *testing.T) {
	timePtr := func(t time.Time) *time.Time {
		return &t
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	oldDate := now.Add(-48 * time.Hour)

	tests := []struct {
		name         string
		projectInfos []model.ProjectInfo
		cfg          config.ProviderConfig
		startTime    time.Time
		expected     []model.ProjectInfo
		expectedErr  string
	}{
		{
			name: "filter projects within time range",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: timePtr(now)},
				{ProjectID: "2", LastActivityAt: timePtr(oldDate)},
			},
			startTime: yesterday,
			expected: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: timePtr(now)},
			},
		},
		{
			name: "skip nil dates",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: nil},
				{ProjectID: "2", LastActivityAt: timePtr(now)},
			},
			startTime: yesterday,
			expected: []model.ProjectInfo{
				{ProjectID: "2", LastActivityAt: timePtr(now)},
			},
		},
		{
			name:         "empty project list",
			projectInfos: []model.ProjectInfo{},
			startTime:    yesterday,
			expected:     []model.ProjectInfo{},
		},
		{
			name: "all projects filtered out",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: timePtr(oldDate)},
				{ProjectID: "2", LastActivityAt: timePtr(oldDate)},
			},
			startTime: yesterday,
			expected:  []model.ProjectInfo{},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			assert := require.New(t)

			filterIncludeExclude := func(_ context.Context, _ config.ProviderConfig, projects []model.ProjectInfo) ([]model.ProjectInfo, error) {
				return projects, nil
			}

			isInInterval := func(_ context.Context, t time.Time) (bool, error) {
				return t.After(tabletest.startTime), nil
			}

			service := filterService{
				isInInterval: isInInterval,
			}

			result, err := service.FilterProjectinfos(
				context.Background(),
				tabletest.cfg,
				tabletest.projectInfos,
				filterIncludeExclude,
				isInInterval,
			)

			if tabletest.expectedErr != "" {
				assert.EqualError(err, tabletest.expectedErr)
				assert.Nil(result)
			} else {
				assert.NoError(err)
				assert.Equal(tabletest.expected, result)
			}
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
		startTime    time.Time
		expected     []model.ProjectInfo
		expectedErr  string
	}{
		{
			name: "filter by date success",
			projectInfos: []model.ProjectInfo{
				{ProjectID: "1", LastActivityAt: &now},
				{ProjectID: "2", LastActivityAt: &oldDate},
			},
			startTime: yesterday,
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
			startTime: yesterday,
			expected: []model.ProjectInfo{
				{ProjectID: "2", LastActivityAt: &now},
			},
		},
		{
			name:         "empty list",
			projectInfos: []model.ProjectInfo{},
			startTime:    yesterday,
			expected:     []model.ProjectInfo{},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			assert := require.New(t)

			isInInterval := func(_ context.Context, t time.Time) (bool, error) {
				return t.After(tabletest.startTime), nil
			}

			result, err := filterByDate(context.Background(), tabletest.projectInfos, isInInterval)

			if tabletest.expectedErr != "" {
				assert.EqualError(err, tabletest.expectedErr)
				assert.Nil(result)
			} else {
				assert.NoError(err)
				assert.Equal(tabletest.expected, result)
			}
		})
	}
}
