package main

import (
	"fmt"
	"github.com/google/uuid"
)

var alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

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
	config      *Config
}

func NewCrackService() *CrackService {
	config, err := LoadConfig("config.json")
	if err != nil {
		panic(err)
	}
	return &CrackService{
		taskStorage: NewTaskStorage(),
		config:      config,
	}
}

func (cs *CrackService) StartCrackHash(hash string, maxLength int) (string, error) {
	requestId := uuid.New().String()

	cs.taskStorage.CreateTask(requestId, 1)

	words := countWordsInAlphabet(alphabet, maxLength)
	for i := 0; i < cs.config.WorkerCount; i++ {
		workerURL := cs.config.WorkerUrls[i]
		task := WorkerTask{
			RequestId:   requestId,
			Hash:        hash,
			MaxLength:   maxLength,
			WorkerCount: cs.config.WorkerCount,
			PartNumber:  i,
			PartCount:   countPartSize(words, cs.config.WorkerCount, i),
		}

		err := sendWorkerTask(workerURL, task)
		if err != nil {
			return "", fmt.Errorf("send task to worker failed: %v", err)
		}
	}

	return requestId, nil
}

func (cs *CrackService) ProcessWorkerResponse(requestId string, words []string) error {
	return cs.taskStorage.UpdateTask(requestId, words)
}

func (cs *CrackService) GetTask(requestId string) (*TaskStatus, error) {
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
