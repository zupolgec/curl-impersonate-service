package handlers

// docsTemplate is the public API documentation page. Inline styles only, so it
// works offline with no external requests.
const docsTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>curl-impersonate-service · API</title>
<style>
  :root { color-scheme: light dark; --bg:#0f1115; --panel:#171a21; --border:#2a2f3a; --fg:#e6e8ee; --muted:#9aa4b2; --accent:#4f8cff; --ok:#3fb950; }
  @media (prefers-color-scheme: light) { :root { --bg:#f6f7f9; --panel:#fff; --border:#e2e5ea; --fg:#1a1d23; --muted:#5b6472; --accent:#2563eb; } }
  * { box-sizing: border-box; }
  body { margin:0; font:15px/1.6 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif; background:var(--bg); color:var(--fg); }
  header { padding:28px 24px; border-bottom:1px solid var(--border); background:var(--panel); }
  header .wrap, main { max-width:900px; margin:0 auto; }
  main { padding:24px; }
  h1 { margin:0; font-size:22px; }
  h1 .v { color:var(--muted); font-size:14px; font-weight:400; }
  .tag { color:var(--muted); margin-top:6px; }
  h2 { font-size:18px; margin:32px 0 12px; padding-bottom:6px; border-bottom:1px solid var(--border); }
  h3 { font-size:15px; margin:20px 0 8px; }
  code { font-family:ui-monospace,SFMono-Regular,Menlo,monospace; background:color-mix(in srgb,var(--fg) 8%,transparent); padding:1px 6px; border-radius:4px; }
  pre { background:var(--panel); border:1px solid var(--border); border-radius:10px; padding:14px 16px; overflow-x:auto; }
  pre code { background:none; padding:0; }
  table { width:100%; border-collapse:collapse; background:var(--panel); border:1px solid var(--border); border-radius:10px; overflow:hidden; margin:8px 0; }
  th,td { text-align:left; padding:8px 12px; border-bottom:1px solid var(--border); font-size:13px; }
  th { color:var(--muted); font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.04em; }
  tr:last-child td { border-bottom:none; }
  .method { display:inline-block; font-family:ui-monospace,monospace; font-size:12px; font-weight:700; padding:2px 8px; border-radius:5px; background:color-mix(in srgb,var(--accent) 18%,transparent); color:var(--accent); }
  .cols { display:grid; grid-template-columns:repeat(auto-fill,minmax(150px,1fr)); gap:6px; }
  .cols code { text-align:center; }
  a { color:var(--accent); }
  .muted { color:var(--muted); }
</style>
</head>
<body>
<header><div class="wrap">
  <h1>curl-impersonate-service <span class="v">v{{.Version}}</span></h1>
  <div class="tag">HTTP requests that look like a real browser (TLS/HTTP2 fingerprints), as a REST API.</div>
</div></header>
<main>

<h2>Authentication</h2>
<p>All endpoints except <code>/health</code> require an API token, sent as
either a bearer header or a query parameter:</p>
<pre><code>Authorization: Bearer &lt;token&gt;
# or
?token=&lt;token&gt;</code></pre>
{{if .AdminEnabled}}<p class="muted">Tokens are managed from the <a href="/admin/">admin UI</a>.</p>{{end}}

<h2>Endpoints</h2>

<h3><span class="method">GET</span> <code>/health</code></h3>
<p>Liveness check, no auth. Returns <code>{"status":"ok","version":"…"}</code>.</p>

<h3><span class="method">GET</span> <code>/browsers</code></h3>
<p>Lists available browser profiles and aliases.</p>

<h3><span class="method">GET</span> <code>/metrics</code></h3>
<p>Service metrics: request counts, success/failure, average duration, per-browser usage.</p>

<h3><span class="method">POST</span> <code>/impersonate</code></h3>
<p>Performs an HTTP request impersonating the chosen browser.</p>
<table>
  <tr><th>Field</th><th>Type</th><th>Default</th><th>Description</th></tr>
  <tr><td><code>url</code></td><td>string</td><td>—</td><td>Target URL (required)</td></tr>
  <tr><td><code>browser</code></td><td>string</td><td><code>{{.Default}}</code></td><td>Profile or alias</td></tr>
  <tr><td><code>method</code></td><td>string</td><td><code>GET</code></td><td>HTTP method</td></tr>
  <tr><td><code>headers</code></td><td>object</td><td>—</td><td>Custom request headers</td></tr>
  <tr><td><code>query_params</code></td><td>object</td><td>—</td><td>Query params merged into the URL</td></tr>
  <tr><td><code>body</code></td><td>string</td><td>—</td><td>Request body (text)</td></tr>
  <tr><td><code>body_base64</code></td><td>string</td><td>—</td><td>Request body (base64, for binary)</td></tr>
  <tr><td><code>follow_redirects</code></td><td>bool</td><td><code>true</code></td><td>Follow redirects</td></tr>
  <tr><td><code>insecure</code></td><td>bool</td><td><code>false</code></td><td>Skip TLS verification</td></tr>
  <tr><td><code>timeout</code></td><td>int</td><td><code>30</code></td><td>Timeout (seconds)</td></tr>
</table>

<h3>Example</h3>
<pre><code>curl -X POST http://localhost:8080/impersonate \
  -H "Authorization: Bearer &lt;token&gt;" \
  -H "Content-Type: application/json" \
  -d '{"browser":"{{.Default}}","url":"https://example.com"}'</code></pre>

<p>Success responses always return <code>200</code> with a JSON envelope
(<code>success</code>, <code>status_code</code>, <code>headers</code>,
<code>body</code>, <code>timing</code>, …). Network problems return
<code>success:false</code> with an <code>error_type</code> of
<code>network</code>, <code>dns</code>, <code>timeout</code>, <code>ssl</code> or
<code>size</code>.</p>

{{if .AdminEnabled}}
<h3><span class="method">GET</span> <code>/admin/</code></h3>
<p>Admin dashboard (separate Basic-auth login): manage tokens, CORS and usage logs.</p>
{{end}}

<h2>Browsers</h2>
<p>Default: <code>{{.Default}}</code>. Aliases:</p>
<table>
  <tr><th>Alias</th><th>Resolves to</th></tr>
  {{range .Aliases}}<tr><td><code>{{.Name}}</code></td><td><code>{{.Target}}</code></td></tr>{{end}}
</table>
<h3>All profiles ({{len .Browsers}})</h3>
<div class="cols">
  {{range .Browsers}}<code>{{.Name}}</code>{{end}}
</div>

</main>
</body>
</html>`
