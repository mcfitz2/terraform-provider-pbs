# Integration Test Scripts

This directory contains scripts and configuration for running integration tests with Docker services.

## Quick Start

### Local Testing

Run all integration tests with Docker services:

```bash
./scripts/run-integration-tests.sh
```

Run a specific test:

```bash
./scripts/run-integration-tests.sh "TestIntegration/Metrics/InfluxDBHTTP"
```

### Prerequisites

- Docker and Docker Compose installed
- Go 1.22+ installed
- Terraform 1.5+ installed
- Access to a PBS 4.0 server (default: https://192.168.1.108:8007)

### Environment Variables

You can customize the PBS connection:

```bash
export PBS_ADDRESS="https://your-pbs-server:8007"
export PBS_USERNAME="root@pam"
export PBS_PASSWORD="your-password"
./scripts/run-integration-tests.sh
```

## Docker Services

The test infrastructure includes:

### InfluxDB (Port 8086)
- **Image**: `influxdb:2.7-alpine`
- **Purpose**: Testing metrics server configurations
- **Credentials**:
  - Username: `admin`
  - Password: `testpass123`
  - Organization: `testorg`
  - Bucket: `pbs-metrics`
  - Token: `test-token-for-pbs-provider`

### Gotify (Port 8080)
- **Image**: `gotify/server:latest`
- **Purpose**: Testing Gotify notification targets
- **Credentials**:
  - Username: `admin`
  - Password: `admin`

### Webhook Receiver (Port 8081)
- **Image**: `tarampampam/webhook-tester:latest`
- **Purpose**: Testing webhook notification targets
- **URL**: `http://localhost:8081/`

### NFS Server (Port 2049)
- **Image**: `erichough/nfs-server:latest`
- **Purpose**: Testing NFS datastore configurations
- **Export**: `/exports/test`

### Samba/CIFS (Port 445)
- **Image**: `dperson/samba:latest`
- **Purpose**: Testing CIFS datastore configurations
- **Credentials**:
  - Username: `testuser`
  - Password: `testpass`
  - Share: `testshare`

## Manual Docker Management

Start services only:
```bash
docker-compose -f scripts/docker-compose.test.yml up -d
```

Stop services:
```bash
docker-compose -f scripts/docker-compose.test.yml down
```

Stop services and remove volumes:
```bash
docker-compose -f scripts/docker-compose.test.yml down -v
```

View logs:
```bash
docker-compose -f scripts/docker-compose.test.yml logs -f
```

## GitHub Actions

The services are automatically started in GitHub Actions workflows using the `services` configuration. See `.github/workflows/vm-integration-tests.yml` for the configuration.

## Troubleshooting

### Services not starting

Check if ports are already in use:
```bash
lsof -i :8086  # InfluxDB
lsof -i :8080  # Gotify
lsof -i :8081  # Webhook
lsof -i :2049  # NFS
lsof -i :445   # CIFS
```

### Permission errors with NFS/CIFS

The NFS and CIFS containers require privileged mode for proper operation. On some systems, you may need to adjust Docker settings or run with elevated privileges.

### Tests failing with connection errors

Ensure all services are healthy before running tests:
```bash
docker-compose -f scripts/docker-compose.test.yml ps
```

All services should show "healthy" status.

## Test Configuration

Tests automatically detect the Docker services using environment variables:

- `TEST_INFLUXDB_HOST` - InfluxDB hostname (default: localhost)
- `TEST_INFLUXDB_PORT` - InfluxDB port (default: 8086)
- `TEST_GOTIFY_HOST` - Gotify hostname (default: localhost)
- `TEST_GOTIFY_PORT` - Gotify port (default: 8080)
- `TEST_WEBHOOK_HOST` - Webhook receiver hostname (default: localhost)
- `TEST_WEBHOOK_PORT` - Webhook receiver port (default: 8081)
- `TEST_NFS_HOST` - NFS server hostname (default: localhost)
- `TEST_CIFS_HOST` - CIFS server hostname (default: localhost)
