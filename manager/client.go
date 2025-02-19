package main

import (
	"bytes"
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

func sendWorkerTask(workerURL string, task WorkerTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task to json failed: %w", err)
	}

	resp, err := http.Post(workerURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("send request to worker failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("worker returned status code: %s", resp.Status)
	}

	return nil
}
