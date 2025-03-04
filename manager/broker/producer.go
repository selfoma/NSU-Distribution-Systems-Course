package broker

import (
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"github.com/selfoma/crackhash/manager/config"
	"github.com/selfoma/crackhash/manager/service"
	"github.com/selfoma/crackhash/manager/storage"
	"log"
)

func publishTask(task *storage.WorkerTask) {
	body, _ := json.Marshal(task)
	err := rabbitChannel.Publish("", config.Cfg.TaskQueueName, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})
	if err != nil {
		return
	}

	log.Printf("publish task success: %v", task)

	err = service.CrackService.SetTaskStatusSent(task)
	if err != nil {
		log.Fatal(err)
	}
}
