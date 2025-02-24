package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/hash/crack", handleCrackRequest)
	http.HandleFunc("/api/hash/status", handleStatusRequest)
	http.HandleFunc("/internal/api/manager/hash/crack/request", handleWorkerResponse)

	port := "8080"
	fmt.Printf("Manager service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
