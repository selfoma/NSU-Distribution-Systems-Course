package main

import (
	"encoding/json"
	"github.com/selfoma/crackhash/manager/service"
	"github.com/selfoma/crackhash/manager/storage"
	"log"
	"net/http"
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

	resp := ClientCrackResponse{}
	requestId, err := service.CrackService.StartCrackHash(req.Hash, req.MaxLength)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.RequestId = requestId

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v", err)
		return
	}
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

	task, err := service.CrackService.GetTask(requestId)
	if err != nil {
		log.Printf("Failed to get task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data []string
	if task.Status == storage.StatusReady {
		data = task.Words
	}
	resp := TaskStatusResponse{
		Status: task.Status,
		Data:   data,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Failed to encode response: [R] %v | [E] %v", resp, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
