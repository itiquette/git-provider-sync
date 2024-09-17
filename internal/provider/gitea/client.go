// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package gitea provides a client for interacting with Gitea repositories
// using the go-git-providers library. It offers a range of functionalities including:
//   - Creating and listing repositories
//   - Filtering repository metadata based on various criteria
//   - Validating repository names according to GitHub's rules
//   - Performing common operations on repositories
//
// This package aims to simplify Gitea interactions in Go applications, providing
// a interface for repository management and metadata handling.
package gitea

// import (
// 	"context"
// 	"fmt"
//
// 	"itiquette/git-provider-sync/internal/configuration"
// 	"itiquette/git-provider-sync/internal/interfaces"
// 	"itiquette/git-provider-sync/internal/log"
// 	"itiquette/git-provider-sync/internal/model"
//
// 	"github.com/pkg/errors"
//
// 	"code.gitea.io/sdk/gitea"
// 	giteasdk "code.gitea.io/sdk/gitea"
// )
//
// // Client represents a Gitea client that can perform various operations
// // on Gitea repositories.
// type Client struct {
// 	client  giteasdk.Client
// 	filter         Filter
// }
//
// // Client returns the underlying gitprovider.Client.
// func (c *Client) Client() giteasdk.Client {
// 	return c.client
// }
//
// // Create creates a new repository in Gitea.
// // It supports creating repositories for both users and organizations.
// //
// // Parameters:
// // - ctx: The context for the operation.
// // - config: Configuration for the provider, including domain and user/group information.
// // - option: Options for creating the repository, including name, visibility, and description.
// //
// // Returns an error if the creation fails.
// func (c Client) Create(ctx context.Context, config configuration.ProviderConfig, option model.CreateOption) error {
// 	logger := log.Logger(ctx)
// 	logger.Trace().Msg("Entering Gitea:Create:")
// 	config.DebugLog(logger).Msg("Gitea:Create:")
//
// 	repoInfo := gitprovider.RepositoryInfo{
// 		Visibility:    gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibility(option.Visibility)),
// 		Description:   &option.Description,
// 		DefaultBranch: &option.DefaultBranch,
// 	}
//
// 	var err error
//
// 	if config.IsGroup() {
// 		_, err = c.providerClient.OrgRepositories().Create(
// 			ctx,
// 			gitprovider.OrgRepositoryRef{
// 				OrganizationRef: gitprovider.OrganizationRef{
// 					Domain:       config.Domain,
// 					Organization: config.Group,
// 				},
// 				RepositoryName: option.RepositoryName,
// 			},
// 			repoInfo,
// 			&gitprovider.RepositoryCreateOptions{},
// 		)
// 	} else {
// 		_, err = c.providerClient.UserRepositories().Create(
// 			ctx,
// 			gitprovider.UserRepositoryRef{
// 				UserRef: gitprovider.UserRef{
// 					Domain:    config.Domain,
// 					UserLogin: config.User,
// 				},
// 				RepositoryName: option.RepositoryName,
// 			},
// 			repoInfo,
// 			&gitprovider.RepositoryCreateOptions{},
// 		)
// 	}
//
// 	if err != nil {
// 		return fmt.Errorf("failed to create repository %s: %w", option.RepositoryName, err)
// 	}
//
// 	logger.Trace().Msg("Repository created successfully")
//
// 	return nil
// }
//
// // Name returns the name of the provider, which is "GITEA".
// func (c Client) Name() string {
// 	return configuration.GITEA
// }
//
// // Metainfos retrieves metadata information for repositories.
// // It can list repositories for both users and organizations.
// //
// // Parameters:
// // - ctx: The context for the operation.
// // - config: Configuration for the provider, including domain and user/group information.
// // - filtering: If true, applies additional filtering to the results.
// //
// // Returns a slice of RepositoryMetainfo and an error if the operation fails.
// func (c Client) Metainfos(ctx context.Context, config configuration.ProviderConfig, filtering bool) ([]model.RepositoryMetainfo, error) {
// 	var (
// 		repositories interface{}
// 		err          error
// 	)
//
// 	if config.IsGroup() {
// 		repositories, err = c.providerClient.OrgRepositories().List(ctx, gitprovider.OrganizationRef{
// 			Domain:       config.Domain,
// 			Organization: config.Group,
// 		})
// 	} else {
// 		repositories, err = c.providerClient.UserRepositories().List(ctx, gitprovider.UserRef{
// 			Domain:    config.Domain,
// 			UserLogin: config.User,
// 		})
// 	}
//
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
// 	}
//
// 	var metainfos []model.RepositoryMetainfo
//
// 	switch repos := repositories.(type) {
// 	case []gitprovider.OrgRepository:
// 		for _, repo := range repos {
// 			rm, _ := newRepositoryMeta(ctx, config, c, repo.Repository().GetRepository())
// 			metainfos = append(metainfos, rm)
// 		}
// 	case []gitprovider.UserRepository:
// 		for _, repo := range repos {
// 			rm, _ := newRepositoryMeta(ctx, config, c, repo.Repository().GetRepository())
// 			metainfos = append(metainfos, rm)
// 		}
// 	default:
// 		return nil, errors.New("unknown repository type returned")
// 	}
//
// 	if filtering {
// 		return c.filter.FilterMetainfo(ctx, config, metainfos)
// 	}
//
// 	return metainfos, nil
// }
//
// // IsValidRepositoryName checks if the given repository name is valid for Gitea.
// // It applies Gitea-specific naming rules.
// //
// // Parameters:
// // - ctx: The context for the operation.
// // - name: The repository name to validate.
// //
// // Returns true if the name is valid, false otherwise.
// func (c Client) IsValidRepositoryName(ctx context.Context, name string) bool {
// 	logger := log.Logger(ctx)
// 	logger.Trace().Msg("Entering Gitea:Validate:")
// 	logger.Debug().Str("name", name).Msg("Gitea:Validate:")
//
// 	return IsValidGiteaRepositoryName(name)
// }
//
// // NewGiteaClient creates a new Gitea client.
// //
// // Parameters:
// // - ctx: The context for the operation.
// // - option: Options for creating the client, including domain and authentication token.
// //
// // Returns a new Client and an error if the creation fails.
// func NewGiteaClient(ctx context.Context, option model.GitProviderClientOption) (Client, error) {
// 	logger := log.Logger(ctx)
// 	logger.Trace().Msg("Entering NewGiteaClient:")
//
// 	cliOption, _ := ctx.Value(model.CLIOptionKey{}).(model.CLIOption)
// 	domain := option.Domain
//
// 	if cliOption.PlainHTTP {
// 		domain = "http://" + domain
// 	}
//
// 	clientOpts := &gitprovider.ClientOptions{
// 		CommonClientOptions: gitprovider.CommonClientOptions{Domain: &domain},
// 	}
//
// 	var (
// 		client gitprovider.Client
// 		err    error
// 	)
//
// 	if option.Token == "" {
// 		client, err = gitea.NewClient("", clientOpts)
// 	} else {
// 		client, err = gitea.NewClient(option.Token, clientOpts)
// 	}
//
// 	if err != nil {
// 		return Client{}, fmt.Errorf("failed to create a new Gitea client: %w", err)
// 	}
//
// 	identity, err := client.UserRepositories().GetUserLogin(ctx)
// 	if err != nil || identity.GetIdentity() == "" {
// 		return Client{}, fmt.Errorf("failed to get user identity: %w", err)
// 	}
//
// 	return Client{providerClient: client}, nil
// }
//
// // newRepositoryMeta creates a new RepositoryMetainfo struct for a given repository.
// // It fetches detailed information about the repository from Gitea.
// //
// // Parameters:
// // - ctx: The context for the operation.
// // - config: Configuration for the provider.
// // - gitClient: The Gitea client to use for fetching repository information.
// // - repositoryName: The name of the repository to fetch information for.
// //
// // Returns a RepositoryMetainfo and an error if the operation fails.
// func newRepositoryMeta(ctx context.Context, config configuration.ProviderConfig, gitClient interfaces.GitClient, repositoryName string) (model.RepositoryMetainfo, error) {
// 	logger := log.Logger(ctx)
// 	logger.Trace().Msg("Entering newRepositoryMeta:")
//
// 	rawClient, ok := gitClient.Client().Raw().(*giteasdk.Client)
// 	if !ok {
// 		return model.RepositoryMetainfo{}, errors.New("failed to get raw client for gitea")
// 	}
//
// 	owner := config.Group
// 	if !config.IsGroup() {
// 		owner = config.User
// 	}
//
// 	giteaProject, _, err := rawClient.GetRepo(owner, repositoryName)
// 	if err != nil {
// 		return model.RepositoryMetainfo{}, fmt.Errorf("failed to get project info for %s: %w", repositoryName, err)
// 	}
//
// 	return model.RepositoryMetainfo{
// 		OriginalName:   repositoryName,
// 		HTTPSURL:       giteaProject.CloneURL,
// 		SSHURL:         giteaProject.SSHURL,
// 		Description:    giteaProject.Description,
// 		DefaultBranch:  giteaProject.DefaultBranch,
// 		LastActivityAt: &giteaProject.Updated,
// 		Visibility:     string(giteaProject.Owner.Visibility),
// 	}, nil
// }
//
// // TODO: Implement isValidGiteaRepositoryName and isValidGiteaRepositoryNameCharacters functions
// // These functions should contain the logic for validating Gitea repository names
// // according to Gitea's specific naming rules and allowed characters.
//
