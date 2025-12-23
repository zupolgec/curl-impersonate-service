package models

import (
	"encoding/json"
	"net/http"
)

type Timing struct {
	Total         float64 `json:"total"`
	NameLookup    float64 `json:"namelookup"`
	Connect       float64 `json:"connect"`
	StartTransfer float64 `json:"starttransfer"`
}

type ImpersonateResponse struct {
	Success    bool                `json:"success"`
	StatusCode int                 `json:"status_code,omitempty"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       string              `json:"body,omitempty"`
	BodyBase64 bool                `json:"body_base64,omitempty"`
	FinalURL   string              `json:"final_url,omitempty"`
	Timing     *Timing             `json:"timing,omitempty"`
	Error      string              `json:"error,omitempty"`
	ErrorType  string              `json:"error_type,omitempty"`
}

type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	ErrorType string `json:"error_type"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type BrowsersResponse struct {
	Browsers []BrowserConfig   `json:"browsers"`
	Aliases  map[string]string `json:"aliases"`
	Default  string            `json:"default"`
}

type MetricsResponse struct {
	UptimeSeconds     int64            `json:"uptime_seconds"`
	RequestsTotal     int64            `json:"requests_total"`
	RequestsSuccess   int64            `json:"requests_success"`
	RequestsFailed    int64            `json:"requests_failed"`
	AverageDurationMs float64          `json:"average_duration_ms"`
	BrowsersUsed      map[string]int64 `json:"browsers_used"`
}

// Helper functions to create responses
func NewErrorResponse(errorType, message string) ErrorResponse {
	return ErrorResponse{
		Success:   false,
		Error:     message,
		ErrorType: errorType,
	}
}

func WriteJSONError(w http.ResponseWriter, statusCode int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := NewErrorResponse(errorType, message)
	_ = json.NewEncoder(w).Encode(resp)
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
