package broker

import (
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"github.com/selfoma/crackhash/manager/config"
	"github.com/selfoma/crackhash/manager/database"
	"log"
)

func PublishTask(task database.WorkerTask) {
	queueName := config.Cfg.TaskQueueName

	_, err := rabbitChannel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	body, _ := json.Marshal(task)
	err = rabbitChannel.Publish("", queueName, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})
	if err != nil {
		return
	}

	database.SetTaskStatusSent(task)
}
