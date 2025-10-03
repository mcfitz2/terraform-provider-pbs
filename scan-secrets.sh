#!/bin/bash

# Comprehensive Secret Scanning Script using gitleaks
# This script scans the entire repository and git history for secrets

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Scanning repository for secrets using gitleaks...${NC}"
echo ""

# Check if gitleaks is installed
if ! command -v gitleaks &> /dev/null; then
    echo -e "${RED}❌ gitleaks is not installed${NC}"
    echo ""
    echo "Install it with:"
    echo "  macOS:    brew install gitleaks"
    echo "  Linux:    https://github.com/gitleaks/gitleaks#installing"
    echo "  Windows:  https://github.com/gitleaks/gitleaks#installing"
    exit 1
fi

echo "Gitleaks version:"
gitleaks version
echo ""

# Scan types
SCAN_TYPE="${1:-all}"

case $SCAN_TYPE in
    "history")
        echo -e "${BLUE}Scanning entire git history...${NC}"
        if gitleaks detect --config .gitleaks.toml --report-path gitleaks-report.json --redact --verbose; then
            echo -e "${GREEN}✅ No secrets found in git history${NC}"
            rm -f gitleaks-report.json
            exit 0
        else
            echo -e "${RED}❌ Secrets found in git history!${NC}"
            echo ""
            echo "Report saved to: gitleaks-report.json"
            echo ""
            echo "⚠️  CRITICAL: These secrets are in your git history."
            echo "You must run ./cleanup-git-history.sh to remove them."
            exit 1
        fi
        ;;
    
    "staged")
        echo -e "${BLUE}Scanning staged files...${NC}"
        if gitleaks protect --staged --config .gitleaks.toml --redact --verbose; then
            echo -e "${GREEN}✅ No secrets in staged files${NC}"
            exit 0
        else
            echo -e "${RED}❌ Secrets found in staged files!${NC}"
            exit 1
        fi
        ;;
    
    "uncommitted")
        echo -e "${BLUE}Scanning uncommitted changes...${NC}"
        if gitleaks protect --config .gitleaks.toml --redact --verbose; then
            echo -e "${GREEN}✅ No secrets in uncommitted changes${NC}"
            exit 0
        else
            echo -e "${RED}❌ Secrets found in uncommitted changes!${NC}"
            exit 1
        fi
        ;;
    
    "all"|*)
        echo -e "${BLUE}Running comprehensive scan...${NC}"
        echo ""
        
        # Scan git history
        echo -e "${YELLOW}1/2: Scanning git history...${NC}"
        if gitleaks detect --config .gitleaks.toml --report-path gitleaks-report.json --redact --verbose; then
            echo -e "${GREEN}✅ No secrets in git history${NC}"
            rm -f gitleaks-report.json
        else
            echo -e "${RED}❌ Secrets found in git history!${NC}"
            echo "   Report: gitleaks-report.json"
            HISTORY_FAIL=1
        fi
        echo ""
        
        # Scan uncommitted changes
        echo -e "${YELLOW}2/2: Scanning uncommitted changes...${NC}"
        if gitleaks protect --config .gitleaks.toml --redact --verbose; then
            echo -e "${GREEN}✅ No secrets in uncommitted changes${NC}"
        else
            echo -e "${RED}❌ Secrets in uncommitted changes!${NC}"
            UNCOMMITTED_FAIL=1
        fi
        echo ""
        
        # Summary
        echo "=========================================="
        if [ -n "$HISTORY_FAIL" ] || [ -n "$UNCOMMITTED_FAIL" ]; then
            echo -e "${RED}❌ Secret scan failed!${NC}"
            echo ""
            if [ -n "$HISTORY_FAIL" ]; then
                echo "  - Secrets found in git history"
                echo "    Action: Run ./cleanup-git-history.sh"
            fi
            if [ -n "$UNCOMMITTED_FAIL" ]; then
                echo "  - Secrets found in uncommitted files"
                echo "    Action: Remove secrets before committing"
            fi
            exit 1
        else
            echo -e "${GREEN}✅ All scans passed!${NC}"
            echo ""
            echo "Repository is clean of secrets."
        fi
        echo "=========================================="
        ;;
esac

echo ""
echo "Usage: $0 [history|staged|uncommitted|all]"
echo "  history      - Scan entire git history"
echo "  staged       - Scan only staged files"
echo "  uncommitted  - Scan all uncommitted changes"
echo "  all          - Comprehensive scan (default)"
