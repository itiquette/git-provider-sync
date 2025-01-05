// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package stringconvert

import (
	"context"
	"itiquette/git-provider-sync/internal/log"
	"net/url"
)

func AddBasicAuthToURL(ctx context.Context, urlStr, username, password string) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("AddBasicAuthToURL")

	parsedURL, _ := url.Parse(urlStr)
	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String()
}

func RemoveBasicAuthFromURL(ctx context.Context, urlStr string, stripInsteadMask bool) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("RemoveBasicAuthFromURL")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	if stripInsteadMask {
		if parsedURL.User != nil {
			parsedURL.User = nil
		}
	} else {
		if parsedURL.User != nil {
			username := parsedURL.User.Username()
			if _, hasPassword := parsedURL.User.Password(); hasPassword {
				parsedURL.User = url.UserPassword(username, "SECRET")
			}
		}
	}

	return parsedURL.String()
}
