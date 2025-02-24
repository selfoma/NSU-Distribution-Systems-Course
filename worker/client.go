package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
)

var httpClient = &http.Client{}

type WorkerResponse struct {
	RequestId  string   `xml:"requestId"`
	Words      []string `xml:"words"`
	PartNumber int      `xml:"partNumber"`
}

func sendResultToManager(managerUrl string, response WorkerResponse) error {
	data, err := xml.Marshal(response)
	if err != nil {
		return fmt.Errorf("xml marshal error: %v", err)
	}

	rq, err := http.NewRequest(http.MethodPatch, managerUrl, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := httpClient.Do(rq)
	if err != nil {
		return fmt.Errorf("send request to manager failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("manager returned status code: %v", resp.Status)
	}

	return nil
}
