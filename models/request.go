package models

import (
	"encoding/json"
	"fmt"
)

type ImpersonateRequest struct {
	Browser         string            `json:"browser"`
	URL             string            `json:"url"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers"`
	QueryParams     map[string]string `json:"query_params"`
	Body            string            `json:"body"`
	BodyBase64      string            `json:"body_base64"`
	FollowRedirects bool              `json:"follow_redirects"`
	Insecure        bool              `json:"insecure"`
	Timeout         int               `json:"timeout"`
}

// Validate validates the request
func (r *ImpersonateRequest) Validate(maxTimeout int) error {
	if r.URL == "" {
		return fmt.Errorf("url is required")
	}

	if r.Method == "" {
		r.Method = "GET"
	}

	if r.Body != "" && r.BodyBase64 != "" {
		return fmt.Errorf("body and body_base64 are mutually exclusive")
	}

	if r.Timeout <= 0 {
		r.Timeout = 30 // default
	}

	if r.Timeout > maxTimeout {
		return fmt.Errorf("timeout exceeds maximum allowed (%d seconds)", maxTimeout)
	}

	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling with defaults
func (r *ImpersonateRequest) UnmarshalJSON(data []byte) error {
	type Alias ImpersonateRequest
	aux := &struct {
		FollowRedirects *bool `json:"follow_redirects"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	// Set default
	followRedirects := true
	aux.FollowRedirects = &followRedirects

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.FollowRedirects != nil {
		r.FollowRedirects = *aux.FollowRedirects
	}

	return nil
}
