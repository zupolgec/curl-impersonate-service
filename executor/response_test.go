package executor

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/zupolgec/curl-impersonate-service/models"
)

func TestBuildCurlArgsEnablesCompressedResponses(t *testing.T) {
	req := &models.ImpersonateRequest{Method: "GET", Timeout: 30}

	args, err := buildCurlArgs(req, "https://example.com", 0)
	if err != nil {
		t.Fatalf("buildCurlArgs() error = %v", err)
	}

	for _, arg := range args {
		if arg == "--compressed" {
			return
		}
	}
	t.Fatalf("buildCurlArgs() = %q, want --compressed", args)
}

func TestParseSuccessResponseReturnsDecodedBody(t *testing.T) {
	output := []byte("HTTP/2 200\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Encoding: gzip\r\n" +
		"Content-Length: 123\r\n\r\n" +
		"<h1>Città</h1>" +
		"\n---TIMING---\n" +
		`{"time_total":0.1,"time_namelookup":0.01,"time_connect":0.02,"time_starttransfer":0.05}`)

	response, err := parseSuccessResponse(output, "https://example.com")
	if err != nil {
		t.Fatalf("parseSuccessResponse() error = %v", err)
	}

	if response.Body != "<h1>Città</h1>" {
		t.Errorf("Body = %q, want decoded UTF-8 body", response.Body)
	}
	if response.BodyBase64 {
		t.Error("BodyBase64 = true, want false")
	}
	if _, ok := response.Headers["Content-Encoding"]; ok {
		t.Error("Content-Encoding was not removed")
	}
	if _, ok := response.Headers["Content-Length"]; ok {
		t.Error("stale Content-Length was not removed")
	}
}

func TestSetResponseBodyBase64EncodesInvalidUTF8(t *testing.T) {
	body := []byte{0x1f, 0x8b, 0x08, 0xff}
	response := &models.ImpersonateResponse{
		Headers: map[string][]string{"Content-Type": {"text/html"}},
	}

	setResponseBody(response, body)

	if !response.BodyBase64 {
		t.Error("BodyBase64 = false, want true")
	}
	if response.Body != base64.StdEncoding.EncodeToString(body) {
		t.Errorf("Body = %q, want lossless base64 encoding", response.Body)
	}
	decoded, err := base64.StdEncoding.DecodeString(response.Body)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	if !reflect.DeepEqual(decoded, body) {
		t.Errorf("decoded body = %x, want %x", decoded, body)
	}
}
