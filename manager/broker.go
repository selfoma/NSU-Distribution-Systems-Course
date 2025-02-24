package main

import (
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

func sendRabbitMq(task WorkerTask) {

}
