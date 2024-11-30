// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"itiquette/git-provider-sync/internal/log"
	"net/url"
)

type authService struct{}

func NewAuthService() *authService { //nolint
	return &authService{}
}

func (s *authService) AddBasicAuthToURL(ctx context.Context, urlStr, username, password string) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("AddBasicAuthToURL")

	parsedURL, _ := url.Parse(urlStr)
	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String()
}

func (s *authService) RemoveBasicAuthFromURL(ctx context.Context, urlStr string) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("RemoveBasicAuthFromURL")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	parsedURL.User = nil

	return parsedURL.String()
}
