package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VerificationStore interface {
	Get(id primitive.ObjectID) (*Verification, error)
	Insert(verification *Verification) (primitive.ObjectID, error)
	Delete(id primitive.ObjectID) error
	Update(id primitive.ObjectID, verification *Verification) error
}
