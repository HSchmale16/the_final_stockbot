package main

import (
	"fmt"
	"log"

	_ "embed"

	"gopkg.in/yaml.v3"
)

var config struct {
	GovtFeedUrls []FeedUrl `yaml:"govt_feed_urls"`
	Prompts      []Prompt  `yaml:"prompts"`
}

type FeedUrl struct {
	Url         string `yaml:"url"`
	Description string `yaml:"description"`
}

type Prompt struct {
	Name       string `yaml:"name"`
	PromptText string `yaml:"prompt_text"`
}

//go:embed prompts.yaml
var yamlFile []byte

func init() {
	fmt.Println("Prompt yaml length", len(yamlFile))
	// Parse the YAML file into the Config struct
	err := yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Failed to parse YAML file: %v", err)
	}

	log.Println("Loaded", len(config.Prompts), "prompts")
}

func GetPrompt() Prompt {
	// Return the configuration
	// Example:
	// return config

	return config.Prompts[0]
}
