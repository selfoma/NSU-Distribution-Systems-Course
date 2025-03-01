package broker

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/selfoma/crackhash/manager/config"
	"github.com/selfoma/crackhash/manager/database"
	"log"
)

var rabbitConn *amqp091.Connection
var rabbitChannel *amqp091.Channel

func ConnectRabbitMq() error {
	var err error
	rabbitConn, err = amqp091.Dial("amqp://user:password@rabbitmq:5672")
	if err != nil {
		return fmt.Errorf("connect rabbitmq error: %v", err)
	}

	rabbitChannel, err = rabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("connect rabbitmq channel error: %v", err)
	}

	err = rabbitChannel.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("connect rabbitmq qos error: %v", err)
	}

	_, err = rabbitChannel.QueueDeclare(config.Cfg.TaskQueueName, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = rabbitChannel.QueueDeclare(config.Cfg.ResponseQueueName, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("RABBITMQ: SUCCEEDED")

	return nil
}

type RabbitMqBroker struct{}

func (rq *RabbitMqBroker) Consume() {
	consumeResponse()
}

func (rq *RabbitMqBroker) Publish(t *database.WorkerTask) {
	publishTask(t)
}
