package main

import (
	"fmt"
	"github.com/selfoma/crackhash/worker/broker"
	"github.com/selfoma/crackhash/worker/config"
	"log"
	"net/http"
)

func main() {
	err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatal(err)
	}

	err = broker.ConnectRabbit()
	if err != nil {
		log.Fatal(err)
	}

	go broker.ConsumeTask()

	port := "8081"
	fmt.Printf("Worker running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
