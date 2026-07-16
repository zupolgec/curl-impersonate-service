package handlers

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/zupolgec/curl-impersonate-service/config"
	"github.com/zupolgec/curl-impersonate-service/models"
)

// DocsHandler serves a public, server-rendered API documentation page.
type DocsHandler struct {
	tmpl         *template.Template
	adminEnabled bool
}

// NewDocsHandler builds the API docs handler. adminEnabled controls whether the
// admin UI is mentioned in the docs.
func NewDocsHandler(adminEnabled bool) *DocsHandler {
	return &DocsHandler{
		tmpl:         template.Must(template.New("docs").Parse(docsTemplate)),
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
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
