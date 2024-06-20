package application

import (
	"encoding/json"
	"errors"
	booking "github.com/ZMS-DevOps/booking-service/proto"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/mmmajder/zms-devops-auth-service/application/external"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/dto"
	"github.com/mmmajder/zms-devops-auth-service/util"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type UserService struct {
	HttpClient           *http.Client
	AuthService          *AuthService
	KeycloakService      *KeycloakService
	IdentityProviderHost string
	bookingClient        booking.BookingServiceClient
}

func NewUserService(httpClient *http.Client, authService *AuthService, keycloakService *KeycloakService, identityProviderHost string, bookingClient booking.BookingServiceClient) *UserService {

	return &UserService{
		HttpClient:           httpClient,
		AuthService:          authService,
		KeycloakService:      keycloakService,
		IdentityProviderHost: identityProviderHost,
		bookingClient:        bookingClient,
	}
}

func (service *UserService) GetUser(authorizationHeader string, span trace.Span, loki promtail.Client) (*dto.UserDTO, error) {
	util.HttpTraceInfo("Getting user...", span, loki, "GetSpecialPrices", "")
	responseBody, err := service.KeycloakService.GetKeycloakUser(authorizationHeader, span, loki)
	if err != nil {
		return nil, err
	}

	userDTO := &dto.UserDTO{}
	if err := json.NewDecoder(responseBody).Decode(userDTO); err != nil {
		return nil, errors.New("can't decode user payload")
	}

	return userDTO, nil
}

func (service *UserService) GetUserById(authorizationHeader, id string, span trace.Span, loki promtail.Client) (*dto.UserDTO, error) {
	util.HttpTraceInfo("Getting user by id...", span, loki, "GetSpecialPrices", "")
	responseBody, err := service.KeycloakService.GetKeycloakUserById(authorizationHeader, id, span, loki)
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

func (service *UserService) UpdateUser(authorizationHeader, id, firstName, lastName, address string, span trace.Span, loki promtail.Client) error {
	util.HttpTraceInfo("Updating user...", span, loki, "GetSpecialPrices", "")
	requestBody := dto.NewUpdateKeycloakUserDTO(id, firstName, lastName, address)
	if _, err := service.KeycloakService.UpdateKeycloakUser(authorizationHeader, id, requestBody, span, loki); err != nil {
		return err
	}

	return nil
}

func (service *UserService) DeleteUser(authorizationHeader, id, group string, span trace.Span, loki promtail.Client) error {
	util.HttpTraceInfo("Deleting user...", span, loki, "GetSpecialPrices", "")
	var canDeleteUser bool
	if group == domain.HostRole {
		response, err := external.IfHostCanBeDeleted(service.bookingClient, id, span, loki)
		if err != nil {
			return err
		}
		canDeleteUser = response.Success
	} else {
		response, err := external.IfGuestCanBeDeleted(service.bookingClient, id, span, loki)
		if err != nil {
			return err
		}
		canDeleteUser = response.Success
	}

	if canDeleteUser {
		if err := service.KeycloakService.DeleteKeycloakUser(authorizationHeader, id, span, loki); err != nil {
			return err
		}
		return nil
	}
	return errors.New(domain.UserCouldNotBeDeletedErrorMessage)
}

func (service *UserService) ResetPassword(authorizationHeader, id, password string, span trace.Span, loki promtail.Client) error {
	util.HttpTraceInfo("Resetting password...", span, loki, "GetSpecialPrices", "")
	requestBody := dto.NewCredentialsDTO(password)
	if err := service.KeycloakService.ResetPasswordOfKeycloakUser(authorizationHeader, id, requestBody, span, loki); err != nil {
		return err
	}

	return nil
}
