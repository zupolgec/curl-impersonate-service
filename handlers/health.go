package handlers

import (
	"net/http"

	"github.com/zupolgec/curl-impersonate-service/config"
	"github.com/zupolgec/curl-impersonate-service/models"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:  "ok",
		Version: config.Version,
	}
	models.WriteJSON(w, http.StatusOK, response)
}
