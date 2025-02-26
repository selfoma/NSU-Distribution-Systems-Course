package broker

import (
	"encoding/json"
	"github.com/selfoma/crackhash/worker/service"
	"log"
)

type WorkerTask struct {
	RequestId   string `json:"requestId"`
	Hash        string `json:"hash"`
	MaxLength   int    `json:"maxLength"`
	WorkerCount int    `json:"workerCount"`
	PartNumber  int    `json:"partNumber"`
	PartCount   int    `json:"partCount"`
}

func ConsumeTask() {
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	msgs, err := rabbitChannel.Consume(
		"worker_tasks",
		"",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	var task WorkerTask
	for msg := range msgs {
		err = json.Unmarshal(msg.Body, &task)
		if err != nil {
			log.Fatalf("Failed to unmarshal message: %v", err)
		}

		service.BruteForce(task)

		msg.Ack(false)
	}
}
