package dto

type UserDTO struct {
	Id            string   `json:"sub"`
	Email         string   `json:"email"`
	GivenName     string   `json:"given_name"`
	FamilyName    string   `json:"family_name"`
	Address       string   `json:"address"`
	EmailVerified bool     `json:"email_verified"`
	Role          []string `json:"role"`
}
