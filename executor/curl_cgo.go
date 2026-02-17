//go:build cgo
// +build cgo

package executor

/*
#cgo LDFLAGS: -lcurl-impersonate-chrome
#include <stdlib.h>
#include <string.h>
#include "curl_wrappers.h"

// Callback to write response data to a buffer
size_t write_callback(void *ptr, size_t size, size_t nmemb, void *userdata) {
    size_t realsize = size * nmemb;
    struct {
        char *data;
        size_t size;
    } *mem = userdata;

    char *ptr2 = realloc(mem->data, mem->size + realsize + 1);
    if (ptr2 == NULL) return 0;

    mem->data = ptr2;
    memcpy(&(mem->data[mem->size]), ptr, realsize);
    mem->size += realsize;
    mem->data[mem->size] = 0;

    return realsize;
}
*/
import "C"

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"unsafe"

	"github.com/zupolgec/curl-impersonate-service/models"
)

type responseBuffer struct {
	data *C.char
	size C.size_t
}

func Execute(req *models.ImpersonateRequest, browserConfig models.BrowserConfig) (*models.ImpersonateResponse, error) {
	// CGO executor only supports Chrome-based browsers (uses libcurl-impersonate-chrome)
	// Firefox browsers need the Firefox SSL library, so fall back to shell execution
	if browserConfig.Binary == "curl-impersonate-ff" {
		return executeShell(req, browserConfig)
	}

	// Initialize curl
	curl := C.curl_easy_init()
	if curl == nil {
		return nil, fmt.Errorf("failed to initialize curl")
	}
	defer C.curl_easy_cleanup(curl)

	// Set URL
	finalURL, err := mergeQueryParams(req.URL, req.QueryParams)
	if err != nil {
		return nil, err
	}
	cURL := C.CString(finalURL)
	defer C.free(unsafe.Pointer(cURL))
	C._curl_easy_setopt_ptr(curl, C.CURLOPT_URL, unsafe.Pointer(cURL))

	// Set Method
	cMethod := C.CString(req.Method)
	defer C.free(unsafe.Pointer(cMethod))
	C._curl_easy_setopt_ptr(curl, C.CURLOPT_CUSTOMREQUEST, unsafe.Pointer(cMethod))

	// Set Impersonate
	cBrowser := C.CString(browserConfig.Name)
	defer C.free(unsafe.Pointer(cBrowser))
	// 1 means add default headers
	impersonateRes := C.curl_easy_impersonate(curl, cBrowser, 1)
	if impersonateRes != 0 {
		return nil, fmt.Errorf("failed to impersonate %s: %d", browserConfig.Name, int(impersonateRes))
	}

	// Set Headers
	var headerList *C.struct_curl_slist
	for k, v := range req.Headers {
		header := fmt.Sprintf("%s: %s", k, v)
		cHeader := C.CString(header)
		headerList = C.curl_slist_append(headerList, cHeader)
		C.free(unsafe.Pointer(cHeader))
	}
	if headerList != nil {
		C._curl_easy_setopt_ptr(curl, C.CURLOPT_HTTPHEADER, unsafe.Pointer(headerList))
		defer C.curl_slist_free_all(headerList)
	}

	// Set Body
	var bodyBytes []byte
	if req.Body != "" {
		bodyBytes = []byte(req.Body)
	} else if req.BodyBase64 != "" {
		bodyBytes, err = base64.StdEncoding.DecodeString(req.BodyBase64)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 body: %w", err)
		}
	}
	if len(bodyBytes) > 0 {
		C._curl_easy_setopt_ptr(curl, C.CURLOPT_POSTFIELDS, unsafe.Pointer(&bodyBytes[0]))
		C._curl_easy_setopt_long(curl, C.CURLOPT_POSTFIELDSIZE, C.long(len(bodyBytes)))
	}

	// Set Redirects
	if req.FollowRedirects {
		C._curl_easy_setopt_long(curl, C.CURLOPT_FOLLOWLOCATION, 1)
	}

	// Set Timeout
	C._curl_easy_setopt_long(curl, C.CURLOPT_TIMEOUT, C.long(req.Timeout))

	// Disable SSL verification if requested
	// CURLOPT_SSL_VERIFYPEER = 64, CURLOPT_SSL_VERIFYHOST = 81
	if req.Insecure {
		C._curl_easy_setopt_long(curl, 64, 0)
		C._curl_easy_setopt_long(curl, 81, 0)
	}

	// Setup Response Buffers
	var respBuf responseBuffer
	var headerBuf responseBuffer

	C._curl_easy_setopt_ptr(curl, C.CURLOPT_WRITEFUNCTION, unsafe.Pointer(C.write_callback))
	C._curl_easy_setopt_ptr(curl, C.CURLOPT_WRITEDATA, unsafe.Pointer(&respBuf))

	C._curl_easy_setopt_ptr(curl, C.CURLOPT_HEADERFUNCTION, unsafe.Pointer(C.write_callback))
	C._curl_easy_setopt_ptr(curl, C.CURLOPT_HEADERDATA, unsafe.Pointer(&headerBuf))

	// Execute
	res := C.curl_easy_perform(curl)

	// Handle Cleanup for buffers
	defer func() {
		if respBuf.data != nil {
			C.free(unsafe.Pointer(respBuf.data))
		}
		if headerBuf.data != nil {
			C.free(unsafe.Pointer(headerBuf.data))
		}
	}()

	if res != C.CURLE_OK {
		errStr := C.GoString(C.curl_easy_strerror(res))
		errorType := "network"
		if res == C.CURLE_OPERATION_TIMEDOUT {
			errorType = "timeout"
		} else if res == C.CURLE_COULDNT_RESOLVE_HOST {
			errorType = "dns"
		} else if res == C.CURLE_SSL_CONNECT_ERROR || res == C.CURLE_PEER_FAILED_VERIFICATION {
			errorType = "ssl"
		}

		return &models.ImpersonateResponse{
			Success:   false,
			Error:     errStr,
			ErrorType: errorType,
		}, nil
	}

	// Get Status Code
	var statusCode C.long
	C._curl_easy_getinfo_ptr(curl, C.CURLINFO_RESPONSE_CODE, unsafe.Pointer(&statusCode))

	// Get Final URL
	var finalURLPtr *C.char
	C._curl_easy_getinfo_ptr(curl, C.CURLINFO_EFFECTIVE_URL, unsafe.Pointer(&finalURLPtr))

	// Get Timings
	var tTotal, tName, tConnect, tStart C.double
	C._curl_easy_getinfo_ptr(curl, C.CURLINFO_TOTAL_TIME, unsafe.Pointer(&tTotal))
	C._curl_easy_getinfo_ptr(curl, C.CURLINFO_NAMELOOKUP_TIME, unsafe.Pointer(&tName))
	C._curl_easy_getinfo_ptr(curl, C.CURLINFO_CONNECT_TIME, unsafe.Pointer(&tConnect))
	C._curl_easy_getinfo_ptr(curl, C.CURLINFO_STARTTRANSFER_TIME, unsafe.Pointer(&tStart))

	// Parse headers
	headers := make(map[string][]string)
	if headerBuf.data != nil {
		hStr := C.GoStringN(headerBuf.data, C.int(headerBuf.size))
		hLines := strings.Split(hStr, "\r\n")
		for i, line := range hLines {
			if i == 0 || line == "" {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := http.CanonicalHeaderKey(strings.TrimSpace(parts[0]))
				val := strings.TrimSpace(parts[1])
				headers[key] = append(headers[key], val)
			}
		}
	}

	response := &models.ImpersonateResponse{
		Success:    true,
		StatusCode: int(statusCode),
		Headers:    headers,
		FinalURL:   C.GoString(finalURLPtr),
		Timing: &models.Timing{
			Total:         float64(tTotal),
			NameLookup:    float64(tName),
			Connect:       float64(tConnect),
			StartTransfer: float64(tStart),
		},
	}

	// Body handling
	if respBuf.data != nil {
		bodyBytes := C.GoBytes(unsafe.Pointer(respBuf.data), C.int(respBuf.size))
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
	}

	return response, nil
}
