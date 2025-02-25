package main

import (
	"encoding/json"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
)

var rabbitConn *amqp091.Connection
var rabbitChannel *amqp091.Channel

func connectRabbitMq() error {
	var err error
	rabbitConn, err = amqp091.Dial("amqp://user:password@rabbitmq:5672")
	if err != nil {
		return fmt.Errorf("connect rabbitmq error: %v", err)
	}

	rabbitChannel, err = rabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("connect rabbitmq channel error: %v", err)
	}

	fmt.Println("RABBITMQ: SUCCEEDED")

	return nil
}

func sendRabbitMq(task WorkerTask) error {
	queueName := "worker_tasks"

	_, err := rabbitChannel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(task)
	err = rabbitChannel.Publish("", queueName, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp091.Persistent,
	})
	if err != nil {
		return err
	}

	err = setTaskStatusSent(task)
	if err != nil {
		return fmt.Errorf("send task to rabbitmq: %v", err)
	}

	return nil
}
