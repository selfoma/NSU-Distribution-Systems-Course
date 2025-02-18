package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
)

type ClientCrackRequest struct {
	Hash      string `json:"hash"`
	MaxLength int    `json:"maxLength"`
}

type ClientCrackResponse struct {
	RequestId string `json:"requestId"`
}

func handleCrackRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req ClientCrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestId := uuid.New().String()
	taskResults.tasks[requestId] = &TaskStatus{
		Words:    nil,
		Parts:    1,
		Received: 0,
		Status:   StatusInProgress,
	}

	workerURL := "http://localhost:8081/internal/api/worker/hash/crack/task"
	task := WorkerTask{
		RequestId:  requestId,
		Hash:       req.Hash,
		MaxLength:  req.MaxLength,
		PartNumber: 0,
		PartCount:  1,
	}
	err := sendWorkerTask(workerURL, task)
	if err != nil {
		http.Error(w, fmt.Sprintf("Send task to worker failed: %v", err), http.StatusInternalServerError)
		return
	}

	resp := ClientCrackResponse{RequestId: requestId}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Encode task to worker failed: %v", err), http.StatusInternalServerError)
		return
	}
}

type WorkerResponse struct {
	RequestId  string   `xml:"requestId"`
	Words      []string `xml:"words"`
	PartNumber int      `xml:"partNumber"`
}

const (
	StatusInProgress = "IN_PROGRESS"
	StatusReady      = "READY"
	StatusError      = "ERROR"
)

type TaskStatus struct {
	Words    []string `json:"words"`
	Parts    int      `json:"parts"`
	Received int      `json:"received"`
	Status   string   `json:"status"`
}

var taskResults = struct {
	sync.Mutex
	tasks map[string]*TaskStatus
}{tasks: make(map[string]*TaskStatus)}

func handleWorkerResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var resp WorkerResponse
	if err := xml.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskResults.Mutex.Lock()
	defer taskResults.Mutex.Unlock()

	task, e := taskResults.tasks[resp.RequestId]
	if !e {
		http.Error(w, fmt.Sprintf("task [%s] not found", resp.RequestId), http.StatusNotFound)
		return
	}

	task.Words = append(task.Words, resp.Words...)
	task.Received++

	if task.Received == task.Parts {
		task.Status = StatusReady
		log.Printf("Task [%s] completed, found words: %v", resp.RequestId, resp.Words)
	}

	w.WriteHeader(http.StatusOK)
}

type TaskStatusResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"tasks"`
}

func handleStatusRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	requestId := r.URL.Query().Get("requestId")
	if requestId == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	task, e := taskResults.tasks[requestId]
	if !e {
		http.Error(w, fmt.Sprintf("task [%s] not found", requestId), http.StatusNotFound)
		return
	}

	var data []string
	if task.Status == StatusReady {
		data = task.Words
	}
	resp := TaskStatusResponse{
		Status: task.Status,
		Data:   data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
