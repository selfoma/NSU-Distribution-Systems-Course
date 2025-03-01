package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Cfg *Config

type Config struct {
	WorkerCount       int    `json:"workerCount"`
	MongoUrl          string `json:"mongoUrl"`
	ResponseQueueName string `json:"responseQueueName"`
	TaskQueueName     string `json:"taskQueueName"`
}

func LoadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening config file: %s", err)
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&Cfg); err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	fmt.Println("CONFIG: SUCCEEDED")

	return nil
}
