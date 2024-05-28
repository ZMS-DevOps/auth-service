package dto

type GettingKeycloakUserDTO struct {
	ID               string `json:"id"`
	CreatedTimestamp int64  `json:"createdTimestamp"`
	Username         string `json:"username"`
	Enabled          bool   `json:"enabled"`
	Totp             bool   `json:"totp"`
	EmailVerified    bool   `json:"emailVerified"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Email            string `json:"email"`
	Attributes       struct {
		Address []string `json:"address"`
	} `json:"attributes"`
}
