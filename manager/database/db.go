package database

import (
	"context"
	"fmt"
	"github.com/selfoma/crackhash/manager/broker"
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
	RequestId   string `bson:"_id,omitempty" json:"requestId"`
	Hash        string `bson:"hash"          json:"hash"`
	MaxLength   int    `bson:"maxLength"     json:"maxLength"`
	WorkerCount int    `bson:"workerCount"   json:"workerCount"`
	PartNumber  int    `bson:"partNumber"    json:"partNumber"`
	PartCount   int    `bson:"partCount"     json:"partCount"`
	Status      string `bson:"status"        json:"status"`
}

func ConnectMongo() error {
	clientOptions := options.Client().
		ApplyURI("mongodb://mongo-primary:21017,mongo-secondary-1:21017,mongo-secondary-2:21017/?replicaSet=rs0").
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Majority())
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

func SaveWorkerTask(task WorkerTask) error {
	collection := db.Collection("workertasks")

	for i := 0; i < maxRetries; i++ {
		_, err := collection.InsertOne(context.TODO(), task)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("task save failed: max retries exceeded")
}

func SetTaskStatusSent(task WorkerTask) error {
	collection := db.Collection("workertasks")

	for i := 0; i < maxRetries; i++ {
		_, err := collection.UpdateOne(context.TODO(),
			bson.M{"_id": task.RequestId},
			bson.M{"$set": bson.M{"status": "sent"}},
		)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("update task status failed: max retries exceeded")
}

func RetryPendingTask() {
	for {
		time.Sleep(10 * time.Second)

		cursor, err := db.Collection("workertasks").Find(context.TODO(), bson.M{"status": "pending"})
		if err != nil {
			log.Printf("Error finding pending tasks: %v", err)
			continue
		}

		for cursor.Next(context.TODO()) {
			var task WorkerTask
			if err = cursor.Decode(&task); err == nil {
				broker.PublishTask(task)
			}
		}
	}
}
