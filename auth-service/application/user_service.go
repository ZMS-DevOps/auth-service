package application

import (
	"encoding/json"
	"errors"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/dto"
	"log"
	"net/http"
)

type UserService struct {
	HttpClient           *http.Client
	AuthService          *AuthService
	KeycloakService      *KeycloakService
	IdentityProviderHost string
}

func NewUserService(httpClient *http.Client, authService *AuthService, keycloakService *KeycloakService, identityProviderHost string) *UserService {

	return &UserService{
		HttpClient:           httpClient,
		AuthService:          authService,
		KeycloakService:      keycloakService,
		IdentityProviderHost: identityProviderHost,
	}
}

func (service *UserService) GetUser(authorizationHeader string) (*dto.UserDTO, error) {
	responseBody, err := service.KeycloakService.GetKeycloakUser(authorizationHeader)
	if err != nil {
		return nil, err
	}

	userDTO := &dto.UserDTO{}
	if err := json.NewDecoder(responseBody).Decode(userDTO); err != nil {
		return nil, errors.New("can't decode user payload")
	}

	return userDTO, nil
}

func (service *UserService) GetUserById(authorizationHeader, id string) (*dto.UserDTO, error) {
	log.Printf("GETTING USER ID: %s", id)
	responseBody, err := service.KeycloakService.GetKeycloakUserById(authorizationHeader, id)
	if err != nil {
		return nil, err
	}

	keycloakDTO := &dto.GettingKeycloakUserDTO{}
	if err := json.NewDecoder(responseBody).Decode(keycloakDTO); err != nil {
		return nil, errors.New("can't decode keycloak user payload")
	}

	return &dto.UserDTO{
		Id:         id,
		GivenName:  keycloakDTO.FirstName,
		FamilyName: keycloakDTO.LastName,
		Email:      keycloakDTO.Email,
		Address:    keycloakDTO.Attributes.Address[0],
	}, nil
}

func (service *UserService) UpdateUser(authorizationHeader, id, firstName, lastName, address string) error {
	requestBody := dto.NewUpdateKeycloakUserDTO(id, firstName, lastName, address)
	if _, err := service.KeycloakService.UpdateKeycloakUser(authorizationHeader, id, requestBody); err != nil {
		return err
	}

	return nil
}

func (service *UserService) DeleteUser(authorizationHeader, id string) error {
	if err := service.KeycloakService.DeleteKeycloakUser(authorizationHeader, id); err != nil {
		return err
	}

	return nil
}

func (service *UserService) ResetPassword(authorizationHeader, id, password string) error {
	requestBody := dto.NewCredentialsDTO(password)
	if err := service.KeycloakService.ResetPasswordOfKeycloakUser(authorizationHeader, id, requestBody); err != nil {
		return err
	}

	return nil
}
