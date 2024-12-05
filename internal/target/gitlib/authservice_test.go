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
		gitOpt       gpsconfig.GitOption
		httpOpt      gpsconfig.HTTPClientOption
		sshOpt       gpsconfig.SSHClientOption
		wantAuthType string
		wantErr      bool
		errMsg       string
	}{
		{
			name: "Valid HTTPS auth with token",
			gitOpt: gpsconfig.GitOption{
				Type: "https",
			},
			httpOpt: gpsconfig.HTTPClientOption{
				Token: "test-token",
			},
			wantAuthType: "*http.BasicAuth",
			wantErr:      false,
		},
		{
			name: "Empty type defaults to HTTPS",
			gitOpt: gpsconfig.GitOption{
				Type: "",
			},
			httpOpt: gpsconfig.HTTPClientOption{
				Token: "test-token",
			},
			wantAuthType: "*http.BasicAuth",
			wantErr:      false,
		},
		// {
		// 	name: "Valid SSH agent auth",
		// 	gitOpt: gpsconfig.GitOption{
		// 		Type: "sshagent",
		// 	},
		// 	wantAuthType: "*ssh.PublicKeysCallback",
		// 	wantErr:      false,
		// },
		{
			name: "Invalid auth type",
			gitOpt: gpsconfig.GitOption{
				Type: "invalid-type",
			},
			wantErr: true,
			errMsg:  "invalid authentication configuration: invalid-type",
		},
		{
			name: "Case insensitive HTTPS",
			gitOpt: gpsconfig.GitOption{
				Type: "HTTPS",
			},
			httpOpt: gpsconfig.HTTPClientOption{
				Token: "test-token",
			},
			wantAuthType: "*http.BasicAuth",
			wantErr:      false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			svc := NewAuthService()
			auth, err := svc.GetAuthMethod(context.Background(), tabletest.gitOpt, tabletest.httpOpt, tabletest.sshOpt)

			if tabletest.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tabletest.errMsg)
				require.Nil(t, auth)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, auth)
			require.Equal(t, tabletest.wantAuthType, fmt.Sprintf("%T", auth))

			// Additional type-specific assertions
			switch auth := auth.(type) {
			case *http.BasicAuth:
				require.Equal(t, "anyUser", auth.Username)
				require.Equal(t, tabletest.httpOpt.Token, auth.Password)
			case *ssh.PublicKeysCallback:
				require.Equal(t, "git", auth.User)
			}
		})
	}
}

// TestNewAuthService verifies the constructor.
func TestNewAuthService(t *testing.T) {
	svc := NewAuthService()
	require.NotNil(t, svc)
	require.IsType(t, &authService{}, svc)
}
