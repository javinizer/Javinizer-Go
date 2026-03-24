# Docker Deployment Guide

This guide explains how to deploy Javinizer using Docker and Docker Compose.

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Building the Image](#building-the-image)
- [Running with Docker Compose](#running-with-docker-compose)
- [Volume Structure](#volume-structure)
- [Configuration](#configuration)
- [Development Mode](#development-mode)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

The fastest way to get Javinizer running:

```bash
# 1. Clone the repository
git clone https://github.com/javinizer/javinizer-go.git
cd javinizer-go

# 2. Configure environment variables
cp .env.example .env
# Edit .env to set your USER_ID, GROUP_ID, and MEDIA_PATH

# 3. Copy the Docker Compose template
cp docker-compose.yml.example docker-compose.yml

# 4. Build the Docker image
docker build -t javinizer:latest .

# 5. Run with Docker Compose
docker-compose up -d

# 6. Access the web UI
open http://localhost:8080
```

---

## Prerequisites

- **Docker**: 20.10+ ([Install Docker](https://docs.docker.com/get-docker/))
- **Docker Compose**: 2.0+ (included with Docker Desktop)
- **Disk Space**: ~500MB for image + your JAV library

---

## Building the Image

### Local Build

Build the image locally with version information:

```bash
docker build \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S') \
  -t javinizer:latest \
  .
```

### Build Process

The Dockerfile uses a multi-stage build:

1. **Stage 1 (frontend-builder)**: Builds the SvelteKit frontend with Node.js 20
2. **Stage 2 (go-builder)**: Compiles the Go binary with SQLite support
3. **Stage 3 (runtime)**: Creates a minimal Alpine runtime image (~80MB)

**Build time**: ~2-3 minutes on modern hardware

---

## Running with Docker Compose

### Basic Usage

Copy the example file and customize as needed:

```bash
cp docker-compose.yml.example docker-compose.yml
```

The `docker-compose.yml` provides a production-ready setup:

```bash
# Start the container
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the container
docker-compose down

# Restart the container
docker-compose restart
```

### Updating the Application

```bash
# Pull latest changes
git pull

# Rebuild the image
docker-compose build

# Restart with new image
docker-compose up -d
```

---

## Configuration with .env File

Javinizer uses a `.env` file to configure Docker Compose variables. This makes it easy to customize your deployment without editing `docker-compose.yml`.

### Setup

1. **Copy the example files**:
   ```bash
   cp .env.example .env
   cp docker-compose.yml.example docker-compose.yml
   ```

2. **Edit `.env` with your settings**:
   ```bash
   # Required: Match container user to your host user (prevents permission issues)
   USER_ID=1000        # Run: id -u
   GROUP_ID=1000       # Run: id -g

   # Required: Set your JAV library path
   MEDIA_PATH=/Users/you/JAV

   # Optional: Change port if 8080 is in use
   HOST_PORT=8080

   # Optional: Set your timezone
   TZ=America/New_York
   ```

3. **Start the container**:
   ```bash
   docker-compose up -d
   ```

### Available Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `USER_ID` | User ID for container (run `id -u`) | 1000 | Recommended |
| `GROUP_ID` | Group ID for container (run `id -g`) | 1000 | Recommended |
| `MEDIA_PATH` | Path to your JAV library on host | `/path/to/your/jav-library` | Yes |
| `HOST_PORT` | Port to expose on host | 8080 | No |
| `TZ` | Timezone (IANA format) | UTC | No |

### Alternative: Command-Line Variables

You can also set variables on the command line (overrides `.env`):

```bash
USER_ID=$(id -u) GROUP_ID=$(id -g) docker-compose up -d
```

---

## Volume Structure

Javinizer uses a **2-volume architecture**:

### Volume 1: Application State (`/javinizer`)

Contains all application data:
- `config.yaml` - Configuration file
- `javinizer.db` - SQLite database (cached metadata)
- `logs/` - Application logs
- `cache/` - Temporary cache files

**Host mount**: `./data:/javinizer`

### Volume 2: Media Files (`/media`)

Your JAV library for scanning and organizing:
- Read-write access required for organize operations
- Can be any directory on your host

**Host mount**: `/path/to/your/jav-library:/media`

### Example Directory Structure

```
/Users/you/
├── javinizer-go/          # Project root
│   └── data/              # Application state (created by Docker)
│       ├── config.yaml
│       ├── javinizer.db
│       └── logs/
└── JAV/                   # Your media library
    ├── IPX-123.mp4
    └── ABW-456.mkv
```

**docker-compose.yml**:
```yaml
volumes:
  - ./data:/javinizer                         # Application state (relative path)
  - /Users/you/JAV:/media                     # Media files (absolute path)
```

---

## Configuration

### Environment Variables

Configure via the `.env` file (recommended) or directly in `docker-compose.yml`:

```yaml
environment:
  # Application home directory
  - JAVINIZER_HOME=/javinizer

  # Configuration file location
  - JAVINIZER_CONFIG=/javinizer/config.yaml

  # Database location
  - JAVINIZER_DB=/javinizer/javinizer.db

  # Log directory
  - JAVINIZER_LOG_DIR=/javinizer/logs

  # Timezone (affects log timestamps)
  - TZ=America/New_York
```

`JAVINIZER_LOG_DIR` only rewrites file paths already configured in `logging.output`.
If `logging.output` is only `stdout`/`stderr`, no log file will be created.

Examples:
- `logging.output: stdout` + `JAVINIZER_LOG_DIR=/javinizer/logs` -> container logs only
- `logging.output: "stdout,data/logs/javinizer.log"` + `JAVINIZER_LOG_DIR=/javinizer/logs` -> container logs + `/javinizer/logs/javinizer.log`

**Note**: The media directory mounted at `/media` is automatically detected and added to allowed directories. No additional configuration needed.

### Configuration File

Edit `./data/config.yaml` on the host:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

scrapers:
  priority: ["r18dev", "dmm"]

  # Proxy configuration (optional)
  proxy:
    enabled: true
    url: "http://proxy.example.com:8080"

output:
  organize_directory: "/media/organized"
  folder_template: "<ID> [<STUDIO>] - <TITLE> (<YEAR>)"
```

**Changes take effect** after restarting the container:
```bash
docker-compose restart
```

### Port Mapping

To use a different port, set `HOST_PORT` in `.env`:

```bash
HOST_PORT=9090
```

Or edit `docker-compose.yml` directly:

```yaml
ports:
  - "9090:8080"  # Access at http://localhost:9090
```

---

## Development Mode

For frontend development with live reload:

```bash
# Start development container
docker-compose --profile dev up

# Changes to web/frontend/ will trigger hot reload
```

This mounts the frontend source directory into the container for live development.

---

## Troubleshooting

### Container Won't Start

Check logs:
```bash
docker-compose logs
```

Common issues:
- **Port 8080 in use**: Set `HOST_PORT=9090` in `.env` file
- **Permission denied**: Ensure the `./data` directory is writable and check `USER_ID`/`GROUP_ID` in `.env`
- **Volume mount failed**: Check that `MEDIA_PATH` in `.env` points to an existing directory

### Health Check Failing

The health check endpoint is `/health`. Test manually:
```bash
curl http://localhost:8080/health
```

### Database Locked

If SQLite database is locked:
```bash
# Stop container
docker-compose down

# Remove lock file
rm ./data/javinizer.db-shm ./data/javinizer.db-wal

# Restart
docker-compose up -d
```

### Viewing Container Internals

```bash
# Enter the running container
docker-compose exec javinizer sh

# Check binary version
javinizer --version

# Check file permissions
ls -la /javinizer

# Check running processes
ps aux
```

### Reset Application State

To start fresh (⚠️ **deletes all cached metadata**):
```bash
# Stop container
docker-compose down

# Remove application state
rm -rf ./data

# Restart (will create fresh state)
docker-compose up -d
```

---

## Docker Commands Reference

### Image Management

```bash
# List images
docker images javinizer

# Remove old images
docker rmi javinizer:old-tag

# Prune unused images
docker image prune
```

### Container Management

```bash
# List running containers
docker ps

# View container resource usage
docker stats javinizer

# View container filesystem changes
docker diff javinizer

# Export container filesystem
docker export javinizer > javinizer-backup.tar
```

### Logs and Debugging

```bash
# Follow logs in real-time
docker-compose logs -f --tail=100

# View logs for specific timeframe
docker-compose logs --since 30m

# View resource usage
docker-compose stats

# Inspect container
docker inspect javinizer
```

---

## Security Considerations

### Running as Non-Root User

The container runs as user `javinizer` for security. By default, the user is created with UID 1000 and GID 1000, but this can be customized to match your host user.

**Recommended method**: Use the `.env` file (see [Configuration with .env File](#configuration-with-env-file)):
```bash
# In .env file:
USER_ID=1000   # Get with: id -u
GROUP_ID=1000  # Get with: id -g
```

**Alternative**: Set via command line:
```bash
USER_ID=$(id -u) GROUP_ID=$(id -g) docker-compose up -d
```

**Why this matters**: Matching the container UID/GID to your host user prevents permission issues when the container writes to mounted volumes (`./data` and `/media`). Without this, you may see "permission denied" errors or files owned by the wrong user.

### Network Security

The default configuration binds to `0.0.0.0:8080` (all interfaces). For production:

1. **Use a reverse proxy** (nginx, Caddy) with HTTPS
2. **Restrict binding** to localhost:
   ```yaml
   ports:
     - "127.0.0.1:8080:8080"
   ```
3. **Complete first-run authentication setup** in Web UI (built-in single-user auth)

---

## Production Deployment

### Recommended Setup

```yaml
services:
  javinizer:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        - USER_ID=${USER_ID:-1000}   # Match host user for permissions
        - GROUP_ID=${GROUP_ID:-1000}
    image: javinizer:latest
    container_name: javinizer
    restart: unless-stopped
    user: "${USER_ID:-1000}:${GROUP_ID:-1000}"

    ports:
      - "127.0.0.1:8080:8080"  # Localhost only

    volumes:
      - ./data:/javinizer
      - /mnt/media/jav:/media  # Read-write for organize operations

    environment:
      - TZ=UTC
      - JAVINIZER_LOG_LEVEL=info  # Reduce log verbosity

    healthcheck:
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s

    # Resource limits
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 256M
```

**Usage**:
```bash
# Set user/group to match your host user
USER_ID=$(id -u) GROUP_ID=$(id -g) docker-compose up -d
```

---

## Next Steps

- [Configuration Guide](./02-configuration.md) - Detailed configuration options
- [CLI Usage](./03-cli-reference.md) - Command-line interface reference
- [API Documentation](http://localhost:8080/docs) - REST API reference (when running)

---

## Support

- **Issues**: [GitHub Issues](https://github.com/javinizer/javinizer-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/javinizer/javinizer-go/discussions)
