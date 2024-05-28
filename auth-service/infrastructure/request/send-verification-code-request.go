package request

import "github.com/go-playground/validator/v10"

type SendVerificationCodeRequest struct {
	VerificationId string `json:"verificationId" validate:"required"`
	UserEmail      string `json:"userEmail" validate:"required,email"`
}

func (request SendVerificationCodeRequest) AreValidRequestData() error {
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}
