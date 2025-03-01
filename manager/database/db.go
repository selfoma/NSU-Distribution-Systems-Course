package database

import (
	"context"
	"fmt"
	"github.com/selfoma/crackhash/manager/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"log"
	"time"
)

const (
	maxRetries = 5
)

var db *mongo.Database

type WorkerTask struct {
	ID          string `bson:"_id,omitempty" json:"id"`
	RequestId   string `bson:"requestId"     json:"requestId"`
	Hash        string `bson:"hash"          json:"hash"`
	MaxLength   int    `bson:"maxLength"     json:"maxLength"`
	WorkerCount int    `bson:"workerCount"   json:"workerCount"`
	PartNumber  int    `bson:"partNumber"    json:"partNumber"`
	PartCount   int    `bson:"partCount"     json:"partCount"`
	Status      string `bson:"status"        json:"status"`
}

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

func SaveWorkerTask(collection *mongo.Collection, task *WorkerTask) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	log.Printf("Saved worker task: %v", task)

	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = collection.InsertOne(ctx, task)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("task save failed: max retries exceeded: %v", err)
}

func SetTaskStatusSent(collection *mongo.Collection, task *WorkerTask) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = collection.UpdateOne(ctx,
			bson.M{"requestId": task.RequestId},
			bson.M{"$set": bson.M{"status": "sent"}},
		)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("update task status failed: max retries exceeded: %v", err)
}
