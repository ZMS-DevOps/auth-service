package dto

type CredentialsDTO struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

func NewCredentialsDTO(value string) *CredentialsDTO {
	return &CredentialsDTO{
		Type:      "password",
		Value:     value,
		Temporary: false,
	}
}
