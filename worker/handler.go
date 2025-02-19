package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WorkerTask struct {
	RequestId  string `json:"requestId"`
	Hash       string `json:"hash"`
	MaxLength  int    `json:"maxLength"`
	PartNumber int    `json:"partNumber"`
	PartCount  int    `json:"partCount"`
}

func handleWorkerTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req WorkerTask
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Received task: ", req)

	foundWords := bruteForce(req.Hash, req.MaxLength, req.PartNumber, req.PartCount)

	managerUrl := "http://localhost:8080/internal/api/manager/hash/crack/request"
	resp := WorkerResponse{
		RequestId:  req.RequestId,
		Words:      foundWords,
		PartNumber: req.PartNumber,
	}
	err := sendResultToManager(managerUrl, resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
