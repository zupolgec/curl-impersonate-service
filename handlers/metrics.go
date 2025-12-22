package handlers

import (
	"net/http"

	"github.com/zupolgec/curl-impersonate-service/metrics"
	"github.com/zupolgec/curl-impersonate-service/models"
)

type MetricsHandler struct {
	collector *metrics.Collector
}

func NewMetricsHandler(collector *metrics.Collector) *MetricsHandler {
	return &MetricsHandler{collector: collector}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uptime, total, success, failed, avgDuration, browsers := h.collector.GetMetrics()

	response := models.MetricsResponse{
		UptimeSeconds:     uptime,
		RequestsTotal:     total,
		RequestsSuccess:   success,
		RequestsFailed:    failed,
		AverageDurationMs: avgDuration,
		BrowsersUsed:      browsers,
	}

	models.WriteJSON(w, http.StatusOK, response)
}
