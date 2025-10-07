# Files to Clean Up

The following files are now obsolete after switching to the pre-built PBS box approach:

## Obsolete Files (Can be Deleted)

### Packer Build Infrastructure
- **`pbs-box.pkr.hcl`** - Packer template for building PBS boxes locally
  - No longer needed - boxes are pre-built in external repo
  
- **`Makefile.pbs-box`** - Make targets for Packer builds
  - No longer needed - CI downloads pre-built boxes

- **`pbs-box.log`** - Build log from failed Packer attempt
  - Can be deleted

### Old Debian-Based Provisioning
- **`provision-pbs.sh`** - Two-stage provisioning script for Debian + PBS
  - No longer needed - pre-built boxes come with PBS installed
  
- **`provision-pbs.sh.backup-debian`** - Backup of original provisioning script
  - Keep as reference or delete

- **`Vagrantfile.backup-debian`** - Backup of Debian-based Vagrantfile
  - Keep as reference or delete

### Experimental Files
- **`Vagrantfile.iso`** - Early experiment with ISO boot
  - Can be deleted (incomplete/non-functional)

- **`Vagrantfile.pbs-box`** - Vagrantfile for locally-built Packer boxes
  - No longer needed - main Vagrantfile updated for pre-built boxes

- **`http/pbs-answer.toml`** - Answer file experiment
  - Can be deleted (not used)

- **`http/` directory** - May be empty after removing answer file
  - Delete if empty

## Files to Keep

### Active Testing Infrastructure
- **`Vagrantfile`** - ✅ Updated for pre-built boxes
- **`docker-compose.yml`** - For Docker-based PBS testing
- **`run_docker_tests.sh`** - Docker test runner
- **`run_integration_tests.sh`** - General integration test runner
- **`run_vagrant_tests.sh`** - Vagrant test runner

### Documentation
- **`README.md`** - Main test documentation
- **`README-VM-APPROACHES.md`** - ✅ Updated with current approach
- **`TESTS.md`** - Test documentation
- **`CLEANUP-NOTES.md`** - This file

### Test Code
- **`integration/`** - Integration test code
- **`unit/`** - Unit test code

### Configuration
- **`config.env.example`** - Example configuration
- **`Makefile`** - Test automation

## Recommended Cleanup Commands

```bash
cd test

# Remove obsolete Packer/provisioning files
rm -f pbs-box.pkr.hcl
rm -f Makefile.pbs-box
rm -f pbs-box.log
rm -f provision-pbs.sh
rm -f provision-pbs.sh.backup-debian

# Remove backup/experimental Vagrantfiles
rm -f Vagrantfile.backup-debian
rm -f Vagrantfile.iso
rm -f Vagrantfile.pbs-box

# Remove experimental files
rm -f http/pbs-answer.toml
rmdir http 2>/dev/null || true  # Remove if empty

# Commit cleanup
cd ..
git add -A
git commit -m "Clean up obsolete PBS provisioning and Packer files

Removed files that are no longer needed after switching to 
pre-built PBS boxes from external repository."
git push
```

## Post-Cleanup File Structure

After cleanup, the test directory should contain:

```
test/
├── docker-compose.yml          # Docker testing
├── run_docker_tests.sh
├── run_integration_tests.sh
├── run_vagrant_tests.sh
├── config.env.example
├── Makefile
├── Vagrantfile                 # Pre-built box approach
├── README.md
├── README-VM-APPROACHES.md
├── TESTS.md
├── integration/                # Test code
└── unit/                       # Test code
```

Clean, focused, and maintainable! 🎉
