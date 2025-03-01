package broker

import (
	"encoding/json"
	"github.com/selfoma/crackhash/worker/config"
	"github.com/selfoma/crackhash/worker/service"
	"log"
)

func consumeTask() {
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	msgs, err := rabbitChannel.Consume(
		config.Cfg.TaskQueueName,
		"",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	for msg := range msgs {
		task := &service.WorkerTask{}
		err = json.Unmarshal(msg.Body, task)
		if err != nil {
			log.Fatalf("Failed to unmarshal message: [M] %v | [E] %v", msg.Body, err)
		}

		service.WorkerService.BruteForce(task)

		msg.Ack(false)
	}
}
