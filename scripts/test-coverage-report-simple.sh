#!/bin/sh
# Test Coverage Report Generator for gimage (POSIX-compliant)
# Shows which MCP tools are exposed, which have tests, and end-to-end test status

set -e

echo ""
echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║                                                                               ║"
echo "║                    📊  GIMAGE TEST COVERAGE REPORT  📊                        ║"
echo "║                                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"
echo ""

# MCP Tools Analysis
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ MCP TOOLS STATUS                                                                │"
echo "├─────────────────────────────────────────────────────────────────────────────────┤"
echo "│ Tool              │ Exposed │ Unit Tests │ E2E Tests                            │"
echo "├─────────────────────────────────────────────────────────────────────────────────┤"

# Check each tool
check_tool() {
    tool_name=$1
    tool_file="internal/mcp/tools/${tool_name}.go"
    test_file="internal/mcp/tools/${tool_name}_test.go"

    # Check if exposed
    if [ -f "$tool_file" ]; then
        exposed="✅"
    else
        exposed="❌"
    fi

    # Check for unit tests
    if [ -f "$test_file" ]; then
        unit_tests="✅"
    else
        unit_tests="❌"
    fi

    # E2E tests (only for generate)
    if [ "$tool_name" = "generate" ]; then
        if [ -f "test/integration/generate_e2e_test.go" ]; then
            e2e_tests="✅"
        else
            e2e_tests="⚠️  MISSING"
        fi
    else
        e2e_tests="N/A"
    fi

    printf "│ %-17s │ %-7s │ %-10s │ %-36s │\n" "$tool_name" "$exposed" "$unit_tests" "$e2e_tests"
}

# List all tools
check_tool "generate"
check_tool "resize"
check_tool "scale"
check_tool "crop"
check_tool "compress"
check_tool "convert"
check_tool "batch"
check_tool "models"

echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""

# Count tools
total_tools=8
tools_with_tests=$(ls internal/mcp/tools/*_test.go 2>/dev/null | wc -l | tr -d ' ')

# Summary
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ COVERAGE SUMMARY                                                                │"
echo "├─────────────────────────────────────────────────────────────────────────────────┤"
printf "│ Total MCP Tools:           %-50s │\n" "$total_tools"
printf "│ Tools with Unit Tests:     %-50s │\n" "$tools_with_tests"

# Calculate percentage
if [ "$total_tools" -gt 0 ]; then
    coverage=$((tools_with_tests * 100 / total_tools))
    printf "│ Test Coverage:             %-50s │\n" "${coverage}%"
fi

echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""

# Run tests
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ RUNNING UNIT TESTS                                                              │"
echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""

go test -v -race ./internal/... ./test/... 2>&1 | grep -E "(PASS|FAIL|ok|coverage:|===)" || true

echo ""
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ END-TO-END TEST STATUS                                                          │"
echo "├─────────────────────────────────────────────────────────────────────────────────┤"

if [ -f "test/integration/generate_e2e_test.go" ]; then
    echo "│ ✅  E2E test file exists: test/integration/generate_e2e_test.go                │"
    echo "│                                                                                 │"
    echo "│ ⚠️  E2E tests require real API credentials and cost money                      │"
    echo "│     Run manually: make test-e2e                                                │"
    echo "│                                                                                 │"
    echo "│ APIs Tested:                                                                    │"

    if grep -q "Gemini" test/integration/generate_e2e_test.go 2>/dev/null; then
        echo "│   ✅  Gemini API                                                               │"
    else
        echo "│   ❌  Gemini API                                                               │"
    fi

    if grep -q "Vertex" test/integration/generate_e2e_test.go 2>/dev/null; then
        echo "│   ✅  Vertex AI                                                                │"
    else
        echo "│   ❌  Vertex AI                                                                │"
    fi

    if grep -q "Bedrock\|Nova" test/integration/generate_e2e_test.go 2>/dev/null; then
        echo "│   ✅  AWS Bedrock Nova Canvas                                                  │"
    else
        echo "│   ❌  AWS Bedrock Nova Canvas                                                  │"
    fi
else
    echo "│ ❌  E2E test file NOT FOUND: test/integration/generate_e2e_test.go             │"
    echo "│                                                                                 │"
    echo "│ Recommendation: Create E2E tests for:                                          │"
    echo "│   - Gemini API real image generation                                           │"
    echo "│   - Vertex AI real image generation                                            │"
    echo "│   - AWS Bedrock Nova Canvas real image generation                              │"
fi

echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""

echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║                            REPORT COMPLETE                                    ║"
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"
echo ""
