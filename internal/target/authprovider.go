// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"itiquette/git-provider-sync/internal/log"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

type authProvider interface {
	getAuthMethod(ctx context.Context, gitOpt gpsconfig.GitOption, httpClient gpsconfig.HTTPClientOption, sshClient gpsconfig.SSHClientOption) (transport.AuthMethod, error)
}

type authProviderImpl struct {
}

func newAuthProvider() authProvider {
	return &authProviderImpl{}
}

func (a *authProviderImpl) getAuthMethod(ctx context.Context, gitOpt gpsconfig.GitOption, httpClient gpsconfig.HTTPClientOption, _ gpsconfig.SSHClientOption) (transport.AuthMethod, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("getAuthMethod")

	switch strings.ToLower(gitOpt.Type) {
	case gpsconfig.SSHAGENT:
		return ssh.NewSSHAgentAuth("git") //nolint
	case gpsconfig.HTTPS, "":
		return &http.BasicAuth{Username: "anyUser", Password: httpClient.Token}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedGitType, gitOpt.Type)
	}
}

func addBasicAuthToURL(ctx context.Context, urlStr, username, password string) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("addBasicAuthToURL")

	parsedURL, _ := url.Parse(urlStr)
	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String()
}

func removeBasicAuthFromURL(ctx context.Context, urlStr string) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("removeBasicAuthFromURL")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	parsedURL.User = nil

	return parsedURL.String()
}
