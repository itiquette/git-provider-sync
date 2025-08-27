#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Installs binary, man page, and shell completions locally
# Usage: ./scripts/install-local-all.sh [--brief] [bin_dir] [executable_name]
# Dependencies: install-binary.sh, manpage.sh, completions.sh, install-completions.sh

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly CYAN=$'\033[1;36m'
readonly NC=$'\033[0m' # No Color

readonly PROJECT_MARKERS=("go.mod" "justfile" "go.sum")

# Parse arguments
BRIEF_MODE=false
BIN_DIR=""
EXECUTABLE=""

while [[ $# -gt 0 ]]; do
  case $1 in
  --brief)
    BRIEF_MODE=true
    shift
    ;;
  *)
    if [[ -z "$BIN_DIR" ]]; then
      BIN_DIR="$1"
    elif [[ -z "$EXECUTABLE" ]]; then
      EXECUTABLE="$1"
    fi
    shift
    ;;
  esac
done

# Set defaults
BIN_DIR="${BIN_DIR:-./bin}"
EXECUTABLE="${EXECUTABLE:-$(basename "$(pwd)")}"

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

# Validates that required project marker files exist in current directory
validate_project_directory() {
  local missing=()
  for marker in "${PROJECT_MARKERS[@]}"; do
    if [[ ! -f "$marker" ]]; then
      missing+=("$marker")
    fi
  done
  if ((${#missing[@]} > 0)); then
    fail "Missing project markers: ${missing[*]}"
    return 1
  fi
  return 0
}

# Installs binary component using dedicated script
install_binary_component() {
  log "Installing binary..."
  if _binary_output=$(./scripts/install/install-binary.sh --brief "$BIN_DIR" "$EXECUTABLE" 2>&1); then
    printf "✓ Binary installed: ~/.local/bin/%s\n" "$EXECUTABLE"
    return 0
  else
    printf "✗ Binary installation failed\n"
    if [[ "$BRIEF_MODE" != "true" ]]; then
      fail "Binary installation failed"
    fi
    return 1
  fi
}

# Generates and installs man page component
install_manpage_component() {
  log "Generating man page..."
  if _manpage_output=$(./scripts/docs/manpage.sh --brief 2>&1); then
    log "Installing man page..."
    if [[ -d ~/.local/share/man/man1 ]]; then
      if cp "generated/manpages/$EXECUTABLE.1.gz" ~/.local/share/man/man1/; then
        printf "✓ Man page installed: ~/.local/share/man/man1/%s.1.gz\n" "$EXECUTABLE"
        return 0
      else
        printf "✗ Man page copy failed\n"
        return 1
      fi
    else
      printf "! Man page generated but ~/.local/share/man/man1 missing\n"
      return 0
    fi
  else
    printf "✗ Man page generation failed\n"
    return 1
  fi
}

# Generates and installs shell completions component
install_completions_component() {
  log "Generating completions..."
  if _completions_output=$(./scripts/docs/completions.sh --brief 2>&1); then
    log "Installing completions..."
    if install_output=$(./scripts/install/install-completions.sh --brief "$EXECUTABLE" 2>&1); then
      # Parse completion results
      installed_shells=$(echo "$install_output" | grep -o "success:[^,]*" | cut -d: -f2 || echo "")
      skipped_shells=$(echo "$install_output" | grep -o "skipped:[^,]*" | cut -d: -f2 || echo "")

      if [[ -n "$installed_shells" ]]; then
        printf "✓ Completions installed for: %s\n" "$installed_shells"
        if [[ -n "$skipped_shells" ]]; then
          printf "! Completions skipped for: %s (directories missing)\n" "$skipped_shells"
        fi
        return 0
      else
        printf "! No completions were installed (missing directories)\n"
        return 0
      fi
    else
      printf "✗ Completion installation failed\n"
      return 1
    fi
  else
    printf "✗ Completion generation failed\n"
    return 1
  fi
}

# Displays unified installation summary with next steps
display_summary() {
  local results=("$@")
  local success_count=0
  local has_failures=false

  if [[ "$BRIEF_MODE" != "true" ]]; then
    printf "\n${CYAN}→${NC} Installation Summary:\n\n"
  fi

  # Show results and count successes
  for result in "${results[@]}"; do
    if [[ "$BRIEF_MODE" != "true" ]]; then
      # Handle multi-line results properly
      while IFS= read -r line; do
        printf "  %s\n" "$line"
        # Count first line that starts with ✓ as success
        if [[ "$line" == ✓* ]]; then
          ((success_count++)) || true
        elif [[ "$line" == ✗* ]]; then
          has_failures=true
        fi
      done <<<"$result"
    else
      # For brief mode, just count successes from first lines
      local first_line
      first_line=$(echo "$result" | head -1)
      if [[ "$first_line" == ✓* ]]; then
        ((success_count++)) || true
      elif [[ "$first_line" == ✗* ]]; then
        has_failures=true
      fi
    fi
  done

  if [[ "$BRIEF_MODE" != "true" ]]; then
    printf "\n"
  fi

  # Display final status and next steps
  if [[ $success_count -gt 0 ]]; then
    if [[ "$BRIEF_MODE" != "true" ]]; then
      success "Installation completed with $success_count/3 components successful"
      printf "\n${YELLOW}Next steps:${NC}\n"
      printf "  • Add to PATH: ${GREEN}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}\n"
      printf "  • Test binary: ${GREEN}$EXECUTABLE --version${NC}\n"
      printf "  • View manual: ${GREEN}man $EXECUTABLE${NC}\n"
      printf "  • Restart shell to activate completions\n"
    else
      printf "INSTALL_STATUS=success:$success_count/3_components\n"
    fi

    if [[ "$has_failures" == "true" ]]; then
      return 1
    fi
    return 0
  else
    if [[ "$BRIEF_MODE" != "true" ]]; then
      fail "Installation failed - no components were successfully installed"
    else
      printf "INSTALL_STATUS=failed:0/3_components\n"
    fi
    return 1
  fi
}

main() {
  if [[ "$BRIEF_MODE" != "true" ]]; then
    printf "${CYAN}→${NC} Installing complete local development environment...\n\n"
  fi

  # Validate environment
  validate_project_directory || exit 1

  # Collect results from each component
  declare -a results

  # Install binary component
  if binary_result=$(install_binary_component); then
    results+=("$binary_result")
  else
    results+=("$binary_result")
    exit 1 # Binary is critical - fail immediately
  fi

  # Install man page component
  if manpage_result=$(install_manpage_component); then
    results+=("$manpage_result")
  else
    results+=("$manpage_result")
  fi

  # Install completions component
  if completions_result=$(install_completions_component); then
    results+=("$completions_result")
  else
    results+=("$completions_result")
  fi

  # Display unified summary
  if ! display_summary "${results[@]}"; then
    exit 1
  fi
}

main "$@"
