package database

import (
	"context"
	"fmt"
	"github.com/selfoma/crackhash/manager/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var db *mongo.Database

func ConnectMongo() (*mongo.Collection, error) {
	clientOptions := options.Client().
		ApplyURI(config.Cfg.MongoUrl).
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Majority())
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, fmt.Errorf("database connection verification failed: %v", err)
	}

	db = client.Database("crackhash")

	fmt.Println("MONGODB: SUCCEEDED")

	return db.Collection("tasks"), nil
}
