package persistence

import (
	"context"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DATABASE   = "authdb"
	COLLECTION = "verification"
)

type VerificationMongoDBStore struct {
	verifications *mongo.Collection
}

func NewVerificationMongoDBStore(client *mongo.Client) domain.VerificationStore {
	verifications := client.Database(DATABASE).Collection(COLLECTION)
	return &VerificationMongoDBStore{
		verifications: verifications,
	}
}

func (store *VerificationMongoDBStore) Get(id primitive.ObjectID) (*domain.Verification, error) {
	filter := bson.M{"_id": id}
	return store.filterOne(filter)
}

func (store *VerificationMongoDBStore) Insert(verification *domain.Verification) (primitive.ObjectID, error) {
	verification.Id = primitive.NewObjectID()
	result, err := store.verifications.InsertOne(context.TODO(), verification)
	if err != nil {
		return primitive.NilObjectID, err
	}
	verification.Id = result.InsertedID.(primitive.ObjectID)
	return verification.Id, nil
}

func (store *VerificationMongoDBStore) Delete(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := store.verifications.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (store *VerificationMongoDBStore) Update(id primitive.ObjectID, verification *domain.Verification) error {
	filter := bson.M{"_id": id}
	update := bson.D{
		{"$set", bson.D{
			{"user_id", verification.UserId},
			{"code", verification.Code},
		}},
	}
	_, err := store.verifications.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (store *VerificationMongoDBStore) filter(filter interface{}) ([]*domain.Verification, error) {
	cursor, err := store.verifications.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())

	if err != nil {
		return nil, err
	}
	return decode(cursor)
}

func (store *VerificationMongoDBStore) filterOne(filter interface{}) (verification *domain.Verification, err error) {
	result := store.verifications.FindOne(context.TODO(), filter)
	err = result.Decode(&verification)
	return
}

func decode(cursor *mongo.Cursor) (verifications []*domain.Verification, err error) {
	for cursor.Next(context.TODO()) {
		var verification domain.Verification
		err = cursor.Decode(&verification)
		if err != nil {
			return
		}
		verifications = append(verifications, &verification)
	}
	err = cursor.Err()
	return
}
