package application

import (
	"encoding/json"
	"errors"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/dto"
	"net/http"
)

type AuthService struct {
	store           domain.VerificationStore
	HttpClient      *http.Client
	KeycloakService *KeycloakService
}

func NewAuthService(store domain.VerificationStore, httpClient *http.Client, keycloakService *KeycloakService) *AuthService {
	return &AuthService{
		store:           store,
		HttpClient:      httpClient,
		KeycloakService: keycloakService,
	}
}

func (service *AuthService) Login(email string, password string) (*dto.LoginDTO, error) {
	responseBody, err := service.KeycloakService.LoginKeycloakUser(email, password)
	if err != nil {
		return nil, err
	}

	loginDTO := &dto.LoginDTO{}
	if err := json.NewDecoder(responseBody).Decode(loginDTO); err != nil {
		return nil, errors.New("can't decode login payload")
	}
	if !service.userIsEnabled(loginDTO) {
		return nil, errors.New(domain.UserNotVerifiedErrorMessage)
	}

	return loginDTO, nil
}

func (service *AuthService) userIsEnabled(loginDTO *dto.LoginDTO) bool {
	responseBody, err := service.KeycloakService.GetKeycloakUser(domain.BearerSchema + loginDTO.AccessToken)
	if err != nil {
		return false
	}
	userDTO := &dto.UserDTO{}
	if err := json.NewDecoder(responseBody).Decode(userDTO); err != nil {
		return false
	}

	if !userDTO.EmailVerified {
		return false
	}
	return true
}
