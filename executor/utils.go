package executor

import (
	"net/url"
	"strings"
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
