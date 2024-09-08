// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/model"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

type Client struct{}

func (Client) Create(_ context.Context, _ configuration.ProviderConfig, _ model.CreateOption) error {
	return nil
}

func (Client) Name() string {
	return configuration.ARCHIVE
}

func (Client) IsValidRepositoryName(_ context.Context, _ string) bool {
	return true
}

func (Client) Metainfos(_ context.Context, _ configuration.ProviderConfig, _ bool) ([]model.RepositoryMetainfo, error) {
	return nil, nil
}

func (Client) Client() gitprovider.Client {
	return nil
}
