package main

import (
	"log"

	_ "embed"

	"gopkg.in/yaml.v3"
)

type Prompt struct {
	Name       string `yaml:"name"`
	PromptText string `yaml:"prompt_text"`
}

// go:embed prompts.yaml
var yamlFile []byte

var prompts []Prompt

func init() {
	// Parse the YAML file into the Config struct
	err := yaml.Unmarshal(yamlFile, &prompts)
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
