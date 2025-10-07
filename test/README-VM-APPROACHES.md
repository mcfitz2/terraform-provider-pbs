# PBS Testing VM Approaches

This directory contains multiple approaches for setting up a PBS (Proxmox Backup Server) VM for integration testing. Each has different tradeoffs.

## Current Approaches

### 1. Debian + PBS Installation (Current - Problematic)
**Files:** `Vagrantfile`, `provision-pbs.sh`  
**Status:** ⚠️ Failing due to kernel/ZFS issues

Uses `debian/bookworm64` box and installs PBS on top. Issues:
- Kernel mismatch between box and available headers
- Complex two-stage provisioning with reboot
- ZFS module build failures

**Backup:** `Vagrantfile.backup-debian`, `provision-pbs.sh.backup-debian`

### 2. PBS ISO Direct Build (Recommended)
**Files:** `pbs-box.pkr.hcl`, `Vagrantfile.pbs-box`, `Makefile.pbs-box`  
**Status:** ✅ Ready to test

Builds a custom Vagrant box from the official PBS ISO using Packer. This gives you a clean PBS installation with ZFS pre-configured.

**Advantages:**
- Official PBS installation (not retrofitted onto Debian)
- ZFS is installed correctly from the start
- No kernel version mismatches
- Matches your production setup (if using PBS ISO)
- Can specify exact PBS version

**Disadvantages:**
- One-time ~20-minute Packer build required
- Requires Packer installed locally or in CI
- ~1.5GB box file to store/cache

**Usage:**

```bash
# One-time setup (do this locally or in CI setup phase)
cd test
make -f Makefile.pbs-box build-pbs-box  # Takes ~20 mins
make -f Makefile.pbs-box add-pbs-box

# Then for testing (fast)
cp Vagrantfile.pbs-box Vagrantfile
vagrant up
# Run tests...
vagrant destroy
```

**For CI:**
```yaml
# In GitHub Actions workflow:
- name: Cache PBS Vagrant Box
  uses: actions/cache@v3
  with:
    path: test/pbs-test.box
    key: pbs-box-3.4-1-${{ hashFiles('test/pbs-box.pkr.hcl') }}

- name: Build PBS Box (if not cached)
  if: steps.cache.outputs.cache-hit != 'true'
  working-directory: ./test
  run: |
    make -f Makefile.pbs-box build-pbs-box

- name: Add PBS Box to Vagrant
  working-directory: ./test
  run: |
    make -f Makefile.pbs-box add-pbs-box
    
- name: Start PBS VM
  working-directory: ./test
  run: |
    cp Vagrantfile.pbs-box Vagrantfile
    vagrant up --provider libvirt
```

### 3. Pre-built PBS Vagrant Box from External Repo (✅ **CURRENT APPROACH**)
**Repository:** https://github.com/mcfitz2/proxmox-backup-server  
**Status:** ✅ In Production

Pre-built Vagrant boxes hosted on GitHub Releases. CI automatically discovers and tests against all available PBS versions in parallel.

**Advantages:**
- ✅ Instant VM startup in CI (just download)
- ✅ No build time or Packer dependency
- ✅ Supports testing multiple PBS versions in parallel (3.4, 4.0, etc.)
- ✅ Automatic version discovery from releases
- ✅ ZFS works perfectly (native PBS ISO-based boxes)
- ✅ Separate repo means boxes can be reused by other projects

**Disadvantages:**
- Depends on external repository for box availability
- Need to update external repo when new PBS versions release

## Comparison

| Approach | Setup Time | VM Start Time | ZFS Support | Reliability | Multi-Version | Maintenance |
|----------|------------|---------------|-------------|-------------|---------------|-------------|
| Debian + PBS | 0 min | 8-10 min | ❌ Broken | ❌ Poor | ❌ | 🔧 High |
| PBS ISO Build | ~20 min (once) | 2-3 min | ✅ Native | ✅ Excellent | ⚠️ Manual | ⚠️ Medium |
| Pre-built Box | 0 min | 2-3 min | ✅ Native | ✅ Excellent | ✅ Automatic | ✅ Low |

## Current Approach: Pre-built Boxes ✅

**We now use pre-built PBS Vagrant boxes from the external repository.**

### How it Works

1. **Box Repository**: https://github.com/mcfitz2/proxmox-backup-server
   - Contains Packer templates and build automation
   - Publishes pre-built boxes to GitHub Releases
   - Currently provides PBS 3.4 and PBS 4.0 boxes

2. **CI Workflow** (`.github/workflows/vm-integration-tests.yml`):
   - **Step 1**: Discover available PBS versions from latest release
   - **Step 2**: Create test matrix with all found versions
   - **Step 3**: Run tests in parallel for each PBS version
   - Each matrix job downloads its specific box and runs tests

3. **Local Testing**:
   ```bash
   # Download a specific PBS version box (libvirt provider for Linux)
   cd test
   wget https://github.com/mcfitz2/proxmox-backup-server/releases/latest/download/proxmox-backup-server-3.4-amd64-libvirt.box
   vagrant box add pbs-3.4 proxmox-backup-server-3.4-amd64-libvirt.box
   
   # Update Vagrantfile to use that box
   sed -i '' 's/config.vm.box = .*/config.vm.box = "pbs-3.4"/' Vagrantfile
   
   # Start VM and run tests
   vagrant up --provider virtualbox  # or libvirt on Linux
   # Run tests...
   vagrant destroy
   ```

### Benefits Achieved

✅ **Parallel Testing**: Automatically test against PBS 3.4 and 4.0 simultaneously  
✅ **No Build Time**: CI downloads pre-built boxes, starts instantly  
✅ **ZFS Native Support**: Boxes built from PBS ISO with ZFS working correctly  
✅ **Automatic Version Discovery**: Adding new PBS versions to external repo auto-updates CI  
✅ **Simplified Maintenance**: One repo for box building, one for provider testing  
✅ **Faster CI**: No 20-minute Packer builds, just download and test

## Files Overview

### Active Development
- `pbs-box.pkr.hcl` - Packer template to build PBS box from ISO
- `Vagrantfile.pbs-box` - Vagrantfile that uses the pre-built PBS box
- `Makefile.pbs-box` - Helper commands for building/managing PBS box

### Backup/Reference
- `Vagrantfile.backup-debian` - Original Debian-based approach
- `provision-pbs.sh.backup-debian` - Original provisioning script
- `Vagrantfile.iso` - Early experiment with ISO boot (incomplete)
- `http/pbs-answer.toml` - Answer file experiment (not used)

### Current (to be replaced)
- `Vagrantfile` - Current two-stage Debian approach (failing)
- `provision-pbs.sh` - Current provisioning script (complex)

## Migration History

1. ✅ **Backup current approach** (Done)
2. ✅ **Create PBS ISO build approach** (Done)
3. ✅ **External box repository created** (mcfitz2/proxmox-backup-server)
4. ✅ **Update CI workflow** (Done - parallel testing with version discovery)
5. ✅ **Replace Vagrantfile** (Done - simplified for pre-built boxes)
6. ⏳ **Test in CI and verify all tests pass**
7. ⏳ **Clean up obsolete files** (Packer templates, old provisioning scripts)

## References

- Based on https://github.com/rgl/proxmox-backup-server
- PBS ISO downloads: https://www.proxmox.com/en/downloads/proxmox-backup-server
- Current PBS version: 3.4-1
- ZFS support: Native in PBS ISO
