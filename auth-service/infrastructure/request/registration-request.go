package request

import (
	"errors"
	"github.com/go-playground/validator/v10"
)

type RegistrationRequest struct {
	Email           string `json:"email" validate:"required,email"`
	FirstName       string `json:"firstName" validate:"required"`
	LastName        string `json:"lastName" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
	Address         string `json:"address" validate:"required"`
	Group           string `json:"group" validate:"required"`
}

func (request RegistrationRequest) AreValidRequestData() error {
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		return err.(validator.ValidationErrors)
	}

	if err := request.arePasswordsMatch(); err != nil {
		return errors.New("password did not match")
	}

	return nil
}

func (request RegistrationRequest) arePasswordsMatch() error {
	return checkPasswordsMatch(request.Password, request.ConfirmPassword)
}
