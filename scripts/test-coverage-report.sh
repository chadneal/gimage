#!/usr/bin/env bash
# Test Coverage Report Generator for gimage
# Shows which MCP tools are exposed, which have tests, and end-to-end test status

set -e

# Check bash version (need 4.0+ for associative arrays)
if [ "${BASH_VERSINFO[0]}" -lt 4 ]; then
    echo "This script requires bash 4.0 or higher"
    echo "On macOS, install with: brew install bash"
    echo ""
    echo "For now, running basic tests only..."
    echo ""
    go test -v -race ./...
    exit 0
fi

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║                                                                               ║"
echo "║                    📊  GIMAGE TEST COVERAGE REPORT  📊                        ║"
echo "║                                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Function to check if a file exists
file_exists() {
    [ -f "$1" ]
}

# Function to check if tests exist for a tool
has_tests() {
    local tool_name=$1
    local test_file="internal/mcp/tools/${tool_name}_test.go"
    file_exists "$test_file"
}

# Function to check if integration tests exist
has_integration_tests() {
    local tool_name=$1
    grep -q "$tool_name" test/integration/*_test.go 2>/dev/null
}

# MCP Tools Analysis
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ MCP TOOLS STATUS                                                                │"
echo "├─────────────────────────────────────────────────────────────────────────────────┤"
echo "│ Tool              │ Exposed │ Unit Tests │ Integration Tests │ E2E Tests       │"
echo "├─────────────────────────────────────────────────────────────────────────────────┤"

# Define all MCP tools
declare -A tools=(
    ["generate"]="internal/mcp/tools/generate.go"
    ["resize"]="internal/mcp/tools/resize.go"
    ["scale"]="internal/mcp/tools/scale.go"
    ["crop"]="internal/mcp/tools/crop.go"
    ["compress"]="internal/mcp/tools/compress.go"
    ["convert"]="internal/mcp/tools/convert.go"
    ["batch"]="internal/mcp/tools/batch.go"
    ["list_models"]="internal/mcp/tools/models.go"
)

total_tools=0
tools_with_unit_tests=0
tools_with_integration_tests=0
tools_with_e2e_tests=0

for tool in "${!tools[@]}"; do
    total_tools=$((total_tools + 1))

    # Check if tool is exposed (file exists)
    if file_exists "${tools[$tool]}"; then
        exposed="✅"
    else
        exposed="❌"
    fi

    # Check for unit tests
    if has_tests "$tool"; then
        unit_tests="✅"
        tools_with_unit_tests=$((tools_with_unit_tests + 1))
    else
        unit_tests="❌"
    fi

    # Check for integration tests
    if has_integration_tests "$tool"; then
        integration_tests="✅"
        tools_with_integration_tests=$((tools_with_integration_tests + 1))
    else
        integration_tests="❌"
    fi

    # E2E tests (only for generate tool with real APIs)
    if [ "$tool" = "generate" ]; then
        if grep -q "E2E\|e2e\|end-to-end" test/integration/*_test.go 2>/dev/null; then
            e2e_tests="✅"
            tools_with_e2e_tests=$((tools_with_e2e_tests + 1))
        else
            e2e_tests="⚠️  MISSING"
        fi
    else
        e2e_tests="N/A"
    fi

    printf "│ %-17s │ %-7s │ %-10s │ %-17s │ %-15s │\n" "$tool" "$exposed" "$unit_tests" "$integration_tests" "$e2e_tests"
done

echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""

# Summary Statistics
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ COVERAGE SUMMARY                                                                │"
echo "├─────────────────────────────────────────────────────────────────────────────────┤"
unit_coverage=$((tools_with_unit_tests * 100 / total_tools))
integration_coverage=$((tools_with_integration_tests * 100 / total_tools))
printf "│ Total MCP Tools:           %-50s │\n" "$total_tools"
printf "│ Tools with Unit Tests:     %-40s (%d%%) │\n" "$tools_with_unit_tests/$total_tools" "$unit_coverage"
printf "│ Tools with Integration:    %-40s (%d%%) │\n" "$tools_with_integration_tests/$total_tools" "$integration_coverage"
printf "│ Tools with E2E Tests:      %-50s │\n" "$tools_with_e2e_tests (generate only)"
echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""

# Run actual tests and show results
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ RUNNING TESTS                                                                   │"
echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""

# Run tests with coverage
echo -e "${BLUE}Running unit tests...${NC}"
go test -v -cover ./internal/... 2>&1 | grep -E "(PASS|FAIL|ok|coverage:)" || true
echo ""

echo -e "${BLUE}Running MCP integration tests...${NC}"
go test -v ./internal/mcp/... 2>&1 | grep -E "(PASS|FAIL|ok)" || true
echo ""

# Check for E2E tests
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ END-TO-END TEST STATUS                                                          │"
echo "├─────────────────────────────────────────────────────────────────────────────────┤"

if file_exists "test/integration/generate_e2e_test.go"; then
    echo "│ ✅  E2E test file exists: test/integration/generate_e2e_test.go                │"
    echo "│                                                                                 │"
    echo -e "│ ${YELLOW}⚠️  E2E tests require real API credentials and cost money${NC}                     │"
    echo "│     Run manually: make test-e2e                                                │"
    echo "│                                                                                 │"
    echo "│ APIs Tested:                                                                    │"
    if grep -q "Gemini" test/integration/generate_e2e_test.go 2>/dev/null; then
        echo "│   ✅  Gemini API                                                               │"
    fi
    if grep -q "Vertex" test/integration/generate_e2e_test.go 2>/dev/null; then
        echo "│   ✅  Vertex AI                                                                │"
    fi
    if grep -q "Bedrock\|Nova" test/integration/generate_e2e_test.go 2>/dev/null; then
        echo "│   ✅  AWS Bedrock Nova Canvas                                                  │"
    fi
else
    echo -e "│ ${RED}❌  E2E test file NOT FOUND: test/integration/generate_e2e_test.go${NC}            │"
    echo "│                                                                                 │"
    echo "│ Recommendation: Create E2E tests for:                                          │"
    echo "│   - Gemini API real image generation                                           │"
    echo "│   - Vertex AI real image generation                                            │"
    echo "│   - AWS Bedrock Nova Canvas real image generation                              │"
fi

echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""

# Test Coverage Details
echo "┌─────────────────────────────────────────────────────────────────────────────────┐"
echo "│ DETAILED COVERAGE BY PACKAGE                                                    │"
echo "└─────────────────────────────────────────────────────────────────────────────────┘"
echo ""
go test -cover ./... 2>&1 | grep -E "coverage:" || echo "No coverage data available"
echo ""

echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
echo "║                            REPORT COMPLETE                                    ║"
echo "╚═══════════════════════════════════════════════════════════════════════════════╝"
echo ""
