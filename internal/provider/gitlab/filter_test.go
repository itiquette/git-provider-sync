// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"errors"
	"testing"
	"time"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockFilterIncludedExcluded is a mock for the FilterIncludedExcludedFunc.
type MockFilterIncludedExcluded struct {
	mock.Mock
}

func (m *MockFilterIncludedExcluded) FilterIncludedExcluded(ctx context.Context, config config.ProviderConfig, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	args := m.Called(ctx, config, projectinfos)

	//nolint:forcetypeassert
	return args.Get(0).([]model.ProjectInfo), args.Error(1) //nolint:wrapcheck
}

func TestFilter_FilterProjectinfos(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	now := time.Now()
	oldTime := now.Add(-24 * time.Hour)

	tests := []struct {
		name              string
		config            config.ProviderConfig
		inputProjectinfos []model.ProjectInfo
		mockFilterResult  []model.ProjectInfo
		mockFilterErr     error
		isInInterval      IsInIntervalFunc
		expectedResult    []model.ProjectInfo
		expectedErr       bool
	}{
		{
			name:   "Success - All repositories included",
			config: config.ProviderConfig{},
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			mockFilterResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			mockFilterErr: nil,
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			expectedErr: false,
		},
		{
			name:   "Success - Some repositories filtered out",
			config: config.ProviderConfig{},
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &oldTime},
			},
			mockFilterResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &oldTime},
			},
			mockFilterErr: nil,
			isInInterval: func(_ context.Context, t time.Time) (bool, error) {
				return t.After(oldTime), nil
			},
			expectedResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			expectedErr: false,
		},
		{
			name:              "Error - Filter function fails",
			config:            config.ProviderConfig{},
			inputProjectinfos: []model.ProjectInfo{{HTTPSURL: "https://example.com/repo1"}},
			mockFilterResult:  nil,
			mockFilterErr:     errors.New("mock error"),
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: nil,
			expectedErr:    true,
		},
		{
			name:              "Edge Case - Empty input",
			config:            config.ProviderConfig{},
			inputProjectinfos: []model.ProjectInfo{},
			mockFilterResult:  []model.ProjectInfo{},
			mockFilterErr:     nil,
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.ProjectInfo{},
			expectedErr:    false,
		},
		{
			name:   "Edge Case - Nil LastActivityAt",
			config: config.ProviderConfig{},
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: nil},
			},
			mockFilterResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: nil},
			},
			mockFilterErr: nil,
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.ProjectInfo{},
			expectedErr:    false,
		},
		{
			name:   "Edge Case - IsInInterval returns error",
			config: config.ProviderConfig{},
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			mockFilterResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			mockFilterErr: nil,
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return false, errors.New("mock error")
			},
			expectedResult: nil,
			expectedErr:    true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(_ *testing.T) {
			mockFilter := new(MockFilterIncludedExcluded)
			mockFilter.On("FilterIncludedExcluded", mock.Anything, tabletest.config, tabletest.inputProjectinfos).Return(tabletest.mockFilterResult, tabletest.mockFilterErr)

			f := NewFilter(tabletest.isInInterval)
			result, err := f.FilterProjectinfos(ctx, tabletest.config, tabletest.inputProjectinfos, mockFilter.FilterIncludedExcluded, tabletest.isInInterval)

			// mockFilter.AssertExpectations(t)
			if tabletest.expectedErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Equal(tabletest.expectedResult, result)
			}
		})
	}
}

func TestFilterByDate(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	now := time.Now()
	oldTime := now.Add(-24 * time.Hour)

	tests := []struct {
		name              string
		inputProjectinfos []model.ProjectInfo
		isInInterval      IsInIntervalFunc
		expectedResult    []model.ProjectInfo
		expectedErr       bool
	}{
		{
			name: "Success - All repositories within date range",
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			expectedErr: false,
		},
		{
			name: "Success - Some repositories filtered out",
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &oldTime},
			},
			isInInterval: func(_ context.Context, t time.Time) (bool, error) {
				return t.After(oldTime), nil
			},
			expectedResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			expectedErr: false,
		},
		{
			name:              "Edge Case - Empty input",
			inputProjectinfos: []model.ProjectInfo{},
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.ProjectInfo{},
			expectedErr:    false,
		},
		{
			name: "Edge Case - Nil LastActivityAt",
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: nil},
			},
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.ProjectInfo{},
			expectedErr:    false,
		},
		{
			name: "Edge Case - IsInInterval returns error",
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return false, errors.New("mock error")
			},
			expectedResult: nil,
			expectedErr:    true,
		},
		{
			name: "Edge Case - Mixed nil and non-nil LastActivityAt",
			inputProjectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: nil},
				{HTTPSURL: "https://example.com/repo3", LastActivityAt: &oldTime},
			},
			isInInterval: func(_ context.Context, t time.Time) (bool, error) {
				return t.After(oldTime), nil
			},
			expectedResult: []model.ProjectInfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			expectedErr: false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result, err := filterByDate(ctx, tabletest.inputProjectinfos, tabletest.isInInterval)

			if tabletest.expectedErr {
				assert.Error(t, err)
			} else {
				require.NoError(err)
				require.Equal(tabletest.expectedResult, result)
			}
		})
	}
}
