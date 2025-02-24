package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var config *Config

type Config struct {
	WorkerCount int      `json:"workerCount"`
	WorkerUrls  []string `json:"workerUrls"`
}

func loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening config file: %s", err)
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	fmt.Println("CONFIG: SUCCEEDED")

	return nil
}
