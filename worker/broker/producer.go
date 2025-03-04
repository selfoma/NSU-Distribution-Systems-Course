package broker

import (
	"encoding/xml"
	"github.com/rabbitmq/amqp091-go"
	"github.com/selfoma/crackhash/worker/config"
	"github.com/selfoma/crackhash/worker/service"
	"log"
	"time"
)

const (
	maxRetries = 5
)

func publishResponse(resp *service.WorkerResponse) {
	body, _ := xml.Marshal(resp)
	for i := 0; i < maxRetries; i++ {
		err := rabbitChannel.Publish("", config.Cfg.ResponseQueueName, false, false, amqp091.Publishing{
			ContentType:  "text/xml",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		})
		if err == nil {
			return
		}
		log.Printf("publish failed: %v", err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	log.Fatal("failed to publish response")
}
