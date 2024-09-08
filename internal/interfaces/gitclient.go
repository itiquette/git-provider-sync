// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"github.com/fluxcd/go-git-providers/gitprovider"
)

// GitClient defines an interface for accessing a gitprovider.Client.
// This interface serves as a wrapper around the gitprovider.Client,
// allowing for easier mocking in tests and providing a layer of abstraction
// between the application and the specific Git provider implementation.
type GitClient interface {
	// Client returns the underlying gitprovider.Client.
	//
	// Returns:
	//   - gitprovider.Client: The underlying client for interacting with a Git provider.
	//
	// This method provides access to the full functionality of the gitprovider.Client,
	// which includes operations such as managing repositories, branches, and pull requests.
	// The exact capabilities depend on the specific implementation of gitprovider.Client.
	Client() gitprovider.Client
}

// Example usage:
//
//	type MyGitClient struct {
//		client gitprovider.Client
//	}
//
//	func (m *MyGitClient) Client() gitprovider.Client {
//		return m.client
//	}
//
//	func NewMyGitClient(token, hostname string) (*MyGitClient, error) {
//		client, err := gitprovider.NewClient(token, gitprovider.WithHostname(hostname))
//		if err != nil {
//			return nil, err
//		}
//		return &MyGitClient{client: client}, nil
//	}
//
//	func UseGitClient(gc GitClient) error {
//		// Use the client to perform operations
//		client := gc.Client()
//
//		// Example: List repositories for the authenticated user
//		repos, err := client.UserRepositories().List(context.Background())
//		if err != nil {
//			return err
//		}
//
//		for _, repo := range repos {
//			fmt.Printf("Repository: %s\n", repo.Repository().GetRepository())
//		}
//
//		return nil
//	}
//
//	func main() {
//		client, err := NewMyGitClient("your-token", "github.com")
//		if err != nil {
//			log.Fatalf("Failed to create client: %v", err)
//		}
//
//		if err := UseGitClient(client); err != nil {
//			log.Fatalf("Error using Git client: %v", err)
//		}
//	}
