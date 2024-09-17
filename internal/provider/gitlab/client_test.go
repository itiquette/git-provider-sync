// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlab

// func gitLabClientMock() Client {
// 	return Client{
// 		providerClient: nil,
// 	}
// }
//
// func TestValidate(t *testing.T) {
// 	ctx := context.Background()
//
// 	input := model.CLIOption{ForcePush: true}
// 	ctx = model.WithCLIOption(ctx, input)
//
// 	require := require.New(t)
//
// 	type args struct {
// 		repositoryName string
// 	}
//
// 	tests := map[string]struct {
// 		args args
// 		want bool
// 	}{
// 		"is valid success": {
// 			args: args{
// 				repositoryName: "aname",
// 			},
// 			want: true,
// 		},
// 		"is valid fail": {
// 			args: args{
// 				repositoryName: "commits",
// 			},
// 			want: false,
// 		},
// 	}
// 	for name, tableTest := range tests {
// 		t.Run(name, func(_ *testing.T) {
// 			glc := Client{
// 				providerClient: gitLabClientMock().Client(),
// 			}
// 			got := glc.IsValidRepositoryName(ctx, tableTest.args.repositoryName)
// 			require.Equal(tableTest.want, got)
// 		})
// 	}
// }
//
// func (m *MockGitProviderClient) SupportedDomain() string {
// 	args := m.Called()
// 	return args.Get(0).(string)
// }
// func (m *MockGitProviderClient) ProviderID() gitprovider.ProviderID {
// 	args := m.Called()
// 	return args.Get(0).(gitprovider.ProviderID)
// }
//
// func (m *MockGitProviderClient) Raw() interface{} {
// 	args := m.Called()
// 	return args.Get(0)
// }
//
// // MockGitProviderClient is a mock for the gitprovider.Client.
// type MockGitProviderClient struct {
// 	mock.Mock
// }
//
// func (m *MockGitProviderClient) OrgRepositories() gitprovider.OrgRepositoriesClient {
// 	args := m.Called()
// 	return args.Get(0).(gitprovider.OrgRepositoriesClient)
// }
//
// func (m *MockGitProviderClient) UserRepositories() gitprovider.UserRepositoriesClient {
// 	args := m.Called()
// 	return args.Get(0).(gitprovider.UserRepositoriesClient)
// }
//
// func (m *MockGitProviderClient) HasTokenPermission(ctx context.Context, permission gitprovider.TokenPermission) (bool, error) {
// 	args := m.Called(ctx, permission)
// 	return args.Bool(0), args.Error(1)
// }
//
// func (m *MockGitProviderClient) Organizations() gitprovider.OrganizationsClient {
// 	args := m.Called()
// 	return args.Get(0).(gitprovider.OrganizationsClient)
// }
//
// // MockOrgRepositoriesClient is a mock for the gitprovider.OrgRepositoriesClient.
// type MockOrgRepositoriesClient struct {
// 	mock.Mock
// }
//
// func (m *MockOrgRepositoriesClient) Create(ctx context.Context, orgRepoRef gitprovider.OrgRepositoryRef, repoInfo gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryCreateOption) (gitprovider.OrgRepository, error) {
// 	args := m.Called(ctx, orgRepoRef, repoInfo, opts)
// 	return args.Get(0).(gitprovider.OrgRepository), args.Error(1)
// }
//
// func (m *MockOrgRepositoriesClient) Get(ctx context.Context, ref gitprovider.OrgRepositoryRef) (gitprovider.OrgRepository, error) {
// 	args := m.Called(ctx, ref)
// 	return args.Get(0).(gitprovider.OrgRepository), args.Error(1)
// }
//
// func (m *MockOrgRepositoriesClient) Reconcile(ctx context.Context, ref gitprovider.OrgRepositoryRef, info gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryReconcileOption) (gitprovider.OrgRepository, bool, error) {
// 	args := m.Called(ctx, ref, info, opts)
// 	return args.Get(0).(gitprovider.OrgRepository), false, args.Error(1)
// }
// func (m *MockOrgRepositoriesClient) List(ctx context.Context, org gitprovider.OrganizationRef) ([]gitprovider.OrgRepository, error) {
// 	args := m.Called(ctx, org)
// 	return args.Get(0).([]gitprovider.OrgRepository), args.Error(1)
// }
//
// // MockUserRepositoryClient is a mock for the gitprovider.UserRepositoryClient.
// type MockUserRepositoryClient struct {
// 	mock.Mock
// }
//
// func (m *MockUserRepositoryClient) Create(ctx context.Context, userRepoRef gitprovider.UserRepositoryRef, repoInfo gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryCreateOption) (gitprovider.UserRepository, error) {
// 	args := m.Called(ctx, userRepoRef, repoInfo, opts)
// 	return args.Get(0).(gitprovider.UserRepository), args.Error(1)
// }
//
// func (m *MockUserRepositoryClient) Get(ctx context.Context, ref gitprovider.UserRepositoryRef) (gitprovider.UserRepository, error) {
// 	args := m.Called(ctx, ref)
// 	return args.Get(0).(gitprovider.UserRepository), args.Error(1)
// }
//
// func (m *MockUserRepositoryClient) GetUserLogin(ctx context.Context) (gitprovider.IdentityRef, error) {
// 	args := m.Called(ctx)
// 	return args.Get(0).(gitprovider.IdentityRef), args.Error(1)
// }
//
// type MockUserRepositoriesClient struct {
// 	mock.Mock
// }
//
// func (m *MockUserRepositoriesClient) List(ctx context.Context, o gitprovider.UserRef) ([]gitprovider.UserRepository, error) {
// 	args := m.Called(ctx, o)
// 	return args.Get(0).([]gitprovider.UserRepository), args.Error(1)
// }
//
// func (m *MockUserRepositoriesClient) Get(ctx context.Context, ref gitprovider.UserRepositoryRef) (gitprovider.UserRepository, error) {
// 	args := m.Called(ctx, ref)
// 	return args.Get(0).(gitprovider.UserRepository), args.Error(1)
// }
//
// func (m *MockUserRepositoryClient) List(ctx context.Context, user gitprovider.UserRef) ([]gitprovider.UserRepository, error) {
// 	args := m.Called(ctx, user)
// 	return args.Get(0).([]gitprovider.UserRepository), args.Error(1)
// }
//
// func (m *MockUserRepositoriesClient) Reconcile(ctx context.Context, ref gitprovider.UserRepositoryRef, info gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryReconcileOption) (gitprovider.UserRepository, bool, error) {
// 	args := m.Called(ctx, ref, info, opts)
// 	return args.Get(0).(gitprovider.UserRepository), args.Bool(1), args.Error(2)
// }
//
// func (m *MockUserRepositoryClient) Reconcile(ctx context.Context, ref gitprovider.UserRepositoryRef, info gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryReconcileOption) (gitprovider.UserRepository, bool, error) {
// 	args := m.Called(ctx, ref, info, opts)
// 	return args.Get(0).(gitprovider.UserRepository), args.Bool(1), args.Error(2)
// }
//
// func (m *MockUserRepositoriesClient) GetUserLogin(ctx context.Context) (gitprovider.IdentityRef, error) {
// 	args := m.Called(ctx)
// 	return args.Get(0).(gitprovider.IdentityRef), args.Error(1)
// }
//
// // MockOrganizationClient is a mock for the gitprovider.OrganizationClient.
// type MockOrganizationClient struct {
// 	mock.Mock
// }
//
// func (m *MockOrganizationClient) Get(ctx context.Context, ref gitprovider.OrganizationRef) (gitprovider.Organization, error) {
// 	args := m.Called(ctx, ref)
// 	return args.Get(0).(gitprovider.Organization), args.Error(1)
// }
//
// func TestCreate(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		config      configuration.ProviderConfig
// 		option      model.CreateOption
// 		setupMock   func(*MockGitProviderClient)
// 		expectError bool
// 		errorMsg    string
// 	}{
// 		{
// 			name: "Create group repository success",
// 			config: configuration.ProviderConfig{
// 				Domain: "gitlab.com",
// 				Group:  "testgroup",
// 			},
// 			option: model.CreateOption{
// 				RepositoryName: "testrepo",
// 				Visibility:     "public",
// 				Description:    "Test repository",
// 			},
// 			setupMock: func(m *MockGitProviderClient) {
// 				orgRepoClient := new(MockOrgRepositoriesClient)
// 				orgRepoClient.On("Create",
// 					mock.Anything,
// 					mock.AnythingOfType("gitprovider.OrgRepositoryRef"),
// 					mock.AnythingOfType("gitprovider.RepositoryInfo"),
// 					mock.AnythingOfType("[]gitprovider.RepositoryCreateOption")).
// 					Return(struct{ gitprovider.OrgRepository }{}, nil)
// 				m.On("OrgRepositories").Return(orgRepoClient)
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "Create user repository success",
// 			config: configuration.ProviderConfig{
// 				Domain: "gitlab.com",
// 				User:   "testuser",
// 			},
// 			option: model.CreateOption{
// 				RepositoryName: "testrepo",
// 				Visibility:     "private",
// 				Description:    "Test user repository",
// 			},
// 			setupMock: func(m *MockGitProviderClient) {
// 				userRepoClient := new(MockUserRepositoryClient)
// 				userRepoClient.On("Create",
// 					mock.Anything,
// 					mock.AnythingOfType("gitprovider.UserRepositoryRef"),
// 					mock.AnythingOfType("gitprovider.RepositoryInfo"),
// 					mock.AnythingOfType("[]gitprovider.RepositoryCreateOption")).
// 					Return(struct{ gitprovider.UserRepository }{}, nil)
// 				m.On("UserRepositories").Return(userRepoClient)
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "Create group repository failure",
// 			config: configuration.ProviderConfig{
// 				Domain: "gitlab.com",
// 				Group:  "testgroup",
// 			},
// 			option: model.CreateOption{
// 				RepositoryName: "testrepo",
// 				Visibility:     "public",
// 				Description:    "Test repository",
// 			},
// 			setupMock: func(m *MockGitProviderClient) {
// 				orgRepoClient := new(MockOrgRepositoriesClient)
// 				orgRepoClient.On("Create",
// 					mock.Anything,
// 					mock.AnythingOfType("gitprovider.OrgRepositoryRef"),
// 					mock.AnythingOfType("gitprovider.RepositoryInfo"),
// 					mock.AnythingOfType("[]gitprovider.RepositoryCreateOption")).
// 					Return(struct{ gitprovider.OrgRepository }{}, errors.New("failed to create repository"))
// 				m.On("OrgRepositories").Return(orgRepoClient)
// 			},
// 			expectError: true,
// 			errorMsg:    "failed to create repository",
// 		},
// 		{
// 			name: "Create with empty repository name",
// 			config: configuration.ProviderConfig{
// 				Domain: "gitlab.com",
// 				User:   "testuser",
// 			},
// 			option: model.CreateOption{
// 				RepositoryName: "",
// 				Visibility:     "public",
// 				Description:    "Test repository",
// 			},
// 			setupMock: func(m *MockGitProviderClient) {
// 				userRepoClient := new(MockUserRepositoryClient)
// 				userRepoClient.On("Create",
// 					mock.Anything,
// 					mock.AnythingOfType("gitprovider.UserRepositoryRef"),
// 					mock.AnythingOfType("gitprovider.RepositoryInfo"),
// 					mock.AnythingOfType("[]gitprovider.RepositoryCreateOption")).
// 					Return(struct{ gitprovider.UserRepository }{}, errors.New("repository name cannot be empty"))
// 				m.On("UserRepositories").Return(userRepoClient)
// 			},
// 			expectError: true,
// 			errorMsg:    "repository name cannot be empty",
// 		},
// 		{
// 			name: "Create with invalid visibility",
// 			config: configuration.ProviderConfig{
// 				Domain: "gitlab.com",
// 				Group:  "testgroup",
// 			},
// 			option: model.CreateOption{
// 				RepositoryName: "testrepo",
// 				Visibility:     "invalid",
// 				Description:    "Test repository",
// 			},
// 			setupMock: func(m *MockGitProviderClient) {
// 				orgRepoClient := new(MockOrgRepositoriesClient)
// 				orgRepoClient.On("Create",
// 					mock.Anything,
// 					mock.AnythingOfType("gitprovider.OrgRepositoryRef"),
// 					mock.AnythingOfType("gitprovider.RepositoryInfo"),
// 					mock.AnythingOfType("[]gitprovider.RepositoryCreateOption")).
// 					Return(struct{ gitprovider.OrgRepository }{}, errors.New("invalid visibility"))
// 				m.On("OrgRepositories").Return(orgRepoClient)
// 			},
// 			expectError: true,
// 			errorMsg:    "invalid visibility",
// 		},
// 		{
// 			name: "Create with very long repository name",
// 			config: configuration.ProviderConfig{
// 				Domain: "gitlab.com",
// 				User:   "testuser",
// 			},
// 			option: model.CreateOption{
// 				RepositoryName: strings.Repeat("a", 256), // 256 character name
// 				Visibility:     "public",
// 				Description:    "Test repository",
// 			},
// 			setupMock: func(m *MockGitProviderClient) {
// 				userRepoClient := new(MockUserRepositoryClient)
// 				userRepoClient.On("Create",
// 					mock.Anything,
// 					mock.AnythingOfType("gitprovider.UserRepositoryRef"),
// 					mock.AnythingOfType("gitprovider.RepositoryInfo"),
// 					mock.AnythingOfType("[]gitprovider.RepositoryCreateOption")).
// 					Return(struct{ gitprovider.UserRepository }{}, errors.New("repository name too long"))
// 				m.On("UserRepositories").Return(userRepoClient)
// 			},
// 			expectError: true,
// 			errorMsg:    "repository name too long",
// 		},
// 	}
//
// 	for _, tabletest := range tests {
// 		t.Run(tabletest.name, func(t *testing.T) {
// 			require := require.New(t)
// 			ctx := context.Background()
//
// 			mockProviderClient := new(MockGitProviderClient)
// 			tabletest.setupMock(mockProviderClient)
//
// 			client := Client{
// 				providerClient: mockProviderClient,
// 			}
//
// 			err := client.Create(ctx, tabletest.config, tabletest.option)
//
// 			if tabletest.expectError {
// 				require.Error(err)
// 				require.Contains(err.Error(), tabletest.errorMsg)
// 			} else {
// 				require.NoError(err)
// 			}
//
// 			mockProviderClient.AssertExpectations(t)
// 		})
// 	}
// }
// func TestNewGitLabClient(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		option      model.GitProviderClientOption
// 		expectError bool
// 	}{
// 		{
// 			name: "Valid client creation",
// 			option: model.GitProviderClientOption{
// 				Domain: "gitlab.com",
// 				Token:  "valid-token",
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "Invalid domain",
// 			option: model.GitProviderClientOption{
// 				Domain: "",
// 				Token:  "valid-token",
// 			},
// 			expectError: true,
// 		},
// 	}
//
// 	for _, tabletest := range tests {
// 		t.Run(tabletest.name, func(t *testing.T) {
// 			ctx := context.Background()
// 			client, err := NewGitLabClient(ctx, tabletest.option)
//
// 			if tabletest.expectError {
// 				assert.Error(t, err)
// 				assert.Nil(t, client.providerClient)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, client.providerClient)
// 			}
// 		})
// 	}
// }
//
// func TestIsValidRepositoryName(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		repoName string
// 		expected bool
// 	}{
// 		{"Valid name", "valid-repo", true},
// 		{"Valid name with numbers", "repo123", true},
// 		{"Valid name with underscore", "valid_repo", true},
// 		{"Invalid name - reserved word", "create", false},
// 		{"Invalid name - starts with dash", "-invalid", false},
// 		{"Invalid name - special characters", "invalid@repo", false},
// 		{"Empty name", "", false},
// 		{"Very long name", string(make([]byte, 256)), false},
// 	}
//
// 	for _, tabletest := range tests {
// 		t.Run(tabletest.name, func(t *testing.T) {
// 			client := Client{}
// 			ctx := context.Background()
// 			result := client.IsValidRepositoryName(ctx, tabletest.repoName)
// 			assert.Equal(t, tabletest.expected, result)
// 		})
// 	}
// }
//
