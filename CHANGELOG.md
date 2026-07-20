# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.2] - 2026-07-20

### Fixed
- Automatically decode compressed upstream responses before serializing the
  JSON envelope, and remove stale `Content-Encoding`/`Content-Length` headers.
- Base64-encode textual responses containing invalid UTF-8 instead of allowing
  JSON serialization to replace and irreversibly corrupt their bytes.

## [1.3.1] - 2026-07-16

### Changed
- Moved the API docs page from `/` to `/docs` and put it behind API-token auth
  (open it in a browser with `?token=<token>`).

## [1.3.0] - 2026-07-16

### Added
- Public API documentation page served at `/` (toggle with `API_DOCS_ENABLED`,
  default enabled). Lists endpoints, request fields and available browsers.

### Changed (stricter SSRF defaults)
- Only `https://` targets are allowed by default; enable `http://` with
  `SSRF_ALLOW_HTTP=true`.
- Targets addressed by a raw IP are rejected by default; allow them with
  `SSRF_ALLOW_IP=true`.

## [1.2.0] - 2026-07-16

### Changed
- Migrated to the actively-maintained [lexiforest/curl-impersonate](https://github.com/lexiforest/curl-impersonate)
  fork (v1.2.2); the original lwthiker project is archived.
- Browser profiles expanded from 19 to 31, including recent versions
  (Chrome up to 136, Firefox 133/135, Safari up to 26.0 with iOS variants, Tor 14.5).
- Unified `libcurl-impersonate` means the CGO executor now handles **all**
  browsers (including Firefox) natively, removing the shell fallback path.
- Default browser is now `chrome136`; `firefox-latest`→`firefox135`,
  `safari-latest`→`safari260`; added `tor-latest`→`tor145`.

### Note
- Old Firefox profile names (`ff91esr`…`ff117`) are replaced by `firefox133`/
  `firefox135`. Update any hardcoded browser names accordingly.

## [1.1.1] - 2026-07-16

### Fixed
- Bumped Go to 1.26.5 to pick up the `crypto/tls` security fix (GO-2026-5856).
- CI hardening: valid Docker image tags on tag pushes, CGO-free release binaries,
  pinned `trivy-action`, and hermetic integration tests (local go-httpbin sidecar).

## [1.1.0] - 2026-07-16

### Added
- SSRF protection: blocks loopback, private, link-local and cloud-metadata
  destinations and restricts schemes to http/https (including across redirects).
  Configurable via `SSRF_ALLOW_PRIVATE`, `SSRF_DENY_HOSTS`, `SSRF_ALLOW_HOSTS`.
- SQLite datastore (`DATA_DIR`) for API tokens, settings and usage logs.
- Admin UI at `/admin/` (enabled with `ADMIN_TOKEN`, HTTP Basic auth): manage
  API tokens, edit CORS origins at runtime, browse usage logs and metrics.
- Multiple API tokens managed from the admin UI; `TOKEN` env is seeded as a
  legacy token for backward compatibility.
- Usage logging with automatic retention (`LOG_RETENTION_HOURS`, default 72h).
- Configurable CORS origins (`CORS_ALLOWED_ORIGINS`), editable from the admin UI.
- `SECURITY.md`, Dependabot config and a Security workflow (govulncheck + Trivy).

### Changed
- Enforced `MAX_RESPONSE_BODY_SIZE` (previously declared but unused).
- Constant-time token comparison for authentication.
- Bumped Go to 1.26 and updated GitHub Actions and lint configuration.

### Notes
- The datastore lives in a `/data` volume — persist it across deployments.
- Multi-architecture Docker images (AMD64 and ARM64) continue to be published.

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

[1.3.2]: https://github.com/zupolgec/curl-impersonate-service/releases/tag/v1.3.2
[1.3.1]: https://github.com/zupolgec/curl-impersonate-service/releases/tag/v1.3.1
[1.3.0]: https://github.com/zupolgec/curl-impersonate-service/releases/tag/v1.3.0
[1.2.0]: https://github.com/zupolgec/curl-impersonate-service/releases/tag/v1.2.0
[1.1.1]: https://github.com/zupolgec/curl-impersonate-service/releases/tag/v1.1.1
[1.1.0]: https://github.com/zupolgec/curl-impersonate-service/releases/tag/v1.1.0
[1.0.0]: https://github.com/zupolgec/curl-impersonate-service/releases/tag/v1.0.0
