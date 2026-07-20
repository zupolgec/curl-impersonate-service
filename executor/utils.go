package executor

import (
	"encoding/base64"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/zupolgec/curl-impersonate-service/models"
)

func mergeQueryParams(urlStr string, queryParams map[string]string) (string, error) {
	if len(queryParams) == 0 {
		return urlStr, nil
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for key, value := range queryParams {
		q.Set(key, value) // Set overwrites existing
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func isTextContent(contentType string) bool {
	if contentType == "" {
		return true // default to text
	}

	textPrefixes := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/javascript",
		"application/x-www-form-urlencoded",
		"application/ld+json",
	}

	contentType = strings.ToLower(contentType)
	for _, prefix := range textPrefixes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}

	return false
}

// setResponseBody stores decoded text directly and safely falls back to base64
// for binary data or malformed text. curl has already decoded Content-Encoding
// at this point, so representation headers referring to the wire body no longer
// describe the body returned in the JSON envelope.
func setResponseBody(response *models.ImpersonateResponse, body []byte) {
	if _, decoded := response.Headers["Content-Encoding"]; decoded {
		delete(response.Headers, "Content-Encoding")
		delete(response.Headers, "Content-Length")
	}

	if len(body) == 0 {
		return
	}

	contentType := ""
	if values := response.Headers["Content-Type"]; len(values) > 0 {
		contentType = values[0]
	}

	if isTextContent(contentType) && utf8.Valid(body) {
		response.Body = string(body)
		response.BodyBase64 = false
		return
	}

	response.Body = base64.StdEncoding.EncodeToString(body)
	response.BodyBase64 = true
}
