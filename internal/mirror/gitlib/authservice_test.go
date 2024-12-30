// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlib

import (
	"context"
	"fmt"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/stretchr/testify/require"
)

func TestGetAuthMethod(t *testing.T) {
	tests := []struct {
		name         string
		authConfig   gpsconfig.AuthConfig
		wantAuthType string
		wantErr      bool
		errMsg       string
	}{
		{
			name: "Valid TLS auth with token",
			authConfig: gpsconfig.AuthConfig{
				Protocol: "tls",
				Token:    "test-token",
			},
			wantAuthType: "*http.BasicAuth",
		},
		{
			name: "Empty protocol defaults to TLS",
			authConfig: gpsconfig.AuthConfig{
				Protocol: "",
				Token:    "test-token",
			},
			wantAuthType: "*http.BasicAuth",
		},
		// {
		// 	name: "Valid SSH agent auth",
		// 	authConfig: gpsconfig.AuthConfig{
		// 		Method: "sshagent",
		// 	},
		// 	wantAuthType: "*ssh.PublicKeysCallback",
		// },
		{
			name: "Invalid auth method",
			authConfig: gpsconfig.AuthConfig{
				Protocol: "invalid-type",
			},
			wantErr: true,
			errMsg:  "invalid authentication configuration",
		},
		{
			name: "Case insensitive TLS",
			authConfig: gpsconfig.AuthConfig{
				Protocol: "TLs",
				Token:    "test-token",
			},
			wantAuthType: "*http.BasicAuth",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			svc := NewAuthService()
			auth, err := svc.GetAuthMethod(context.Background(), tabletest.authConfig)

			if tabletest.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tabletest.errMsg)
				require.Nil(t, auth)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, auth)
			require.Equal(t, tabletest.wantAuthType, fmt.Sprintf("%T", auth))

			switch auth := auth.(type) {
			case *http.BasicAuth:
				require.Equal(t, "anyUser", auth.Username)
				require.Equal(t, tabletest.authConfig.Token, auth.Password)
			case *ssh.PublicKeysCallback:
				require.Equal(t, "git", auth.User)
			}
		})
	}
}

func TestNewAuthService(t *testing.T) {
	svc := NewAuthService()
	require.NotNil(t, svc)
	require.IsType(t, &authService{}, svc)
}
