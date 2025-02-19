package main

import (
	"fmt"
	"github.com/google/uuid"
)

type CrackService struct {
	taskStorage *TaskStorage
}

func NewCrackService() *CrackService {
	return &CrackService{
		taskStorage: NewTaskStorage(),
	}
}

func (cs *CrackService) StartCrackHash(hash string, maxLength, workerCount int) (string, error) {
	requestId := uuid.New().String()

	cs.taskStorage.CreateTask(requestId, 1)

	workerURL := "http://localhost:8081/internal/api/worker/hash/crack/task"
	task := WorkerTask{
		RequestId:  requestId,
		Hash:       hash,
		MaxLength:  maxLength,
		PartNumber: 0,
		PartCount:  1,
	}

	err := sendWorkerTask(workerURL, task)
	if err != nil {
		return "", fmt.Errorf("send task to worker failed: %v", err)
	}

	return requestId, nil
}

func (cs *CrackService) ProcessWorkerResponse(requestId string, words []string) error {
	return cs.taskStorage.UpdateTask(requestId, words)
}

func (cs *CrackService) GetTask(requestId string) (*TaskStatus, error) {
	return cs.taskStorage.GetTask(requestId)
}
