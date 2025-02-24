package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

func connectMongo() error {
	clientOptions := options.Client().
		ApplyURI("mongodb://mongo-primary:21017,mongo-secondary-1:21017,mongo-secondary-2:21017/?replicaSet=rs0")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("database connection verification failed: %v", err)
	}

	db = client.Database("crackhash")

	fmt.Println("MONGODB: SUCCEEDED")

	return nil
}

func saveTask(task WorkerTask) error {
	collection := db.Collection("workerTasks")
	_, err := collection.InsertOne(context.TODO(), task)
	return err
}
