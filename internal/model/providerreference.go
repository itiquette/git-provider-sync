// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"github.com/fluxcd/go-git-providers/gitprovider"
)

// NewOrgRepositoryRef creates a new OrgRepositoryRef.
// This reference type is used to identify a repository within an organization.
//
// Parameters:
//   - domain: The domain of the git hosting service (e.g., "github.com").
//   - org: The name of the organization.
//   - repo: The name of the repository.
//
// Returns:
//   - A gitprovider.OrgRepositoryRef struct.
func NewOrgRepositoryRef(domain, org, repo string) gitprovider.OrgRepositoryRef {
	return gitprovider.OrgRepositoryRef{
		RepositoryName: repo,
		OrganizationRef: gitprovider.OrganizationRef{
			Domain:       domain,
			Organization: org,
		},
	}
}

// NewUserRepositoryRef creates a new UserRepositoryRef.
// This reference type is used to identify a repository owned by a user.
//
// Parameters:
//   - domain: The domain of the git hosting service (e.g., "github.com").
//   - user: The username of the repository owner.
//   - repo: The name of the repository.
//
// Returns:
//   - A gitprovider.UserRepositoryRef struct.
func NewUserRepositoryRef(domain, user, repo string) gitprovider.UserRepositoryRef {
	return gitprovider.UserRepositoryRef{
		RepositoryName: repo,
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: user,
		},
	}
}

// NewOrgRef creates a new OrganizationRef.
// This reference type is used to identify an organization.
//
// Parameters:
//   - domain: The domain of the git hosting service (e.g., "github.com").
//   - org: The name of the organization.
//
// Returns:
//   - A gitprovider.OrganizationRef struct.
func NewOrgRef(domain, org string) gitprovider.OrganizationRef {
	return gitprovider.OrganizationRef{
		Domain:       domain,
		Organization: org,
	}
}

// NewUserRef creates a new UserRef.
// This reference type is used to identify a user.
//
// Parameters:
//   - domain: The domain of the git hosting service (e.g., "github.com").
//   - user: The username of the user.
//
// Returns:
//   - A gitprovider.UserRef struct.
func NewUserRef(domain, user string) gitprovider.UserRef {
	return gitprovider.UserRef{
		Domain:    domain,
		UserLogin: user,
	}
}

// Example usage:
//
//	orgRepoRef := NewOrgRepositoryRef("github.com", "myorg", "myrepo")
//	userRepoRef := NewUserRepositoryRef("gitlab.com", "johndoe", "project")
//	orgRef := NewOrgRef("bitbucket.org", "companyteam")
//	userRef := NewUserRef("github.com", "janedoe")
