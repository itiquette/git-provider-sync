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

func (m *MockFilterIncludedExcluded) FilterIncludedExcluded(ctx context.Context, config config.ProviderConfig, metainfos []model.RepositoryMetainfo) ([]model.RepositoryMetainfo, error) {
	args := m.Called(ctx, config, metainfos)

	//nolint:forcetypeassert
	return args.Get(0).([]model.RepositoryMetainfo), args.Error(1) //nolint:wrapcheck
}

func TestFilter_FilterMetainfo(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	now := time.Now()
	oldTime := now.Add(-24 * time.Hour)

	tests := []struct {
		name             string
		config           config.ProviderConfig
		inputMetainfos   []model.RepositoryMetainfo
		mockFilterResult []model.RepositoryMetainfo
		mockFilterErr    error
		isInInterval     IsInIntervalFunc
		expectedResult   []model.RepositoryMetainfo
		expectedErr      bool
	}{
		{
			name:   "Success - All repositories included",
			config: config.ProviderConfig{},
			inputMetainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			mockFilterResult: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			mockFilterErr: nil,
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			expectedErr: false,
		},
		{
			name:   "Success - Some repositories filtered out",
			config: config.ProviderConfig{},
			inputMetainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &oldTime},
			},
			mockFilterResult: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &oldTime},
			},
			mockFilterErr: nil,
			isInInterval: func(_ context.Context, t time.Time) (bool, error) {
				return t.After(oldTime), nil
			},
			expectedResult: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			expectedErr: false,
		},
		{
			name:             "Error - Filter function fails",
			config:           config.ProviderConfig{},
			inputMetainfos:   []model.RepositoryMetainfo{{HTTPSURL: "https://example.com/repo1"}},
			mockFilterResult: nil,
			mockFilterErr:    errors.New("mock error"),
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: nil,
			expectedErr:    true,
		},
		{
			name:             "Edge Case - Empty input",
			config:           config.ProviderConfig{},
			inputMetainfos:   []model.RepositoryMetainfo{},
			mockFilterResult: []model.RepositoryMetainfo{},
			mockFilterErr:    nil,
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.RepositoryMetainfo{},
			expectedErr:    false,
		},
		{
			name:   "Edge Case - Nil LastActivityAt",
			config: config.ProviderConfig{},
			inputMetainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: nil},
			},
			mockFilterResult: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: nil},
			},
			mockFilterErr: nil,
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.RepositoryMetainfo{},
			expectedErr:    false,
		},
		{
			name:   "Edge Case - IsInInterval returns error",
			config: config.ProviderConfig{},
			inputMetainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			mockFilterResult: []model.RepositoryMetainfo{
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
			mockFilter.On("FilterIncludedExcluded", mock.Anything, tabletest.config, tabletest.inputMetainfos).Return(tabletest.mockFilterResult, tabletest.mockFilterErr)

			f := NewFilter(tabletest.isInInterval)
			result, err := f.FilterMetainfo(ctx, tabletest.config, tabletest.inputMetainfos, mockFilter.FilterIncludedExcluded, tabletest.isInInterval)

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
		name           string
		inputMetainfos []model.RepositoryMetainfo
		isInInterval   IsInIntervalFunc
		expectedResult []model.RepositoryMetainfo
		expectedErr    bool
	}{
		{
			name: "Success - All repositories within date range",
			inputMetainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &now},
			},
			expectedErr: false,
		},
		{
			name: "Success - Some repositories filtered out",
			inputMetainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: &oldTime},
			},
			isInInterval: func(_ context.Context, t time.Time) (bool, error) {
				return t.After(oldTime), nil
			},
			expectedResult: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			expectedErr: false,
		},
		{
			name:           "Edge Case - Empty input",
			inputMetainfos: []model.RepositoryMetainfo{},
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.RepositoryMetainfo{},
			expectedErr:    false,
		},
		{
			name: "Edge Case - Nil LastActivityAt",
			inputMetainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: nil},
			},
			isInInterval: func(_ context.Context, _ time.Time) (bool, error) {
				return true, nil
			},
			expectedResult: []model.RepositoryMetainfo{},
			expectedErr:    false,
		},
		{
			name: "Edge Case - IsInInterval returns error",
			inputMetainfos: []model.RepositoryMetainfo{
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
			inputMetainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
				{HTTPSURL: "https://example.com/repo2", LastActivityAt: nil},
				{HTTPSURL: "https://example.com/repo3", LastActivityAt: &oldTime},
			},
			isInInterval: func(_ context.Context, t time.Time) (bool, error) {
				return t.After(oldTime), nil
			},
			expectedResult: []model.RepositoryMetainfo{
				{HTTPSURL: "https://example.com/repo1", LastActivityAt: &now},
			},
			expectedErr: false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result, err := filterByDate(ctx, tabletest.inputMetainfos, tabletest.isInInterval)

			if tabletest.expectedErr {
				assert.Error(t, err)
			} else {
				require.NoError(err)
				require.Equal(tabletest.expectedResult, result)
			}
		})
	}
}
