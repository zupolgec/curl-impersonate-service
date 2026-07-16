package handlers

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zupolgec/curl-impersonate-service/metrics"
	"github.com/zupolgec/curl-impersonate-service/store"
)

// corsSettingKey is the settings key holding the comma-separated CORS origins.
const corsSettingKey = "cors_allowed_origins"

// AdminHandler serves the admin UI and its form actions.
type AdminHandler struct {
	store     *store.Store
	collector *metrics.Collector
	tmpl      *template.Template
}

// NewAdminHandler builds the admin UI handler and returns an http.Handler
// mounted under /admin/.
func NewAdminHandler(st *store.Store, collector *metrics.Collector) http.Handler {
	h := &AdminHandler{
		store:     st,
		collector: collector,
		tmpl:      template.Must(template.New("admin").Funcs(adminFuncs).Parse(adminTemplates)),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/", h.dashboard)
	mux.HandleFunc("GET /admin/tokens", h.tokens)
	mux.HandleFunc("POST /admin/tokens", h.createToken)
	mux.HandleFunc("POST /admin/tokens/delete", h.deleteToken)
	mux.HandleFunc("POST /admin/tokens/toggle", h.toggleToken)
	mux.HandleFunc("GET /admin/cors", h.cors)
	mux.HandleFunc("POST /admin/cors", h.saveCORS)
	mux.HandleFunc("GET /admin/logs", h.logs)
	return mux
}

var adminFuncs = template.FuncMap{
	"fmtTime": func(t time.Time) string {
		if t.IsZero() {
			return "—"
		}
		return t.UTC().Format("2006-01-02 15:04:05 UTC")
	},
	"fmtTimePtr": func(t *time.Time) string {
		if t == nil {
			return "never"
		}
		return t.UTC().Format("2006-01-02 15:04:05 UTC")
	},
}

func (h *AdminHandler) render(w http.ResponseWriter, page string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["Page"] = page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AdminHandler) dashboard(w http.ResponseWriter, r *http.Request) {
	uptime, total, success, failed, avg, browsers := h.collector.GetMetrics()

	type kv struct {
		Name  string
		Count int64
	}
	var used []kv
	for k, v := range browsers {
		used = append(used, kv{k, v})
	}
	sort.Slice(used, func(i, j int) bool { return used[i].Count > used[j].Count })

	logs, _ := h.store.ListLogs(15)

	h.render(w, "dashboard", map[string]any{
		"Uptime":   uptime,
		"Total":    total,
		"Success":  success,
		"Failed":   failed,
		"Avg":      avg,
		"Browsers": used,
		"Logs":     logs,
	})
}

func (h *AdminHandler) tokens(w http.ResponseWriter, r *http.Request) {
	toks, err := h.store.ListTokens()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, "tokens", map[string]any{
		"Tokens":  toks,
		"Created": r.URL.Query().Get("created"),
	})
}

func (h *AdminHandler) createToken(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = "token"
	}
	tok, err := h.store.CreateToken(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Show the new token value once, via query param.
	http.Redirect(w, r, "/admin/tokens?created="+tok.Token, http.StatusSeeOther)
}

func (h *AdminHandler) deleteToken(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err := h.store.DeleteToken(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/tokens", http.StatusSeeOther)
}

func (h *AdminHandler) toggleToken(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	enabled := r.FormValue("enabled") == "true"
	if err := h.store.SetTokenEnabled(id, enabled); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/tokens", http.StatusSeeOther)
}

func (h *AdminHandler) cors(w http.ResponseWriter, r *http.Request) {
	origins := h.store.GetSetting(corsSettingKey, "*")
	h.render(w, "cors", map[string]any{
		"Origins": origins,
		"Saved":   r.URL.Query().Get("saved") == "1",
	})
}

func (h *AdminHandler) saveCORS(w http.ResponseWriter, r *http.Request) {
	origins := strings.TrimSpace(r.FormValue("origins"))
	if origins == "" {
		origins = "*"
	}
	if err := h.store.SetSetting(corsSettingKey, origins); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/cors?saved=1", http.StatusSeeOther)
}

func (h *AdminHandler) logs(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 2000 {
		limit = l
	}
	logs, err := h.store.ListLogs(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, "logs", map[string]any{"Logs": logs, "Limit": limit})
}

// CORSOriginProvider returns an origin provider backed by the datastore setting,
// falling back to def when unset.
func CORSOriginProvider(st *store.Store, def []string) func() []string {
	defStr := strings.Join(def, ",")
	return func() []string {
		raw := st.GetSetting(corsSettingKey, defStr)
		var out []string
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				out = append(out, o)
			}
		}
		if len(out) == 0 {
			return []string{"*"}
		}
		return out
	}
}

// SeedCORSSetting stores the initial CORS setting if none exists yet.
func SeedCORSSetting(st *store.Store, def []string) error {
	if st.GetSetting(corsSettingKey, "") != "" {
		return nil
	}
	return st.SetSetting(corsSettingKey, strings.Join(def, ","))
}
