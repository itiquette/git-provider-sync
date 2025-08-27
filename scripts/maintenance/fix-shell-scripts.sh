#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Automatically fixes shell script violations using shellcheck and shfmt
# Usage: ./scripts/fix-shell-scripts.sh [--brief] [scripts_dir]
# Dependencies: shellcheck, shfmt, git

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly CYAN=$'\033[1;36m'
readonly NC=$'\033[0m' # No Color

# Parse arguments
BRIEF_MODE=false
SCRIPTS_DIR=""

while [[ $# -gt 0 ]]; do
  case $1 in
  --brief)
    BRIEF_MODE=true
    shift
    ;;
  *)
    if [[ -z "$SCRIPTS_DIR" ]]; then
      SCRIPTS_DIR="$1"
    fi
    shift
    ;;
  esac
done

# Set defaults
SCRIPTS_DIR="${SCRIPTS_DIR:-scripts}"

log() {
  if [[ "$BRIEF_MODE" != "true" ]]; then
    printf "${YELLOW}▸${NC} %s\n" "$1"
  fi
}

success() {
  printf "${GREEN}✓${NC} %s\n" "$1"
}

fail() {
  printf "${RED}✗${NC} %s\n" "$1" >&2
}

validate_scripts_directory() {
  if [[ ! -d "$SCRIPTS_DIR" ]]; then
    fail "Scripts directory does not exist: $SCRIPTS_DIR"
    return 1
  fi

  if ! find "$SCRIPTS_DIR" -name "*.sh" -type f | head -1 >/dev/null 2>&1; then
    fail "No shell scripts found in directory: $SCRIPTS_DIR"
    return 1
  fi

  return 0
}

# Applies shellcheck fixes to a single script using diff format
apply_shellcheck_fixes() {
  local script="$1"
  local fixes_applied=false

  if [[ ! -f "$script" ]]; then
    return 1
  fi

  # Generate shellcheck diff and apply if available
  local diff_output
  if diff_output=$(shellcheck -f diff "$script" 2>/dev/null || true); then
    if [[ -n "$diff_output" ]]; then
      if [[ "$BRIEF_MODE" != "true" ]]; then
        log "Applying shellcheck fixes to $(basename "$script")..."
      fi

      if printf "%s" "$diff_output" | git apply; then
        fixes_applied=true
      else
        if [[ "$BRIEF_MODE" != "true" ]]; then
          fail "Failed to apply shellcheck fixes to $(basename "$script")"
        fi
        return 1
      fi
    fi
  fi

  if [[ "$fixes_applied" == "true" ]]; then
    printf "shellcheck:%s\n" "$(basename "$script")"
  fi
  return 0
}

# Applies shfmt formatting to a single script
apply_shfmt_formatting() {
  local script="$1"

  if [[ ! -f "$script" ]]; then
    return 1
  fi

  # Check if formatting changes are needed
  if ! shfmt -i 2 -d "$script" >/dev/null 2>&1; then
    if [[ "$BRIEF_MODE" != "true" ]]; then
      log "Formatting $(basename "$script") with shfmt..."
    fi

    if shfmt -i 2 -w "$script"; then
      printf "shfmt:%s\n" "$(basename "$script")"
      return 0
    else
      if [[ "$BRIEF_MODE" != "true" ]]; then
        fail "Failed to format $(basename "$script") with shfmt"
      fi
      return 1
    fi
  fi

  return 0
}

# Processes all shell scripts in the directory
process_all_scripts() {
  local shellcheck_fixes=()
  local shfmt_fixes=()
  local failed_scripts=()

  while IFS= read -r -d '' script; do
    if [[ -f "$script" ]]; then
      # Apply shellcheck fixes
      if shellcheck_result=$(apply_shellcheck_fixes "$script"); then
        if [[ -n "$shellcheck_result" ]]; then
          shellcheck_fixes+=("$shellcheck_result")
        fi
      else
        failed_scripts+=("shellcheck:$(basename "$script")")
      fi

      # Apply shfmt formatting
      if shfmt_result=$(apply_shfmt_formatting "$script"); then
        if [[ -n "$shfmt_result" ]]; then
          shfmt_fixes+=("$shfmt_result")
        fi
      else
        failed_scripts+=("shfmt:$(basename "$script")")
      fi
    fi
  done < <(find "$SCRIPTS_DIR" -name "*.sh" -type f -print0)

  # Output results
  if [[ "$BRIEF_MODE" == "true" ]]; then
    printf "SHELLCHECK_FIXES=%d\n" "${#shellcheck_fixes[@]}"
    printf "SHFMT_FIXES=%d\n" "${#shfmt_fixes[@]}"
    printf "FAILED_SCRIPTS=%d\n" "${#failed_scripts[@]}"
  else
    if [[ ${#shellcheck_fixes[@]} -gt 0 ]]; then
      success "Applied shellcheck fixes to ${#shellcheck_fixes[@]} scripts"
      for fix in "${shellcheck_fixes[@]}"; do
        printf "  • %s\n" "$fix"
      done
    fi

    if [[ ${#shfmt_fixes[@]} -gt 0 ]]; then
      success "Applied shfmt formatting to ${#shfmt_fixes[@]} scripts"
      for fix in "${shfmt_fixes[@]}"; do
        printf "  • %s\n" "$fix"
      done
    fi

    if [[ ${#failed_scripts[@]} -gt 0 ]]; then
      fail "Failed to process ${#failed_scripts[@]} script operations"
      for failure in "${failed_scripts[@]}"; do
        printf "  • %s\n" "$failure"
      done
      return 1
    fi

    if [[ ${#shellcheck_fixes[@]} -eq 0 && ${#shfmt_fixes[@]} -eq 0 ]]; then
      success "All scripts are already properly formatted and compliant"
    fi
  fi

  return 0
}

main() {
  if [[ "$BRIEF_MODE" != "true" ]]; then
    printf "${CYAN}→${NC} Auto-fixing shell script violations...\n"
  fi

  # Validate environment
  validate_scripts_directory || exit 1

  # Check dependencies
  if ! command -v shellcheck >/dev/null 2>&1; then
    fail "shellcheck not found - install it to apply automatic fixes"
    exit 1
  fi

  if ! command -v shfmt >/dev/null 2>&1; then
    fail "shfmt not found - install it to apply formatting fixes"
    exit 1
  fi

  if ! command -v git >/dev/null 2>&1; then
    fail "git not found - needed to apply shellcheck diff patches"
    exit 1
  fi

  # Process all scripts
  if ! process_all_scripts; then
    exit 1
  fi

  if [[ "$BRIEF_MODE" != "true" ]]; then
    success "Shell script auto-fixes completed"
  fi
}

main "$@"
