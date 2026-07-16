# Security Policy

## Reporting a Vulnerability

Please report security vulnerabilities privately via
[GitHub Security Advisories](https://github.com/zupolgec/curl-impersonate-service/security/advisories/new)
rather than opening a public issue.

We aim to acknowledge reports within a few days and will keep you updated on
remediation progress.

## Security model

This service proxies HTTP requests to arbitrary targets on behalf of
authenticated clients. Keep the following in mind when deploying:

- **Authentication**: all API endpoints (except `/health`) require a valid API
  token. The admin UI requires the fixed `ADMIN_TOKEN`. Use strong, random
  tokens and treat them as secrets.
- **SSRF protection**: by default the service blocks requests to loopback,
  private (RFC1918), link-local and cloud-metadata addresses, and restricts
  schemes to `http`/`https` (including across redirects). Only set
  `SSRF_ALLOW_PRIVATE=true` if you intentionally proxy to internal hosts.
- **Response limits**: responses larger than `MAX_RESPONSE_BODY_SIZE` are
  rejected to bound memory usage.
- **Transport**: run behind a reverse proxy that terminates TLS. Avoid passing
  tokens via the `?token=` query parameter where proxy logs may capture them;
  prefer the `Authorization: Bearer` header.
- **Persistence**: the SQLite datastore holds API tokens. Protect the `/data`
  volume accordingly.

## Supported Versions

The latest released version on the `main` branch is supported.
