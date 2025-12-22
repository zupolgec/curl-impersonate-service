package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/zupolgec/curl-impersonate-service/config"
	"github.com/zupolgec/curl-impersonate-service/executor"
	"github.com/zupolgec/curl-impersonate-service/metrics"
	"github.com/zupolgec/curl-impersonate-service/models"
)

type ImpersonateHandler struct {
	cfg       *config.Config
	collector *metrics.Collector
}

func NewImpersonateHandler(cfg *config.Config, collector *metrics.Collector) *ImpersonateHandler {
	return &ImpersonateHandler{
		cfg:       cfg,
		collector: collector,
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
	defer r.Body.Close()

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

	// Resolve browser name
	browserName := models.ResolveBrowserName(req.Browser)

	// Get browser config
	browserConfig, err := models.GetBrowserConfig(browserName)
	if err != nil {
		models.WriteJSONError(w, http.StatusBadRequest, "validation", err.Error())
		return
	}

	// Execute curl-impersonate
	response, err := executor.Execute(&req, browserConfig)
	if err != nil {
		// Internal service error
		h.collector.RecordRequest(browserName, false, time.Since(start))
		models.WriteJSONError(w, http.StatusInternalServerError, "internal", "failed to execute request: "+err.Error())
		return
	}

	// Record metrics
	h.collector.RecordRequest(browserName, response.Success, time.Since(start))

	// Return response (always 200, even for network errors)
	models.WriteJSON(w, http.StatusOK, response)
}
