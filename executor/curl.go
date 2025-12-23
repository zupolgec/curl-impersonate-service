//go:build !cgo
// +build !cgo

package executor

import (
	"github.com/zupolgec/curl-impersonate-service/models"
)

// Execute runs curl-impersonate with the given request (non-CGO version uses shell)
func Execute(req *models.ImpersonateRequest, browserConfig models.BrowserConfig) (*models.ImpersonateResponse, error) {
	return executeShell(req, browserConfig)
}
