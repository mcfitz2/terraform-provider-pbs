#!/bin/bash

# Git History Cleanup Script
# This script creates a fresh git history without any previous commits
# Use this to remove all traces of committed secrets before making the repo public

set -e

echo "⚠️  WARNING: This will completely rewrite git history!"
echo "⚠️  All commit history will be lost permanently."
echo ""
echo "Before running this script, ensure:"
echo "  1. All credentials in SECURITY_CLEANUP.md have been rotated"
echo "  2. You have a backup of the repository"
echo "  3. All collaborators are aware of the history rewrite"
echo ""
read -p "Have you completed the checklist above? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted. Complete the checklist first."
    exit 1
fi

echo ""
echo "This will:"
echo "  1. Create a new orphan branch with clean history"
echo "  2. Commit all current files (minus ignored files)"
echo "  3. Replace main branch with the clean branch"
echo "  4. Delete all old history"
echo ""
read -p "Type 'DELETE_HISTORY' to proceed: " final_confirm

if [ "$final_confirm" != "DELETE_HISTORY" ]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo "Starting cleanup process..."
echo ""

# Step 1: Create a new orphan branch (no parent commits)
echo "Step 1/6: Creating orphan branch..."
git checkout --orphan clean-history

# Step 2: Add all files (respecting .gitignore)
echo "Step 2/6: Adding all files..."
git add -A

# Step 3: Create the initial commit
echo "Step 3/6: Creating initial commit..."
git commit -m "Initial commit - clean history

This repository has been reinitialized with a clean history to remove
sensitive data that was accidentally committed.

All previous commits have been removed. This is the first commit in the
new history.

terraform-provider-pbs - Terraform Provider for Proxmox Backup Server
License: MPL-2.0"

# Step 4: Delete the old main branch
echo "Step 4/6: Deleting old main branch..."
git branch -D main

# Step 5: Rename the clean branch to main
echo "Step 5/6: Renaming clean-history to main..."
git branch -m main

# Step 6: Garbage collect to remove old objects
echo "Step 6/6: Running garbage collection..."
git gc --aggressive --prune=now

echo ""
echo "✅ Git history cleanup complete!"
echo ""
echo "Repository statistics:"
git count-objects -vH
echo ""
echo "Next steps:"
echo "  1. Review the changes: git log"
echo "  2. Verify no secrets remain: git grep -i 'password\|secret\|key' || echo 'No secrets found'"
echo "  3. Force push to GitHub: git push --force origin main"
echo "  4. Update GitHub Secrets with rotated credentials"
echo ""
echo "⚠️  IMPORTANT: After force pushing, all collaborators must re-clone:"
echo "   git clone <repository-url>"
echo ""
