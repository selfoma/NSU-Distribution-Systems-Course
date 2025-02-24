package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var config *Config

type Config struct {
	ManagerUrl string `json:"managerUrl"`
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

	return nil
}
