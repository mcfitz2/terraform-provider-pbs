# Terraform Provider for Proxmox Backup Server

[![License](https://img.shields.io/badge/License-MPL--2.0-blue.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/mcfitz2/terraform-provider-pbs)](https://goreportcard.com/report/github.com/mcfitz2/terraform-provider-pbs)

A Terraform provider for managing [Proxmox Backup Server](https://www.proxmox.com/en/proxmox-backup-server) resources.

## Features

- **Datastore Management**: Create and manage PBS datastores (directory, ZFS, CIFS)
- **S3 Provider Support**: Configure S3-compatible storage backends (AWS, Backblaze B2, Scaleway)
- **S3 Endpoints**: Manage S3 storage endpoints
- **Full Terraform Lifecycle**: Complete support for create, read, update, delete, and import operations

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- [Proxmox Backup Server](https://www.proxmox.com/en/proxmox-backup-server) >= 2.0

## Using the Provider

### From Terraform Registry

```hcl
terraform {
  required_providers {
    pbs = {
      source  = "mcfitz2/pbs"
      version = "~> 0.1"
    }
  }
}

provider "pbs" {
  address     = "https://pbs.example.com:8007"
  username    = "admin@pbs"
  password    = var.pbs_password  # Or use API token
  insecure    = false             # Set to true for self-signed certs
}
```

### Provider Configuration

The provider can be configured using:

1. **HCL Configuration** (as shown above)
2. **Environment Variables**:
   ```bash
   export PBS_ADDRESS="https://pbs.example.com:8007"
   export PBS_USERNAME="admin@pbs"
   export PBS_PASSWORD="your-password"
   export PBS_API_TOKEN="your-api-token"  # Alternative to password
   export PBS_INSECURE_TLS="false"
   ```

## Example Usage

### Create a Directory Datastore

```hcl
resource "pbs_datastore" "backup" {
  name    = "backup-store"
  path    = "/mnt/datastore/backup"
  comment = "Primary backup datastore"
}
```

### Create an S3 Datastore with AWS

```hcl
resource "pbs_s3_endpoint" "aws" {
  name              = "aws-s3"
  bucket            = "my-backup-bucket"
  region            = "us-east-1"
  access_key_id     = var.aws_access_key
  secret_access_key = var.aws_secret_key
  comment           = "AWS S3 storage"
}

resource "pbs_datastore" "s3_backup" {
  name    = "s3-backup"
  s3_endpoint = pbs_s3_endpoint.aws.name
  comment = "S3-backed datastore"
  
  depends_on = [pbs_s3_endpoint.aws]
}
```

### Create a ZFS Datastore

```hcl
resource "pbs_datastore" "zfs_backup" {
  name    = "zfs-backup"
  path    = "rpool/backup"
  comment = "ZFS-backed datastore"
}
```

## Available Resources

- `pbs_datastore` - Manage PBS datastores
- `pbs_s3_endpoint` - Manage S3 storage endpoints

For detailed documentation on each resource and data source, see the [Terraform Registry documentation](https://registry.terraform.io/providers/mcfitz2/pbs/latest/docs).

## Building from Source

```bash
# Clone the repository
git clone https://github.com/mcfitz2/terraform-provider-pbs.git
cd terraform-provider-pbs

# Build the provider
make build

# Install locally for testing
make install
```

## Development

### Running Tests

```bash
# Unit tests
make test-unit

# Integration tests (requires PBS instance)
export PBS_ADDRESS="https://pbs.local:8007"
export PBS_USERNAME="admin@pbs"
export PBS_PASSWORD="your-password"
export PBS_INSECURE_TLS="true"
make test
```

### Generating Documentation

```bash
make docs
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `make lint` and `make test`
5. Submit a pull request

## Security

This repository uses automated secret scanning with [gitleaks](https://github.com/gitleaks/gitleaks). A pre-commit hook prevents accidental credential commits.

To report security vulnerabilities, please email security contact or create a private security advisory on GitHub.

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Proxmox Backup Server](https://www.proxmox.com/en/proxmox-backup-server) team
- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
- All contributors to this project

## Support

- **Issues**: [GitHub Issues](https://github.com/mcfitz2/terraform-provider-pbs/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mcfitz2/terraform-provider-pbs/discussions)
- **Documentation**: [Terraform Registry](https://registry.terraform.io/providers/mcfitz2/pbs/latest/docs)
