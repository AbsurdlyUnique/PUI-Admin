// Julian: This file handles saving and loading PostgreSQL configurations.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Julian: Structure to hold the database configuration.
type Config struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// Julian: File path where the configuration will be stored.
const configFilePath = "configs.json"

// Julian: Function to load configurations from the JSON file.
func loadConfigs() ([]Config, error) {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return []Config{}, nil // Julian: Return empty list if no file exists
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("Julian: Could not read config file: %v", err)
	}

	var configs []Config
	err = json.Unmarshal(data, &configs)
	if err != nil {
		return nil, fmt.Errorf("Julian: Error parsing config file: %v", err)
	}

	return configs, nil
}

// Julian: Function to save configurations to the JSON file.
func saveConfigs(configs []Config) error {
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("Julian: Error marshalling configs to JSON: %v", err)
	}

	err = ioutil.WriteFile(configFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("Julian: Could not write config file: %v", err)
	}

	return nil
}

