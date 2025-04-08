#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
#
# SPDX-License-Identifier: CC0-1.0

# Description: This script generates shell completions for the gitprovidersync binary.
#              It creates completion files for bash, zsh, and fish shells.
#
# Usage: ./generate_completions.sh
#
# The script performs the following actions:
# 1. Removes the existing completions directory (if it exists)
# 2. Creates a new completions directory
# 3. Generates completion files for bash, zsh, and fish shells
# 4. Saves the completion files in the ./generated/completions directory
#
# Note: This script requires the gitprovidersync binary to be built and available
#       in the same directory as this script.
#

set -euo pipefail

# Define color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Define constants
COMPLETIONS_DIR="./generated/completions"
BINARY_NAME="gitprovidersync"

remove_directory() {
  local dir_to_remove="$1"
  # Check if the directory exists
  if [[ ! -d "$dir_to_remove" ]]; then
    echo -e "${YELLOW}Directory does not exist: $dir_to_remove${NC}" >&2
    return 0
  fi
  # Check if the directory is within the project
  if [[ ! "$dir_to_remove" == "./generated/"* ]]; then
    echo -e "${RED}Safety check failed. Directory is not within ./generated/: $dir_to_remove${NC}" >&2
    return 1
  fi
  # Remove the directory
  if ! rm -rf "${dir_to_remove:?}"; then
    echo -e "${RED}Failed to remove directory: $dir_to_remove${NC}" >&2
    return 1
  fi
  echo -e "${GREEN}Successfully removed directory: $dir_to_remove${NC}"
}

generate_completions() {
  local shells=("bash" "zsh" "fish")
  echo -e "${BLUE}Generating completions...${NC}"
  if ! remove_directory "$COMPLETIONS_DIR"; then
    echo -e "${RED}Failed to remove existing completions directory${NC}" >&2
    return 1
  fi
  mkdir -p "$COMPLETIONS_DIR"
  # Generate completions for each shell
  for shell in "${shells[@]}"; do
    echo -e "${YELLOW}Generating $shell completion...${NC}"
    if ! go run main.go completion "$shell" >"$COMPLETIONS_DIR/$BINARY_NAME.$shell"; then
      echo -e "${RED}Error generating $shell completion${NC}" >&2
      return 1
    fi
  done
  echo -e "${GREEN}Completions generated successfully in $COMPLETIONS_DIR${NC}"
}

main() {
  generate_completions
}

main "$@"
