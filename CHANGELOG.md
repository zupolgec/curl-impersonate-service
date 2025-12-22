# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-12-23

### Added
- Initial release of curl-impersonate-service
- REST API with 4 endpoints: `/health`, `/browsers`, `/metrics`, `/impersonate`
- Support for 19 browser profiles (Chrome, Firefox, Edge, Safari)
- Token-based authentication (Bearer token or query parameter)
- Browser aliases: `chrome-latest`, `firefox-latest`, `edge-latest`, `safari-latest`
- Default browser: `chrome-latest` (chrome116)
- Binary data support via base64 encoding for requests and responses
- Custom headers and query parameters support
- Configurable timeouts and redirect following
- Comprehensive error handling with error types: `auth`, `validation`, `network`, `dns`, `timeout`, `ssl`, `internal`
- Metrics collection: request counts, success/failure rates, average duration, browser usage
- Alpine-based Docker image (~200MB)
- Docker Compose support for easy deployment
- GitHub Actions workflows for testing and Docker builds
- Automatic image publishing to GitHub Container Registry (ghcr.io)
- Health checks in Docker
- Graceful shutdown support
- Request logging with unique request IDs
- Comprehensive unit tests (12 tests)
- Complete API documentation
- Usage examples and test scripts

### Configuration
- `TOKEN` - Required authentication token
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Logging level (default: info)
- `MAX_REQUEST_BODY_SIZE` - Maximum request body size (default: 10MB)
- `MAX_RESPONSE_BODY_SIZE` - Maximum response body size (default: 50MB)
- `MAX_TIMEOUT` - Maximum request timeout (default: 120s)
- `DEFAULT_TIMEOUT` - Default request timeout (default: 30s)

### Supported Browsers
- Chrome: 99, 100, 101, 104, 107, 110, 116
- Firefox: 91esr, 95, 98, 100, 102, 109, 117
- Edge: 99, 101
- Safari: 15.3, 15.5

### Documentation
- Comprehensive README with API documentation
- Docker deployment guide
- Production deployment recommendations
- Example requests and usage patterns
- Contributing guidelines
- MIT License

[1.0.0]: https://github.com/zupolgec/curl-impersonate-service/releases/tag/v1.0.0
