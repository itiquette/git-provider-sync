#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Project-agnostic shell completion installer
# Usage: ./scripts/install-completions.sh [binary_name] [completion_dir]
# Defaults: binary_name=$(basename "$(pwd)"), completion_dir=generated/completions

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly NC=$'\033[0m' # No Color

# Parse arguments with defaults
readonly BINARY_NAME="${1:-${BINARY_NAME:-$(basename "$(pwd)")}}"
readonly COMPLETION_DIR="${2:-${COMPLETION_DIR:-generated/completions}}"

log() {
  printf "${YELLOW}▸${NC} %s\n" "$1"
}

success() {
  printf "${GREEN}✓${NC} %s\n" "$1"
}

fail() {
  printf "${RED}✗${NC} %s\n" "$1" >&2
}

install_bash_completion() {
  local source_file="${COMPLETION_DIR}/${BINARY_NAME}.bash"
  if [[ ! -f "$source_file" ]]; then
    log "Bash completion not found: $source_file"
    return 1
  fi

  local dest_dir="$HOME/.local/share/bash-completion/completions"
  if [[ ! -d "$dest_dir" ]]; then
    log "Creating bash completion directory: $dest_dir"
    mkdir -p "$dest_dir"
  fi

  # Simple atomic copy
  cp "$source_file" "${dest_dir}/${BINARY_NAME}.tmp" &&
    mv -f "${dest_dir}/${BINARY_NAME}.tmp" "${dest_dir}/${BINARY_NAME}"

  success "Installed bash completion to $dest_dir/"
  return 0
}

install_zsh_completion() {
  local source_file="${COMPLETION_DIR}/${BINARY_NAME}.zsh"
  if [[ ! -f "$source_file" ]]; then
    log "Zsh completion not found: $source_file"
    return 1
  fi

  local dest_dir="${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions"
  if [[ ! -d "$dest_dir" ]]; then
    log "Creating zsh completion directory: $dest_dir"
    mkdir -p "$dest_dir"
  fi

  # Simple atomic copy
  cp "$source_file" "${dest_dir}/_${BINARY_NAME}.tmp" &&
    mv -f "${dest_dir}/_${BINARY_NAME}.tmp" "${dest_dir}/_${BINARY_NAME}"

  success "Installed zsh completion to $dest_dir/"
  return 0
}

install_fish_completion() {
  local source_file="${COMPLETION_DIR}/${BINARY_NAME}.fish"
  if [[ ! -f "$source_file" ]]; then
    log "Fish completion not found: $source_file"
    return 1
  fi

  local dest_dir="${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions"
  if [[ ! -d "$dest_dir" ]]; then
    log "Creating fish completion directory: $dest_dir"
    mkdir -p "$dest_dir"
  fi

  # Simple atomic copy
  cp "$source_file" "${dest_dir}/${BINARY_NAME}.fish.tmp" &&
    mv -f "${dest_dir}/${BINARY_NAME}.fish.tmp" "${dest_dir}/${BINARY_NAME}.fish"

  success "Installed fish completion to $dest_dir/"
  return 0
}

main() {
  log "Installing shell completions for ${BINARY_NAME}"

  if [[ ! -d "$COMPLETION_DIR" ]]; then
    fail "Completion directory not found: $COMPLETION_DIR"
    echo "Generate completions first"
    exit 1
  fi

  local installed_count=0

  # Try to install each shell's completion
  if command -v bash >/dev/null 2>&1; then
    if install_bash_completion; then
      ((installed_count++))
    fi
  fi

  if command -v zsh >/dev/null 2>&1; then
    if install_zsh_completion; then
      ((installed_count++))
    fi
  fi

  if command -v fish >/dev/null 2>&1; then
    if install_fish_completion; then
      ((installed_count++))
    fi
  fi

  if [[ $installed_count -eq 0 ]]; then
    fail "No shell completions were installed"
    exit 1
  fi

  success "Installed $installed_count shell completion(s)"
}

main "$@"
