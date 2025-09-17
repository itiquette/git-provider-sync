#!/bin/bash

# SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
#
# SPDX-License-Identifier: EUPL-1.2

# analyze_mock_usage.sh - Identifies tests that would benefit from memory filesystem conversion

echo "================================================================="
echo "Mock Usage Analysis for git-provider-sync"
echo "================================================================="
echo ""

# Function to calculate complexity score
calculate_score() {
    local file=$1
    local mock_on_count=$(grep -c "\.On(" "$file" 2>/dev/null || echo 0)
    local matched_by_count=$(grep -c "mock\.MatchedBy" "$file" 2>/dev/null || echo 0)
    local mock_objects=$(grep -c "Mock[A-Z].*{}" "$file" 2>/dev/null || echo 0)
    local assert_expect=$(grep -c "AssertExpectations" "$file" 2>/dev/null || echo 0)
    
    # Calculate score
    local score=$((mock_on_count + (matched_by_count * 3) + (mock_objects * 2) + assert_expect))
    echo $score
}

echo "=== High Priority Tests (Score > 15) ==="
echo "These tests would benefit most from memory filesystem conversion"
echo ""

high_priority=()
medium_priority=()
low_priority=()

# Find all test files
for file in $(find . -name "*_test.go" -path "*/internal/domain/sync/*" -o -name "*_test.go" -path "*/internal/integrationtest/*" 2>/dev/null); do
    # Skip files we've already converted
    if [[ "$file" == *"_improved_test.go" ]] || [[ "$file" == *"_memfs_test.go" ]] || [[ "$file" == *"comparison_example"* ]]; then
        continue
    fi
    
    score=$(calculate_score "$file")
    
    if [ "$score" -gt 15 ]; then
        high_priority+=("$score:$file")
    elif [ "$score" -gt 8 ]; then
        medium_priority+=("$score:$file")
    elif [ "$score" -gt 0 ]; then
        low_priority+=("$score:$file")
    fi
done

# Sort and display high priority
if [ ${#high_priority[@]} -gt 0 ]; then
    IFS=$'\n' sorted_high=($(sort -rn <<<"${high_priority[*]}"))
    for item in "${sorted_high[@]}"; do
        score=${item%%:*}
        file=${item#*:}
        filename=$(basename "$file")
        echo "📍 $filename (Score: $score)"
        echo "   Path: $file"
        
        # Show mock usage details
        mock_on=$(grep -c "\.On(" "$file" 2>/dev/null || echo 0)
        matched_by=$(grep -c "mock\.MatchedBy" "$file" 2>/dev/null || echo 0)
        
        echo "   - Mock expectations: $mock_on"
        echo "   - Complex matchers: $matched_by"
        echo "   - Estimated reduction: ~$(((score * 3)))% less code"
        echo ""
    done
else
    echo "   No high priority files found"
    echo ""
fi

echo "=== Medium Priority Tests (Score 8-15) ==="
if [ ${#medium_priority[@]} -gt 0 ]; then
    IFS=$'\n' sorted_medium=($(sort -rn <<<"${medium_priority[*]}"))
    for item in "${sorted_medium[@]}"; do
        score=${item%%:*}
        file=${item#*:}
        filename=$(basename "$file")
        echo "📌 $filename (Score: $score)"
    done
else
    echo "   No medium priority files found"
fi
echo ""

echo "=== Low Priority Tests (Score < 8) ==="
if [ ${#low_priority[@]} -gt 0 ]; then
    echo "   ${#low_priority[@]} files with minimal mocking (keep as is)"
else
    echo "   No low priority files found"
fi
echo ""

echo "=== Conversion Status Summary ==="
echo ""

# Count already converted files
converted_count=$(find . -name "*_improved_test.go" -o -name "*_memfs_test.go" | wc -l)
echo "✅ Already converted: $converted_count files"

# Show which patterns have been established
if [ -f "MEMFS_TESTING_PATTERN.md" ]; then
    echo "📚 Pattern documentation: MEMFS_TESTING_PATTERN.md ✓"
fi

if [ -f "comparison_example_test.go" ] || [ -f "internal/domain/sync/comparison_example_test.go" ]; then
    echo "🔍 Comparison examples: comparison_example_test.go ✓"
fi

echo ""
echo "=== Recommendations ==="
echo ""
echo "1. Start with HIGH priority files for maximum impact"
echo "2. Each conversion typically reduces test code by 60-80%"
echo "3. Use established patterns from fetch_improved_test.go"
echo "4. Keep provider mocks, only replace git/file operations"
echo ""

# Calculate total potential improvement
total_high=${#high_priority[@]}
total_medium=${#medium_priority[@]}
potential_loc_reduction=$((total_high * 70 + total_medium * 40))

echo "=== Potential Impact ==="
echo "Converting high priority files: ~$((total_high * 70)) lines reduced"
echo "Converting medium priority files: ~$((total_medium * 40)) lines reduced"
echo "Total potential reduction: ~$potential_loc_reduction lines of test code"
echo ""

echo "================================================================="
echo "Run this script periodically to track conversion progress"
echo "================================================================="