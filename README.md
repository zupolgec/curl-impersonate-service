# curl-impersonate-service

A Docker-based REST API service that wraps [curl-impersonate](https://github.com/lwthiker/curl-impersonate) to provide browser impersonation as a service. Make HTTP requests that appear to come from real browsers (Chrome, Firefox, Edge, Safari) to bypass TLS and HTTP/2 fingerprinting.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Available-2496ED?logo=docker)](https://ghcr.io/zupolgec/curl-impersonate-service)

## Features

- **19 Browser Profiles**: Impersonate Chrome, Firefox, Edge, and Safari across different versions
- **Full HTTP Support**: All HTTP methods (GET, POST, PUT, PATCH, DELETE, etc.)
- **Custom Headers & Query Params**: Full control over request customization
- **Binary Data Support**: Base64 encoding for request and response bodies
- **Token Authentication**: Secure API access via Bearer token or query parameter
- **Metrics & Monitoring**: Built-in metrics endpoint for monitoring
- **Lightweight**: Debian slim-based Docker image (~250MB)
- **Multi-Architecture**: Supports both AMD64 and ARM64 (Apple Silicon, AWS Graviton, etc.)
- **Production Ready**: Graceful shutdown, health checks, comprehensive logging

## Quick Start

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/zupolgec/curl-impersonate-service.git
cd curl-impersonate-service

# Start the service
docker-compose up -d

# Test it
curl http://localhost:8080/health
```

### Using Docker

```bash
# Works on both AMD64 and ARM64 (Apple Silicon, AWS Graviton, etc.)
docker run -d \
  -p 8080:8080 \
  -e TOKEN=your-secret-token \
  ghcr.io/zupolgec/curl-impersonate-service:latest
```

## API Documentation

### Authentication

All endpoints except `/health` require authentication via:
- **Bearer Token**: `Authorization: Bearer <token>`
- **Query Parameter**: `?token=<token>`

Set your token via the `TOKEN` environment variable.

### Endpoints

#### `GET /health`

Health check endpoint (no authentication required).

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

#### `GET /browsers`

List all available browser profiles and aliases (authentication required).

**Response:**
```json
{
  "browsers": [
    {
      "name": "chrome116",
      "browser": {
        "name": "chrome",
        "version": "116.0.5845.180",
        "os": "win10"
      },
      "binary": "curl-impersonate-chrome",
      "wrapper_script": "curl_chrome116"
    }
  ],
  "aliases": {
    "chrome-latest": "chrome116",
    "firefox-latest": "ff117",
    "edge-latest": "edge101",
    "safari-latest": "safari15_5"
  },
  "default": "chrome116"
}
```

#### `GET /metrics`

Get service metrics (authentication required).

**Response:**
```json
{
  "uptime_seconds": 3600,
  "requests_total": 1234,
  "requests_success": 1200,
  "requests_failed": 34,
  "average_duration_ms": 456.78,
  "browsers_used": {
    "chrome116": 800,
    "ff109": 400
  }
}
```

#### `POST /impersonate`

Make an HTTP request impersonating a browser (authentication required).

**Request Body:**
```json
{
  "browser": "chrome116",
  "url": "https://example.com",
  "method": "GET",
  "headers": {
    "X-Custom-Header": "value"
  },
  "query_params": {
    "key": "value"
  },
  "body": "request body",
  "body_base64": "base64-encoded-binary-data",
  "follow_redirects": true,
  "insecure": false,
  "timeout": 30
}
```

**Fields:**
- `browser` (optional): Browser to impersonate. Default: `chrome-latest`. Supports aliases.
- `url` (required): Target URL
- `method` (optional): HTTP method. Default: `GET`
- `headers` (optional): Custom headers as key-value pairs
- `query_params` (optional): Query parameters to merge with URL
- `body` (optional): Request body as string (mutually exclusive with `body_base64`)
- `body_base64` (optional): Request body as base64 string for binary data
- `follow_redirects` (optional): Follow HTTP redirects. Default: `true`
- `insecure` (optional): Skip SSL certificate verification. Default: `false`
- `timeout` (optional): Request timeout in seconds. Default: `30`, Max: `120`

**Success Response (200 OK):**
```json
{
  "success": true,
  "status_code": 200,
  "headers": {
    "content-type": ["text/html; charset=utf-8"]
  },
  "body": "response body or base64 encoded data",
  "body_base64": false,
  "final_url": "https://example.com",
  "timing": {
    "total": 1.234,
    "namelookup": 0.123,
    "connect": 0.234,
    "starttransfer": 0.456
  }
}
```

**Network Error Response (200 OK):**
```json
{
  "success": false,
  "error": "connection timeout",
  "error_type": "network",
  "status_code": 0
}
```

Error types: `network`, `dns`, `timeout`, `ssl`

**Validation Error (400 Bad Request):**
```json
{
  "success": false,
  "error": "unknown browser: invalid-browser",
  "error_type": "validation"
}
```

**Authentication Error (401 Unauthorized):**
```json
{
  "success": false,
  "error": "invalid or missing authentication token",
  "error_type": "auth"
}
```

## Supported Browsers

| Browser | Versions | Alias |
|---------|----------|-------|
| Chrome | 99, 100, 101, 104, 107, 110, 116 | `chrome-latest` |
| Firefox | 91esr, 95, 98, 100, 102, 109, 117 | `firefox-latest` |
| Edge | 99, 101 | `edge-latest` |
| Safari | 15.3, 15.5 | `safari-latest` |

Default browser: `chrome116` (chrome-latest)

See full list: [browsers.json](browsers.json)

## Examples

### Simple GET Request

```bash
curl -X POST http://localhost:8080/impersonate?token=your-token \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://httpbin.org/get"
  }'
```

### POST with Custom Headers

```bash
curl -X POST http://localhost:8080/impersonate \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "browser": "ff109",
    "url": "https://httpbin.org/post",
    "method": "POST",
    "headers": {
      "X-API-Key": "secret"
    },
    "body": "{\"key\": \"value\"}"
  }'
```

### Using Query Parameters

```bash
curl -X POST http://localhost:8080/impersonate \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "browser": "edge101",
    "url": "https://api.example.com/search",
    "query_params": {
      "q": "search term",
      "limit": "10"
    }
  }'
```

### Binary Data Upload

```bash
# Encode file to base64
FILE_B64=$(base64 -i image.png)

curl -X POST http://localhost:8080/impersonate \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d "{
    \"browser\": \"chrome116\",
    \"url\": \"https://api.example.com/upload\",
    \"method\": \"POST\",
    \"body_base64\": \"$FILE_B64\"
  }"
```

See more examples in [examples/test-requests.sh](examples/test-requests.sh)

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TOKEN` | Yes | - | Authentication token for API access |
| `PORT` | No | `8080` | Server port |
| `LOG_LEVEL` | No | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `MAX_REQUEST_BODY_SIZE` | No | `10485760` | Max request body size in bytes (10MB) |
| `MAX_RESPONSE_BODY_SIZE` | No | `52428800` | Max response body size in bytes (50MB) |
| `MAX_TIMEOUT` | No | `120` | Maximum timeout in seconds |
| `DEFAULT_TIMEOUT` | No | `30` | Default timeout in seconds |

### Docker Compose Example

```yaml
services:
  impersonate-service:
    image: ghcr.io/zupolgec/curl-impersonate-service:latest
    ports:
      - "8080:8080"
    environment:
      - TOKEN=your-secret-token-here
      - LOG_LEVEL=info
      - MAX_TIMEOUT=120
    restart: unless-stopped
```

## Building from Source

```bash
# Clone repository
git clone https://github.com/zupolgec/curl-impersonate-service.git
cd curl-impersonate-service

# Build Docker image
docker build -t curl-impersonate-service .

# Run
docker run -d -p 8080:8080 -e TOKEN=your-token curl-impersonate-service
```

## Development

```bash
# Install Go 1.21+
# Clone repository
git clone https://github.com/zupolgec/curl-impersonate-service.git
cd curl-impersonate-service

# Install dependencies
go mod download

# Run locally (requires curl-impersonate binaries)
TOKEN=test-token go run main.go

# Run tests
go test ./...

# Build binary
go build -o impersonate-service .
```

## Testing

Run the test suite:

```bash
# Start service
docker-compose up -d

# Run tests
TOKEN=test-token-123 ./examples/test-requests.sh

# View logs
docker-compose logs -f
```

## Production Deployment

### Security Recommendations

1. **Use a strong TOKEN**: Generate a cryptographically secure random token
2. **Use HTTPS**: Put the service behind a reverse proxy (nginx, Traefik, etc.)
3. **Rate Limiting**: Implement rate limiting at the reverse proxy level
4. **Network Isolation**: Run in a private network, expose via reverse proxy only
5. **Monitoring**: Enable metrics endpoint and monitor service health

### Example nginx Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name impersonate.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Rate limiting
        limit_req zone=api burst=10 nodelay;
    }
}
```

## Architecture

```
┌─────────────────┐
│   HTTP Client   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Auth Middleware│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Handlers     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Executor     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│curl-impersonate │
└─────────────────┘
```

## Troubleshooting

### Binary response appears corrupted
- Check if `body_base64` is `true` in the response
- Decode base64 before using: `echo $BODY | base64 -d > file.bin`

### SSL certificate errors
- If you get `error_type: "ssl"` with a message about peer certificate verification, the target server may have an invalid, self-signed, or expired certificate
- Set `"insecure": true` in your request to skip SSL certificate verification
- **Warning**: Only use `insecure` when you understand the security implications

### Timeout errors
- Increase timeout in request or `MAX_TIMEOUT` environment variable
- Check if target server is slow or unreachable

### Authentication fails
- Verify `TOKEN` environment variable is set correctly
- Check Bearer token format: `Authorization: Bearer <token>`
- Or use query parameter: `?token=<token>`

### Container fails to start
- Check logs: `docker-compose logs`
- Verify `TOKEN` environment variable is set
- Ensure port 8080 is available

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

- Built on [curl-impersonate](https://github.com/lwthiker/curl-impersonate) by lwthiker
- Inspired by the need for browser impersonation in web scraping and testing

## Support

- Issues: [GitHub Issues](https://github.com/zupolgec/curl-impersonate-service/issues)
- Discussions: [GitHub Discussions](https://github.com/zupolgec/curl-impersonate-service/discussions)

---

**Note**: This tool is for legitimate testing and development purposes. Always respect websites' Terms of Service and robots.txt. Use responsibly.
