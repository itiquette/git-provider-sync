#!/bin/bash

# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: EUPL-1.2

CMD="./bin/gitprovidersync-linux-amd64 sync"

# Colors for better readability
GREEN='\033[0;32m'
NC='\033[0m' # No Color
YELLOW='\033[1;33m'

# Function to run command and wait for user input
run_test() {
  printf "\n${GREEN}Running command:${NC}\n"
  printf "${YELLOW}$1${NC}\n"
  printf "\nPress Enter to execute, Ctrl+C to cancel...\n"
  read -r
  eval "$1"
  printf "\nPress Enter to return to menu...\n"
  read -r
}

while true; do
  clear
  printf "%s\n" "Git Provider Sync Test Menu - All Possible Flag Combinations"
  printf "%s\n" "========================================================"
  printf "%s\n" "Single Flags:"
  printf "%s\n" "1.  Basic sync (no flags)"
  printf "%s\n" "2.  Active from limit only"
  printf "%s\n" "3.  ASCII name only"
  printf "%s\n" "4.  Dry run only"
  printf "%s\n" "5.  Force push only"
  printf "%s\n" "6.  Ignore invalid name only"
  printf "\n"
  printf "%s\n" "Two Flag Combinations:"
  printf "%s\n" "7.  Active from limit + ASCII name"
  printf "%s\n" "8.  Active from limit + Dry run"
  printf "%s\n" "9.  Active from limit + Force push"
  printf "%s\n" "10. Active from limit + Ignore invalid name"
  printf "%s\n" "11. ASCII name + Dry run"
  printf "%s\n" "12. ASCII name + Force push"
  printf "%s\n" "13. ASCII name + Ignore invalid name"
  printf "%s\n" "14. Dry run + Force push"
  printf "%s\n" "15. Dry run + Ignore invalid name"
  printf "%s\n" "16. Force push + Ignore invalid name"
  printf "\n"
  printf "%s\n" "Three Flag Combinations:"
  printf "%s\n" "17. Active from limit + ASCII name + Dry run"
  printf "%s\n" "18. Active from limit + ASCII name + Force push"
  printf "%s\n" "19. Active from limit + ASCII name + Ignore invalid name"
  printf "%s\n" "20. Active from limit + Dry run + Force push"
  printf "%s\n" "21. Active from limit + Dry run + Ignore invalid name"
  printf "%s\n" "22. Active from limit + Force push + Ignore invalid name"
  printf "%s\n" "23. ASCII name + Dry run + Force push"
  printf "%s\n" "24. ASCII name + Dry run + Ignore invalid name"
  printf "%s\n" "25. ASCII name + Force push + Ignore invalid name"
  printf "%s\n" "26. Dry run + Force push + Ignore invalid name"
  printf "\n"
  printf "%s\n" "Four Flag Combinations:"
  printf "%s\n" "27. Active from limit + ASCII name + Dry run + Force push"
  printf "%s\n" "28. Active from limit + ASCII name + Dry run + Ignore invalid name"
  printf "%s\n" "29. Active from limit + ASCII name + Force push + Ignore invalid name"
  printf "%s\n" "30. Active from limit + Dry run + Force push + Ignore invalid name"
  printf "%s\n" "31. ASCII name + Dry run + Force push + Ignore invalid name"
  printf "\n"
  printf "%s\n" "All Flags:"
  printf "%s\n" "32. All flags combined"
  printf "\n"
  printf "%s\n" "Global Flag Combinations:"
  printf "%s\n" "33. All sync flags + Console output"
  printf "%s\n" "34. All sync flags + JSON output"
  printf "%s\n" "35. All sync flags + Quiet mode"
  printf "%s\n" "36. All sync flags + Debug verbosity"
  printf "%s\n" "37. All sync flags + Trace + Caller info"
  printf "%s\n" "38. All sync flags + Custom config"
  printf "%s\n" "39. All sync flags + Config file only"
  printf "%s\n" "40. Everything combined"
  printf "\n"
  printf "%s\n" "0. Exit"
  printf "\n"
  read -r -p "Select an option (0-40): " choice

  case $choice in
  0)
    printf "%s\n" "Exiting..."
    exit 0
    ;;
  1)
    run_test "$CMD"
    ;;
  2)
    run_test "$CMD --since=\"-1h\""
    ;;
  3)
    run_test "$CMD --sanitize-names"
    ;;
  4)
    run_test "$CMD --dry-run"
    ;;
  5)
    run_test "$CMD --force-push"
    ;;
  6)
    run_test "$CMD --skip-invalid"
    ;;
  7)
    run_test "$CMD --since=\"-1h\" --sanitize-names"
    ;;
  8)
    run_test "$CMD --since=\"-1h\" --dry-run"
    ;;
  9)
    run_test "$CMD --since=\"-1h\" --force-push"
    ;;
  10)
    run_test "$CMD --since=\"-1h\" --skip-invalid"
    ;;
  11)
    run_test "$CMD --sanitize-names --dry-run"
    ;;
  12)
    run_test "$CMD --sanitize-names --force-push"
    ;;
  13)
    run_test "$CMD --sanitize-names --skip-invalid"
    ;;
  14)
    run_test "$CMD --dry-run --force-push"
    ;;
  15)
    run_test "$CMD --dry-run --skip-invalid"
    ;;
  16)
    run_test "$CMD --force-push --skip-invalid"
    ;;
  17)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run"
    ;;
  18)
    run_test "$CMD --since=\"-1h\" --sanitize-names --force-push"
    ;;
  19)
    run_test "$CMD --since=\"-1h\" --sanitize-names --skip-invalid"
    ;;
  20)
    run_test "$CMD --since=\"-1h\" --dry-run --force-push"
    ;;
  21)
    run_test "$CMD --since=\"-1h\" --dry-run --skip-invalid"
    ;;
  22)
    run_test "$CMD --since=\"-1h\" --force-push --skip-invalid"
    ;;
  23)
    run_test "$CMD --sanitize-names --dry-run --force-push"
    ;;
  24)
    run_test "$CMD --sanitize-names --dry-run --skip-invalid"
    ;;
  25)
    run_test "$CMD --sanitize-names --force-push --skip-invalid"
    ;;
  26)
    run_test "$CMD --dry-run --force-push --skip-invalid"
    ;;
  27)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push"
    ;;
  28)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --skip-invalid"
    ;;
  29)
    run_test "$CMD --since=\"-1h\" --sanitize-names --force-push --skip-invalid"
    ;;
  30)
    run_test "$CMD --since=\"-1h\" --dry-run --force-push --skip-invalid"
    ;;
  31)
    run_test "$CMD --sanitize-names --dry-run --force-push --skip-invalid"
    ;;
  32)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid"
    ;;
  33)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid --output-format=\"console\""
    ;;
  34)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid --output-format=\"json\""
    ;;
  35)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid --quiet"
    ;;
  36)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid --verbosity=\"debug\""
    ;;
  37)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid --verbosity=\"trace\" --verbosity-with-caller"
    ;;
  38)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid --config-file=\"custom.yaml\""
    ;;
  39)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid --config-file=\"custom.yaml\" --config-file-only"
    ;;
  40)
    run_test "$CMD --since=\"-1h\" --sanitize-names --dry-run --force-push --skip-invalid --config-file=\"custom.yaml\" --config-file-only --output-format=\"json\" --verbosity=\"trace\" --verbosity-with-caller"
    ;;
  *)
    printf "%s\n" "Invalid option. Press Enter to continue..."
    read -r
    ;;
  esac
done
