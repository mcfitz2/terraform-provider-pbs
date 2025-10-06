#!/bin/bash
# Provision Proxmox Backup Server on Debian with ZFS support

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Export for non-interactive mode
export DEBIAN_FRONTEND=noninteractive

# Check which stage we're in
STAGE_FILE="/var/lib/pbs-provision-stage"

if [ ! -f "$STAGE_FILE" ]; then
    #
    # STAGE 1: Upgrade kernel and prepare for reboot
    #
    log_info "=== STAGE 1: Upgrading kernel ==="
    
    RUNNING_KERNEL=$(uname -r)
    log_info "Current kernel: $RUNNING_KERNEL"
    
    # Enable contrib repository for ZFS support
    log_info "Enabling contrib repository..."
    sed -i 's/main$/main contrib/g' /etc/apt/sources.list
    
    # Update package lists
    log_info "Updating package lists..."
    apt-get update -qq
    
    # Upgrade kernel to latest available
    log_info "Upgrading kernel and headers..."
    apt-get install -y -qq linux-image-amd64 linux-headers-amd64
    
    # Mark that stage 1 is complete
    echo "1" > "$STAGE_FILE"
    
    log_success "Stage 1 complete - kernel upgraded"
    log_info "System will reboot, then continue with Stage 2..."
    
else
    #
    # STAGE 2: Install ZFS and PBS on upgraded kernel
    #
    log_info "=== STAGE 2: Installing ZFS and PBS ==="
    
    RUNNING_KERNEL=$(uname -r)
    log_info "Running kernel: $RUNNING_KERNEL"
    
    # Update package lists again
    log_info "Updating package lists..."
    apt-get update -qq
    
    # Install base packages
    log_info "Installing base packages..."
    apt-get install -y -qq \
        curl \
        gnupg \
        lsb-release \
        ca-certificates \
        software-properties-common \
        dkms
    
    # Verify we have headers for the running kernel
    log_info "Verifying kernel headers for $RUNNING_KERNEL..."
    if [ ! -d "/lib/modules/$RUNNING_KERNEL/build" ]; then
        log_error "Kernel headers not found for running kernel $RUNNING_KERNEL"
        log_error "Available kernel modules:"
        ls -1 /lib/modules/
        exit 1
    fi
    log_success "Kernel headers present for $RUNNING_KERNEL"
    
    # Install ZFS which will use DKMS to build modules for running kernel
    log_info "Installing ZFS packages (this may take a few minutes)..."
    apt-get install -y -qq zfs-dkms zfsutils-linux
    
    # Get ZFS version
    ZFS_VERSION=$(dkms status zfs | head -1 | cut -d',' -f1 | cut -d'/' -f2 || echo "")
    if [ -z "$ZFS_VERSION" ]; then
        log_error "Could not determine ZFS version from DKMS"
        exit 1
    fi
    log_info "ZFS version: $ZFS_VERSION"
    
    # Check DKMS status
    log_info "Checking DKMS build status:"
    dkms status zfs
    
    # Verify ZFS modules were built for running kernel
    log_info "Verifying ZFS modules for running kernel $RUNNING_KERNEL..."
    if ! find /lib/modules/$RUNNING_KERNEL -name "zfs.ko*" 2>/dev/null | grep -q .; then
        log_error "ZFS modules were not built for running kernel $RUNNING_KERNEL"
        log_error "DKMS status:"
        dkms status
        log_error "Available kernel modules:"
        find /lib/modules -name "zfs.ko*" 2>/dev/null || echo "No ZFS modules found"
        exit 1
    fi
    log_success "ZFS modules found for kernel $RUNNING_KERNEL"
    
    # Load ZFS modules
    log_info "Loading ZFS kernel modules..."
    if ! modprobe zfs 2>/dev/null; then
        log_error "Failed to load ZFS kernel modules"
        log_error "Module info:"
        modinfo zfs 2>&1 || echo "modinfo failed"
        log_error "Kernel messages:"
        dmesg | tail -20
        exit 1
    fi
    log_success "ZFS kernel modules loaded successfully"
    
    # Verify ZFS is functional
    log_info "Verifying ZFS is functional..."
    zpool version >/dev/null 2>&1 || {
        log_error "ZFS tools not working correctly"
        exit 1
    }
    log_success "ZFS is fully functional"
    
    log_success "ZFS packages installed and verified"
    
    # Upgrade the rest of the system
    log_info "Upgrading remaining system packages..."
    apt-get upgrade -y -qq
    
    log_success "Base packages installed"
fi

# Only continue with PBS installation in stage 2
if [ -f "$STAGE_FILE" ]; then

# Add Proxmox repository
log_info "Adding Proxmox repository..."
echo "deb http://download.proxmox.com/debian/pbs $(lsb_release -sc) pbs-no-subscription" > /etc/apt/sources.list.d/pbs.list

# Add Proxmox repository key
log_info "Adding Proxmox GPG key..."
curl -fsSL https://enterprise.proxmox.com/debian/proxmox-release-bookworm.gpg -o /etc/apt/trusted.gpg.d/proxmox-release-bookworm.gpg

# Update package list
log_info "Updating package list with Proxmox repository..."
apt-get update -qq

# Install Proxmox Backup Server
log_info "Installing Proxmox Backup Server (this may take several minutes)..."
apt-get install -y -qq proxmox-backup-server

log_success "Proxmox Backup Server installed"

# Configure PBS admin user
log_info "Configuring PBS admin user..."

# Create admin user with password 'password123' (min 8 chars required)
# PBS uses PAM authentication, but we need to set up the PBS realm
proxmox-backup-manager user create admin@pbs --password password123 --email admin@example.com || true

# Make sure admin has proper permissions
proxmox-backup-manager acl update / Admin --auth-id admin@pbs || true

log_success "PBS admin user configured (admin@pbs / password123)"

# Set up ZFS pool for testing
log_info "Setting up ZFS pool for testing..."

# Check if additional disk exists (should be /dev/vdb for libvirt or /dev/sdb for virtualbox)
if [ -b /dev/vdb ]; then
    DISK="/dev/vdb"
elif [ -b /dev/sdb ]; then
    DISK="/dev/sdb"
else
    DISK=""
fi

if [ -z "$DISK" ]; then
    log_error "No additional disk found for ZFS pool!"
    log_error "Expected /dev/vdb (libvirt) or /dev/sdb (virtualbox)"
    log_error "Available disks:"
    lsblk
    exit 1
fi

log_info "Found additional disk $DISK, creating ZFS pool 'testpool'..."

# Check if pool already exists
if zpool list testpool >/dev/null 2>&1; then
    log_info "ZFS pool 'testpool' already exists"
    zpool status testpool
else
    # Create ZFS pool
    if ! zpool create -f testpool $DISK; then
        log_error "Failed to create ZFS pool 'testpool' on $DISK"
        log_error "Disk information:"
        lsblk $DISK
        log_error "ZFS status:"
        zpool status 2>&1 || echo "No pools exist"
        exit 1
    fi
    log_success "ZFS pool 'testpool' created successfully"
fi

# Set ZFS properties for PBS compatibility
log_info "Configuring ZFS pool properties..."
zfs set compression=lz4 testpool
zfs set atime=off testpool
log_success "ZFS pool configured (compression=lz4, atime=off)"

# Create a dataset for backups
zfs create testpool/backup 2>/dev/null || log_info "Dataset testpool/backup already exists"

# Show ZFS status
log_info "Final ZFS configuration:"
zpool status testpool
zfs list -r testpool

# Create default datastore directory
log_info "Creating default datastore directories..."
mkdir -p /var/lib/proxmox-backup/datastore
chmod 700 /var/lib/proxmox-backup/datastore

# Start and enable PBS services
log_info "Starting PBS services..."
systemctl enable proxmox-backup.service
systemctl enable proxmox-backup-proxy.service
systemctl restart proxmox-backup.service
systemctl restart proxmox-backup-proxy.service

# Wait for services to be ready
log_info "Waiting for PBS services to start..."
sleep 5

# Check service status
if systemctl is-active --quiet proxmox-backup.service && \
   systemctl is-active --quiet proxmox-backup-proxy.service; then
    log_success "PBS services are running"
else
    log_error "PBS services failed to start"
    systemctl status proxmox-backup.service --no-pager
    systemctl status proxmox-backup-proxy.service --no-pager
    exit 1
fi

# Wait for API to be responsive
log_info "Waiting for PBS API to become responsive..."
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if curl -k -f -s https://localhost:8007 >/dev/null 2>&1; then
        log_success "PBS API is responsive!"
        break
    fi
    attempt=$((attempt + 1))
    if [ $attempt -eq $max_attempts ]; then
        log_error "PBS API did not become responsive after $max_attempts attempts"
        exit 1
    fi
    sleep 2
done

# Display system information
log_info "=== PBS VM Information ==="
echo "Hostname: $(hostname)"
echo "IP Address: $(hostname -I | awk '{print $1}')"
echo "PBS Version: $(proxmox-backup-manager version --verbose)"
echo ""
echo "ZFS Pools:"
zpool list || echo "No ZFS pools configured"
echo ""
echo "PBS Access:"
echo "  URL: https://localhost:8007"
echo "  Username: admin@pbs"
echo "  Password: password123"
echo "========================="

log_success "PBS provisioning completed successfully!"
