package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/dto"
	"github.com/mmmajder/zms-devops-auth-service/util"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type KeycloakService struct {
	HttpClient           *http.Client
	IdentityProviderHost string
}

func NewKeycloakService(httpClient *http.Client, identityProviderHost string) *KeycloakService {

	return &KeycloakService{
		HttpClient:           httpClient,
		IdentityProviderHost: identityProviderHost,
	}
}

func (service *KeycloakService) LoginKeycloakUser(email, password string, span trace.Span, loki promtail.Client) (io.ReadCloser, error) {
	util.HttpTraceInfo("Logging in Keycloak user...", span, loki, "LoginKeycloakUser", "")
	formData := url.Values{
		"client_id":  {"Istio"},
		"username":   {email},
		"password":   {password},
		"grant_type": {"password"},
		"scope":      {"openid"},
	}
	encodedData := formData.Encode()

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://keycloak.backend.svc.cluster.local/realms/Istio/protocol/openid-connect/token"),
		strings.NewReader(encodedData),
	)
	if err != nil {
		return nil, err
	}

	service.setRequestHeader(req, "application/x-www-form-urlencoded")

	resp, err := service.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("login failed")
	}

	return resp.Body, nil
}

func (service *KeycloakService) CreateKeycloakUser(signupDTO *dto.KeycloakDTO, authorizationHeader string, span trace.Span, loki promtail.Client) (string, error) {
	util.HttpTraceInfo("Creating Keycloak user...", span, loki, "CreateKeycloakUser", "")
	jsonBody, err := json.Marshal(signupDTO)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(
		http.MethodPost,
		service.getDefaultAdminConsoleUrl(),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", err
	}
	setContentType(req, domain.JsonContentType)
	req.Header.Set(domain.Authorization, authorizationHeader)

	resp, err := service.HttpClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusConflict {
			return "", errors.New("user exists with same email")
		}
		return "", errors.New("registration failed")
	}

	return service.getUserIdFromLocation(resp, span, loki)
}

func (service *KeycloakService) GetKeycloakUser(authorizationHeader string, span trace.Span, loki promtail.Client) (io.ReadCloser, error) {
	util.HttpTraceInfo("Getting Keycloak user...", span, loki, "GetKeycloakUser", "")
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/protocol/openid-connect/userinfo", service.getDefaultUserConsoleUrl()),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set(domain.Authorization, authorizationHeader)

	resp, err := service.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("getting user failed")
	}

	return resp.Body, nil
}

func (service *KeycloakService) GetKeycloakUserById(authorizationHeader, id string, span trace.Span, loki promtail.Client) (io.ReadCloser, error) {
	util.HttpTraceInfo("Getting Keycloak user by id...", span, loki, "GetKeycloakUserById", "")
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/%s?q=address", service.getDefaultAdminConsoleUrl(), id),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(domain.Authorization, authorizationHeader)

	resp, err := service.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("getting user failed")
	}

	return resp.Body, nil
}

func (service *KeycloakService) UpdateKeycloakUser(authorizationHeader string, id string, updateUser *dto.UpdateKeycloakUserDTO, span trace.Span, loki promtail.Client) (io.ReadCloser, error) {
	util.HttpTraceInfo("Updating Keycloak user...", span, loki, "UpdateKeycloakUser", "")
	jsonBody, err := json.Marshal(updateUser)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/%s", service.getDefaultAdminConsoleUrl(), id),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, err
	}
	setContentType(req, domain.JsonContentType)
	req.Header.Set(domain.Authorization, authorizationHeader)

	resp, err := service.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return nil, errors.New("updating user failed")
	}

	return resp.Body, nil
}

func (service *KeycloakService) DeleteKeycloakUser(authorizationHeader, id string, span trace.Span, loki promtail.Client) error {
	util.HttpTraceInfo("Deleting Keycloak user...", span, loki, "DeleteKeycloakUser", "")
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/%s", service.getDefaultAdminConsoleUrl(), id),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set(domain.Authorization, authorizationHeader)

	resp, err := service.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("deleting user failed")
	}
	return nil
}

func (service *KeycloakService) ResetPasswordOfKeycloakUser(authorizationHeader string, id string, requestBody *dto.CredentialsDTO, span trace.Span, loki promtail.Client) error {
	util.HttpTraceInfo("Resetting Keycloak user password...", span, loki, "ResetPasswordOfKeycloakUser", "")
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/%s/reset-password", service.getDefaultAdminConsoleUrl(), id),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return err
	}
	setContentType(req, domain.JsonContentType)
	req.Header.Set(domain.Authorization, authorizationHeader)

	resp, err := service.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("resetting user password failed")
	}
	return nil
}

func (service *KeycloakService) getUserIdFromLocation(resp *http.Response, span trace.Span, loki promtail.Client) (string, error) {
	util.HttpTraceInfo("Getting user id from location...", span, loki, "getUserIdFromLocation", "")
	if location := resp.Header.Get(domain.LocationHeader); location != "" {
		parts := strings.Split(location, "/")
		if value := service.checkIfFieldHasValue(parts, span, loki); value != "" {
			return value, nil
		}
	}

	return "", errors.New("user not created")
}

func (service *KeycloakService) checkIfFieldHasValue(parts []string, span trace.Span, loki promtail.Client) string {
	util.HttpTraceInfo("Checking if field has value...", span, loki, "checkIfFieldHasValue", "")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		return lastPart
	}
	return ""
}

func setContentType(req *http.Request, contentType string) {
	req.Header.Set(domain.ContentType, contentType)
}

func (service *KeycloakService) setRequestHeader(req *http.Request, contentType string) {
	req.Header.Set(domain.ContentType, contentType)
	req.Header.Add("Host", service.IdentityProviderHost)
}

func (service *KeycloakService) getDefaultAdminConsoleUrl() string {
	return fmt.Sprintf("http://%s/admin/realms/Istio/users", service.IdentityProviderHost)
}

func (service *KeycloakService) getDefaultUserConsoleUrl() string {
	return fmt.Sprintf("http://%s/realms/Istio", service.IdentityProviderHost)
}
