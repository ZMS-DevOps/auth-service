package request

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VerificationRequest struct {
	VerificationId primitive.ObjectID `json:"verificationId" validate:"required"`
	UserId         string             `json:"userId" validate:"required"`
	SecurityCode   int                `json:"securityCode" validate:"required,numeric,min=1000,max=9999"`
}

func (request VerificationRequest) AreValidRequestData() error {
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}
