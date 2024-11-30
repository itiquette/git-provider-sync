// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlib

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"itiquette/git-provider-sync/internal/log"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

type authService struct {
}

func NewAuthService() *authService { //nolint
	return &authService{}
}

func (p *authService) GetAuthMethod(ctx context.Context, gitOpt gpsconfig.GitOption, httpOpt gpsconfig.HTTPClientOption, _ gpsconfig.SSHClientOption) (transport.AuthMethod, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("getAuthMethod")

	switch strings.ToLower(gitOpt.Type) {
	case gpsconfig.SSHAGENT:
		return ssh.NewSSHAgentAuth("git") //nolint
	case gpsconfig.HTTPS, "":
		return &http.BasicAuth{Username: "anyUser", Password: httpOpt.Token}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidAuth, gitOpt.Type)
	}
}
