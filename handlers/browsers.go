package handlers

import (
	"net/http"

	"github.com/zupolgec/curl-impersonate-service/models"
)

func BrowsersHandler(w http.ResponseWriter, r *http.Request) {
	response := models.BrowsersResponse{
		Browsers: models.GetAllBrowsers(),
		Aliases:  models.GetAliases(),
		Default:  models.GetDefaultBrowser(),
	}
	models.WriteJSON(w, http.StatusOK, response)
}
