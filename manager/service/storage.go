package service

import (
	"context"
	"fmt"
	"github.com/selfoma/crackhash/manager/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
	"time"
)

const (
	StatusInProgress = "IN_PROGRESS"
	StatusReady      = "READY"
	StatusError      = "ERROR"
)

type TaskResult struct {
	TaskId   string   `bson:"_id"      json:"taskId"`
	Words    []string `bson:"words"    json:"words"`
	Parts    int      `bson:"parts"    json:"parts"`
	Received int      `bson:"received" json:"received"`
	Status   string   `bson:"status"   json:"status"`
}

type taskStorage struct {
	mu    sync.RWMutex
	tasks *mongo.Collection
}

func newTaskStorage() (*taskStorage, error) {
	tasks, err := database.ConnectMongo()
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %v", err)
	}
	return &taskStorage{
		tasks: tasks,
	}, nil
}

func (s *taskStorage) CreateTask(requestId string, parts int) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	task := &TaskResult{
		TaskId:   requestId,
		Words:    nil,
		Parts:    parts,
		Received: 0,
		Status:   StatusInProgress,
	}

	_, err := s.tasks.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("create tasks: %v", err)
	}

	return nil
}

func (s *taskStorage) UpdateTask(requestId string, workerFoundWords []string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	id := bson.M{"_id": requestId}
	upd := bson.M{
		"$push": bson.M{"words": bson.M{"$each": workerFoundWords}},
		"$inc":  bson.M{"parts": 1},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var task TaskResult
	err := s.tasks.FindOneAndUpdate(ctx, id, upd, opts).Decode(&task)
	if err != nil {
		return fmt.Errorf("find mongo: %v", err)
	}

	if task.Received == task.Parts {
		_, err = s.tasks.UpdateOne(ctx, id, bson.M{"$set": bson.M{"status": StatusReady}})
		if err != nil {
			return fmt.Errorf("update status: %v", err)
		}
		log.Printf("TaskResult [%s] completed, found words: %v", requestId, workerFoundWords)
	}

	return nil
}

func (s *taskStorage) GetTaskStatusById(requestId string) (string, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	opts := options.FindOne().SetProjection(bson.M{"_id": 0, "words": 0, "parts": 0, "received": 0})

	var status string
	err := s.tasks.FindOne(ctx, bson.M{"_id": requestId}, opts).Decode(&status)
	if err != nil {
		return "", fmt.Errorf("get status: %v", err)
	}

	return status, nil
}

func (s *taskStorage) UpdateTaskStatus(requestId, status string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	_, err := s.tasks.UpdateOne(ctx, bson.M{"_id": requestId}, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		return fmt.Errorf("update status: %v", err)
	}

	return nil
}

func (s *taskStorage) GetTask(requestId string) (*TaskResult, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	var task *TaskResult
	err := s.tasks.FindOne(ctx, bson.M{"_id": requestId}).Decode(task)
	if err != nil {
		return nil, fmt.Errorf("get tasks: %v", err)
	}

	return task, nil
}
