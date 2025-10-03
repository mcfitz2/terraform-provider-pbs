#!/bin/bash

# Pre-Publication Security Verification Script
# Run this before making the repository public to ensure no secrets are present

set -e

echo "🔍 Running security verification checks..."
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ISSUES_FOUND=0

# Check 1: Search for common secret patterns (excluding safe patterns)
echo "Check 1: Scanning for hardcoded secrets..."
SECRETS=$(git grep -iE 'password.*=.*["\x27][^"\x27]{3,}|secret.*=.*["\x27][^"\x27]{3,}|api[_-]?key.*=.*["\x27][^"\x27]{3,}' -- '*.go' '*.yml' '*.yaml' '*.tf' '*.sh' '*.env*' 2>/dev/null | grep -v 'example\|sample\|template\|pbspbs\|your-\|xxx\|os.Getenv\|backup-password' || true)
if [ -n "$SECRETS" ]; then
    echo -e "${RED}❌ Found potential hardcoded secrets:${NC}"
    echo "$SECRETS"
    ISSUES_FOUND=1
else
    echo -e "${GREEN}✅ No hardcoded secrets found${NC}"
fi
echo ""

# Check 2: Look for AWS keys (excluding documentation)
echo "Check 2: Scanning for AWS access keys..."
AWS_KEYS=$(git grep -E 'AKIA[0-9A-Z]{16}' 2>/dev/null | grep -v 'SECURITY_CLEANUP.md' || true)
if [ -n "$AWS_KEYS" ]; then
    echo -e "${RED}❌ Found AWS access key pattern:${NC}"
    echo "$AWS_KEYS"
    ISSUES_FOUND=1
else
    echo -e "${GREEN}✅ No AWS access keys found (excluding cleanup docs)${NC}"
fi
echo ""

# Check 3: Look for private keys
echo "Check 3: Scanning for private keys..."
if git grep -l 'BEGIN.*PRIVATE KEY' 2>/dev/null; then
    echo -e "${RED}❌ Found private key files${NC}"
    ISSUES_FOUND=1
else
    echo -e "${GREEN}✅ No private keys found${NC}"
fi
echo ""

# Check 4: Check for .env files
echo "Check 4: Checking for .env files..."
ENV_FILES=$(git ls-files | grep -E '\.env$|\.env\..*[^example]$' || true)
if [ -n "$ENV_FILES" ]; then
    echo -e "${RED}❌ Found .env files in git:${NC}"
    echo "$ENV_FILES"
    ISSUES_FOUND=1
else
    echo -e "${GREEN}✅ No .env files in git${NC}"
fi
echo ""

# Check 5: Verify .gitignore is comprehensive
echo "Check 5: Verifying .gitignore..."
if grep -q "\.env" .gitignore && grep -q "\*\.env" .gitignore; then
    echo -e "${GREEN}✅ .gitignore includes .env patterns${NC}"
else
    echo -e "${YELLOW}⚠️  .gitignore may not cover all .env files${NC}"
fi
echo ""

# Check 6: Check commit history length
echo "Check 6: Checking commit history..."
COMMIT_COUNT=$(git rev-list --count HEAD)
if [ "$COMMIT_COUNT" -eq 1 ]; then
    echo -e "${GREEN}✅ Single commit (clean history)${NC}"
else
    echo -e "${YELLOW}⚠️  Multiple commits ($COMMIT_COUNT) - history may contain secrets${NC}"
    echo "   Consider running ./cleanup-git-history.sh"
fi
echo ""

# Check 7: Verify README exists
echo "Check 7: Checking for README..."
if [ -f "README.md" ]; then
    echo -e "${GREEN}✅ README.md exists${NC}"
else
    echo -e "${YELLOW}⚠️  No README.md found${NC}"
fi
echo ""

# Check 8: Verify LICENSE exists
echo "Check 8: Checking for LICENSE..."
if [ -f "LICENSE" ]; then
    echo -e "${GREEN}✅ LICENSE file exists${NC}"
else
    echo -e "${YELLOW}⚠️  No LICENSE file found${NC}"
fi
echo ""

# Check 9: Look for TODO/FIXME comments with sensitive info
echo "Check 9: Scanning for sensitive TODOs..."
if git grep -iE 'TODO.*password|FIXME.*secret|TODO.*key' 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Found TODOs mentioning credentials${NC}"
else
    echo -e "${GREEN}✅ No sensitive TODOs found${NC}"
fi
echo ""

# Check 10: Verify GitHub workflows use secrets properly
echo "Check 10: Checking GitHub workflows..."
if git grep -l 'secrets\.' .github/workflows/ >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Workflows use GitHub Secrets${NC}"
    # Check for hardcoded values
    if git grep -E '(AWS_SECRET|B2_SECRET|SCALEWAY_SECRET).*:.*["\x27][A-Za-z0-9/+=]{20,}' .github/workflows/ 2>/dev/null; then
        echo -e "${RED}❌ Found hardcoded secrets in workflows${NC}"
        ISSUES_FOUND=1
    fi
else
    echo -e "${YELLOW}⚠️  No GitHub Secrets usage found in workflows${NC}"
fi
echo ""

# Final summary
echo "=========================================="
if [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${GREEN}✅ All critical checks passed!${NC}"
    echo ""
    echo "Repository appears ready for public release."
    echo ""
    echo "Final checklist:"
    echo "  [ ] All credentials from SECURITY_CLEANUP.md rotated"
    echo "  [ ] Git history cleaned (if needed)"
    echo "  [ ] README.md is complete and professional"
    echo "  [ ] LICENSE is appropriate for public use"
    echo "  [ ] GitHub Secrets configured for CI/CD"
    echo "  [ ] Repository description and topics set"
else
    echo -e "${RED}❌ Issues found! Do not make repository public yet.${NC}"
    echo ""
    echo "Fix the issues above before proceeding."
fi
echo "=========================================="
