#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 itiquette/git-provider-sync
#
# SPDX-License-Identifier: CC0-1.0

# Description: This script generates mock interfaces for the hexagonal architecture
#              using the mockery tool with configuration file.
#
# Usage: ./generatemock.sh
#
# The script generates mocks for:
# 1. Hexagonal architecture ports (domain interfaces)
# 2. External dependencies (GitLab API client, etc.)
#
# Mock files are generated in the following directory structure:
# - generated/mocks/mockhexagonal: Mocks for hexagonal ports
# - generated/mocks/mockgitlab: Mocks for external dependencies
#
# Dependencies:
# - mockery v2.x: This script requires mockery tool to be installed
# - Go: The project and its dependencies should be properly set up
# - .mockery.yaml: Configuration file for mockery
#
# Note: Ensure you run this script from the root directory of the project
#
set -euo pipefail

# Define color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Define directories
OUTPUT_DIR="generated/mocks"

# Check if mockery config file exists
check_mockery_config() {
  if [ ! -f ".mockery.yaml" ]; then
    echo -e "${RED}Error: .mockery.yaml configuration file not found${NC}" >&2
    echo -e "${YELLOW}Please ensure the .mockery.yaml file exists in the project root${NC}" >&2
    return 1
  fi
}

# Function to create output directories
prepare_directories() {
  echo -e "${BLUE}Preparing output directories...${NC}"
  mkdir -p "${OUTPUT_DIR}/mockhexagonal"
  mkdir -p "${OUTPUT_DIR}/mockgitlab"
  mkdir -p "${OUTPUT_DIR}/mockdomain"
}

# Function to clean old auto-generated mocks (preserve manual mocks)
clean_old_mocks() {
  echo -e "${BLUE}Cleaning old auto-generated mock files...${NC}"
  if [ -d "${OUTPUT_DIR}" ]; then
    # Only clean files that don't have "manual creation" in their header
    find "${OUTPUT_DIR}" -name "*.go" -type f -exec grep -L "manual creation" {} + -print0 | xargs -0 rm -f 2>/dev/null || true
  fi
}

echo -e "${BLUE}Starting hexagonal architecture mock generation...${NC}"

# Check prerequisites
check_mockery_config

# Prepare environment
prepare_directories
clean_old_mocks

# Check for existing manual mocks first
echo -e "${BLUE}Checking for existing manual mocks...${NC}"
if [ -f "${OUTPUT_DIR}/mockhexagonal/MockLogger.go" ] && [ -f "${OUTPUT_DIR}/mockhexagonal/MockRepositoryProvider.go" ]; then
  echo -e "${GREEN}✓ Manual hexagonal mocks found${NC}"
fi

# Try mockery generation (may fail due to complex imports)
echo -e "${YELLOW}Attempting automatic mock generation...${NC}"

# Generate mocks using mockery v2.53+ with command line approach
echo -e "${BLUE}Generating Logger mock...${NC}"
if mockery --name Logger --dir internal/domain/ports --output "${OUTPUT_DIR}/mockhexagonal" --outpkg mockhexagonal --filename MockLogger.go --disable-version-string --with-expecter 2>/dev/null; then
  echo -e "${GREEN}✓ Logger mock generated automatically${NC}"
else
  echo -e "${YELLOW}⚠ Logger mock failed - will use manual approach${NC}"
fi

echo -e "${BLUE}Generating RepositoryProvider mock...${NC}"
if mockery --name RepositoryProvider --dir internal/domain/ports --output "${OUTPUT_DIR}/mockhexagonal" --outpkg mockhexagonal --filename MockRepositoryProvider.go --disable-version-string --with-expecter 2>/dev/null; then
  echo -e "${GREEN}✓ RepositoryProvider mock generated automatically${NC}"
else
  echo -e "${YELLOW}⚠ RepositoryProvider mock failed - will use manual approach${NC}"
fi

echo -e "${BLUE}Generating GitOperations mock...${NC}"
if mockery --name GitOperations --dir internal/domain/ports --output "${OUTPUT_DIR}/mockhexagonal" --outpkg mockhexagonal --filename MockGitOperations.go --disable-version-string --with-expecter 2>/dev/null; then
  echo -e "${GREEN}✓ GitOperations mock generated automatically${NC}"
else
  echo -e "${YELLOW}⚠ GitOperations mock failed - will use manual approach${NC}"
fi

echo -e "${BLUE}Generating Configuration mock...${NC}"
if mockery --name Configuration --dir internal/domain/ports --output "${OUTPUT_DIR}/mockhexagonal" --outpkg mockhexagonal --filename MockConfiguration.go --disable-version-string --with-expecter 2>/dev/null; then
  echo -e "${GREEN}✓ Configuration mock generated automatically${NC}"
else
  echo -e "${YELLOW}⚠ Configuration mock failed - will use manual approach${NC}"
fi

# Generate external library mocks
echo -e "${YELLOW}Generating external library mocks...${NC}"
if mockery --all --srcpkg gitlab.com/gitlab-org/api/client-go --output "${OUTPUT_DIR}/mockgitlab" 2>/dev/null; then
  echo -e "${GREEN}✓ GitLab API mocks generated${NC}"
else
  echo -e "${YELLOW}⚠ GitLab API mocks failed (external dependency may not be available)${NC}"
fi

# Summary
echo -e "${GREEN}Mock generation process completed!${NC}"
echo -e "${BLUE}Generated files can be found in:${NC}"
echo -e "  - ${OUTPUT_DIR}/mockhexagonal/ (hexagonal architecture ports)"
echo -e "  - ${OUTPUT_DIR}/mockgitlab/ (external dependencies)"

# Check what was actually generated
if [ -d "${OUTPUT_DIR}" ]; then
  MOCK_COUNT=$(find "${OUTPUT_DIR}" -name "*.go" -type f | wc -l)
  echo -e "${GREEN}Total mock files generated: ${MOCK_COUNT}${NC}"
else
  echo -e "${YELLOW}No mock files were generated${NC}"
fi
