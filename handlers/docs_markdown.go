package handlers

// docsMarkdownTemplate is the API documentation in Markdown, for agents. Keep
// it in sync with docsTemplate.
const docsMarkdownTemplate = `# curl-impersonate-service v{{.Version}}

HTTP requests that look like a real browser (TLS/HTTP2 fingerprints), as a REST API.

Base URL: {{.BaseURL}}

## Authentication

All API endpoints except ` + "`/health`" + ` require an API token, as a header or a query parameter:

    Authorization: Bearer <token>
    ?token=<token>
{{if .AdminEnabled}}
Tokens are issued by the service administrator.
{{end}}
## Endpoints

- ` + "`GET /health`" + `: liveness, no auth. Returns ` + "`{\"status\":\"ok\",\"version\":\"…\"}`" + `.
- ` + "`GET /browsers`" + `: available browser profiles, aliases and default.
- ` + "`GET /metrics`" + `: request counts, success/failure, average duration, per-browser usage.
- ` + "`POST /impersonate`" + `: performs an HTTP request impersonating a browser.

### POST /impersonate

JSON body:

| Field | Type | Default | Description |
|---|---|---|---|
| ` + "`url`" + ` | string | (required) | Target URL |
| ` + "`browser`" + ` | string | ` + "`{{.Default}}`" + ` | Profile or alias |
| ` + "`method`" + ` | string | ` + "`GET`" + ` | HTTP method |
| ` + "`headers`" + ` | object | | Custom request headers |
| ` + "`query_params`" + ` | object | | Query params merged into the URL |
| ` + "`body`" + ` | string | | Request body (text) |
| ` + "`body_base64`" + ` | string | | Request body (base64, for binary) |
| ` + "`follow_redirects`" + ` | bool | ` + "`true`" + ` | Follow redirects |
| ` + "`insecure`" + ` | bool | ` + "`false`" + ` | Skip TLS verification |
| ` + "`timeout`" + ` | int | ` + "`30`" + ` | Timeout in seconds |

Example:

    curl -X POST {{.BaseURL}}/impersonate \
      -H "Authorization: Bearer <token>" \
      -H "Content-Type: application/json" \
      -d '{"browser":"{{.Default}}","url":"https://example.com"}'

The response is always HTTP 200 with a JSON envelope: ` + "`success`" + `, ` + "`status_code`" + `,
` + "`headers`" + `, ` + "`body`" + `, ` + "`timing`" + `, … When the request could not be made,
` + "`success`" + ` is false and ` + "`error_type`" + ` is one of ` + "`network`" + `, ` + "`dns`" + `,
` + "`timeout`" + `, ` + "`ssl`" + ` or ` + "`size`" + `. Bad input returns HTTP 400 with ` + "`error_type: validation`" + `,
a missing or wrong token HTTP 401 with ` + "`error_type: auth`" + `.

## Browsers

Default: ` + "`{{.Default}}`" + `. Aliases:
{{range .Aliases}}
- ` + "`{{.Name}}`" + ` → ` + "`{{.Target}}`" + `{{end}}

All profiles: {{range $i, $b := .Browsers}}{{if $i}}, {{end}}` + "`{{$b.Name}}`" + `{{end}}
`
