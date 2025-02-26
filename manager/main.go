package main

import (
	"fmt"
	"github.com/selfoma/crackhash/manager/broker"
	"github.com/selfoma/crackhash/manager/config"
	"github.com/selfoma/crackhash/manager/database"
	"log"
	"net/http"
)

func main() {
	err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatal(err)
	}

	err = database.ConnectMongo()
	if err != nil {
		log.Fatal(err)
	}

	err = broker.ConnectRabbitMq()
	if err != nil {
		log.Fatal(err)
	}

	go database.RetryPendingTask()

	http.HandleFunc("/api/hash/crack", handleCrackRequest)
	http.HandleFunc("/api/hash/status", handleStatusRequest)

	port := "8080"
	fmt.Printf("Manager service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
