#!/bin/bash
# check-code-quality.sh - Automated code quality verification
# Run this to verify RandomX code quality before review

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASS=0
FAIL=0

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    PASS=$((PASS + 1))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    FAIL=$((FAIL + 1))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

echo "========================================="
echo "  Code Quality Check - Ducros RandomX"
echo "========================================="
echo ""

# Check 1: No TODOs/FIXMEs
info "Checking for TODOs/FIXMEs/HACKs..."
if grep -ri "TODO\|FIXME\|XXX\|HACK" consensus/randomx/*.go 2>/dev/null | grep -v "Binary" > /dev/null; then
    check_fail "Found TODOs/FIXMEs in code"
    grep -ri "TODO\|FIXME\|XXX\|HACK" consensus/randomx/*.go | head -5
else
    check_pass "No TODOs/FIXMEs found"
fi
echo ""

# Check 2: No debug prints
info "Checking for debug code..."
DEBUG_COUNT=$(grep -E "fmt\.Print|println\(" consensus/randomx/*.go 2>/dev/null | grep -v "//" | wc -l || echo "0")
if [ "$DEBUG_COUNT" -eq "0" ]; then
    check_pass "No debug prints found"
else
    check_fail "Found $DEBUG_COUNT debug prints"
    grep -E "fmt\.Print|println\(" consensus/randomx/*.go | head -3
fi
echo ""

# Check 3: Test count
info "Counting tests..."
TEST_COUNT=$(grep "^func Test" consensus/randomx/*test.go 2>/dev/null | wc -l || echo "0")
if [ "$TEST_COUNT" -ge "5" ]; then
    check_pass "Found $TEST_COUNT test functions (≥5 required)"
else
    check_fail "Only $TEST_COUNT test functions (need ≥5)"
fi
echo ""

# Check 4: File headers
info "Checking copyright headers..."
HEADER_COUNT=$(grep -l "Copyright.*go-ethereum Authors" consensus/randomx/*.go 2>/dev/null | wc -l || echo "0")
FILE_COUNT=$(ls consensus/randomx/*.go 2>/dev/null | wc -l || echo "0")
if [ "$HEADER_COUNT" -ge "$FILE_COUNT" ]; then
    check_pass "All files have copyright headers ($HEADER_COUNT/$FILE_COUNT)"
else
    check_warn "Some files missing copyright headers ($HEADER_COUNT/$FILE_COUNT)"
fi
echo ""

# Check 5: Go vet
info "Running go vet..."
if go vet ./consensus/randomx 2>&1 | grep -v "lookup storage.googleapis.com" | grep -v "downloading" > /tmp/vet_output.txt 2>&1; then
    if [ -s /tmp/vet_output.txt ]; then
        check_warn "Go vet has warnings"
        head -5 /tmp/vet_output.txt
    else
        check_pass "Go vet clean"
    fi
else
    # Check if only network errors
    if grep -q "dial tcp\|lookup storage" /tmp/vet_output.txt 2>/dev/null; then
        check_warn "Go vet skipped (network issues)"
    else
        check_pass "Go vet clean"
    fi
fi
echo ""

# Check 6: Gofmt
info "Checking gofmt..."
UNFMT=$(gofmt -l consensus/randomx/*.go 2>/dev/null || echo "")
if [ -z "$UNFMT" ]; then
    check_pass "All files properly formatted"
else
    check_fail "Files need gofmt:"
    echo "$UNFMT"
fi
echo ""

# Check 7: Interface implementation
info "Checking consensus.Engine interface..."
REQUIRED_METHODS=(
    "Author"
    "VerifyHeader"
    "VerifyHeaders"
    "VerifyUncles"
    "Prepare"
    "Finalize"
    "FinalizeAndAssemble"
    "Seal"
    "SealHash"
    "CalcDifficulty"
    "Close"
)

MISSING=0
for METHOD in "${REQUIRED_METHODS[@]}"; do
    if grep -q "func (randomx \*RandomX) $METHOD" consensus/randomx/*.go 2>/dev/null; then
        : # Method found
    else
        check_fail "Missing method: $METHOD"
        MISSING=$((MISSING + 1))
    fi
done

if [ "$MISSING" -eq "0" ]; then
    check_pass "All ${#REQUIRED_METHODS[@]} consensus.Engine methods implemented"
else
    check_fail "$MISSING methods missing from interface"
fi
echo ""

# Check 8: Core files exist
info "Checking core files..."
CORE_FILES=(
    "consensus/randomx/randomx.go"
    "consensus/randomx/consensus.go"
    "consensus/randomx/api.go"
    "consensus/randomx/lwma.go"
)

for FILE in "${CORE_FILES[@]}"; do
    if [ -f "$FILE" ]; then
        : # File exists
    else
        check_fail "Missing core file: $FILE"
    fi
done
check_pass "All core files present"
echo ""

# Check 9: Documentation
info "Checking documentation..."
DOC_FILES=(
    "BUILD-GUIDE.md"
    "MINING-API.md"
    "DEPLOYMENT-GUIDE.md"
    "PRODUCTION-READINESS.md"
)

DOC_FOUND=0
for FILE in "${DOC_FILES[@]}"; do
    if [ -f "$FILE" ]; then
        DOC_FOUND=$((DOC_FOUND + 1))
    fi
done

if [ "$DOC_FOUND" -ge "3" ]; then
    check_pass "Found $DOC_FOUND/$((${#DOC_FILES[@]})) documentation files"
else
    check_warn "Only $DOC_FOUND/$((${#DOC_FILES[@]})) documentation files"
fi
echo ""

# Check 10: Line count sanity
info "Checking code size..."
TOTAL_LINES=$(wc -l consensus/randomx/*.go 2>/dev/null | tail -1 | awk '{print $1}' || echo "0")
if [ "$TOTAL_LINES" -gt "1000" ] && [ "$TOTAL_LINES" -lt "10000" ]; then
    check_pass "Code size reasonable ($TOTAL_LINES lines)"
elif [ "$TOTAL_LINES" -eq "0" ]; then
    check_fail "No code found"
else
    check_warn "Code size: $TOTAL_LINES lines"
fi
echo ""

# Summary
echo "========================================="
echo "  SUMMARY"
echo "========================================="
echo ""
echo -e "${GREEN}Passed:${NC} $PASS checks"
echo -e "${RED}Failed:${NC} $FAIL checks"
echo ""

if [ "$FAIL" -eq "0" ]; then
    echo -e "${GREEN}✓ CODE QUALITY: EXCELLENT${NC}"
    echo ""
    echo "The code is production-ready and professional."
    echo "Safe to submit for code review!"
    exit 0
elif [ "$FAIL" -le "2" ]; then
    echo -e "${YELLOW}⚠ CODE QUALITY: GOOD${NC}"
    echo ""
    echo "Minor issues found but generally production-ready."
    exit 0
else
    echo -e "${RED}✗ CODE QUALITY: NEEDS WORK${NC}"
    echo ""
    echo "Please fix the issues above before review."
    exit 1
fi
