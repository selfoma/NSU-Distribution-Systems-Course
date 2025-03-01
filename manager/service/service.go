package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/selfoma/crackhash/manager/config"
	"github.com/selfoma/crackhash/manager/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	timeOut  = 3 * time.Minute
)

type Broker interface {
	Consume()
	Publish(t *database.WorkerTask)
}

var CrackService *crackService

type crackService struct {
	b           Broker
	taskStorage *taskStorage
}

func InitService(b Broker) error {
	storage, err := newTaskStorage()
	if err != nil {
		return fmt.Errorf("create tasks storage: %v", err)
	}

	CrackService = &crackService{b: b, taskStorage: storage}

	return nil
}

func (cs *crackService) StartCrackHash(hash string, maxLength int) (string, error) {
	requestId := uuid.New().String()

	err := cs.taskStorage.CreateTask(requestId, config.Cfg.WorkerCount)
	if err != nil {
		return "", fmt.Errorf("create tasks: %v", err)
	}

	words := countWordsInAlphabet(alphabet, maxLength)
	for i := 0; i < config.Cfg.WorkerCount; i++ {
		task := &database.WorkerTask{
			ID:          uuid.New().String(),
			RequestId:   requestId,
			Hash:        hash,
			MaxLength:   maxLength,
			WorkerCount: config.Cfg.WorkerCount,
			PartNumber:  i,
			PartCount:   countPartSize(words, config.Cfg.WorkerCount, i),
			Status:      "pending",
		}

		err = database.SaveWorkerTask(cs.taskStorage.tasks, task)
		if err != nil {
			log.Fatal(err)
		}

		cs.b.Publish(task)
	}

	go cs.monitorTaskTimeOut(requestId)

	return requestId, nil
}

func (cs *crackService) monitorTaskTimeOut(requestId string) {
	timer := time.NewTimer(timeOut)
	defer timer.Stop()

	<-timer.C

	s, err := cs.taskStorage.GetTaskStatusById(requestId)
	if err != nil {
		log.Printf("get tasks status failed: %v", err)
		return
	}

	if s != StatusReady {
		err = cs.taskStorage.UpdateTaskStatus(requestId, StatusError)
		if err != nil {
			log.Printf("update tasks status failed: %v", err)
		}
	}
}

func (cs *crackService) SetTaskStatusSent(task *database.WorkerTask) error {
	return database.SetTaskStatusSent(cs.taskStorage.tasks, task)
}

func (cs *crackService) ProcessWorkerResponse(requestId string, words []string) error {
	return cs.taskStorage.UpdateTask(requestId, words)
}

func (cs *crackService) GetTask(requestId string) (*TaskResult, error) {
	return cs.taskStorage.GetTask(requestId)
}

func countWordsInAlphabet(alphabet string, length int) int {
	n := len(alphabet)
	wordsCount := 0
	for i := 1; i <= length; i++ {
		wordsCount += pow(n, i)
	}
	return wordsCount
}

func pow(x, n int) int {
	if n < 0 {
		return 1 / pow(x, -n)
	}
	if n == 0 {
		return 1
	}
	a := pow(x, n/2)
	if n&1 == 0 {
		return a * a
	}
	return a * a * x
}

func countPartSize(part, n, r int) int {
	base := part / n
	rem := part % n
	if r < rem {
		return base + 1
	}
	return base
}

func (cs *crackService) RetryPendingTask() {
	for {
		time.Sleep(10 * time.Second)

		cursor, err := cs.taskStorage.tasks.Find(context.TODO(), bson.M{"status": "pending"})
		if err != nil {
			log.Printf("Error finding pending tasks: %v", err)
			continue
		}

		for cursor.Next(context.TODO()) {
			var task *database.WorkerTask
			if err = cursor.Decode(task); err == nil {
				cs.b.Publish(task)
			}
		}
	}
}
