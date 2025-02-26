package service

import (
	"github.com/google/uuid"
	"github.com/selfoma/crackhash/manager/broker"
	"github.com/selfoma/crackhash/manager/config"
	"github.com/selfoma/crackhash/manager/database"
	"log"
	"time"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	timeOut  = 3 * time.Minute
)

var CrackService = newCrackService()

type crackService struct {
	taskStorage *TaskStorage
}

func newCrackService() *crackService {
	return &crackService{
		taskStorage: NewTaskStorage(),
	}
}

func (cs *crackService) StartCrackHash(hash string, maxLength int) (string, error) {
	requestId := uuid.New().String()

	cs.taskStorage.CreateTask(requestId, 1)

	words := countWordsInAlphabet(alphabet, maxLength)
	for i := 0; i < config.Cfg.WorkerCount; i++ {
		task := database.WorkerTask{
			RequestId:   requestId,
			Hash:        hash,
			MaxLength:   maxLength,
			WorkerCount: config.Cfg.WorkerCount,
			PartNumber:  i,
			PartCount:   countPartSize(words, config.Cfg.WorkerCount, i),
			Status:      "pending",
		}

		err := database.SaveWorkerTask(task)
		if err != nil {
			log.Fatal(err)
		}

		broker.PublishTask(task)
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
		log.Printf("get task status failed: %v", err)
	}

	if s != StatusReady {
		err = cs.taskStorage.UpdateTaskStatus(requestId, StatusError)
		if err != nil {
			log.Printf("update task status failed: %v", err)
		}
	}
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
