package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	WorkerCount int      `json:"workerCount"`
	WorkerUrls  []string `json:"workerUrls"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err = json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
