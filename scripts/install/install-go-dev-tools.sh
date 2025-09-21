#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Project-agnostic Go development tools installer
# Usage: ./scripts/install-go-dev-tools.sh [tools_dir]
# Default: tools_dir=tools

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly CYAN=$'\033[1;36m'
readonly NC=$'\033[0m' # No Color

# Parse arguments with defaults
readonly TOOLS_DIR="${1:-${TOOLS_DIR:-tools}}"

log() {
  printf "${YELLOW}▸${NC} %s\n" "$1"
}

success() {
  printf "${GREEN}✓${NC} %s\n" "$1"
}

fail() {
  printf "${RED}✗${NC} %s\n" "$1" >&2
  exit 1
}

# Efficient command finder
find_command() {
  local cmd="$1"
  shift
  local search_paths=("$@" "$cmd")

  for path in "${search_paths[@]}"; do
    if command -v "$path" >/dev/null 2>&1; then
      echo "$path"
      return 0
    fi
  done
  return 1
}

check_go() {
  local go_cmd
  go_cmd=$(find_command go "$HOME/go/bin/go" "/usr/local/go/bin/go") || {
    fail "Go not found. Please install Go 1.25 or later"
  }

  local go_version
  go_version=$("$go_cmd" env GOVERSION | cut -c3-)
  success "Go $go_version found at $go_cmd" >&2
  echo "$go_cmd"
}

install_go_tools() {
  local go_cmd="$1"

  if [[ ! -d "$TOOLS_DIR" ]] || [[ ! -f "$TOOLS_DIR/go.mod" ]]; then
    fail "Tools module not found at $TOOLS_DIR/go.mod"
  fi

  log "Installing Go development tools from $TOOLS_DIR"

  # Install tools from tools module
  (
    cd "$TOOLS_DIR"

    # Extract tools from tool directive (Go 1.25+ format)
    local tools
    tools=$(awk '/^tool \(/{flag=1; next} /^\)/{flag=0} flag && /^\s+[a-zA-Z]/' go.mod 2>/dev/null | sed 's/^\s*//' | sed 's/\s*$//' || true)

    if [[ -z "$tools" ]]; then
      log "No tools found in $TOOLS_DIR/go.mod"
      return 0
    fi

    # Install each tool
    echo "$tools" | while IFS= read -r tool; do
      if [[ -n "$tool" ]]; then
        log "Installing $tool"
        "$go_cmd" install -v "$tool" || log "Failed to install $tool (continuing)"
      fi
    done
  )

  success "Go tools installation completed"
}

check_external_tools() {
  log "Checking for recommended external tools..."

  local external_tools=(
    "goreleaser" "cosign" "scorecard" "syft"
    "actionlint" "gitleaks" "shellcheck" "shfmt"
    "yamlfmt" "hadolint" "dockle" "trivy" "rumdl"
  )

  local missing=()
  for tool in "${external_tools[@]}"; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      missing+=("$tool")
    fi
  done

  if ((${#missing[@]} > 0)); then
    printf "${YELLOW}Missing external tools:${NC} ${missing[*]}\n"
    printf "These tools need to be installed separately via your package manager\n"
  else
    success "All recommended external tools found"
  fi
}

main() {
  printf "${CYAN}→${NC} Setting up Go development tools\n\n"

  local go_cmd
  go_cmd=$(check_go)

  install_go_tools "$go_cmd"
  check_external_tools

  printf "\n"
  success "Development tools setup completed"
}

main "$@"
