package main

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
)

var crackService = NewCrackService()

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
	requestId, err := crackService.StartCrackHash(req.Hash, req.MaxLength)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.RequestId = requestId

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	}
}

type WorkerResponse struct {
	RequestId  string   `xml:"requestId"`
	Words      []string `xml:"words"`
	PartNumber int      `xml:"partNumber"`
}

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

	e := crackService.ProcessWorkerResponse(resp.RequestId, resp.Words)
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
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

	task, err := crackService.GetTask(requestId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
