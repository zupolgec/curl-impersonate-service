package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zupolgec/curl-impersonate-service/config"
	"github.com/zupolgec/curl-impersonate-service/executor"
	"github.com/zupolgec/curl-impersonate-service/metrics"
	"github.com/zupolgec/curl-impersonate-service/middleware"
	"github.com/zupolgec/curl-impersonate-service/models"
	"github.com/zupolgec/curl-impersonate-service/security"
	"github.com/zupolgec/curl-impersonate-service/store"
)

type ImpersonateHandler struct {
	cfg       *config.Config
	collector *metrics.Collector
	guard     *security.Guard
	store     *store.Store
}

func NewImpersonateHandler(cfg *config.Config, collector *metrics.Collector, st *store.Store) *ImpersonateHandler {
	return &ImpersonateHandler{
		cfg:       cfg,
		collector: collector,
		guard:     security.NewGuard(cfg.SSRFAllowPrivate, cfg.SSRFDenyHosts, cfg.SSRFAllowHosts),
		store:     st,
	}
}

func (h *ImpersonateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Read and parse request body
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, h.cfg.MaxRequestBodySize))
	if err != nil {
		models.WriteJSONError(w, http.StatusBadRequest, "validation", "failed to read request body")
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req models.ImpersonateRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		models.WriteJSONError(w, http.StatusBadRequest, "validation", "invalid JSON: "+err.Error())
		return
	}

	// Validate request
	if err := req.Validate(h.cfg.MaxTimeout); err != nil {
		models.WriteJSONError(w, http.StatusBadRequest, "validation", err.Error())
		return
	}

	// SSRF protection: block internal/metadata destinations.
	if err := h.guard.ValidateURL(req.URL); err != nil {
		models.WriteJSONError(w, http.StatusBadRequest, "validation", err.Error())
		return
	}

	// Resolve browser name
	browserName := models.ResolveBrowserName(req.Browser)

	// Get browser config
	browserConfig, err := models.GetBrowserConfig(browserName)
	if err != nil {
		models.WriteJSONError(w, http.StatusBadRequest, "validation", err.Error())
		return
	}

	// Execute curl-impersonate
	response, err := executor.Execute(&req, browserConfig, h.cfg.MaxResponseBodySize)
	if err != nil {
		// Internal service error
		h.collector.RecordRequest(browserName, false, time.Since(start))
		models.WriteJSONError(w, http.StatusInternalServerError, "internal", "failed to execute request: "+err.Error())
		return
	}

	// Record metrics
	duration := time.Since(start)
	h.collector.RecordRequest(browserName, response.Success, duration)
	h.recordUsage(r, &req, browserName, response, duration)

	// Return response (always 200, even for network errors)
	models.WriteJSON(w, http.StatusOK, response)
}

// recordUsage persists a usage-log entry. Only the target host is stored, never
// the full URL, body or headers.
func (h *ImpersonateHandler) recordUsage(r *http.Request, req *models.ImpersonateRequest, browser string, resp *models.ImpersonateResponse, d time.Duration) {
	if h.store == nil {
		return
	}
	host := ""
	if u, err := url.Parse(req.URL); err == nil {
		host = u.Hostname()
	}
	_ = h.store.AddLog(store.LogEntry{
		TokenName:  middleware.TokenName(r.Context()),
		Browser:    browser,
		Method:     req.Method,
		TargetHost: host,
		StatusCode: resp.StatusCode,
		Success:    resp.Success,
		DurationMs: d.Milliseconds(),
		ErrorType:  resp.ErrorType,
	})
}
