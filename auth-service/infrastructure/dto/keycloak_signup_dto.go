package dto

type Access struct {
	ManageGroupMembership bool `json:"manageGroupMembership"`
	View                  bool `json:"view"`
	MapRoles              bool `json:"mapRoles"`
	Impersonate           bool `json:"impersonate"`
	Manage                bool `json:"manage"`
}

type KeycloakDTO struct {
	Username      string              `json:"username"`
	Email         string              `json:"email"`
	FirstName     string              `json:"firstName"`
	LastName      string              `json:"lastName"`
	Enabled       bool                `json:"enabled"`
	EmailVerified bool                `json:"emailVerified"`
	Groups        []string            `json:"groups"`
	Credentials   []CredentialsDTO    `json:"credentials"`
	Attributes    map[string][]string `json:"attributes,omitempty"`
	Access        *Access             `json:"access"`
}

func NewKeycloakDTO(email, firstName, lastName, password, address, group string) *KeycloakDTO {
	return &KeycloakDTO{
		Username:      email,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		Enabled:       true,
		EmailVerified: false,
		Groups:        []string{group},
		Credentials: []CredentialsDTO{
			*NewCredentialsDTO(password),
		},
		Attributes: map[string][]string{
			"address": {address},
		},
		Access: &Access{
			ManageGroupMembership: true,
			View:                  true,
			MapRoles:              true,
			Impersonate:           true,
			Manage:                true,
		},
	}
}
