package handlers

// adminTemplates holds the full admin UI as a set of named templates. Styles
// are inline so the UI works offline with no external requests.
const adminTemplates = `
{{define "layout"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>impersonate · admin</title>
<style>
  :root { color-scheme: light dark; --bg:#0f1115; --panel:#171a21; --border:#2a2f3a; --fg:#e6e8ee; --muted:#9aa4b2; --accent:#4f8cff; --ok:#3fb950; --bad:#f85149; }
  @media (prefers-color-scheme: light) { :root { --bg:#f6f7f9; --panel:#fff; --border:#e2e5ea; --fg:#1a1d23; --muted:#5b6472; --accent:#2563eb; } }
  * { box-sizing: border-box; }
  body { margin:0; font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif; background:var(--bg); color:var(--fg); }
  header { display:flex; align-items:center; gap:20px; padding:14px 24px; border-bottom:1px solid var(--border); background:var(--panel); }
  header h1 { font-size:15px; margin:0; font-weight:600; }
  nav { display:flex; gap:6px; }
  nav a { padding:6px 12px; border-radius:6px; text-decoration:none; color:var(--muted); }
  nav a.active, nav a:hover { color:var(--fg); background:color-mix(in srgb, var(--accent) 15%, transparent); }
  main { max-width:1000px; margin:0 auto; padding:24px; }
  .grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(160px,1fr)); gap:12px; margin-bottom:24px; }
  .card { background:var(--panel); border:1px solid var(--border); border-radius:10px; padding:16px; }
  .card .n { font-size:26px; font-weight:700; }
  .card .l { color:var(--muted); font-size:12px; text-transform:uppercase; letter-spacing:.04em; }
  table { width:100%; border-collapse:collapse; background:var(--panel); border:1px solid var(--border); border-radius:10px; overflow:hidden; }
  th,td { text-align:left; padding:10px 12px; border-bottom:1px solid var(--border); font-size:13px; }
  th { color:var(--muted); font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.04em; }
  tr:last-child td { border-bottom:none; }
  code { font-family:ui-monospace,SFMono-Regular,Menlo,monospace; background:color-mix(in srgb,var(--fg) 8%,transparent); padding:1px 6px; border-radius:4px; word-break:break-all; }
  form.inline { display:inline; }
  input[type=text] { background:var(--bg); border:1px solid var(--border); color:var(--fg); border-radius:6px; padding:8px 10px; font:inherit; }
  textarea { width:100%; background:var(--bg); border:1px solid var(--border); color:var(--fg); border-radius:6px; padding:10px; font:13px/1.5 ui-monospace,monospace; }
  button { background:var(--accent); color:#fff; border:0; border-radius:6px; padding:8px 14px; font:inherit; font-weight:600; cursor:pointer; }
  button.ghost { background:transparent; color:var(--muted); border:1px solid var(--border); }
  button.danger { background:transparent; color:var(--bad); border:1px solid var(--border); }
  .ok { color:var(--ok); } .bad { color:var(--bad); } .muted { color:var(--muted); }
  .banner { background:color-mix(in srgb,var(--ok) 15%,transparent); border:1px solid var(--ok); border-radius:8px; padding:12px 16px; margin-bottom:20px; }
  .banner code { background:var(--bg); }
  h2 { font-size:16px; margin:0 0 14px; }
  .row-actions { display:flex; gap:6px; }
</style>
</head>
<body>
<header>
  <h1>🎭 impersonate</h1>
  <nav>
    <a href="/admin/" class="{{if eq .Page "dashboard"}}active{{end}}">Dashboard</a>
    <a href="/admin/tokens" class="{{if eq .Page "tokens"}}active{{end}}">Tokens</a>
    <a href="/admin/cors" class="{{if eq .Page "cors"}}active{{end}}">CORS</a>
    <a href="/admin/logs" class="{{if eq .Page "logs"}}active{{end}}">Logs</a>
  </nav>
</header>
<main>
{{if eq .Page "dashboard"}}{{template "dashboard" .}}{{end}}
{{if eq .Page "tokens"}}{{template "tokens" .}}{{end}}
{{if eq .Page "cors"}}{{template "cors" .}}{{end}}
{{if eq .Page "logs"}}{{template "logs" .}}{{end}}
</main>
</body>
</html>{{end}}

{{define "dashboard"}}
<div class="grid">
  <div class="card"><div class="n">{{.Total}}</div><div class="l">Requests</div></div>
  <div class="card"><div class="n ok">{{.Success}}</div><div class="l">Success</div></div>
  <div class="card"><div class="n bad">{{.Failed}}</div><div class="l">Failed</div></div>
  <div class="card"><div class="n">{{printf "%.0f" .Avg}}<span class="muted" style="font-size:14px"> ms</span></div><div class="l">Avg duration</div></div>
  <div class="card"><div class="n">{{.Uptime}}<span class="muted" style="font-size:14px"> s</span></div><div class="l">Uptime</div></div>
</div>
<h2>Recent requests</h2>
{{template "logtable" .Logs}}
{{end}}

{{define "tokens"}}
<h2>API tokens</h2>
{{if .Created}}
<div class="banner">New token created. Copy it now — it won't be shown again:<br><code>{{.Created}}</code></div>
{{end}}
<form method="post" action="/admin/tokens" style="margin-bottom:20px; display:flex; gap:8px;">
  <input type="text" name="name" placeholder="Token name (e.g. scraper-prod)" required>
  <button type="submit">Create token</button>
</form>
<table>
  <tr><th>Name</th><th>Token</th><th>Status</th><th>Created</th><th>Last used</th><th></th></tr>
  {{range .Tokens}}
  <tr>
    <td>{{.Name}}</td>
    <td><code>{{slice .Token 0 8}}…{{slice .Token 56 64}}</code></td>
    <td>{{if .Enabled}}<span class="ok">enabled</span>{{else}}<span class="muted">disabled</span>{{end}}</td>
    <td class="muted">{{fmtTime .CreatedAt}}</td>
    <td class="muted">{{fmtTimePtr .LastUsedAt}}</td>
    <td><div class="row-actions">
      <form class="inline" method="post" action="/admin/tokens/toggle">
        <input type="hidden" name="id" value="{{.ID}}">
        <input type="hidden" name="enabled" value="{{if .Enabled}}false{{else}}true{{end}}">
        <button class="ghost" type="submit">{{if .Enabled}}Disable{{else}}Enable{{end}}</button>
      </form>
      <form class="inline" method="post" action="/admin/tokens/delete" onsubmit="return confirm('Delete this token?')">
        <input type="hidden" name="id" value="{{.ID}}">
        <button class="danger" type="submit">Delete</button>
      </form>
    </div></td>
  </tr>
  {{else}}
  <tr><td colspan="6" class="muted">No tokens yet. Create one above.</td></tr>
  {{end}}
</table>
{{end}}

{{define "cors"}}
<h2>CORS origins</h2>
{{if .Saved}}<div class="banner">CORS settings saved.</div>{{end}}
<p class="muted">One origin per line, or <code>*</code> to allow any origin.</p>
<form method="post" action="/admin/cors">
  <textarea name="origins" rows="6">{{.Origins}}</textarea>
  <div style="margin-top:12px"><button type="submit">Save</button></div>
</form>
{{end}}

{{define "logs"}}
<h2>Usage logs <span class="muted" style="font-size:13px; font-weight:400">(most recent {{.Limit}})</span></h2>
{{template "logtable" .Logs}}
{{end}}

{{define "logtable"}}
<table>
  <tr><th>Time</th><th>Token</th><th>Browser</th><th>Method</th><th>Host</th><th>Status</th><th>ms</th></tr>
  {{range .}}
  <tr>
    <td class="muted">{{fmtTime .TS}}</td>
    <td>{{.TokenName}}</td>
    <td>{{.Browser}}</td>
    <td>{{.Method}}</td>
    <td><code>{{.TargetHost}}</code></td>
    <td>{{if .Success}}<span class="ok">{{.StatusCode}}</span>{{else}}<span class="bad">{{if .ErrorType}}{{.ErrorType}}{{else}}error{{end}}</span>{{end}}</td>
    <td class="muted">{{.DurationMs}}</td>
  </tr>
  {{else}}
  <tr><td colspan="7" class="muted">No requests logged yet.</td></tr>
  {{end}}
</table>
{{end}}
`
