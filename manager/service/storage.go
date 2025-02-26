package service

import (
	"fmt"
	"log"
	"sync"
)

const (
	StatusInProgress = "IN_PROGRESS"
	StatusReady      = "READY"
	StatusError      = "ERROR"
)

type TaskResult struct {
	Words    []string `json:"words"`
	Parts    int      `json:"parts"`
	Received int      `json:"received"`
	Status   string   `json:"status"`
}

type TaskStorage struct {
	mu    sync.RWMutex
	tasks map[string]*TaskResult
}

func NewTaskStorage() *TaskStorage {
	return &TaskStorage{
		tasks: make(map[string]*TaskResult),
	}
}

func (s *TaskStorage) CreateTask(requestId string, parts int) *TaskResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &TaskResult{
		Words:    nil,
		Parts:    parts,
		Received: 0,
		Status:   StatusInProgress,
	}

	s.tasks[requestId] = task
	return task
}

func (s *TaskStorage) UpdateTask(requestId string, workerFoundWords []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[requestId]
	if !ok {
		return fmt.Errorf("task %s not found", requestId)
	}

	task.Words = append(task.Words, workerFoundWords...)
	task.Received++

	if task.Received == task.Parts {
		task.Status = StatusReady
		log.Printf("TaskResult [%s] completed, found words: %v", requestId, workerFoundWords)
	}

	return nil
}

func (s *TaskStorage) GetTaskStatusById(requestId string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, ok := s.tasks[requestId]; ok {
		return task.Status, nil
	}
	return "", fmt.Errorf("task %s not found", requestId)
}

func (s *TaskStorage) UpdateTaskStatus(requestId, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[requestId]
	if !ok {
		return fmt.Errorf("task %s not found", requestId)
	}

	task.Status = status

	return nil
}

func (s *TaskStorage) GetTask(requestId string) (*TaskResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[requestId]
	if !ok {
		return nil, fmt.Errorf("task %s not found", requestId)
	}
	return task, nil
}
