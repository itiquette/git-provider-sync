<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: EUPL-1.2

-->

# Git Test Environment Utility

This package provides utilities for setting up isolated git environments for integration testing. It allows you to test real git operations without external dependencies or network calls.

## Features

- **Isolated Environment**: All repositories are created in temporary directories, completely isolated from the host system
- **Simulated Git Providers**: Creates bare repositories that simulate GitHub, GitLab, Gitea, etc.
- **Real Git Operations**: Test actual git operations like clone, push, pull, remote management
- **Configurable Setup**: Customize repository names, initial content, remotes, and more
- **Easy Cleanup**: Automatic cleanup of all temporary files and directories

## Quick Start

```go
func TestMyGitFeature(t *testing.T) {
    gitOps := gogit.New(ports.GitConfig{
        UserName:    "Test User",
        UserEmail:   "test@example.com",
        StorageMode: ports.StorageModeFilesystem, // Required for pushable bare repos in tests
    })

    // Setup isolated git environment
    env, err := testutil.SetupSimpleGitTestEnvironment(t, gitOps)
    require.NoError(t, err)
    // No manual cleanup needed - t.TempDir() handles it automatically

    ctx := context.Background()

    // Test git operations
    err = env.WorkingRepo.Repo.AddRemote(ctx, "upstream", env.GetTargetURL())
    require.NoError(t, err)

    // Test remote updates (e.g., GitHub → GitLab sync)
    err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
    require.NoError(t, err)

    // Verify operations
    remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
    require.NoError(t, err)
    // ... assertions
}
```

## Architecture

The utility creates three types of repositories:

1. **Source Bare Repository**: Simulates the source git provider (e.g., GitHub)

- Bare repository (no working directory)
- Acts as the "upstream" or "origin" for cloning

2. **Target Bare Repository**: Simulates the target git provider (e.g., GitLab)

- Bare repository for push operations
- Represents the destination for syncing

3. **Working Repository**: A normal git repository for testing operations

- Cloned from source or initialized with content
- Has remotes configured to point to bare repositories
- Used for actual git operations in tests

```text
tmpdir/
├── source-repo.git/     (bare - simulates GitHub)
├── target-repo.git/     (bare - simulates GitLab)
└── working-repo/        (working - for git operations)
    ├── .git/
    ├── README.md
    ├── src/
    └── ...
```

## Configuration Options

### GitTestOptions

```go
type GitTestOptions struct {
    // Repository names
    SourceRepoName  string
    TargetRepoName  string  
    WorkingRepoName string

    // Initial content
    InitialFiles map[string]string
    InitialBranch string

    // Remote configuration
    AddRemotes map[string]string
}
```

### Example with Custom Options

```go
opts := testutil.GitTestOptions{
    SourceRepoName:  "my-github-repo",
    TargetRepoName:  "my-gitlab-repo", 
    WorkingRepoName: "my-local-clone",
    InitialBranch:   "develop",
    InitialFiles: map[string]string{
        "package.json":    `{"name": "my-app", "version": "1.0.0"}`,
        "src/index.js":    `console.log("Hello World");`,
        "README.md":       "# My Application\n\nBuilt with Node.js",
        ".gitignore":      "node_modules/\n*.log\n",
        "docs/api.md":     "# API Documentation\n\nAPI endpoints...",
    },
    AddRemotes: map[string]string{
        "origin":   "", // Will use source bare repo URL
        "upstream": "https://github.com/upstream/repo.git",
        "fork":     "https://github.com/myuser/repo.git",
    },
}

env, err := testutil.SetupGitTestEnvironment(t, gitOps, opts)
```

## Common Test Scenarios

### Testing GitHub to GitLab Sync

```go
func TestGitHubToGitLabSync(t *testing.T) {
    env, err := testutil.SetupSimpleGitTestEnvironment(t, gitOps)
    require.NoError(t, err)
    // No manual cleanup needed - t.TempDir() handles it automatically

    ctx := context.Background()

    // 1. Verify initial state (origin points to GitHub)
    remotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
    require.NoError(t, err)
    assert.Equal(t, env.GetSourceURL(), remotes[0].URL)

    // 2. Set up backup remote
    err = env.WorkingRepo.Repo.AddRemote(ctx, "GPSUPSTREAM", env.GetSourceURL())
    require.NoError(t, err)

    // 3. Update origin to point to GitLab (THE CRITICAL FIX)
    err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
    require.NoError(t, err)

    // 4. Verify the fix
    updatedRemotes, err := env.WorkingRepo.Repo.ListRemotes(ctx)
    require.NoError(t, err)

    // Assert origin now points to GitLab, GPSUPSTREAM to GitHub
    // ...
}
```

### Testing Push Operations

```go
func TestPushOperations(t *testing.T) {
    env, err := testutil.SetupSimpleGitTestEnvironment(t, gitOps)
    require.NoError(t, err)
    // No manual cleanup needed - t.TempDir() handles it automatically

    ctx := context.Background()

    // Setup target remote
    err = env.WorkingRepo.Repo.UpdateRemote(ctx, "origin", env.GetTargetURL())
    require.NoError(t, err)

    // Test push options
    pushOptions := ports.PushOptions{
        Remote: "origin",
        Force:  false,
        Auth:   ports.AuthOptions{Type: ports.AuthTypeNone},
    }

    // With git binary support, you could do:
    // err = env.WorkingRepo.Repo.Push(ctx, pushOptions)
    // require.NoError(t, err)
}
```

### Testing Multi-Provider Scenarios

```go
func TestMultiProviderSync(t *testing.T) {
    env, err := testutil.SetupSimpleGitTestEnvironment(t, gitOps)
    require.NoError(t, err)
    // No manual cleanup needed - t.TempDir() handles it automatically

    // Add third provider
    giteaPath := env.TmpDir + "/gitea-repo.git"
    giteaRepo, err := env.GitOps.Init(ctx, giteaPath, ports.InitOptions{Bare: true})
    require.NoError(t, err)
    defer giteaRepo.Close()

    // Setup multi-provider remotes
    err = env.WorkingRepo.Repo.AddRemote(ctx, "gitlab", env.GetTargetURL())
    err = env.WorkingRepo.Repo.AddRemote(ctx, "gitea", giteaPath)

    // Test complex sync scenarios...
}
```

## Limitations

1. **Git Binary Operations**: Currently requires git binary adapter for add, commit, push operations
2. **Authentication**: Local file:// URLs don't require authentication, so auth testing needs mocks
3. **Network Simulation**: Doesn't simulate network latency, errors, etc.

## Integration with Existing Tests

Replace manual git setup in integration tests:

```go
// Before: Manual setup
tmpDir, _ := os.MkdirTemp("", "git-test-*")
defer os.RemoveAll(tmpDir)
sourcePath := filepath.Join(tmpDir, "source.git")
// ... lots of manual git setup code

// After: Use utility
env, err := testutil.SetupSimpleGitTestEnvironment(t, gitOps)
require.NoError(t, err)
// No manual cleanup needed - t.TempDir() handles it automatically
// Ready to test!
```

## Best Practices

1. **Automatic Cleanup**: Using `t.TempDir()` ensures automatic cleanup without manual intervention
2. **Parallel Tests**: Use `t.Parallel()` since each test gets isolated directories
3. **Descriptive Names**: Use custom repository names for complex scenarios
4. **Realistic Content**: Add realistic file structures for comprehensive testing
5. **Test Boundaries**: Focus on testing git operations, not business logic

## Examples

See `example_usage.go` for comprehensive usage examples and `git_test_env_test.go` for test cases demonstrating the utility in action.
