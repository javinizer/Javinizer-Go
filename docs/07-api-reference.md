# API Reference

The Javinizer REST API provides programmatic access to all metadata scraping, file organization, and database operations. The API powers the web UI and is fully documented with interactive examples.

## Overview

- **Base URL**: `http://localhost:8080/api/v1`
- **Content Type**: `application/json`
- **Authentication**: None (API is designed for local network use)
- **WebSocket**: Real-time progress updates at `/ws/progress`

## Getting Started

### Start the API Server

**Using Docker (recommended):**
```bash
docker run --rm -p 8080:8080 \
  -v "$(pwd)/data:/javinizer" \
  -v "/path/to/media:/media" \
  ghcr.io/javinizer/javinizer-go:latest
```

**Using CLI:**
```bash
javinizer api
```

**Custom port:**
```bash
PORT=9000 javinizer api
# or
javinizer api --port 9000
```

### Interactive API Documentation

The API server provides two interactive documentation interfaces:

- **Scalar UI**: [http://localhost:8080/docs](http://localhost:8080/docs) - Modern, user-friendly API explorer
- **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) - Traditional OpenAPI spec viewer

These interfaces provide:
- Complete request/response schemas
- Try-it-now functionality
- Code generation for multiple languages
- Authentication testing

## API Endpoints

### Movies

Scrape, retrieve, and manage movie metadata.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/scrape` | Scrape metadata for a JAV ID |
| `GET` | `/api/v1/movies` | List all movies in database |
| `GET` | `/api/v1/movies/:id` | Get movie metadata by ID |
| `POST` | `/api/v1/movies/:id/rescrape` | Force re-scrape movie metadata |
| `POST` | `/api/v1/movies/:id/compare-nfo` | Compare database metadata with NFO file |

**Example - Scrape movie:**
```bash
curl -X POST http://localhost:8080/api/v1/scrape \
  -H "Content-Type: application/json" \
  -d '{"id": "IPX-535"}'
```

**Example - List movies:**
```bash
curl http://localhost:8080/api/v1/movies
```

### Actresses

Manage actress database with images and metadata.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/actresses` | List all actresses in database |
| `GET` | `/api/v1/actresses/:id` | Get actress by ID |
| `GET` | `/api/v1/actresses/search` | Search actresses by name |
| `POST` | `/api/v1/actresses` | Create new actress entry |
| `PUT` | `/api/v1/actresses/:id` | Update actress metadata |
| `DELETE` | `/api/v1/actresses/:id` | Delete actress from database |

**Example - Search actresses:**
```bash
curl "http://localhost:8080/api/v1/actresses/search?q=Sakura"
```

### Batch Operations

Batch scraping workflow with job tracking and WebSocket progress updates.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/batch/scrape` | Start batch scrape job for directory |
| `GET` | `/api/v1/batch/:id` | Get batch job status and results |
| `POST` | `/api/v1/batch/:id/cancel` | Cancel running batch job |
| `POST` | `/api/v1/batch/:id/organize` | Organize files from completed batch job |
| `POST` | `/api/v1/batch/:id/update` | Update batch job settings |
| `PATCH` | `/api/v1/batch/:id/movies/:movieId` | Update movie metadata in batch job |
| `POST` | `/api/v1/batch/:id/movies/:movieId/poster-crop` | Update poster crop settings |
| `POST` | `/api/v1/batch/:id/movies/:movieId/exclude` | Exclude movie from organization |
| `POST` | `/api/v1/batch/:id/movies/:movieId/preview` | Preview organization path for movie |
| `POST` | `/api/v1/batch/:id/movies/:movieId/rescrape` | Re-scrape specific movie in batch |

**Example - Start batch scrape:**
```bash
curl -X POST http://localhost:8080/api/v1/batch/scrape \
  -H "Content-Type: application/json" \
  -d '{"directory": "/media/unsorted"}'
```

**Example - Get batch job status:**
```bash
curl http://localhost:8080/api/v1/batch/abc-123-def
```

### File Operations

Browse filesystem, scan directories, and preview organization results.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/cwd` | Get current working directory |
| `POST` | `/api/v1/scan` | Scan directory for JAV files |
| `POST` | `/api/v1/browse` | Browse filesystem (directory listing) |
| `POST` | `/api/v1/browse/autocomplete` | Get path autocomplete suggestions |

**Example - Scan directory:**
```bash
curl -X POST http://localhost:8080/api/v1/scan \
  -H "Content-Type: application/json" \
  -d '{"path": "/media"}'
```

**Security Note:** File operations respect `allowed_directories` and `denied_directories` in `config.yaml`. Docker deployments auto-detect `/media` as allowed. Manual deployments must explicitly configure allowed paths.

### System

Configuration, proxy testing, and scraper management.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/config` | Get current configuration |
| `PUT` | `/api/v1/config` | Update configuration (saves to file) |
| `GET` | `/api/v1/scrapers` | List available scrapers and status |
| `POST` | `/api/v1/proxy/test` | Test proxy connection |
| `POST` | `/api/v1/translation/models` | Get available translation models |

**Example - Get configuration:**
```bash
curl http://localhost:8080/api/v1/config
```

**Example - Test proxy:**
```bash
curl -X POST http://localhost:8080/api/v1/proxy/test \
  -H "Content-Type: application/json" \
  -d '{"url": "http://proxy.example.com:8080"}'
```

### History

Track and rollback file organization operations.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/history` | List operation history |
| `GET` | `/api/v1/history/stats` | Get history statistics |
| `DELETE` | `/api/v1/history/:id` | Delete single history entry |
| `DELETE` | `/api/v1/history` | Bulk delete history entries |

**Example - List history:**
```bash
curl http://localhost:8080/api/v1/history
```

### Resources

Serve temporary and persistent image files.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/temp/posters/:jobId/:filename` | Serve temporary poster from batch job |
| `GET` | `/api/v1/temp/image` | Serve temporary image with query params |
| `GET` | `/api/v1/posters/:filename` | Serve cropped poster from database |

**Example - Get temp poster:**
```bash
curl http://localhost:8080/api/v1/temp/posters/abc-123/IPX-535-poster.jpg -o poster.jpg
```

**Note:** Temp posters are preserved in `data/temp/posters/{jobID}/` when organization fails, allowing retry without re-scraping.

### Other Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check endpoint (returns `200 OK`) |
| `GET` | `/ws/progress` | WebSocket endpoint for real-time updates |
| `GET` | `/docs` | Scalar interactive API documentation |
| `GET` | `/swagger/*` | Swagger UI and OpenAPI spec |

## WebSocket

The `/ws/progress` endpoint provides real-time progress updates for batch operations.

**Connect to WebSocket:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/progress');

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Progress:', update);
};
```

**Message Format:**
```json
{
  "job_id": "abc-123-def",
  "type": "progress",
  "file": "IPX-535.mp4",
  "progress": 0.75,
  "message": "Downloading poster...",
  "bytes_processed": 1024000
}
```

**Event Types:**
- `progress` - Task progress update (0.0 to 1.0)
- `complete` - Task completed successfully
- `error` - Task failed with error message
- `cancelled` - Job cancelled by user

**Use Cases:**
- Real-time batch scrape progress
- Live download status
- Organization operation feedback
- Multi-client synchronization

## CORS Configuration

The API includes CORS middleware for browser-based frontends. Configure in `config.yaml`:

```yaml
api:
  security:
    # Allow all origins (development only)
    allowed_origins: ["*"]

    # Specific origins (recommended for production)
    # allowed_origins: ["http://localhost:5173", "http://127.0.0.1:5173"]

    # Same-origin only (most secure)
    # allowed_origins: []
```

## Directory Security

File operations (scan, browse, organize) are restricted by `allowed_directories` config:

```yaml
api:
  security:
    allowed_directories:
      - /media
      - ~/Videos
    denied_directories:
      - /etc
      - /root
```

**Behavior:**
- Empty `allowed_directories` = deny all (secure by default)
- Docker auto-detects `/media` as allowed
- Attempts to access denied paths return `403 Forbidden`

## Error Responses

Standard error format:

```json
{
  "error": "Not Found",
  "message": "The requested resource does not exist",
  "path": "/api/v1/movies/INVALID-ID",
  "method": "GET"
}
```

**Common HTTP Status Codes:**
- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request body or parameters
- `403 Forbidden` - Directory access denied
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Complete Documentation

For full request/response schemas, examples, and interactive testing, visit:

- **Scalar UI**: [http://localhost:8080/docs](http://localhost:8080/docs) - Recommended for exploration
- **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) - Full OpenAPI spec

These interfaces provide complete documentation including:
- Request body schemas
- Response models
- Query parameter validation
- Example requests for all endpoints
- Try-it-now functionality

---

**Next**: [Migration Guide](./08-migration-guide.md)
