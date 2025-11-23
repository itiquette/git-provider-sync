#!/usr/bin/env just --justfile

# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Tool versions
reuse_version := "5.0.2-debian"
conform_version := "v0.1.0-alpha.30"

# Dynamic version information
version := `git describe --tags --dirty --always --abbrev=12`                          # Comprehensive version: tag-commits-hash-dirty
commit := `git rev-parse HEAD`                                                          # Current git commit hash
build_date := `date -u +'%Y-%m-%dT%H:%M:%SZ'`                                          # ISO 8601 UTC build timestamp

# Build paths
bin := "./bin"                           # Build output directory
dist := "./dist"                         # Release output directory (goreleaser)
executable := "gitprovidersync"          # Binary name

# Environment variables
compare_to_branch := env_var_or_default("COMPARETOBRANCH", "main")         # Default branch for commit comparison

# ==================================================================================== #
# DEFAULT - Show available recipes
# ==================================================================================== #

# Display available recipes
default:
  @printf "\033[1;36m Just Recipes\033[0m\n"
  @printf "\n"
  @printf "Quick start: \033[1;32mjust dev\033[0m | \033[1;34mjust test\033[0m | \033[1;35mjust lint\033[0m\n"
  @printf "\n"
  @just --list --unsorted

# ==================================================================================== #
# HELPER FUNCTIONS
# ==================================================================================== #

# Print cyan header with optional command in dim
_header text cmd="":
    #!/usr/bin/env bash
    if [[ -n "{{cmd}}" ]]; then
        printf "\033[1;36m{{text}}\033[0m\n"
        printf " \033[2m{{cmd}}\033[0m\n"
    else
        printf "\033[1;36m{{text}}\033[0m\n"
    fi

# Run command with output capture - shows command and output only on failure
_run_with_output cmd desc:
    #!/usr/bin/env bash
    set -euo pipefail
    printf " \033[2m%s\033[0m\n" "{{cmd}}"
    output=$(eval "{{cmd}}" 2>&1) || {
        code=$?
        printf "\033[1;31m✗\033[0m %s failed\n" "{{desc}}"
        printf "%s\n" "$output"
        exit $code
    }
    printf "\033[0;32m✓\033[0m %s completed\n" "{{desc}}"

# ==================================================================================== #
# DEVELOPMENT - Development workflow
# ==================================================================================== #

# ▪ Primary development workflow - verify and build host architecture binary
[group('development')]
dev: verify build-host

# Quality assurance pipeline - clean, fix, lint, and test
[group('development')]
verify: clean-build lint-fix lint test

# ==================================================================================== #
# TEST - Testing and coverage
# ==================================================================================== #

# ▪ Execute all tests - unit and integration tests
[group('test')]
test: test-unit test-integration

# Execute unit tests only
[group('test')]
test-unit:
    @just _header "Run unit tests" "go test -count=1 -race -buildvcs=false \$(go list './...' | grep -v generated)"
    go test -count=1 -race -buildvcs=false `go list './...' | grep -v generated`

# Execute integration tests only
[group('test')]
test-integration:
    @just _header "Run integration tests" "go test -tags=integration -count=1 -race -buildvcs=false ./internal/integrationtest/"
    go test -tags=integration -count=1 -race -buildvcs=false ./internal/integrationtest/



# Generate test coverage - HTML report output (unit + integration tests)
[group('test')]
test-coverage: clean-build
    @just _header "Run all tests with coverage" "go test -coverprofile"
    go test -v -count=1 -race -buildvcs=false -coverprofile={{bin}}/coverage-unit.out `go list './...' | grep -v generated`
    go test -v -tags=integration -count=1 -race -buildvcs=false -coverprofile={{bin}}/coverage-integration.out ./internal/integrationtest/
    @just _header "Merge coverage profiles"
    @just _run_with_output "go run github.com/wadey/gocovmerge {{bin}}/coverage-unit.out {{bin}}/coverage-integration.out > {{bin}}/coverage.out" "Coverage merge"
    @just _header "Generate HTML report"
    @just _run_with_output "go tool cover -html {{bin}}/coverage.out -o {{bin}}/coverage.html" "Coverage report generation"

# ==================================================================================== #
# BUILD - Compilation and packaging
# ==================================================================================== #

# ▪ Build pipeline - verify, binary, and container
[group('build')]
build: verify build-all build-image

# Compile multi-architecture binaries - fast compilation mode
[group('build')]
build-all: clean-build
    @just _header "Multi-arch binaries build"
    @just _run_with_output "just _build_binary linux amd64" "AMD64 binary built"
    @just _run_with_output "just _build_binary linux arm64" "ARM64 binary built"

# Build AMD64 binary - linux amd64 architecture
[group('build')]
build-amd64: clean-build
    @just _header "AMD64 binary build"
    @just _run_with_output "just _build_binary linux amd64" "AMD64 binary built successfully"

# Build ARM64 binary - linux arm64 architecture
[group('build')]
build-arm64: clean-build
    @just _header "ARM64 binary build"
    @just _run_with_output "just _build_binary linux arm64" "ARM64 binary built successfully"

# Build host architecture binary - automatic architecture detection
[group('build')]
build-host: clean-build
    #!/usr/bin/env bash
    set -euo pipefail
    HOST_ARCH=$(uname -m)
    case "$HOST_ARCH" in
        x86_64)
            GOARCH="amd64"
            ;;
        aarch64|arm64)
            GOARCH="arm64"
            ;;
        *)
            printf "\033[1;33m! Unsupported host architecture: %s\033[0m\n" "$HOST_ARCH"
            exit 1
            ;;
    esac
    just _header "Host binary build" "just _build_binary linux $GOARCH"
    if just _build_binary linux "$GOARCH" >/dev/null 2>&1; then
        printf "\033[0;32m✓\033[0m Host binary built successfully (%s)\n" "$GOARCH"
    else
        printf "\033[1;31m✗\033[0m Host binary build failed (%s)\n" "$GOARCH"
        just _build_binary linux "$GOARCH"
        exit 1
    fi

# Build dev container image - multi-architecture support
[group('build')]
build-image: build-all
    #!/usr/bin/env bash
    set -euo pipefail
    just _header "Multi-arch container" "podman buildx build"
    
    # Create manifest list
    podman manifest rm git-provider-sync:dev 2>/dev/null || true
    podman rmi git-provider-sync:dev 2>/dev/null || true
    podman manifest create git-provider-sync:dev
    
    # Build and add AMD64 image to manifest
    printf "Building AMD64 image...\n"
    podman buildx build --platform=linux/amd64 --manifest=git-provider-sync:dev --build-arg DIRPATH={{bin}}/ -f Containerfile .
    
    # Build and add ARM64 image to manifest  
    printf "Building ARM64 image...\n"
    podman buildx build --platform=linux/arm64 --manifest=git-provider-sync:dev --build-arg DIRPATH={{bin}}/ -f Containerfile .
    
    printf "\033[0;32m✓ Multi-arch container manifest created\033[0m\n"


# ==================================================================================== #
# LINT - Quality assurance and code formatting
# ==================================================================================== #

# ▪ Execute linting - all file types
# Note: lint-commit temporarily disabled - will be re-enabled later
[group('lint')]
lint: lint-go lint-shell lint-md lint-yaml lint-actions lint-containers lint-license lint-secrets
    @printf "\033[0;32m✓ All linting checks completed\033[0m\n"

# Lint Go source code - static analysis and verification
[group('lint')]
lint-go:
    @just _header "Verify module checksums"
    @just _run_with_output "go mod verify" "Go mod verify"
    @just _header "Static analysis"
    @just _run_with_output "go vet ./..." "Go vet"
    @just _header "Advanced static analysis"
    @just _run_with_output "staticcheck -checks=all,-ST1000,-U1000,-SA1019 ./..." "Staticcheck"
    @just _header "Vulnerability scanning"
    @just _run_with_output "govulncheck ./..." "Govulncheck"
    @just _header "Multi-linter runner"
    @just _run_with_output "golangci-lint run" "Golangci-lint"
    @just _header "Whitespace linter"
    @just _run_with_output "wsl ./..." "WSL"



# Lint GitHub Actions - workflow syntax validation
[group('lint')]
lint-actions:
    @just _header "GitHub Actions workflow linter"
    @just _run_with_output "actionlint .github/workflows/*.yml" "Actionlint"


# Validate commit messages - branch comparison check
# Temporarily disabled - will be re-enabled later
# [group('lint')]
# lint-commit:
#     #!/usr/bin/env bash
#     set -euo pipefail
#     just _header "Commit message validation" "conform"
#     if [[ $(git rev-list --count {{compare_to_branch}}..) -gt 0 ]]; then
#         just _run_with_output "podman run --rm -i --volume $(pwd):/repo -w /repo ghcr.io/siderolabs/conform:{{conform_version}} enforce --base-branch={{compare_to_branch}}" "Conform completed"
#     else
#         printf "\033[1;33m! No new commits found in branch compared to %s, skipping commit lint\033[0m\n" "{{compare_to_branch}}"
#     fi

# Lint container definitions - Containerfile best practices
[group('lint')]
lint-containers:
    @just _header "Container file linter"
    @just _run_with_output "hadolint Containerfile" "Hadolint"



# Verify license compliance - REUSE specification check
[group('lint')]
lint-license:
    @just _header "License compliance check"
    @just _run_with_output "podman run --rm --volume $(pwd):/data docker.io/fsfe/reuse:{{reuse_version}} lint" "Reuse"


# Lint markdown files - style and format validation
[group('lint')]
lint-md:
    @just _header "Markdown linter"
    @just _run_with_output "rumdl check --config development/rumdl.toml --quiet ." "Rumdl"


# Scan for secrets - repository-wide credential detection
[group('lint')]
lint-secrets:
    @just _header "Secrets scanner"
    @just _run_with_output "gitleaks git --no-banner --log-level error ." "Gitleaks"

# Lint shell scripts - syntax and style validation
[group('lint')]
lint-shell:
    @just _header "Shell script linter"
    @just _run_with_output "find scripts -name '*.sh' -type f | xargs shellcheck -e SC2059" "Shellcheck"
    @just _header "Shell script formatter"
    @just _run_with_output "find scripts -name '*.sh' -type f | xargs shfmt -i 2 -d" "Shfmt"


# Lint YAML files - format and syntax check
[group('lint')]
lint-yaml:
    @just _header "YAML formatter/linter"
    @just _run_with_output "yamlfmt -lint ." "Yamlfmt"



# ==================================================================================== #
# LINT-FIX - Auto-fix linting violations
# ==================================================================================== #
# ▪ Auto-repair code violations - all supported formats
[group('lint-fix')]
lint-fix: lint-fix-go lint-fix-shell lint-fix-md lint-fix-yaml
    @printf "\033[0;32m✓ All auto-fixes completed\033[0m\n"

# Format and organize code - Go source cleanup
[group('lint-fix')]
tidy: clean-build
    @just _header "Format Go code"
    @just _run_with_output "go fmt ./..." "Go fmt"
    @just _header "Clean dependencies verbose"
    @just _run_with_output "go mod tidy -v" "Go mod tidy"

# Auto-repair Go violations - formatting and dependencies
[group('lint-fix')]
lint-fix-go:
    @just _header "Clean dependencies"
    @just _run_with_output "go mod tidy" "Go mod tidy"
    @just _header "Format Go code"
    @just _run_with_output "go fmt ./..." "Go fmt"
    @just _header "Auto-fix Go issues"
    @just _run_with_output "golangci-lint run --fix" "Golangci-lint fix"
    @just _header "Fix whitespace issues"
    @just _run_with_output "wsl --fix ./..." "WSL fix"

# Auto-repair markdown issues - format standardization
[group('lint-fix')]
lint-fix-md:
    @just _header "Auto-fix markdown issues"
    @just _run_with_output "rumdl check --config development/rumdl.toml --fix --quiet ." "Rumdl fix"


# Auto-repair shell scripts - syntax and formatting
[group('lint-fix')]
lint-fix-shell:
    @just _header "Auto-fix shell issues"
    @just _run_with_output "./scripts/maintenance/fix-shell-scripts.sh" "Shell fix"

# Auto-repair YAML files - format normalization
[group('lint-fix')]
lint-fix-yaml:
    @just _header "Format YAML files"
    @just _run_with_output "yamlfmt ." "Yamlfmt fix"

# ==================================================================================== #
# SECURITY - Vulnerability and secret scanning
# ==================================================================================== #
# ▪ Execute security scanning - vulnerability analysis
[group('security')]
security: build-all
    @just _header "Vulnerability analysis" "./scripts/security/security-audit.sh {{bin}} {{executable}}"
    ./scripts/security/security-audit.sh {{bin}} {{executable}}

# Scan container vulnerabilities - dockle and trivy analysis
[group('security')]
containerimage-vuln-scan: build-all
    @just _header "Dockle and trivy analysis" "./scripts/security/container-security-scan.sh {{bin}} {{executable}}"
    ./scripts/security/container-security-scan.sh {{bin}} {{executable}}

# Execute OSSF scorecard - GitHub security assessment
[group('security')]
ossf-scorecard-check:
    @just _header "Security best practices scan" "./scripts/security/ossf-scorecard.sh github.com/itiquette/git-provider-sync"
    ./scripts/security/ossf-scorecard.sh github.com/itiquette/git-provider-sync


# ==================================================================================== #
# RELEASE - Publishing and distribution
# ==================================================================================== #
# ▪ Validate release configuration - dry run with snapshot
[group('release')]
release-dry: clean-build
    @just _header "Validate config" "goreleaser check"
    goreleaser check
    @just _header "Cross-platform build" "goreleaser release --clean --snapshot"
    goreleaser release --clean --snapshot

# Execute production release - build and publish packages
[group('release')]
release: clean-build
    @just _header "Validate config" "goreleaser check"
    goreleaser check
    @just _header "Cross-platform build & publish" "goreleaser release --clean"
    goreleaser release --clean

# Test Docker image builds only - no signing or publishing
[group('release')]
release-docker: clean-build
    @just _header "Docker images only" "goreleaser release --snapshot --clean --skip sign,publish,announce,validate,sbom,archive,nfpm"
    goreleaser release --snapshot --clean --skip sign,publish,announce,validate,sbom,archive,nfpm

# Build packages only for testing signing
[group('release')]
release-packages: clean-build
    @just _header "Test signing packages" "goreleaser release --snapshot --clean --skip publish,announce,validate"
    goreleaser release --snapshot --clean --skip publish,announce,validate

# Build packages only for testing - skip sigstore signing
[group('release')]
release-packages-no-sigstore: clean-build
    @just _header "Test packages without sigstore" "goreleaser release --snapshot --clean --skip publish,announce,validate,sign"
    goreleaser release --snapshot --clean --skip publish,announce,validate,sign

# ==================================================================================== #
# DOCUMENTATION - Documentation and code generation
# ==================================================================================== #
# ▪ Generate project documentation - man pages and completions
[group('documentation')]
generate: generate-manpage generate-completion

# Generate manual pages - compressed Unix format
[group('documentation')]
generate-manpage:
    @just _header "Generate man page" "./scripts/docs/manpage.sh {{executable}}"
    ./scripts/docs/manpage.sh {{executable}} docs/{{executable}}.1.md generated/manpages

# Generate shell completions - bash, zsh, and fish
[group('documentation')]
generate-completion:
    @just _header "Generate shell completions" "./scripts/docs/completions.sh {{executable}}"
    ./scripts/docs/completions.sh {{executable}} generated/completions
    @printf "\n\033[1;36mJust install commands:\033[0m\n"
    @printf "  \033[1;32mjust install-local-completion\033[0m - Install shell completions to user directories\n"
    @printf "  \033[1;32mjust install-local-man\033[0m        - Install manual page\n"
    @printf "  \033[1;32mjust install-local\033[0m            - Install everything (binary + man + completions)\n"

# Generate mocks
[group('documentation')]
mock:
    @just _header "Generate test mocks" "./scripts/tools/generatemock.sh"
    ./scripts/tools/generatemock.sh
# ==================================================================================== #
# DEV-HELPERS - Project installation helpers
# ==================================================================================== #

# ▪ Install dev build to local environment - binary, man, and completions
[group('dev-helpers')]
install-local: build-all
    @just _header "Install all components locally" "./scripts/install/install-local-all.sh {{bin}} {{executable}}"
    ./scripts/install/install-local-all.sh {{bin}} {{executable}}

# Install binary locally - architecture detection and user directory
[group('dev-helpers')]
install-local-binary: build-all
    @just _header "Install binary locally" "./scripts/install/install-binary.sh {{bin}} {{executable}}"
    ./scripts/install/install-binary.sh {{bin}} {{executable}}

# Install manual page - user documentation directory
[group('dev-helpers')]
install-local-man: generate-manpage
    @just _header "Local man page installation" "cp to ~/.local/share/man/man1/"
    # Check if ~/.local/share/man/man1 directory exists
    if [ ! -d ~/.local/share/man/man1 ]; then \
        printf "\033[1;31m✗ Directory ~/.local/share/man/man1 does not exist\033[0m\n"; \
        printf "  Please create it first: mkdir -p ~/.local/share/man/man1\n"; \
        printf "  Or ensure your system follows XDG Base Directory specification\n"; \
        exit 1; \
    fi
    cp generated/manpages/{{executable}}.1.gz ~/.local/share/man/man1/
    printf "\033[0;32m✓ Installed man page to ~/.local/share/man/man1/\033[0m\n"
    printf "  Run 'man {{executable}}' to view\n"

# Install shell completions - user completion directories
[group('dev-helpers')]
install-local-completion: generate-completion
    @just _header "Install shell completions" "./scripts/install/install-completions.sh {{executable}}"
    ./scripts/install/install-completions.sh {{executable}}
    @printf "\n\033[1;36mShell Completion Installation\033[0m\n"
    @printf "\n"
    @printf "\033[1;33mRuntime completion (recommended):\033[0m\n"
    @printf "  Users can generate completions on-demand:\n"
    @printf "  \033[1;32m{{executable}} completion bash\033[0m   # Generate bash completion\n"
    @printf "  \033[1;32m{{executable}} completion zsh\033[0m    # Generate zsh completion\n"
    @printf "  \033[1;32m{{executable}} completion fish\033[0m   # Generate fish completion\n"
    @printf "\n"
    @printf "\033[1;33mPre-generated files (for packaging):\033[0m\n"
    @printf "\n"
    @printf "\033[1;34mBash:\033[0m\n"
    @printf "  System-wide: \033[1;32msudo cp generated/completions/{{executable}}.bash /usr/share/bash-completion/completions/{{executable}}\033[0m\n"
    @printf "  User-local:  \033[1;32mcp generated/completions/{{executable}}.bash ~/.local/share/bash-completion/completions/{{executable}}\033[0m\n"
    @printf "\n"
    @printf "\033[1;34mZsh:\033[0m\n"
    @printf "  System-wide: \033[1;32msudo cp generated/completions/{{executable}}.zsh /usr/local/share/zsh/site-functions/_{{executable}}\033[0m\n"
    @printf "  User-local:  \033[1;32mmkdir -p ~/.local/share/zsh/site-functions && cp generated/completions/{{executable}}.zsh ~/.local/share/zsh/site-functions/_{{executable}}\033[0m\n"
    @printf "\n"
    @printf "\033[1;34mFish:\033[0m\n"
    @printf "  System-wide: \033[1;32msudo cp generated/completions/{{executable}}.fish /usr/share/fish/vendor_completions.d/\033[0m\n"
    @printf "  User-local:  \033[1;32mcp generated/completions/{{executable}}.fish ~/.config/fish/completions/\033[0m\n"
    @printf "\n\033[1;36mJust install equivalents:\033[0m\n"
    @printf "  \033[1;32mjust install-local\033[0m        - Install all components (binary + man + completions)\n"
    @printf "  \033[1;32mjust install-local-binary\033[0m - Install binary only\n"
    @printf "  \033[1;32mjust install-local-man\033[0m    - Install manual page only\n"


# ==================================================================================== #
# DEV-SETUP - Development environment setup
# ==================================================================================== #

# Install Go development tools - Go toolchain components
[group('dev-setup')]
install-go-dev-tools:
    @just _header "Install Go development tools" "./scripts/install/install-go-dev-tools.sh"
    ./scripts/install/install-go-dev-tools.sh

# ==================================================================================== #
# DEPENDENCIES - Dependency management
# ==================================================================================== #

# ▪ Upgrade all dependencies - Go modules and development tools (not upgrade-go) 
[group('dependencies')]
upgrade: upgrade-deps upgrade-go-dev-tools

# Upgrade Go version - updates Go version across all project files
[group('dependencies')]
upgrade-go:
    @just _header "Upgrade Go to latest stable version" "./scripts/tools/upgrade-go-version.sh"
    ./scripts/tools/upgrade-go-version.sh

# Upgrade Go dependencies - Go dependencies update
[group('dependencies')]
upgrade-deps:
    @just _header "Upgrade all dependencies" "go get -u -t ./..."
    go get -u -t ./...
    @just _header "Clean dependencies" "go mod tidy"
    go mod tidy

# Upgrade Go development tools - upgrades tools listed in tools/go.mod 
[group('dependencies')]
upgrade-go-dev-tools:
    @just _header "Upgrade Go development tools" "./scripts/tools/upgrade-go-dev-tools.sh"
    ./scripts/tools/upgrade-go-dev-tools.sh

# List available project updates - dependency, tools, Go status
[group('dependencies')]
upgrade-list:
    @just _header "Check available updates" "./scripts/tools/list-updates.sh"
    ./scripts/tools/list-updates.sh

# ==================================================================================== #
# MAINTENANCE - Cache cleanup and utilities
# ==================================================================================== #

# ▪ Clean all artifacts - build outputs and caches
[group('maintenance')]
clean: clean-build clean-caches

# Clean build artifacts - remove compiled binaries
[group('maintenance')]
clean-build:
    @just _header "Remove compiled binaries"
    @bash -c 'set -euo pipefail; \
    for dir in "{{bin}}" "{{dist}}"; do \
        if [[ -L "$dir" ]]; then \
            echo "Error: $dir is a symlink, refusing to remove" >&2; \
            exit 1; \
        fi; \
        if [[ -e "$dir" ]] && [[ ! -d "$dir" ]]; then \
            echo "Error: $dir exists but is not a directory, refusing to remove" >&2; \
            exit 1; \
        fi; \
    done; \
    rm -rf "{{bin}}" "{{dist}}"; \
    mkdir -p "{{bin}}" "{{dist}}"; \
    go clean -cache; \
    echo "✓ Build artifacts clean completed"'

# Clean development caches - clear Go and Go linter caches
[group('maintenance')]
clean-caches:
    @just _header "Clear all Go caches"
    @just _run_with_output "go clean -cache -modcache -testcache -fuzzcache" "Go cache clean"
    @just _header "Clear linter cache"
    @just _run_with_output "golangci-lint cache clean" "Golangci-lint cache clean"

# ==================================================================================== #
# HELPERS - Internal recipes
# ==================================================================================== #

# Build binary for target - internal cross-compilation helper
_build_binary goos goarch:
    @./scripts/build/build-cross-platform.sh {{goos}} {{goarch}} {{bin}} {{executable}}
