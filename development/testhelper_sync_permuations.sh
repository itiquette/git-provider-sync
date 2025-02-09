#!/bin/bash

# SPDX-FileCopyrightText: 2024 Josef Andersson
#
# SPDX-License-Identifier: EUPL-1.2

CMD="./dist/gitprovidersync-linux-amd64 sync"

# Colors for better readability
GREEN='\033[0;32m'
NC='\033[0m' # No Color
YELLOW='\033[1;33m'

# Function to run command and wait for user input
run_test() {
  echo -e "\n${GREEN}Running command:${NC}"
  echo -e "${YELLOW}$1${NC}"
  echo -e "\nPress Enter to execute, Ctrl+C to cancel..."
  read -r
  eval "$1"
  echo -e "\nPress Enter to return to menu..."
  read -r
}

while true; do
  clear
  echo "Git Provider Sync Test Menu - All Possible Flag Combinations"
  echo "========================================================"
  echo "Single Flags:"
  echo "1.  Basic sync (no flags)"
  echo "2.  Active from limit only"
  echo "3.  ASCII name only"
  echo "4.  Dry run only"
  echo "5.  Force push only"
  echo "6.  Ignore invalid name only"
  echo
  echo "Two Flag Combinations:"
  echo "7.  Active from limit + ASCII name"
  echo "8.  Active from limit + Dry run"
  echo "9.  Active from limit + Force push"
  echo "10. Active from limit + Ignore invalid name"
  echo "11. ASCII name + Dry run"
  echo "12. ASCII name + Force push"
  echo "13. ASCII name + Ignore invalid name"
  echo "14. Dry run + Force push"
  echo "15. Dry run + Ignore invalid name"
  echo "16. Force push + Ignore invalid name"
  echo
  echo "Three Flag Combinations:"
  echo "17. Active from limit + ASCII name + Dry run"
  echo "18. Active from limit + ASCII name + Force push"
  echo "19. Active from limit + ASCII name + Ignore invalid name"
  echo "20. Active from limit + Dry run + Force push"
  echo "21. Active from limit + Dry run + Ignore invalid name"
  echo "22. Active from limit + Force push + Ignore invalid name"
  echo "23. ASCII name + Dry run + Force push"
  echo "24. ASCII name + Dry run + Ignore invalid name"
  echo "25. ASCII name + Force push + Ignore invalid name"
  echo "26. Dry run + Force push + Ignore invalid name"
  echo
  echo "Four Flag Combinations:"
  echo "27. Active from limit + ASCII name + Dry run + Force push"
  echo "28. Active from limit + ASCII name + Dry run + Ignore invalid name"
  echo "29. Active from limit + ASCII name + Force push + Ignore invalid name"
  echo "30. Active from limit + Dry run + Force push + Ignore invalid name"
  echo "31. ASCII name + Dry run + Force push + Ignore invalid name"
  echo
  echo "All Flags:"
  echo "32. All flags combined"
  echo
  echo "Global Flag Combinations:"
  echo "33. All sync flags + Console output"
  echo "34. All sync flags + JSON output"
  echo "35. All sync flags + Quiet mode"
  echo "36. All sync flags + Debug verbosity"
  echo "37. All sync flags + Trace + Caller info"
  echo "38. All sync flags + Custom config"
  echo "39. All sync flags + Config file only"
  echo "40. Everything combined"
  echo
  echo "0. Exit"
  echo
  read -r -p "Select an option (0-40): " choice

  case $choice in
  0)
    echo "Exiting..."
    exit 0
    ;;
  1)
    run_test "$CMD"
    ;;
  2)
    run_test "$CMD --active-from-limit=\"-1h\""
    ;;
  3)
    run_test "$CMD --alphanumhyph-name"
    ;;
  4)
    run_test "$CMD --dry-run"
    ;;
  5)
    run_test "$CMD --force-push"
    ;;
  6)
    run_test "$CMD --ignore-invalid-name"
    ;;
  7)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name"
    ;;
  8)
    run_test "$CMD --active-from-limit=\"-1h\" --dry-run"
    ;;
  9)
    run_test "$CMD --active-from-limit=\"-1h\" --force-push"
    ;;
  10)
    run_test "$CMD --active-from-limit=\"-1h\" --ignore-invalid-name"
    ;;
  11)
    run_test "$CMD --alphanumhyph-name --dry-run"
    ;;
  12)
    run_test "$CMD --alphanumhyph-name --force-push"
    ;;
  13)
    run_test "$CMD --alphanumhyph-name --ignore-invalid-name"
    ;;
  14)
    run_test "$CMD --dry-run --force-push"
    ;;
  15)
    run_test "$CMD --dry-run --ignore-invalid-name"
    ;;
  16)
    run_test "$CMD --force-push --ignore-invalid-name"
    ;;
  17)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run"
    ;;
  18)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --force-push"
    ;;
  19)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --ignore-invalid-name"
    ;;
  20)
    run_test "$CMD --active-from-limit=\"-1h\" --dry-run --force-push"
    ;;
  21)
    run_test "$CMD --active-from-limit=\"-1h\" --dry-run --ignore-invalid-name"
    ;;
  22)
    run_test "$CMD --active-from-limit=\"-1h\" --force-push --ignore-invalid-name"
    ;;
  23)
    run_test "$CMD --alphanumhyph-name --dry-run --force-push"
    ;;
  24)
    run_test "$CMD --alphanumhyph-name --dry-run --ignore-invalid-name"
    ;;
  25)
    run_test "$CMD --alphanumhyph-name --force-push --ignore-invalid-name"
    ;;
  26)
    run_test "$CMD --dry-run --force-push --ignore-invalid-name"
    ;;
  27)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push"
    ;;
  28)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --ignore-invalid-name"
    ;;
  29)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --force-push --ignore-invalid-name"
    ;;
  30)
    run_test "$CMD --active-from-limit=\"-1h\" --dry-run --force-push --ignore-invalid-name"
    ;;
  31)
    run_test "$CMD --alphanumhyph-name --dry-run --force-push --ignore-invalid-name"
    ;;
  32)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name"
    ;;
  33)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name --output-format=\"console\""
    ;;
  34)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name --output-format=\"json\""
    ;;
  35)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name --quiet"
    ;;
  36)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name --verbosity=\"debug\""
    ;;
  37)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name --verbosity=\"trace\" --verbosity-with-caller"
    ;;
  38)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name --config-file=\"custom.yaml\""
    ;;
  39)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name --config-file=\"custom.yaml\" --config-file-only"
    ;;
  40)
    run_test "$CMD --active-from-limit=\"-1h\" --alphanumhyph-name --dry-run --force-push --ignore-invalid-name --config-file=\"custom.yaml\" --config-file-only --output-format=\"json\" --verbosity=\"trace\" --verbosity-with-caller"
    ;;
  *)
    echo "Invalid option. Press Enter to continue..."
    read -r
    ;;
  esac
done
