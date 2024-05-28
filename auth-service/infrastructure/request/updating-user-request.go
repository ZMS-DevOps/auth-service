package request

import (
	"github.com/go-playground/validator/v10"
)

type UpdatingUserRequest struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Address   string `json:"address" validate:"required"`
}

func (request UpdatingUserRequest) AreValidRequestData() error {
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}
