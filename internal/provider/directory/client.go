// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

type Client struct{}

func (Client) Create(_ context.Context, _ config.ProviderConfig, _ model.CreateOption) error {
	return nil
}

func (Client) Name() string {
	return config.DIRECTORY
}

func (Client) DefaultBranch(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

func (Client) IsValidRepositoryName(_ context.Context, _ string) bool {
	return true
}

func (Client) ProjectInfos(_ context.Context, _ config.ProviderConfig, _ bool) ([]model.ProjectInfo, error) {
	return nil, nil
}
