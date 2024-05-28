package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-auth-service/application"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/request"
	"net/http"
)

type AuthHandler struct {
	authService  *application.AuthService
	emailService *application.EmailService
}

func NewAuthHandler(authService *application.AuthService, emailService *application.EmailService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		emailService: emailService,
	}
}

func (handler *AuthHandler) Init(router *mux.Router) {
	router.HandleFunc(domain.AuthContextPath+"/login", handler.Login).Methods(http.MethodPost)
	router.HandleFunc(domain.AuthContextPath+"/health", handler.GetHealthCheck).Methods(http.MethodGet)
}

func (handler *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest request.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid login payload")
		return
	}

	if err := loginRequest.AreValidRequestData(); err != nil {
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := handler.authService.Login(loginRequest.Email, loginRequest.Password)

	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, res)
}

func (handler *AuthHandler) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, domain.HealthCheckMessage)
}
