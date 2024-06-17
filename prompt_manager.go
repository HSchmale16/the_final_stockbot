package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Prompt struct {
	// Define your configuration struct here
	// Example:
	// APIKey string `yaml:"api_key"`

	Name       string `yaml:"name"`
	PromptText string `yaml:"prompt_text"`
}

var prompts []Prompt

func init() {
	// Read the YAML file
	yamlFile, err := os.ReadFile("prompts.yaml")
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	// Parse the YAML file into the Config struct
	err = yaml.Unmarshal(yamlFile, &prompts)
	if err != nil {
		log.Fatalf("Failed to parse YAML file: %v", err)
	}

	// Use the loaded configuration
	// Example:
	// fmt.Println("API Key:", config.APIKey)
}

func GetPrompt() Prompt {
	// Return the configuration
	// Example:
	// return config

	return prompts[0]
}
