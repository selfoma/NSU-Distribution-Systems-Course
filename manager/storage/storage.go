package storage

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
	maxRetries       = 5
)

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

type TaskResult struct {
	TaskId   string   `bson:"_id"      json:"taskId"`
	Words    []string `bson:"words"    json:"words"`
	Parts    int      `bson:"parts"    json:"parts"`
	Received int      `bson:"received" json:"received"`
	Status   string   `bson:"status"   json:"status"`
}

type TaskStorage struct {
	mu    sync.RWMutex
	tasks *mongo.Collection
}

func NewTaskStorage() (*TaskStorage, error) {
	tasks, err := database.ConnectMongo()
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %v", err)
	}
	return &TaskStorage{
		tasks: tasks,
	}, nil
}

func (s *TaskStorage) CreateTask(requestId string, parts int) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	task := &TaskResult{
		TaskId:   requestId,
		Words:    []string{},
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

func (s *TaskStorage) UpdateTask(requestId string, workerFoundWords []string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	if workerFoundWords == nil {
		workerFoundWords = []string{}
	}

	id := bson.M{"_id": requestId}
	upd := bson.M{
		"$push": bson.M{"words": bson.M{"$each": workerFoundWords}},
		"$inc":  bson.M{"received": 1},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	task := &TaskResult{}
	err := s.tasks.FindOneAndUpdate(ctx, id, upd, opts).Decode(task)
	if err != nil {
		return fmt.Errorf("find mongo: %v", err)
	}

	if task.Received == task.Parts {
		_, err = s.tasks.UpdateOne(ctx, id, bson.M{"$set": bson.M{"status": StatusReady}})
		if err != nil {
			return fmt.Errorf("update status: %v", err)
		}
		log.Printf("TaskResult [%s] completed, found words: %v", requestId, task.Words)
	}

	return nil
}

func (s *TaskStorage) GetTaskStatusById(requestId string) (string, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	opts := options.FindOne().SetProjection(bson.M{"_id": 0, "words": 0, "parts": 0, "received": 0})

	status := struct {
		Status string `bson:"status"`
	}{}
	err := s.tasks.FindOne(ctx, bson.M{"_id": requestId}, opts).Decode(&status)
	if err != nil {
		return "", fmt.Errorf("get status: [ID] %v | [E] %v", requestId, err)
	}

	return status.Status, nil
}

func (s *TaskStorage) UpdateTaskStatus(requestId, status string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	_, err := s.tasks.UpdateOne(ctx, bson.M{"_id": requestId}, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		return fmt.Errorf("update status: [ID] %v | [E] %v", requestId, err)
	}

	return nil
}

func (s *TaskStorage) GetTask(requestId string) (*TaskResult, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	task := &TaskResult{}
	err := s.tasks.FindOne(ctx, bson.M{"_id": requestId}).Decode(task)
	if err != nil {
		return nil, fmt.Errorf("get tasks: [ID] %v | [E] %v", requestId, err)
	}

	return task, nil
}

func (s *TaskStorage) SaveWorkerTask(task *WorkerTask) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = s.tasks.InsertOne(ctx, task)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("task save failed: max retries exceeded: %v", err)
}

func (s *TaskStorage) SetTaskStatusSent(task *WorkerTask) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = s.tasks.UpdateOne(ctx,
			bson.M{"_id": task.ID},
			bson.M{"$set": bson.M{"status": "sent"}},
		)
		if err == nil {
			log.Println("SET SENT")
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("update task status failed: max retries exceeded: %v", err)
}

func (s *TaskStorage) FindPendingTasks() ([]*WorkerTask, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	cursor, err := s.tasks.Find(ctx, bson.M{"status": "pending"})
	if err != nil {
		log.Printf("find pending tasks: %v", err)
		return nil, fmt.Errorf("find pending tasks: %v", err)
	}
	defer cursor.Close(ctx)

	tasks := make([]*WorkerTask, 0, cursor.RemainingBatchLength())
	for cursor.Next(ctx) {
		task := &WorkerTask{}
		if err = cursor.Decode(task); err == nil {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}
