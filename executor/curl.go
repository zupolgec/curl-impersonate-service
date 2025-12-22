package executor

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	"github.com/zupolgec/curl-impersonate-service/models"
)

const (
	curlImpersonateChrome = "/usr/local/bin/curl-impersonate-chrome"
	curlImpersonateFF     = "/usr/local/bin/curl-impersonate-ff"
)

type CurlTiming struct {
	TimeTotal         float64 `json:"time_total"`
	TimeNameLookup    float64 `json:"time_namelookup"`
	TimeConnect       float64 `json:"time_connect"`
	TimeStartTransfer float64 `json:"time_starttransfer"`
}

// Execute runs curl-impersonate with the given request
func Execute(req *models.ImpersonateRequest, browserConfig models.BrowserConfig) (*models.ImpersonateResponse, error) {
	// Merge query params
	finalURL, err := mergeQueryParams(req.URL, req.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Determine binary path
	binaryPath, err := getBinaryPath(browserConfig.Binary)
	if err != nil {
		return nil, err
	}

	// Build curl command
	args, err := buildCurlArgs(req, finalURL)
	if err != nil {
		return nil, err
	}

	// Execute curl
	cmd := exec.Command(binaryPath, args...)
	output, err := cmd.CombinedOutput()

	// Parse response even if there's an error (might be network error)
	if err != nil {
		return parseErrorResponse(output, err)
	}

	return parseSuccessResponse(output, finalURL)
}

func getBinaryPath(binaryName string) (string, error) {
	switch binaryName {
	case "curl-impersonate-chrome":
		return curlImpersonateChrome, nil
	case "curl-impersonate-ff":
		return curlImpersonateFF, nil
	default:
		return "", fmt.Errorf("unknown binary: %s", binaryName)
	}
}

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

	// Parse HTTP response
	reader := bufio.NewReader(bytes.NewReader(responseData))
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTTP response: %w", err)
	}
	defer resp.Body.Close()

	// Read body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse timing
	var timing CurlTiming
	if err := json.Unmarshal(timingData, &timing); err != nil {
		return nil, fmt.Errorf("failed to parse timing: %w", err)
	}

	// Build response
	response := &models.ImpersonateResponse{
		Success:    true,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		FinalURL:   requestedURL, // curl doesn't easily provide final URL, use requested
		Timing: &models.Timing{
			Total:         timing.TimeTotal,
			NameLookup:    timing.TimeNameLookup,
			Connect:       timing.TimeConnect,
			StartTransfer: timing.TimeStartTransfer,
		},
	}

	// Determine if response is text or binary
	contentType := resp.Header.Get("Content-Type")
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
