package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/selfoma/crackhash/manager/config"
	"github.com/selfoma/crackhash/manager/storage"
	"log"
	"time"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	timeOut  = 1 * time.Minute
)

type Broker interface {
	Consume()
	Publish(t *storage.WorkerTask)
}

var CrackService *crackService

type crackService struct {
	b           Broker
	taskStorage *storage.TaskStorage
}

func InitService(b Broker) error {
	s, err := storage.NewTaskStorage()
	if err != nil {
		return fmt.Errorf("create tasks storage: %v", err)
	}

	CrackService = &crackService{b: b, taskStorage: s}

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
		task := &storage.WorkerTask{
			ID:          uuid.New().String(),
			RequestId:   requestId,
			Hash:        hash,
			MaxLength:   maxLength,
			WorkerCount: config.Cfg.WorkerCount,
			PartNumber:  i,
			PartCount:   countPartSize(words, config.Cfg.WorkerCount, i),
			Status:      "pending",
		}

		err = cs.taskStorage.SaveWorkerTask(task)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("tasks saved")

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

	if s != storage.StatusReady {
		err = cs.taskStorage.UpdateTaskStatus(requestId, storage.StatusError)
		if err != nil {
			log.Printf("update tasks status failed: %v", err)
		}
	}
}

func (cs *crackService) SetTaskStatusSent(task *storage.WorkerTask) error {
	return cs.taskStorage.SetTaskStatusSent(task)
}

func (cs *crackService) ProcessWorkerResponse(requestId string, words []string) error {
	return cs.taskStorage.UpdateTask(requestId, words)
}

func (cs *crackService) GetTask(requestId string) (*storage.TaskResult, error) {
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

func (cs *crackService) SendPendingTasks() {
	for {
		time.Sleep(10 * time.Second)

		tasks, err := cs.taskStorage.FindPendingTasks()
		if err != nil {
			log.Printf("find pending tasks: %v", err)
			continue
		}

		for _, task := range tasks {
			cs.b.Publish(task)
		}
	}
}
