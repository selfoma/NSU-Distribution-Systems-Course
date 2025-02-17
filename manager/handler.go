package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

type CrackRequest struct {
	Hash      string `json:"hash"`
	MaxLength int    `json:"maxLength"`
}

type CrackResponse struct {
	RequestId string `json:"requestId"`
}

func handleCrackRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req CrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestId := uuid.New().String()

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

	resp := CrackResponse{RequestId: requestId}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Encode task to worker failed: %v", err), http.StatusInternalServerError)
		return
	}
}

func handleStatusRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "IN_PROGRESS", "data": null}`))
}
