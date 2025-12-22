package models

import (
	"encoding/json"
	"testing"
)

func TestImpersonateRequest_Validate(t *testing.T) {
	tests := []struct {
		name       string
		req        ImpersonateRequest
		maxTimeout int
		wantErr    bool
	}{
		{
			name: "valid request",
			req: ImpersonateRequest{
				URL:    "https://example.com",
				Method: "GET",
			},
			maxTimeout: 120,
			wantErr:    false,
		},
		{
			name: "missing URL",
			req: ImpersonateRequest{
				Method: "GET",
			},
			maxTimeout: 120,
			wantErr:    true,
		},
		{
			name: "timeout exceeds max",
			req: ImpersonateRequest{
				URL:     "https://example.com",
				Timeout: 200,
			},
			maxTimeout: 120,
			wantErr:    true,
		},
		{
			name: "both body and body_base64",
			req: ImpersonateRequest{
				URL:        "https://example.com",
				Body:       "test",
				BodyBase64: "dGVzdA==",
			},
			maxTimeout: 120,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(tt.maxTimeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImpersonateRequest_DefaultMethod(t *testing.T) {
	req := ImpersonateRequest{
		URL: "https://example.com",
	}

	if err := req.Validate(120); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("Method = %q, want %q", req.Method, "GET")
	}
}

func TestImpersonateRequest_DefaultTimeout(t *testing.T) {
	req := ImpersonateRequest{
		URL: "https://example.com",
	}

	if err := req.Validate(120); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if req.Timeout != 30 {
		t.Errorf("Timeout = %d, want %d", req.Timeout, 30)
	}
}

func TestImpersonateRequest_FollowRedirectsDefault(t *testing.T) {
	jsonData := `{"url": "https://example.com"}`

	var req ImpersonateRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !req.FollowRedirects {
		t.Error("FollowRedirects should default to true")
	}
}
