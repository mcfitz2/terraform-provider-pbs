# Packer template to build PBS Vagrant box from ISO
# Adapted from https://github.com/rgl/proxmox-backup-server

packer {
  required_plugins {
    qemu = {
      version = "~> 1.0"
      source  = "github.com/hashicorp/qemu"
    }
  }
}

variable "pbs_version" {
  type    = string
  default = "3.4-1"
}

variable "iso_url" {
  type    = string
  default = "https://enterprise.proxmox.com/iso/proxmox-backup-server_3.4-1.iso"
}

variable "iso_checksum" {
  type    = string
  default = "sha256:ed4777f570f2589843765fff9e942288ff16a6cc3728655733899188479b7e08"
}

source "qemu" "pbs-test" {
  accelerator  = "kvm"
  machine_type = "q35"
  cpus         = 2
  memory       = 2048
  
  headless            = true
  use_default_display = false
  net_device          = "virtio-net"
  
  format              = "qcow2"
  disk_size           = "20G"
  disk_interface      = "virtio-scsi"
  disk_cache          = "unsafe"
  disk_discard        = "unmap"
  
  iso_url             = var.iso_url
  iso_checksum        = var.iso_checksum
  
  ssh_username        = "root"
  ssh_password        = "password123"
  ssh_timeout         = "60m"
  
  boot_wait           = "10s"
  
  # Automated PBS installation via boot commands
  # This simulates pressing keys to go through the graphical installer
  boot_command = [
    # Select "Install Proxmox Backup Server (Graphical)"
    "<enter>",
    # Wait for installer to boot
    "<wait3m>",
    # Accept license
    "<enter><wait>",
    # Select target disk (default /dev/sda)
    "<enter><wait>",
    # Country: United States
    "United S<wait>t<wait>a<wait>t<wait>e<wait>s<wait><enter><wait><tab><wait>",
    # Timezone: default
    "<tab><wait>",
    # Keyboard: default (US)
    "<tab><wait>",
    # Next button
    "<tab><wait>",
    # Go to password page
    "<enter><wait5>",
    # Root password
    "password123<tab><wait>",
    # Confirm password
    "password123<tab><wait>",
    # Email
    "admin@example.com<tab><wait>",
    # Next button
    "<tab><wait>",
    # Go to network page
    "<enter><wait5>",
    # Hostname: pbs-test
    "pbs-test<tab><wait>",
    # IP: use DHCP (default)
    "<tab><wait>",
    "<tab><wait>",
    "<tab><wait>",
    "<tab><wait>",
    # Next button
    "<tab><wait>",
    # Go to summary/install page
    "<enter><wait5>",
    # Install button
    "<enter><wait5>",
    # Wait for installation to complete
    "<wait4m>",
    # Login prompt appears - login as root
    "root<enter>",
    "<wait5>",
    # Enter password
    "password123<enter>",
    "<wait5>",
    # Disable enterprise repo (not accessible without subscription)
    "rm -f /etc/apt/sources.list.d/pbs-enterprise.list<enter>",
    # Add no-subscription repo
    "echo 'deb http://download.proxmox.com/debian/pbs bookworm pbs-no-subscription' > /etc/apt/sources.list.d/pbs.list<enter>",
    "apt-get update<enter><wait1m>",
  ]
  
  shutdown_command = "poweroff"
}

build {
  sources = ["source.qemu.pbs-test"]
  
  # Basic cleanup and optimization
  provisioner "shell" {
    inline = [
      "export DEBIAN_FRONTEND=noninteractive",
      # Install qemu-guest-agent for better VM management
      "apt-get install -y qemu-guest-agent",
      # Clean up
      "apt-get clean",
      "rm -rf /var/lib/apt/lists/*",
      # Zero free space for better compression
      "dd if=/dev/zero of=/EMPTY bs=1M || true",
      "sync",
      "rm -f /EMPTY",
    ]
  }
  
  # Create Vagrant box
  post-processor "vagrant" {
    output = "pbs-test.box"
    compression_level = 9
  }
}
