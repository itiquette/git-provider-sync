// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2
package gitlib

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	auth transport.AuthMethod
	err  error
}

func (m *mockAuthService) GetAuthMethod(context.Context, gpsconfig.AuthConfig) (transport.AuthMethod, error) {
	return m.auth, m.err
}

type mockMetadataHandler struct {
	updateCalled bool
	updateStatus string
	updatePath   string
}

func (m *mockMetadataHandler) UpdateSyncMetadata(_ context.Context, status, path string) {
	m.updateCalled = true
	m.updateStatus = status
	m.updatePath = path
}

func TestService_Clone(t *testing.T) {
	tests := []struct {
		name      string
		opt       model.CloneOption
		setupAuth *mockAuthService
		wantErr   bool
		errType   error
	}{
		// {
		// 	name: "successful non-bare clone",
		// 	opt: model.CloneOption{
		// 		URL:         "https://github.com/test/repo.git",
		// 		NonBareRepo: true,
		// 		AuthCfg: gpsconfig.AuthConfig{
		// 			Method: gpsconfig.HTTPS,
		// 			Token:  "test-token",
		// 		},
		// 	},
		// 	setupAuth: &mockAuthService{
		// 		auth: &http.BasicAuth{
		// 			Username: "anyUser",
		// 			Password: "test-token",
		// 		},
		// 	},
		// },
		{
			name: "auth error",
			opt: model.CloneOption{
				URL: "https://github.com/test/repo.git",
			},
			setupAuth: &mockAuthService{
				err: ErrInvalidAuth,
			},
			wantErr: true,
			errType: ErrAuthMethod,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			svc := &Service{
				authService: tabletest.setupAuth,
				Ops:         *NewOperation(),
				metadata:    &mockMetadataHandler{},
			}

			repo, err := svc.Clone(context.Background(), tabletest.opt)

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.errType)
				require.Empty(t, repo)

				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, repo)
		})
	}
}

// func TestService_Pull(t *testing.T) {
// 	tests := []struct {
// 		name      string
// 		opt       model.PullOption
// 		targetDir string
// 		setupAuth *mockAuthService
// 		wantErr   bool
// 		errType   error
// 	}{
// 		{
// 			name: "pull with auth error",
// 			opt: model.PullOption{
// 				AuthCfg: gpsconfig.AuthConfig{
// 					Method: gpsconfig.HTTPS,
// 				},
// 			},
// 			setupAuth: &mockAuthService{
// 				err: ErrInvalidAuth,
// 			},
// 			wantErr: true,
// 			errType: ErrAuthMethod,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			svc := &Service{
// 				authService: tt.setupAuth,
// 				Ops:         *NewOperation(),
// 				metadata:    &mockMetadataHandler{},
// 			}

// 			err := svc.Pull(context.Background(), tt.opt, tt.targetDir)

// 			if tt.wantErr {
// 				require.Error(t, err)
// 				require.ErrorIs(t, err, tt.errType)
// 				return
// 			}
// 			require.NoError(t, err)
// 		})
// 	}
// }

func TestService_Push(t *testing.T) {
	tests := []struct {
		name      string
		opt       model.PushOption
		setupAuth *mockAuthService
		setupRepo func() model.Repository
		wantErr   bool
		errType   error
	}{
		{
			name: "auth error during push",
			opt: model.PushOption{
				Target: "origin",
				AuthCfg: gpsconfig.AuthConfig{
					Protocol: gpsconfig.TLS,
				},
			},
			setupAuth: &mockAuthService{
				err: ErrInvalidAuth,
			},
			wantErr: true,
			errType: ErrAuthMethod,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			svc := &Service{
				authService: tabletest.setupAuth,
				Ops:         *NewOperation(),
				metadata:    &mockMetadataHandler{},
			}

			var repo model.Repository
			if tabletest.setupRepo != nil {
				repo = tabletest.setupRepo()
			}

			err := svc.Push(context.Background(), repo, tabletest.opt)

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.errType)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestNewService(t *testing.T) {
	svc := NewService()
	require.NotNil(t, svc)
	require.NotNil(t, svc.authService)
	require.NotNil(t, svc.metadata)
}
