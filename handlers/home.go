package handlers

import (
	htmltemplate "html/template"
	"net/http"
	texttemplate "text/template"
)

// HomeHandler serves a minimal landing page at "/" pointing to the docs and
// the admin UI. Like the docs, it answers in Markdown when asked to.
type HomeHandler struct {
	html                      *htmltemplate.Template
	md                        *texttemplate.Template
	docsEnabled, adminEnabled bool
}

// NewHomeHandler builds the landing page handler.
func NewHomeHandler(docsEnabled, adminEnabled bool) *HomeHandler {
	return &HomeHandler{
		html:         htmltemplate.Must(htmltemplate.New("home").Parse(homeTemplate)),
		md:           texttemplate.Must(texttemplate.New("home").Parse(homeMarkdownTemplate)),
		docsEnabled:  docsEnabled,
		adminEnabled: adminEnabled,
	}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"DocsEnabled":  h.docsEnabled,
		"AdminEnabled": h.adminEnabled,
		"BaseURL":      baseURL(r),
	}
	w.Header().Set("Vary", "Accept")
	if wantsMarkdown(r) {
		renderText(w, h.md, data)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.html.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const homeMarkdownTemplate = `# curl-impersonate-service

HTTP requests that look like a real browser (TLS/HTTP2 fingerprints), as a REST API.
{{if .DocsEnabled}}
API documentation: {{.BaseURL}}/docs.md
{{end}}`

const homeTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>curl-impersonate-service</title>
<style>
  :root { color-scheme: light dark; --bg:#0f1115; --panel:#171a21; --border:#2a2f3a; --fg:#e6e8ee; --muted:#9aa4b2; --accent:#4f8cff; }
  @media (prefers-color-scheme: light) { :root { --bg:#f6f7f9; --panel:#fff; --border:#e2e5ea; --fg:#1a1d23; --muted:#5b6472; --accent:#2563eb; } }
  body { margin:0; min-height:100vh; display:grid; place-items:center; font:15px/1.6 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif; background:var(--bg); color:var(--fg); }
  .card { max-width:420px; margin:24px; padding:28px; background:var(--panel); border:1px solid var(--border); border-radius:12px; }
  h1 { margin:0 0 6px; font-size:20px; }
  p { margin:0 0 20px; color:var(--muted); }
  nav { display:flex; gap:10px; }
  nav a { padding:8px 14px; border-radius:6px; text-decoration:none; font-weight:600; border:1px solid var(--border); color:var(--fg); }
  nav a.primary { background:var(--accent); border-color:var(--accent); color:#fff; }
</style>
</head>
<body>
<div class="card">
  <h1>🎭 curl-impersonate-service</h1>
  <p>HTTP requests that look like a real browser, as a REST API.</p>
  {{if or .DocsEnabled .AdminEnabled}}<nav>
    {{if .DocsEnabled}}<a class="primary" href="/docs">API docs</a>{{end}}
    {{if .AdminEnabled}}<a href="/admin/">Admin</a>{{end}}
  </nav>{{end}}
</div>
</body>
</html>`
