package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Cfg *Config

type Config struct {
	ManagerUrl        string `json:"managerUrl"`
	ResponseQueueName string `json:"responseQueueName"`
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

	return nil
}
