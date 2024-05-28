package persistence

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetClient(username, password, host, port string) (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/", username, password, host, port)

	options := options.Client().ApplyURI(uri)
	return mongo.Connect(context.TODO(), options)
}
