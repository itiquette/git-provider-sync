# GITPROVIDERSYNC 1 "" "" "User Commands"

## NAME

gitprovidersync - synchronize Git repositories between providers

## SYNOPSIS

**gitprovidersync** [*OPTION*...] *COMMAND* [*args*...]
**gitprovidersync** [*OPTION*...] **sync** [*OPTION*...]
**gitprovidersync** [*OPTION*...] **print** [*OPTION*...]
**gitprovidersync** [*OPTION*...] **man** [*OPTION*...]

## DESCRIPTION

**gitprovidersync** synchronizes Git repositories between different Git providers (GitHub, GitLab, Gitea) and local storage. It can mirror repositories to other providers, archive them as compressed files, or save them to directories.

Configuration is specified in YAML format. Use **--config-file** to specify the configuration file path.

Commands available: **sync** (primary operation), **print** (display configuration), and **man** (generate manual pages).

## OPTIONS

**--help**, **-h**
: Show help message and exit.

**--version**
: Output version information and exit.

**--config-file** *FILE*
: Path to the configuration file (default: gitprovidersync.yaml).

**--log-level** *LEVEL*
: Set logging level: trace, debug, info, warn, error, fatal (default: info).

**--log-format** *FORMAT*
: Set log output format: text, json (default: text).

**--dry-run**
: Show what would be synced without making changes.

**--force-push**
: Force push to target repositories.

**--quiet**
: Suppress all output except errors.

**--verbosity** *LEVEL*
: Set verbosity level: trace, debug, info, warn, error, fatal.

**--format** *FORMAT*
: Select output format: console, json, plain (default: console).

**--sanitize-names**
: Clean repository names to alphanumeric + hyphens only.

**--since** *DATE*
: Only sync repositories active since this date/time.

**--skip-invalid**
: Ignore repositories with invalid names.

## EXIT STATUS

**0**
: Synchronization completed successfully, all repositories in sync.

**1**
: Synchronization failed or conflicts detected.

**2**
: Configuration or system error (invalid arguments, network issues, etc.).

**130**
: Interrupted by signal (SIGINT, SIGTERM, SIGQUIT).

## ENVIRONMENT

**GITPROVIDERSYNC_CONFIG**
: Configuration file path (overrides default locations).

**GITPROVIDERSYNC_LOG_LEVEL**
: Default log level (overridden by --log-level).

**NO_COLOR**
: Disable colored output when set to any value.

**GITHUB_TOKEN**
: GitHub API token for authentication.

**GITLAB_TOKEN**
: GitLab API token for authentication.

**GITEA_TOKEN**
: Gitea API token for authentication.

## FILES

**gitprovidersync.yaml**
: Default configuration file in current directory.

**.gitprovidersync.yaml**
: Alternative configuration file name.

**~/.config/gitprovidersync/config.yaml**
: User configuration file.

**/etc/gitprovidersync/config.yaml**
: System-wide configuration file.

## EXAMPLES

Synchronize repositories using default configuration:

    gitprovidersync sync

Perform a dry-run to see what would be synchronized:

    gitprovidersync sync --dry-run

Force synchronization with a specific configuration file:

    gitprovidersync sync --config-file=my-sync.yaml --force-push

Show effective configuration:

    gitprovidersync print

Display configuration with connectivity check:

    gitprovidersync print --connectivity-check

Show configuration without executing:

    gitprovidersync print

Synchronize with verbose output and custom log level:

    gitprovidersync sync --verbosity=debug --config-file=my-sync.yaml

## CONFIGURATION

**gitprovidersync** loads configuration from YAML files with a "gitprovidersync" root key. Configuration files are searched in the following order:

1. File specified by **--config-file** flag
2. **GITPROVIDERSYNC_CONFIG** environment variable
3. **gitprovidersync.yaml** in current directory
4. **.gitprovidersync.yaml** in current directory
5. **~/.config/gitprovidersync/config.yaml**
6. **/etc/gitprovidersync/config.yaml**

Example configuration:

    gitprovidersync:
      production:
        github-to-gitlab:
          provider_type: github
          domain: github.com
          owner: myorg
          owner_type: group
          auth:
            token: "${GITHUB_TOKEN}"
          repositories:
            include:
              - "repo1"
              - "repo2"
          mirrors:
            gitlab-target:
              provider_type: gitlab
              domain: gitlab.com
              owner: mygroup
              owner_type: group
              auth:
                token: "${GITLAB_TOKEN}"

Configuration sections:

- **providers**: Define Git provider connections
- **synchronizations**: Specify sync operations between providers
- **settings**: Global and per-sync configuration options
- **filters**: Include/exclude patterns for repositories and branches

## BUGS

### Reporting Bugs

Report bugs at: <https://github.com/itiquette/git-provider-sync/issues>

### Known Issues

Large repositories with many branches may experience slower synchronization times. Use filters to limit the scope of synchronization for better performance.

Network timeouts may occur with slow Git providers. Increase **--timeout** value for unreliable connections.

## SEE ALSO

**git**(1), **git-clone**(1), **git-push**(1), **git-remote**(1)

Full documentation: <https://github.com/itiquette/git-provider-sync>

## COPYRIGHT

Copyright © 2025 itiquette/git-provider-sync. Licensed under EUPL-1.2 <https://eupl.eu/1.2/en/>.

This is free software: you are free to change and redistribute it. There is NO WARRANTY, to the extent permitted by law.