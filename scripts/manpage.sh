#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
#
# SPDX-License-Identifier: CC0-1.0

# Description: This script generates man pages for the gitprovidersync binary.
#              It creates a compressed man page file in the specified output directory.
#
# Usage: ./generate_manpages.sh
#
# The script performs the following actions:
# 1. Removes the existing manpages directory (if it exists)
# 2. Creates a new manpages directory
# 3. Runs the gitprovidersync binary with the 'man' command to generate man page content
# 4. Compresses the generated man page and saves it as a .gz file
#
# Output:
# - The generated man page is saved as ./generated/manpages/gitprovidersync.1.gz
#
# Dependencies:
# - Go: The gitprovidersync binary should be built and accessible
# - gzip: Used for compressing the man page
#
# Note:
# - Ensure you run this script from the root directory of the gitprovidersync project
#
set -euo pipefail

# Define color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Define constants
MANPAGES_DIR="./generated/manpages"
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

generate_manpages() {
  echo -e "${YELLOW}Generating manpages...${NC}"
  if ! remove_directory "$MANPAGES_DIR"; then
    echo -e "${RED}Failed to remove existing manpages directory${NC}" >&2
    return 1
  fi
  mkdir -p "$MANPAGES_DIR"
  echo -e "${YELLOW}Generating manpages...${NC}"
  if ! go run . man | gzip -c -9 >"$MANPAGES_DIR/$BINARY_NAME".1.gz; then
    echo -e "${RED}Error generating manpages${NC}" >&2
    return 1
  fi
  echo -e "${GREEN}Manpages generated successfully in $MANPAGES_DIR${NC}"
}

main() {
  generate_manpages
}

main "$@"
