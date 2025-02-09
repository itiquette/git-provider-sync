// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package synccmd

// import (
// 	"context"
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"
// 	"itiquette/git-provider-sync/internal/log"
// 	"itiquette/git-provider-sync/internal/model"
// 	config "itiquette/git-provider-sync/internal/model/configuration"

// 	"github.com/spf13/pflag"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"
// )

// // func TestExecuteSyncCommandFileConfArgSuccess(t *testing.T) {
// // 	require := require.New(t)
// //
// // 	cmd := newSyncCommand()
// //
// // 	require.Empty(cmd.Commands())
// //
// // 	//	flagConfigFile = "testdata/testconfig.yaml"
// // 	//	flagDryRun = true
// //
// // 	_ = cmd.Execute()
// // }

// func TestCreateTmpDir(t *testing.T) {
// 	require := require.New(t)

// 	type args struct {
// 		dir    string
// 		prefix string
// 	}

// 	tests := map[string]struct {
// 		args    args
// 		want    string
// 		wantErr bool
// 	}{
// 		"default os tmp dir with regular name success": {
// 			args: args{dir: "", prefix: "gitprovidersync"},
// 			want: filepath.Join(os.TempDir(), "gitprovidersync."),
// 		},
// 		"invalid os characters for tmp directory fail": {
// 			args:    args{dir: "#¤%&/()=?!", prefix: "guthostsync"},
// 			want:    "",
// 			wantErr: true,
// 		},
// 		"non-existent directory fail": {
// 			args:    args{dir: "/nonexistent", prefix: "test"},
// 			want:    "",
// 			wantErr: true,
// 		},
// 		// "custom directory success": {
// 		// 	args: args{dir: "/tmp/custom", prefix: "gitprovidersync"},
// 		// 	want: filepath.Join("/tmp/custom", "gitprovidersync."),
// 		// },
// 		"empty prefix success": {
// 			args: args{dir: "", prefix: ""},
// 			want: filepath.Join(os.TempDir(), ""),
// 		},
// 	}

// 	for name, tableTest := range tests {
// 		ctx := context.Background()
// 		ctx = log.InitLogger(ctx, newPrintCommand(), false, string(log.CONSOLE))

// 		t.Run(name, func(_ *testing.T) {
// 			ctx, err := model.CreateTmpDir(ctx, tableTest.args.dir, tableTest.args.prefix)

// 			if !tableTest.wantErr {
// 				tmpDir, _ := ctx.Value(model.TmpDirKey{}).(string)
// 				require.DirExists(tmpDir, "tmp directory should exist")
// 				require.Contains(tmpDir, tableTest.want, "tmp directory name should contain prefix")
// 			} else {
// 				require.Error(err, "error should be returned")
// 				require.Contains(err.Error(), "failed to create temporary", "error message should be descriptive")
// 			}
// 		})
// 	}
// }

// func TestNewSyncCommand(t *testing.T) {
// 	cmd := newSyncCommand()
// 	require.NotNil(t, cmd)
// 	require.Equal(t, "sync", cmd.Use)
// 	require.NotEmpty(t, cmd.Short)
// 	require.NotEmpty(t, cmd.Long)
// 	require.NotNil(t, cmd.Run)

// 	// Check if all expected flags are present
// 	expectedFlags := []string{"force-push", "ignore-invalid-name", "alphanumhyph-name", "active-from-limit"}
// 	for _, flag := range expectedFlags {
// 		// Assert that the flag is defined by checking if Lookup returns a non-nil *pflag.Flag
// 		require.NotEqual(t, (*pflag.Flag)(nil), cmd.Flags().Lookup(flag), "Flag %s should be present", flag)
// 	}
// }

// type MockConfigLoader struct {
// 	mock.Mock
// }

// func (m *MockConfigLoader) LoadConfiguration(ctx context.Context) (*config.AppConfiguration, error) {
// 	args := m.Called(ctx)
// 	//nolint:wrapcheck
// 	return args.Get(0).(*config.AppConfiguration), args.Error(1) //nolint:forcetypeassert
// }

// // func TestRunSync(t *testing.T) { TO-Do mock client to test
// //
// //
// // 	cmd := newSyncCommand()
// // 	ctx := context.Background()
// // 	ctx = log.InitLogger(ctx, cmd)
// // 	cmd.SetContext(ctx)
// // 	cmd.Flags().String("config-file","testdata/syncdryrun.yaml","")
// // 	cmd.Root().PersistentFlags().String("config-file","testdata/syncdryrun.yaml","")
// // 	cmd.Flags().Bool("config-file-only",false,"")
// // 	cmd.Flags().Set("dry-run","true")
// // 	cmd.Root().PersistentFlags().Bool("config-file-only",false,"")
// //
// // 	// Capture output
// // 	output := captureOutput(func() {
// // 		runSync(cmd, []string{})
// // 	})
// //
// // 	require.Contains(t, output, "All sync configurations completed")
// // }
// //
// // func captureOutput(f func()) string {
// // 	old := os.Stdout
// // 	r, w, _ := os.Pipe()
// // 	os.Stdout = w
// //
// // 	f()
// //
// // 	w.Close()
// //
// // 	os.Stdout = old
// //
// // 	var buf bytes.Buffer
// //
// // 	io.Copy(&buf, r)
// //
// // 	return buf.String()
// // }

// func TestIsValidRepository(t *testing.T) {
// 	ctx := context.Background()
// 	ctx = model.WithCLIOption(ctx, model.CLIOption{DryRun: false})

// 	tests := []struct {
// 		name     string
// 		repoName string
// 		expected bool
// 	}{
// 		{"Valid name", "valid-repo", true},
// 		{"Invalid name", "invalid/repo", false},
// 	}

// 	for _, tabletest := range tests {
// 		t.Run(tabletest.name, func(t *testing.T) {
// 			mockProvider := &mocks.GitProvider{}
// 			mockRepo := &mocks.GitRepository{}

// 			mockProvider.On("IsValidProjectName", ctx, tabletest.repoName).Return(tabletest.expected)
// 			mockRepo.On("Metainfo").Return(model.RepositoryMetainfo{OriginalName: tabletest.repoName})

// 			result := isValidRepository(ctx, mockProvider, mockRepo)
// 			require.Equal(t, tabletest.expected, result)
// 		})
// 	}
// }

// // type mockGitProvider struct {
// // 	mock.Mock
// // }
// //
// // // Create implements interfaces.GitProvider.
// // func (m *mockGitProvider) Create(_ context.Context, _ model.ProviderConfig, _ model.CreateOption) error {
// // 	panic("unimplemented")
// // }
// //
// // // ProjectInfos implements interfaces.GitProvider.
// // func (m *mockGitProvider) Projectinfos(_ context.Context, _ model.ProviderConfig, _ bool) ([]model.RepositoryMetainfo, error) {
// // 	panic("unimplementedb")
// // }
// //
// // // Name implements interfaces.GitProvider.
// // func (m *mockGitProvider) Name() string {
// // 	return "provider"
// // }
// //
// // func (m *mockGitProvider) IsValidProjectName(ctx context.Context, name string) bool {
// // 	args := m.Called(ctx, name)
// //
// // 	return args.Bool(0)
// // }
// //
// // type mockGitRepository struct {
// // 	mock.Mock
// // }
// //
// // // DeleteRemote implements interfaces.GitRepository.
// // func (m *mockGitRepository) DeleteRemote(_ string) error {
// // 	return nil
// // }
// //
// // // GoGitRepository implements interfaces.GitRepository.
// // func (m *mockGitRepository) GoGitRepository() *git.Repository {
// // 	panic("unimplemented")
// // }
// //
// // // Metainfo implements interfaces.GitRepository.
// // func (m *mockGitRepository) Metainfo() model.RepositoryMetainfo {
// // 	return model.RepositoryMetainfo{OriginalName: "originalName"}
// // }
// //
// // // Remote implements interfaces.GitRepository.
// // func (m *mockGitRepository) Remote(_ string) (model.Remote, error) {
// // 	panic("unimplemented")
// // }
// //
// // func (m *mockGitRepository) CreateRemote(_ string, _ string, _ bool) error {
// // 	args := m.Called()
// //
// // 	return args.Error(1) //nolint:wrapcheck
// // }
