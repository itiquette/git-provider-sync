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

type AuthService interface {
	GetAuthMethod(ctx context.Context, authCfg gpsconfig.AuthConfig) (transport.AuthMethod, error)
}

type authService struct {
}

func NewAuthService() *authService { //nolint
	return &authService{}
}

func (p *authService) GetAuthMethod(ctx context.Context, authCfg gpsconfig.AuthConfig) (transport.AuthMethod, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("getAuthMethod")

	switch strings.ToLower(authCfg.Protocol) {
	case gpsconfig.SSH:
		return ssh.NewSSHAgentAuth("git") //nolint
	case gpsconfig.TLS, "":
		return &http.BasicAuth{Username: "anyUser", Password: authCfg.Token}, nil
	default:
		return nil, fmt.Errorf("%w", ErrInvalidAuth)
	}
}
