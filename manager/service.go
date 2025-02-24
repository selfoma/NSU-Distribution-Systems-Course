package main

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	timeOut  = 3 * time.Minute
)

type WorkerTask struct {
	RequestId   string `json:"requestId"`
	Hash        string `json:"hash"`
	MaxLength   int    `json:"maxLength"`
	WorkerCount int    `json:"workerCount"`
	PartNumber  int    `json:"partNumber"`
	PartCount   int    `json:"partCount"`
}

type CrackService struct {
	taskStorage *TaskStorage
}

func NewCrackService() *CrackService {
	return &CrackService{
		taskStorage: NewTaskStorage(),
	}
}

func (cs *CrackService) StartCrackHash(hash string, maxLength int) (string, error) {
	requestId := uuid.New().String()

	cs.taskStorage.CreateTask(requestId, 1)

	words := countWordsInAlphabet(alphabet, maxLength)
	for i := 0; i < config.WorkerCount; i++ {
		workerURL := config.WorkerUrls[i]
		task := WorkerTask{
			RequestId:   requestId,
			Hash:        hash,
			MaxLength:   maxLength,
			WorkerCount: config.WorkerCount,
			PartNumber:  i,
			PartCount:   countPartSize(words, config.WorkerCount, i),
		}

		err := sendWorkerTask(workerURL, task)
		if err != nil {
			return "", fmt.Errorf("send task to worker failed: %v", err)
		}
	}

	go cs.monitorTaskTimeOut(requestId)

	return requestId, nil
}

func (cs *CrackService) monitorTaskTimeOut(requestId string) {
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

func (cs *CrackService) ProcessWorkerResponse(requestId string, words []string) error {
	return cs.taskStorage.UpdateTask(requestId, words)
}

func (cs *CrackService) GetTask(requestId string) (*Task, error) {
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
