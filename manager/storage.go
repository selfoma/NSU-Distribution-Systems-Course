package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	StatusInProgress = "IN_PROGRESS"
	StatusReady      = "READY"
	StatusError      = "ERROR"
	TimeOut          = 30 * time.Second
)

type TaskStatus struct {
	Words    []string `json:"words"`
	Parts    int      `json:"parts"`
	Received int      `json:"received"`
	Status   string   `json:"status"`
}

type TaskStorage struct {
	mu    sync.Mutex
	tasks map[string]*TaskStatus
}

func NewTaskStorage() *TaskStorage {
	return &TaskStorage{
		tasks: make(map[string]*TaskStatus),
	}
}

func (s *TaskStorage) CreateTask(requestId string, parts int) *TaskStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &TaskStatus{
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
		log.Printf("Task [%s] completed, found words: %v", requestId, task.Words)
		time.AfterFunc(TimeOut, func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			task.Status = StatusError
		})
	}

	return nil
}

func (s *TaskStorage) GetTask(requestId string) (*TaskStatus, error) {
	task, ok := s.tasks[requestId]
	if !ok {
		return nil, fmt.Errorf("task %s not found", requestId)
	}
	return task, nil
}
