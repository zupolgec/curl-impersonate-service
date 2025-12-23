package executor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/zupolgec/curl-impersonate-service/models"
)

type CurlTiming struct {
	TimeTotal         float64 `json:"time_total"`
	TimeNameLookup    float64 `json:"time_namelookup"`
	TimeConnect       float64 `json:"time_connect"`
	TimeStartTransfer float64 `json:"time_starttransfer"`
}

// executeShell runs curl-impersonate via shell wrapper script
func executeShell(req *models.ImpersonateRequest, browserConfig models.BrowserConfig) (*models.ImpersonateResponse, error) {
	// Merge query params
	finalURL, err := mergeQueryParams(req.URL, req.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Use the wrapper script which sets the correct browser signature
	wrapperScript := "/usr/local/bin/" + browserConfig.WrapperScript

	// Build curl command
	args, err := buildCurlArgs(req, finalURL)
	if err != nil {
		return nil, err
	}

	// Execute curl via wrapper script
	cmd := exec.Command(wrapperScript, args...)
	output, err := cmd.CombinedOutput()

	// Parse response even if there's an error (might be network error)
	if err != nil {
		return parseErrorResponse(output, err)
	}

	return parseSuccessResponse(output, finalURL)
}

func buildCurlArgs(req *models.ImpersonateRequest, finalURL string) ([]string, error) {
	args := []string{
		"-i",             // Include response headers
		"-s",             // Silent mode
		"-X", req.Method, // HTTP method
		"--max-time", strconv.Itoa(req.Timeout),
		"-w", "\n---TIMING---\n{\"time_total\":%{time_total},\"time_namelookup\":%{time_namelookup},\"time_connect\":%{time_connect},\"time_starttransfer\":%{time_starttransfer}}",
	}

	if req.FollowRedirects {
		args = append(args, "-L")
	}

	// Add custom headers
	for key, value := range req.Headers {
		args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
	}

	// Add body
	if req.Body != "" {
		args = append(args, "--data", req.Body)
	} else if req.BodyBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.BodyBase64)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 body: %w", err)
		}
		args = append(args, "--data-binary", string(decoded))
	}

	args = append(args, finalURL)
	return args, nil
}

func parseSuccessResponse(output []byte, requestedURL string) (*models.ImpersonateResponse, error) {
	// Split output into response and timing
	parts := bytes.Split(output, []byte("\n---TIMING---\n"))
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected curl output format")
	}

	responseData := parts[0]
	timingData := parts[1]

	// Parse timing first
	var timing CurlTiming
	if err := json.Unmarshal(timingData, &timing); err != nil {
		return nil, fmt.Errorf("failed to parse timing: %w", err)
	}

	// Split headers and body (separated by blank line)
	headerBodySplit := bytes.SplitN(responseData, []byte("\r\n\r\n"), 2)
	if len(headerBodySplit) < 2 {
		headerBodySplit = bytes.SplitN(responseData, []byte("\n\n"), 2)
	}
	if len(headerBodySplit) != 2 {
		return nil, fmt.Errorf("failed to split headers and body")
	}

	headerLines := bytes.Split(headerBodySplit[0], []byte("\n"))
	bodyBytes := headerBodySplit[1]

	// Parse status line (HTTP/1.1 200 OK or HTTP/2 200)
	statusLine := string(bytes.TrimSpace(headerLines[0]))
	var statusCode int
	if strings.HasPrefix(statusLine, "HTTP/2 ") {
		_, _ = fmt.Sscanf(statusLine, "HTTP/2 %d", &statusCode)
	} else if strings.HasPrefix(statusLine, "HTTP/1") {
		_, _ = fmt.Sscanf(statusLine, "HTTP/1.%d %d", new(int), &statusCode)
	} else {
		return nil, fmt.Errorf("unrecognized HTTP version: %s", statusLine)
	}

	// Parse headers
	headers := make(map[string][]string)
	for i := 1; i < len(headerLines); i++ {
		line := string(bytes.TrimSpace(headerLines[i]))
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Normalize header names to canonical form
			key = http.CanonicalHeaderKey(key)
			headers[key] = append(headers[key], value)
		}
	}

	// Build response
	response := &models.ImpersonateResponse{
		Success:    true,
		StatusCode: statusCode,
		Headers:    headers,
		FinalURL:   requestedURL,
		Timing: &models.Timing{
			Total:         timing.TimeTotal,
			NameLookup:    timing.TimeNameLookup,
			Connect:       timing.TimeConnect,
			StartTransfer: timing.TimeStartTransfer,
		},
	}

	// Determine if response is text or binary
	contentType := ""
	if ct, ok := headers["Content-Type"]; ok && len(ct) > 0 {
		contentType = ct[0]
	}

	if isTextContent(contentType) {
		response.Body = string(bodyBytes)
		response.BodyBase64 = false
	} else {
		response.Body = base64.StdEncoding.EncodeToString(bodyBytes)
		response.BodyBase64 = true
	}

	return response, nil
}

func parseErrorResponse(output []byte, cmdErr error) (*models.ImpersonateResponse, error) {
	errorMsg := string(output)
	if errorMsg == "" {
		errorMsg = cmdErr.Error()
	}

	// Determine error type based on curl output
	errorType := "network"
	if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "timed out") {
		errorType = "timeout"
	} else if strings.Contains(errorMsg, "Could not resolve host") {
		errorType = "dns"
	} else if strings.Contains(errorMsg, "SSL") || strings.Contains(errorMsg, "certificate") {
		errorType = "ssl"
	}

	return &models.ImpersonateResponse{
		Success:    false,
		Error:      errorMsg,
		ErrorType:  errorType,
		StatusCode: 0,
	}, nil
}
