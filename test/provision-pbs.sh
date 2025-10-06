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

log_info "Starting PBS provisioning on Debian VM..."

# Update system
log_info "Updating system packages..."
export DEBIAN_FRONTEND=noninteractive

# Enable contrib repository for ZFS support
log_info "Enabling contrib repository for ZFS..."
sed -i 's/main$/main contrib/g' /etc/apt/sources.list

apt-get update -qq
apt-get upgrade -y -qq

# Install required packages (ZFS headers must match kernel)
log_info "Installing required packages (ZFS, curl, gnupg)..."

# Get current running kernel version
RUNNING_KERNEL=$(uname -r)
log_info "Running kernel: $RUNNING_KERNEL"

# Install base packages first
apt-get install -y -qq \
    curl \
    gnupg \
    lsb-release \
    ca-certificates \
    software-properties-common

# Install ZFS packages that match the current kernel
apt-get install -y -qq \
    linux-headers-${RUNNING_KERNEL} \
    zfsutils-linux

log_success "Base packages installed"

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

# Ensure ZFS modules are available for current running kernel
log_info "Ensuring ZFS modules match running kernel..."
RUNNING_KERNEL=$(uname -r)

# If modules aren't available, rebuild them
if ! modinfo zfs >/dev/null 2>&1; then
    log_warning "ZFS modules not found for kernel $RUNNING_KERNEL, attempting to build..."
    apt-get install -y -qq linux-headers-${RUNNING_KERNEL} dkms
    dpkg-reconfigure zfs-dkms || log_warning "Failed to rebuild ZFS modules"
fi

# Configure PBS admin user
log_info "Configuring PBS admin user..."

# Create admin user with password 'password123' (min 8 chars required)
# PBS uses PAM authentication, but we need to set up the PBS realm
proxmox-backup-manager user create admin@pbs --password password123 --email admin@example.com || true

# Make sure admin has proper permissions
proxmox-backup-manager acl update / Admin --auth-id admin@pbs || true

log_success "PBS admin user configured (admin@pbs / password123)"

# Load ZFS kernel modules
log_info "Loading ZFS kernel modules..."
modprobe zfs || log_warning "Failed to load ZFS modules, ZFS functionality may be limited"

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

if [ -n "$DISK" ]; then
    log_info "Found additional disk $DISK, creating ZFS pool 'testpool'..."
    
    # Create ZFS pool
    zpool create -f testpool $DISK || {
        log_warning "ZFS pool creation failed, checking if it already exists..."
        if zpool list testpool >/dev/null 2>&1; then
            log_info "ZFS pool 'testpool' already exists"
        else
            log_error "Failed to create ZFS pool"
            exit 1
        fi
    }
    
    # Set ZFS properties for PBS compatibility
    zfs set compression=lz4 testpool
    zfs set atime=off testpool
    
    log_success "ZFS pool 'testpool' created and configured"
    
    # Create a dataset for backups
    zfs create testpool/backup || {
        log_info "Dataset testpool/backup may already exist"
    }
    
    # Show ZFS status
    log_info "ZFS pool status:"
    zpool status testpool
    zfs list testpool
else
    log_warning "No additional disk found (/dev/vdb or /dev/sdb), ZFS tests will be limited"
    log_info "To enable full ZFS testing, ensure Vagrant creates a second disk"
fi

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
