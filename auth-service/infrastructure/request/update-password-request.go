package request

import (
	"errors"
	"github.com/go-playground/validator/v10"
)

type UpdatePasswordRequest struct {
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
}

func (request UpdatePasswordRequest) AreValidRequestData() error {
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		return err.(validator.ValidationErrors)
	}

	if err := request.arePasswordsMatch(); err != nil {
		return errors.New("password did not match")
	}

	return nil
}

func (request UpdatePasswordRequest) arePasswordsMatch() error {
	return checkPasswordsMatch(request.Password, request.ConfirmPassword)
}
