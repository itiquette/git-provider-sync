#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Updates Go version to latest stable across all project files
# Usage: ./scripts/tools/upgrade-go-version.sh
# Dependencies: curl, jq, sed
# Output: Updated Go version in all relevant files

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
# readonly BLUE=$'\033[0;34m' # Unused
readonly CYAN=$'\033[0;36m'
readonly NC=$'\033[0m' # No Color

# Helper functions
log() { printf "%s\n" "$1"; }
success() { printf "${GREEN}✓${NC} %s\n" "$1"; }
fail() { printf "${RED}✗${NC} %s\n" "$1" >&2; }
info() { printf "${CYAN}→${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}!${NC} %s\n" "$1"; }

# Get the latest stable Go version from official API
get_latest_go_version() {
  local version
  version=$(curl -sL "https://go.dev/dl/?mode=json" | jq -r '.[0].version' | sed 's/^go//')

  if [[ -z "$version" ]]; then
    fail "Failed to fetch latest Go version"
    exit 1
  fi

  echo "$version"
}

# Get current Go version from go.mod
get_current_go_version() {
  if [[ ! -f "go.mod" ]]; then
    fail "go.mod not found"
    exit 1
  fi

  grep "^go " go.mod | awk '{print $2}'
}

# Update Go version in a file
update_file() {
  local file=$1
  # shellcheck disable=SC2034 # Used for documentation
  local old_version=$2
  # shellcheck disable=SC2034 # Used for documentation
  local new_version=$3
  local pattern=$4
  # shellcheck disable=SC2034 # Used for documentation
  local replacement=$5

  if [[ ! -f "$file" ]]; then
    warn "File not found: $file (skipping)"
    return
  fi

  # Use sed with the provided pattern and replacement
  if sed -i.bak "$pattern" "$file"; then
    rm -f "${file}.bak"
    success "Updated $file"
  else
    fail "Failed to update $file"
  fi
}

main() {
  info "Checking for latest stable Go version..."

  # Get versions
  local current_version
  current_version=$(get_current_go_version)

  local latest_version
  latest_version=$(get_latest_go_version)

  # Extract major.minor for go.mod (Go uses major.minor in go.mod)
  local latest_major_minor
  latest_major_minor=$(echo "$latest_version" | cut -d. -f1,2)

  log ""
  info "Current Go version: ${current_version}"
  info "Latest Go version: ${latest_version} (go.mod will use ${latest_major_minor})"

  # Check if update is needed (handle both 1.25 and 1.25.0 formats)
  local current_major_minor
  current_major_minor=$(echo "$current_version" | cut -d. -f1,2)

  if [[ "$current_major_minor" == "$latest_major_minor" ]]; then
    success "Already using the latest Go version"
    exit 0
  fi

  log ""
  info "Updating Go version from ${current_version} to ${latest_major_minor}..."
  log ""

  # Update go.mod files
  info "Updating go.mod files..."
  update_file "go.mod" "$current_version" "$latest_major_minor" \
    "s/^go ${current_version}/go ${latest_major_minor}/" \
    "go ${latest_major_minor}"

  update_file "tools/go.mod" "$current_version" "$latest_major_minor" \
    "s/^go ${current_version}/go ${latest_major_minor}/" \
    "go ${latest_major_minor}"

  # Update GitHub workflow files
  info "Updating GitHub workflow files..."
  for workflow in .github/workflows/*.yml .github/workflows/*.yaml; do
    if [[ -f "$workflow" ]]; then
      # Update go-version fields (handles both quoted and unquoted)
      sed -i.bak "s/go-version: ['\"]\\?${current_version}['\"]\\?/go-version: '${latest_major_minor}'/" "$workflow"
      sed -i.bak "s/go-version: ['\"]\\?${latest_version}['\"]\\?/go-version: '${latest_version}'/" "$workflow"

      # Update Go setup actions that use full version
      sed -i.bak "s/go-version: ['\"]\\?[0-9]\\+\\.[0-9]\\+\\.[0-9]\\+['\"]\\?/go-version: '${latest_version}'/" "$workflow"

      # Update matrix versions
      sed -i.bak "s/\\[${current_version}\\]/[${latest_major_minor}]/" "$workflow"
      sed -i.bak "s/'${current_version}'/'${latest_major_minor}'/" "$workflow"
      sed -i.bak "s/\"${current_version}\"/\"${latest_major_minor}\"/" "$workflow"

      rm -f "${workflow}.bak"
      success "Updated $(basename "$workflow")"
    fi
  done

  # Update documentation files
  info "Updating documentation files..."

  # Update version references in docs
  for doc in README.md docs/*.md CONTRIBUTING.md; do
    if [[ -f "$doc" ]]; then
      # Update Go version mentions (e.g., "Go 1.23" -> "Go 1.24")
      sed -i.bak "s/Go ${current_version}/Go ${latest_major_minor}/g" "$doc"
      sed -i.bak "s/go${current_version}/go${latest_major_minor}/g" "$doc"
      sed -i.bak "s/Go ${current_version%.*}/Go ${latest_major_minor%.*}/g" "$doc"

      # Update version in code blocks
      sed -i.bak "s/^go ${current_version}/go ${latest_major_minor}/g" "$doc"

      if diff -q "$doc" "${doc}.bak" >/dev/null; then
        rm -f "${doc}.bak"
        # File unchanged, skip message
      else
        rm -f "${doc}.bak"
        success "Updated $(basename "$doc")"
      fi
    fi
  done

  # Update Dockerfile/Containerfile
  info "Updating container files..."
  for containerfile in Dockerfile Containerfile; do
    if [[ -f "$containerfile" ]]; then
      # Update FROM golang:version lines
      sed -i.bak "s/FROM golang:${current_version}/FROM golang:${latest_version}/" "$containerfile"
      sed -i.bak "s/FROM golang:${latest_major_minor}/FROM golang:${latest_version}/" "$containerfile"

      # Update any Go version ARGs
      sed -i.bak "s/ARG GO_VERSION=${current_version}/ARG GO_VERSION=${latest_version}/" "$containerfile"
      sed -i.bak "s/ARG GO_VERSION=\"${current_version}\"/ARG GO_VERSION=\"${latest_version}\"/" "$containerfile"

      rm -f "${containerfile}.bak"
      success "Updated $containerfile"
    fi
  done

  # Update .tool-versions if it exists (asdf)
  if [[ -f ".tool-versions" ]]; then
    info "Updating .tool-versions..."
    sed -i.bak "s/golang ${current_version}/golang ${latest_version}/" ".tool-versions"
    sed -i.bak "s/golang [0-9]\\+\\.[0-9]\\+\\.[0-9]\\+/golang ${latest_version}/" ".tool-versions"
    rm -f ".tool-versions.bak"
    success "Updated .tool-versions"
  fi

  # Run go mod tidy on both modules
  info "Running go mod tidy..."
  go mod tidy
  success "Main module tidied"

  if [[ -d "tools" ]]; then
    (cd tools && go mod tidy)
    success "Tools module tidied"
  fi

  log ""
  success "Go version updated successfully!"
  info "Updated from ${current_version} to ${latest_major_minor} (latest: ${latest_version})"
  log ""
  warn "Please review the changes and test the build before committing"
  info "Recommended next steps:"
  log "  1. Run 'just test' to verify tests pass"
  log "  2. Run 'just lint' to check for any issues"
  log "  3. Commit the changes with: git commit -m \"build: upgrade Go to ${latest_major_minor}\""
}

main "$@"
