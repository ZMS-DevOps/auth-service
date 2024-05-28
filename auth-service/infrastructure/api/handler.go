package api

import (
	"encoding/json"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"net/http"
)

func handleError(w http.ResponseWriter, httpStatus int, message string) {
	w.WriteHeader(httpStatus)
	if _, err := w.Write([]byte(message)); err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
	}
}

func writeResponse(w http.ResponseWriter, httpStatus int, data interface{}) {
	w.Header().Set(domain.ContentType, domain.JsonContentType)
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
	}
}
