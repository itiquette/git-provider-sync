<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Agent Instructions for Git Provider Sync

Comprehensive guidance for AI assistants working with the git-provider-sync codebase.

## Core Philosophy

**Be hexagonal, functional, idiomatic, and don't overengineer.**

This codebase follows these principles:
- **Hexagonal**: Clean separation between domain and infrastructure
- **Functional**: Pure functions with explicit dependencies
- **Idiomatic**: Follow Go conventions and patterns
- **Simple**: Don't overengineer - prefer clarity over cleverness

## Architecture Overview

**Hexagonal Architecture** - Domain-driven design with clear boundaries:
- **Domain** (`internal/domain/`) - Pure business logic, no external dependencies
- **Ports** (`internal/domain/ports/`) - Interfaces defined by domain
- **Adapters** (`internal/adapters/`) - Implementations of ports (providers, repository, filesystem, etc.)
- **Use Cases** (`internal/domain/sync/`) - Orchestrate domain operations

### Key Architectural Principles

Remember: **Be hexagonal, functional, idiomatic, and don't overengineer.**

1. **Pure Functions**: Business logic as pure functions with explicit dependencies (functional)
2. **Dependency Injection**: All dependencies passed explicitly via constructors/parameters (hexagonal)
3. **Interface Segregation**: Small, focused interfaces in domain/ports (idiomatic Go)
4. **No Global State**: Never use package-level variables for configuration (functional)
5. **Value Objects**: Prefer immutable value objects over mutable state (don't overengineer)

```go
// ✓ Good: Pure function with explicit dependencies
func (uc RepositoriesUseCase) Execute(ctx context.Context, request Request) (Response, error)

// ✗ Bad: Hidden dependencies or global state
func SyncRepositories() error // Where does config come from?
```

## Project Structure

```text
git-provider-sync/
├── cmd/                     # CLI commands
│   ├── gitprovidersync/    # Main entry point
│   ├── synccmd/            # Sync command implementation
│   ├── printcmd/           # Print/display commands
│   └── statuscmd/          # Status checking commands
├── internal/
│   ├── domain/             # Core business logic (pure, functional)
│   │   ├── sync/          # Sync use cases
│   │   ├── mirror/        # Mirror operations
│   │   ├── config/        # Configuration domain
│   │   ├── entities/      # Domain entities
│   │   └── ports/         # Interface definitions (hexagonal boundaries)
│   ├── adapters/          # External implementations (hexagonal adapters)
│   │   ├── providers/     # Git provider adapters (GitHub, GitLab, Gitea)
│   │   ├── repository/    # Git repository operations
│   │   ├── filesystem/    # File system operations
│   │   ├── auth/         # Authentication handling
│   │   └── cli/          # CLI output formatting
│   ├── composition/       # Dependency injection container
│   ├── configuration/     # Config loading and validation
│   └── model/            # Data models and value objects
└── docs/                   # Documentation

```

## Core Functionality

### Primary Use Cases

1. **Repository Sync** - Mirror repositories between Git providers
2. **Batch Operations** - Clone/archive multiple repositories
3. **Provider Support** - GitHub, GitLab, Gitea with extensible design
4. **Authentication** - TLS/SSH with token-based API access
5. **Filtering** - Include/exclude patterns, fork handling, activity filters

### Key Domain Concepts

- **Source** - Origin Git provider configuration
- **Mirror** - Target Git provider for synchronization  
- **Environment** - Named configuration grouping sources and mirrors
- **Provider** - Git hosting service (GitHub, GitLab, Gitea)
- **Repository** - Git repository with metadata

## Build & Test Commands

```bash
# Primary workflows
just dev                 # Main dev workflow: verify + build
just test               # Run all tests (unit + integration)
just verify             # Complete quality checks

# Building
just build-host         # Build for current architecture
just build-all          # Build for all architectures
just build-image        # Build container image

# Testing
just test-unit          # Unit tests only
just test-integration   # Integration tests only
go test -run TestName ./path/to/package  # Run specific test
go test -tags=integration ./...  # Run integration tests with build tag

# Quality & Linting
just lint               # Run all linters
just lint-fix           # Auto-fix linting issues
just quality            # Complete quality pipeline

# Installation
just install-local      # Install binary, man page, completions locally
just install-go-dev-tools  # Install Go development tools

# Dependencies
just upgrade-deps       # Upgrade Go module dependencies
just upgrade-go-dev-tools  # Upgrade development tools
```

## Code Style Guidelines

### Go Conventions (Be Idiomatic)
- **Module**: `itiquette/git-provider-sync` (Go 1.25.0+)
- **Formatting**: Standard `go fmt` formatting
- **Imports**: Group by stdlib → external → internal with blank lines
- **Naming**:  
  - Exported: `CamelCase`
  - Private: `camelCase`
  - Interfaces: `-er` suffix (e.g., `Provider`, `Syncer`) or `Interface`
  - Files: lowercase with underscores for tests (e.g., `sync.go`, `sync_test.go`)
- **Simplicity**: Don't overengineer - prefer simple, clear code over clever abstractions

### Error Handling

```go
// Wrap errors with context
return fmt.Errorf("failed to sync repository %s: %w", repoName, err)

// Define sentinel errors
var ErrInvalidConfig = errors.New("invalid configuration")

// Check specific errors
if errors.Is(err, ErrNotFound) {
    // Handle not found case
}
```

### Testing Patterns

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:    "success case",
            input:   validInput,
            want:    expectedOutput,
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### File Headers

```go
// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package packagename
```

Config files use CC0-1.0 license instead of EUPL-1.2.

## Configuration

### YAML Structure

```yaml
gitprovidersync:
  environment-name:           # Environment grouping
    source-name:             # Source configuration
      provider_type: github  # Provider: github, gitlab, gitea
      domain: github.com     # Provider domain
      owner: username        # Repository owner
      owner_type: user       # user or group
      auth:
        token: ${ENV_VAR}    # API token (use env vars)
      mirrors:
        target-name:         # Mirror target configuration
          provider_type: gitlab
          domain: gitlab.com
          owner: backup-user
          owner_type: user
```

### Environment Variables
- Config file: `GPS_CONFIG_FILE` or default `gitprovidersync.yaml`
- Tokens: Use `${GITHUB_TOKEN}`, `${GITLAB_TOKEN}`, etc. in config
- Log level: `GPS_LOG_LEVEL` (trace, debug, info, warn, error)

## Domain-Driven Patterns

**Remember: Be hexagonal, functional, idiomatic, and don't overengineer.**

### Use Case Pattern (Hexagonal & Functional)

```go
type UseCaseName struct {
    // Dependencies injected via constructor
    repository ports.Repository
    provider   ports.Provider
    logger     ports.Logger
}

func NewUseCaseName(repo ports.Repository, provider ports.Provider, logger ports.Logger) *UseCaseName {
    return &UseCaseName{
        repository: repo,
        provider:   provider,
        logger:     logger,
    }
}

func (uc *UseCaseName) Execute(ctx context.Context, request Request) (Response, error) {
    // Implementation
}
```

### Port/Adapter Pattern (Hexagonal Architecture)

```go
// Domain defines interface (port) - keep it simple, don't overengineer
type RepositoryProvider interface {
    ListRepositories(ctx context.Context, owner string) ([]Repository, error)
    GetRepository(ctx context.Context, owner, name string) (*Repository, error)
}

// Adapter implements interface - be idiomatic
type GitHubProvider struct {
    client *github.Client
}

func (g *GitHubProvider) ListRepositories(ctx context.Context, owner string) ([]Repository, error) {
    // GitHub-specific implementation - functional, no side effects
}
```

## Common Tasks

**Always: Be hexagonal, functional, idiomatic, and don't overengineer.**

### Adding a New Provider
1. Create adapter in `internal/adapters/providers/newprovider/`
2. Implement `ports.RepositoryProvider` interface
3. Add to provider factory in `internal/composition/providerfactory.go`
4. Add provider type to entity constants
5. Update configuration validation

### Adding a New Command
1. Create package in `cmd/newcmd/`
2. Define command structure with `urfave/cli/v3` (be idiomatic)
3. Implement command logic following hexagonal pattern (keep domain pure)
4. Register in root command (`cmd/root.go`)
5. Add tests for command and logic (don't overengineer tests)

### Modifying Sync Logic
1. Domain logic goes in `internal/domain/sync/` (keep it functional)
2. Keep use cases focused and composable (don't overengineer)
3. Define new ports if external dependencies needed (hexagonal boundaries)
4. Update relevant adapters to implement new ports (be idiomatic)
5. Maintain backward compatibility in configuration

## Testing Guidelines

**Keep tests functional, idiomatic, and don't overengineer.**

### Test Organization
- Unit tests: Same package, `*_test.go` suffix
- Integration tests: `//go:build integration` tag
- Test data: `testdata/` directories
- Mocks: Generated in `generated/mocks/`

### Test Coverage Requirements
- Domain logic: >80% coverage expected
- Use cases: Test all paths including errors
- Adapters: Integration tests for external services
- Commands: Test flag parsing and command execution

### Mock Generation

```bash
# Generate mocks for interfaces
mockery --name=InterfaceName --dir=internal/domain/ports --output=generated/mocks
```

## Security Considerations

1. **Never commit tokens** - Always use environment variables
2. **Validate all inputs** - Especially in configuration loading
3. **Minimal permissions** - Request only necessary API scopes
4. **Secure defaults** - HTTPS by default, validate certificates
5. **Audit logging** - Log security-relevant operations

## Performance Guidelines

1. **Concurrent operations** - Use goroutines for parallel repo processing
2. **Batch API calls** - Minimize API requests to avoid rate limits
3. **Efficient cloning** - Support shallow clones when appropriate
4. **Memory management** - Stream large files, don't load fully into memory
5. **Caching** - Cache provider responses when safe

## Documentation Standards

- **Package docs**: Every package needs a doc.go or package comment
- **Exported items**: Document all exported types, functions, constants
- **Examples**: Provide examples for complex functionality
- **Configuration**: Document all config options with defaults
- **Architecture**: Keep architecture.adoc updated with significant changes

## Common Pitfalls to Avoid

**Remember the opposite: Be hexagonal, functional, idiomatic, and don't overengineer.**

1. **Don't store config in context** - Pass as explicit parameters (be functional)
2. **Don't use global variables** - Use dependency injection (be hexagonal)
3. **Don't ignore errors** - Always handle or explicitly document why ignored (be idiomatic)
4. **Don't mix concerns** - Keep domain logic separate from infrastructure (be hexagonal)
5. **Don't break interfaces** - Extend rather than modify when possible (don't overengineer)
6. **Don't create unnecessary abstractions** - Keep it simple (don't overengineer)
7. **Don't violate hexagonal boundaries** - Domain shouldn't know about adapters

## Helpful Context

- Main entry: `cmd/gitprovidersync/main.go`
- Config examples: `examples/gitprovidersync-complete-example.yaml`
- Architecture docs: `docs/architecture.adoc`
- Testing docs: `docs/testing.md`
- CI/CD examples: `docs/ci-examples.md`

## Final Reminder

**Always: Be hexagonal, functional, idiomatic, and don't overengineer.**

These four principles guide every decision in this codebase. When in doubt, choose the simpler, cleaner solution that maintains clear boundaries and follows Go conventions.
