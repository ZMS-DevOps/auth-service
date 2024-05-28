package application

import (
	"errors"
	"fmt"
	"github.com/mmmajder/zms-devops-auth-service/domain"
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

func (service *KeycloakService) LoginKeycloakUser(email, password string) (io.ReadCloser, error) {
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

func (service *KeycloakService) GetKeycloakUser(authorizationHeader string) (io.ReadCloser, error) {
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
