package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/internal/api/worker/hash/crack/task", handleWorkerTask)

	port := "8081"
	fmt.Printf("Worker running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
