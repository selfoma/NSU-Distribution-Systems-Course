package broker

import (
	"encoding/xml"
	"github.com/rabbitmq/amqp091-go"
	"github.com/selfoma/crackhash/worker/config"
	"github.com/selfoma/crackhash/worker/service"
	"log"
)

func PublishResponse(resp service.WorkerResponse) {
	queueName := config.Cfg.ResponseQueueName

	_, err := rabbitChannel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	body, _ := xml.Marshal(resp)
	err = rabbitChannel.Publish("", queueName, false, false, amqp091.Publishing{
		ContentType:  "text/xml",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})
	if err != nil {
		log.Fatal(err)
	}
}
