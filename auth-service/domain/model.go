package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Verification struct {
	Id        primitive.ObjectID `bson:"_id"`
	UserId    string             `bson:"user_id"`
	Code      int                `bson:"code"`
	FirstName string             `bson:"first_name"`
	LastName  string             `bson:"last_name"`
	Address   string             `bson:"address"`
}
