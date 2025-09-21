#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: CC0-1.0

# Lists available updates for Go version, modules, and development tools
# Usage: ./scripts/list-updates.sh (run from project root)
# Dependencies: go, jq, tools/go.mod with tool directive

set -euo pipefail

# Colors for output
readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly CYAN=$'\033[1;36m'
readonly NC=$'\033[0m' # No Color

readonly PROJECT_MARKERS=("go.mod" "justfile" "go.sum")

log() {
  printf "${YELLOW}▸${NC} %s\n" "$1"
}

success() {
  printf "${GREEN}✓${NC} %s\n" "$1"
}

fail() {
  printf "${RED}✗${NC} %s\n" "$1" >&2
}

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

validate_tools_directory() {
  if [[ ! -d "tools" ]] || [[ ! -f "tools/go.mod" ]]; then
    return 1 # Not an error for this script, just skip tools section
  fi
  return 0
}

main() {
  validate_project_directory || exit 1

  if ! command -v jq >/dev/null 2>&1; then
    fail "jq is required but not installed"
    log "Install jq to parse JSON output: apt install jq / brew install jq"
    exit 1
  fi

  printf "${CYAN}→${NC} Go Version\n"

  # Get current Go version from go.mod (major.minor only)
  current_go_mod=$(grep "^go " go.mod | awk '{print $2}')

  # Try to get the actual installed Go version with patch number
  if command -v go >/dev/null 2>&1; then
    current_go_full=$(go version | sed -n 's/.*go\([0-9]\+\.[0-9]\+\.[0-9]\+\).*/\1/p')
    if [[ -z "$current_go_full" ]]; then
      # Fallback: if patch version not found, append .0
      current_go_full="${current_go_mod}.0"
    fi
  else
    # If go is not installed, assume .0 patch version
    current_go_full="${current_go_mod}.0"
  fi

  # Get latest stable Go version (full version with patch)
  latest_go=$(curl -sL "https://go.dev/dl/?mode=json" 2>/dev/null | jq -r '.[0].version' 2>/dev/null | sed 's/^go//' || echo "")

  if [[ -n "$latest_go" ]]; then
    # Compare full versions including patch
    if [[ "$current_go_full" == "$latest_go" ]]; then
      printf "  %-30s ${GREEN}%s${NC} (up to date)\n" "Go" "$current_go_full"
    else
      # Show full version with patch number for the update
      printf "  %-30s ${YELLOW}%s${NC} → ${GREEN}%s${NC}\n" "Go" "$current_go_full" "$latest_go"
    fi
  else
    printf "  %-30s ${GREEN}%s${NC} (version check failed)\n" "Go" "$current_go_full"
  fi

  printf "\n"

  printf "${CYAN}→${NC} Go Module Dependencies\n"

  # Get outdated dependencies with nice formatting
  temp_file=$(mktemp)

  # Get direct dependencies from go.mod using JSON output
  direct_deps_json=$(go mod edit -json)
  direct_deps=$(echo "$direct_deps_json" | jq -r '.Require[]? | select(.Indirect != true) | .Path' 2>/dev/null)

  # Get update information for all modules
  go list -u -m all 2>/dev/null >"$temp_file"

  # Process all direct dependencies and show their status
  has_deps=false
  while IFS= read -r dep_path; do
    if [[ -n "$dep_path" ]]; then
      has_deps=true
      # Look for this dependency in the update list
      found=false
      while IFS= read -r line; do
        if [[ "$line" =~ ^"$dep_path " ]]; then
          found=true
          read -r module current rest <<<"$line"

          if echo "$line" | grep -q '\['; then
            # Direct dependency has an update available
            latest="${line#*[}"
            latest="${latest%]*}"
            printf "  %-50s ${YELLOW}%s${NC} → ${GREEN}%s${NC}\n" "$module" "$current" "$latest"
          else
            # Direct dependency is up to date
            printf "  %-50s ${GREEN}%s${NC} (up to date)\n" "$module" "$current"
          fi
          break
        fi
      done <"$temp_file"

      # If dependency not found in update list, it might be missing or have issues
      if [[ "$found" == "false" ]]; then
        printf "  %-50s ${RED}version check failed${NC}\n" "$dep_path"
      fi
    fi
  done <<<"$direct_deps"

  # If no direct dependencies were found
  if [[ "$has_deps" == "false" ]]; then
    printf "  %s(no direct dependencies)%s\n" "${GREEN}" "${NC}"
  fi

  printf "\n"

  printf "${CYAN}→${NC} Go Tools (tools/go.mod)\n"

  if validate_tools_directory; then
    # Extract tools from tool directive (stop at closing parenthesis)
    tools=$(awk '/^tool \(/{flag=1; next} /^\)/{flag=0} flag && /^\s+[a-zA-Z]/' tools/go.mod 2>/dev/null | sed 's/^\s*//' | sed 's/\s*$//' || printf "\n")

    if [[ -n "$tools" ]]; then
      # For each tool, get its module path and check version
      echo "$tools" | while IFS= read -r tool_path; do
        if [[ -n "$tool_path" ]]; then
          # Extract the module path (everything before /cmd/ or use full path)
          if [[ "$tool_path" =~ (.+)/cmd/.+ ]]; then
            module_path="${BASH_REMATCH[1]}"
          else
            module_path="$tool_path"
          fi

          # Get tool name for display
          tool_name=$(basename "$tool_path")
          if [[ "$tool_name" == "cmd" ]] || [[ "$tool_name" =~ ^(v[0-9]+|internal|pkg)$ ]]; then
            tool_name=$(echo "$tool_path" | awk -F'/' '{
            for(i=NF; i>0; i--) {
              if($i !~ /^(cmd|v[0-9]+|internal|pkg)$/) {
                print $i
                break
              }
            }
          }')
          fi

          # Check version in tools module
          cd tools 2>/dev/null && {
            version_info=$(go list -m -json "$module_path" 2>/dev/null || echo "{}")
            if [[ "$version_info" != "{}" ]]; then
              current=$(echo "$version_info" | jq -r '.Version // "unknown"')
              update_info=$(go list -m -u -json "$module_path" 2>/dev/null || echo "{}")
              latest=$(echo "$update_info" | jq -r '.Update.Version // empty' || printf "\n")

              if [[ -n "$latest" ]] && [[ "$latest" != "null" ]]; then
                printf "  %-30s ${YELLOW}%s${NC} → ${GREEN}%s${NC}\n" "$tool_name" "$current" "$latest"
              else
                printf "  %-30s ${GREEN}%s${NC} (up to date)\n" "$tool_name" "$current"
              fi
            else
              printf "  %-30s ${RED}version check failed${NC}\n" "$tool_name"
            fi
            cd - >/dev/null 2>&1
          }
        fi
      done
    else
      printf "  No tools found in tools module\n"
    fi
  else
    printf "  Tools module not found (tools/go.mod)\n"
  fi

  printf "\n"
}

main "$@"
