#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Project-agnostic shell completion generator
# Usage: ./scripts/docs/completions.sh [binary_name] [output_dir]
# Defaults: binary_name=$(basename "$(pwd)"), output_dir=generated/completions

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly CYAN=$'\033[1;36m'
readonly NC=$'\033[0m' # No Color

# Parse arguments with defaults
readonly BINARY_NAME="${1:-${BINARY_NAME:-$(basename "$(pwd)")}}"
readonly OUTPUT_DIR="${2:-${OUTPUT_DIR:-generated/completions}}"
readonly BINARY_PATH="${BINARY_PATH:-./bin/${BINARY_NAME}}"

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

check_binary() {
  # Try to find the binary in various locations
  local binary_locations=(
    "$BINARY_PATH"
    "./bin/${BINARY_NAME}"
    "./${BINARY_NAME}"
    "$(command -v "$BINARY_NAME" 2>/dev/null || true)"
  )

  for location in "${binary_locations[@]}"; do
    if [[ -n "$location" ]] && [[ -x "$location" ]]; then
      echo "$location"
      return 0
    fi
  done

  return 1
}

verify_completion() {
  local shell="$1"
  local completion_file="$2"

  # Check file exists and has content
  if [[ ! -s "$completion_file" ]]; then
    return 1
  fi

  local content
  content=$(<"$completion_file")

  # Verify shell-specific patterns
  case "$shell" in
  bash)
    # Check for bash completion patterns
    if ! echo "$content" | grep -qE "complete.*${BINARY_NAME}|_${BINARY_NAME}\(\)"; then
      log "Warning: Bash completion may be invalid (missing expected patterns)"
      return 1
    fi
    # Basic syntax check if bash is available
    if command -v bash >/dev/null 2>&1; then
      if ! bash -n "$completion_file" 2>/dev/null; then
        log "Warning: Bash completion has syntax errors"
        return 1
      fi
    fi
    ;;
  zsh)
    # Check for zsh completion patterns
    if ! echo "$content" | grep -qE "#compdef.*${BINARY_NAME}|_${BINARY_NAME}"; then
      log "Warning: Zsh completion may be invalid (missing expected patterns)"
      return 1
    fi
    ;;
  fish)
    # Check for fish completion patterns
    if ! echo "$content" | grep -qE "complete.*${BINARY_NAME}"; then
      log "Warning: Fish completion may be invalid (missing expected patterns)"
      return 1
    fi
    # Basic syntax check if fish is available
    if command -v fish >/dev/null 2>&1; then
      if ! fish -n "$completion_file" 2>/dev/null; then
        log "Warning: Fish completion has syntax errors"
        return 1
      fi
    fi
    ;;
  esac

  return 0
}

generate_completions() {
  local binary
  binary=$(check_binary) || fail "Binary not found: $BINARY_NAME"

  log "Using binary: $binary"

  # Create output directory
  mkdir -p "$OUTPUT_DIR"

  # Remove existing completions
  rm -f "$OUTPUT_DIR/${BINARY_NAME}".{bash,zsh,fish}

  local generated_count=0
  local shells=("bash" "zsh" "fish")

  for shell in "${shells[@]}"; do
    local output_file="$OUTPUT_DIR/${BINARY_NAME}.$shell"

    log "Generating $shell completion..."

    if "$binary" completion "$shell" >"$output_file" 2>/dev/null; then
      # Verify the generated completion
      if verify_completion "$shell" "$output_file"; then
        success "Generated and verified $shell completion"
        ((generated_count++))
      else
        # Keep the file even if validation warns, user can check it
        log "Generated $shell completion (with warnings)"
        ((generated_count++))
      fi
    else
      log "Failed to generate $shell completion (command may not be supported)"
      rm -f "$output_file"
    fi
  done

  if [[ $generated_count -eq 0 ]]; then
    fail "No completions were generated. Binary may not support completion command."
  fi

  success "Generated $generated_count completion file(s) in $OUTPUT_DIR"
}

show_installation_help() {
  printf "\n${CYAN}Shell Completion Installation${NC}\n\n"

  printf "${YELLOW}Runtime completion (recommended):${NC}\n"
  printf "  Users can generate completions on-demand:\n"
  printf "  ${GREEN}%s completion bash${NC}   # Generate bash completion\n" "$BINARY_NAME"
  printf "  ${GREEN}%s completion zsh${NC}    # Generate zsh completion\n" "$BINARY_NAME"
  printf "  ${GREEN}%s completion fish${NC}   # Generate fish completion\n\n" "$BINARY_NAME"

  printf "${YELLOW}Pre-generated files (for packaging):${NC}\n\n"

  printf "${CYAN}Bash:${NC}\n"
  printf "  System-wide: ${GREEN}sudo cp %s/%s.bash /usr/share/bash-completion/completions/%s${NC}\n" "$OUTPUT_DIR" "$BINARY_NAME" "$BINARY_NAME"
  printf "  User-local:  ${GREEN}cp %s/%s.bash ~/.local/share/bash-completion/completions/%s${NC}\n\n" "$OUTPUT_DIR" "$BINARY_NAME" "$BINARY_NAME"

  printf "${CYAN}Zsh:${NC}\n"
  printf "  System-wide: ${GREEN}sudo cp %s/%s.zsh /usr/local/share/zsh/site-functions/_%s${NC}\n" "$OUTPUT_DIR" "$BINARY_NAME" "$BINARY_NAME"
  printf "  User-local:  ${GREEN}mkdir -p ~/.local/share/zsh/site-functions && cp %s/%s.zsh ~/.local/share/zsh/site-functions/_%s${NC}\n\n" "$OUTPUT_DIR" "$BINARY_NAME" "$BINARY_NAME"

  printf "${CYAN}Fish:${NC}\n"
  printf "  System-wide: ${GREEN}sudo cp %s/%s.fish /usr/share/fish/vendor_completions.d/${NC}\n" "$OUTPUT_DIR" "$BINARY_NAME"
  printf "  User-local:  ${GREEN}cp %s/%s.fish ~/.config/fish/completions/${NC}\n" "$OUTPUT_DIR" "$BINARY_NAME"
}

main() {
  log "Starting completion generation for $BINARY_NAME"

  generate_completions
  show_installation_help

  success "Completion generation completed"
}

main "$@"
