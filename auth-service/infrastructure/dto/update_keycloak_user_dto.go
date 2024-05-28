package dto

type UpdateKeycloakUserDTO struct {
	Id            string              `json:"id"`
	FirstName     string              `json:"firstName"`
	LastName      string              `json:"lastName"`
	Enabled       bool                `json:"enabled"`
	EmailVerified bool                `json:"emailVerified"`
	Attributes    map[string][]string `json:"attributes"`
}

func NewUpdateKeycloakUserDTO(id, firstName, lastName, address string) *UpdateKeycloakUserDTO {
	return &UpdateKeycloakUserDTO{
		Id:            id,
		FirstName:     firstName,
		LastName:      lastName,
		Enabled:       true,
		EmailVerified: true,
		Attributes: map[string][]string{
			"address": {address},
		},
	}
}
