package broker

import (
	"encoding/xml"
	"github.com/selfoma/crackhash/manager/config"
	"github.com/selfoma/crackhash/manager/service"
	"log"
)

type WorkerResponse struct {
	RequestId  string   `xml:"requestId"`
	Words      []string `xml:"words"`
	PartNumber int      `xml:"partNumber"`
}

func consumeResponse() {
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	msgs, err := rabbitChannel.Consume(
		config.Cfg.ResponseQueueName,
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
		resp := &WorkerResponse{}
		err = xml.Unmarshal(msg.Body, resp)
		if err != nil {
			log.Fatalf("Failed to unmarshal message: [M] %v | [E] %v", msg.Body, err)
		}

		err = service.CrackService.ProcessWorkerResponse(resp.RequestId, resp.Words)
		if err != nil {
			log.Fatalf("Failed to process worker response: %v", err)
		}

		msg.Ack(false)
	}
}
