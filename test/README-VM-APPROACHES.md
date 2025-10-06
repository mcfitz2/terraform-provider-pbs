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

### 3. Pre-built PBS Vagrant Box (Future)
**Status:** 🔮 Proposed

If we build the PBS box and upload it to Vagrant Cloud or GitHub Releases, CI can just download it directly.

**Advantages:**
- Instant VM startup in CI
- No Packer build time
- Can be shared across projects

**Disadvantages:**
- Need hosting for ~1.5GB box file
- Manual updates when PBS releases new versions
- Trust/security considerations

## Comparison

| Approach | Setup Time | VM Start Time | ZFS Support | Reliability | Maintenance |
|----------|------------|---------------|-------------|-------------|-------------|
| Debian + PBS | 0 min | 8-10 min | ❌ Broken | ❌ Poor | 🔧 High |
| PBS ISO Build | ~20 min (once) | 2-3 min | ✅ Native | ✅ Excellent | ✅ Low |
| Pre-built Box | 0 min | 2-3 min | ✅ Native | ✅ Excellent | ⚠️ Medium |

## Recommendation

**For immediate use:** Switch to **PBS ISO Direct Build** approach.

1. Build the box once locally: `make -f Makefile.pbs-box build-pbs-box`
2. Add it to vagrant: `make -f Makefile.pbs-box add-pbs-box`
3. Update CI to use the new Vagrantfile: `cp Vagrantfile.pbs-box Vagrantfile`

**For production CI:** Set up GitHub Actions caching for the PBS box:
- First run builds it (~20 min)
- Subsequent runs use cached box (instant)
- Cache invalidates when PBS version changes

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

## Migration Path

1. ✅ **Backup current approach** (Done)
2. ✅ **Create PBS ISO build approach** (Done)
3. ⏳ **Test PBS ISO approach locally**
4. ⏳ **Update CI workflow**
5. ⏳ **Replace Vagrantfile**
6. ⏳ **Verify all tests pass**
7. ⏳ **Clean up old files**

## References

- Based on https://github.com/rgl/proxmox-backup-server
- PBS ISO downloads: https://www.proxmox.com/en/downloads/proxmox-backup-server
- Current PBS version: 3.4-1
- ZFS support: Native in PBS ISO
