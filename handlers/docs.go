package handlers

import (
	htmltemplate "html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	texttemplate "text/template"

	"github.com/zupolgec/curl-impersonate-service/config"
	"github.com/zupolgec/curl-impersonate-service/models"
)

// DocsHandler serves the public API documentation, as HTML for browsers and
// as Markdown for agents (Accept: text/markdown, /docs.md or /llms.txt).
type DocsHandler struct {
	html         *htmltemplate.Template
	md           *texttemplate.Template
	adminEnabled bool
}

// NewDocsHandler builds the API docs handler. adminEnabled controls whether the
// admin UI is mentioned in the docs.
func NewDocsHandler(adminEnabled bool) *DocsHandler {
	return &DocsHandler{
		html:         htmltemplate.Must(htmltemplate.New("docs").Parse(docsTemplate)),
		md:           texttemplate.Must(texttemplate.New("docs").Parse(docsMarkdownTemplate)),
		adminEnabled: adminEnabled,
	}
}

func (h *DocsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	browsers := models.GetAllBrowsers()
	sort.Slice(browsers, func(i, j int) bool { return browsers[i].Name < browsers[j].Name })

	aliases := models.GetAliases()
	aliasKeys := make([]string, 0, len(aliases))
	for k := range aliases {
		aliasKeys = append(aliasKeys, k)
	}
	sort.Strings(aliasKeys)
	type alias struct{ Name, Target string }
	aliasList := make([]alias, 0, len(aliasKeys))
	for _, k := range aliasKeys {
		aliasList = append(aliasList, alias{k, aliases[k]})
	}

	data := map[string]any{
		"Version":      config.Version,
		"Browsers":     browsers,
		"Aliases":      aliasList,
		"Default":      models.GetDefaultBrowser(),
		"AdminEnabled": h.adminEnabled,
		"BaseURL":      baseURL(r),
	}
	if r.URL.Path == "/docs.md" || r.URL.Path == "/llms.txt" {
		renderText(w, h.md, data)
		return
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

func renderText(w http.ResponseWriter, t *texttemplate.Template, data any) {
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// baseURL is the scheme and host the request reached us on, honouring the
// reverse proxy's X-Forwarded-Proto.
func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// wantsMarkdown reports whether the Accept header asks for text/markdown at
// least as strongly as for text/html. Wildcards alone mean HTML.
func wantsMarkdown(r *http.Request) bool {
	var md, html float64
	for _, part := range strings.Split(r.Header.Get("Accept"), ",") {
		fields := strings.Split(part, ";")
		q := 1.0
		for _, p := range fields[1:] {
			if v, ok := strings.CutPrefix(strings.TrimSpace(p), "q="); ok {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					q = f
				}
			}
		}
		switch strings.TrimSpace(fields[0]) {
		case "text/markdown":
			md = max(md, q)
		case "text/html":
			html = max(html, q)
		}
	}
	return md > 0 && md >= html
}
