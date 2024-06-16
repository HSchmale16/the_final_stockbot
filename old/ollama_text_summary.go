package main

/**
 * complete a chat message. Calls a local ollama instance and makes the following request over http
 * curl http://localhost:11434/api/chat -d '{
  "model": "llama3",
  "messages": [
    { "role": "user", "content": "why is the sky blue?" }
  ]
}'
*/

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

type ChatResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func completeChatMessage(chatMessage, model string) (string, error) {
	// Define the request payload
	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": chatMessage,
			},
		},
	}

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to marshal JSON payload:", err)
		return "", err
	}

	// Send the HTTP POST request
	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return "", err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received non-OK status code:", resp.StatusCode)
		return "", fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	// Read the response body line by line
	var responseBody bytes.Buffer
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		var response ChatResponse
		err := json.Unmarshal([]byte(line), &response)
		if err != nil {
			fmt.Println("Failed to unmarshal response:", err)
			return "", err
		}

		responseBody.WriteString(response.Message.Content)

		if response.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Failed to read response body:", err)
		return "", err
	}

	// Print the response body
	fmt.Println(responseBody.String())

	// Return the response body as a string
	return responseBody.String(), nil
}

func parseResponse(response string) []string {
	response_tags := make([]string, 0, 20)
	tags := strings.Split(response, "\n")
	for _, tag := range tags {
		if !strings.HasPrefix(tag, "-") {
			continue
		}
		tag = strings.Trim(tag, "-")
		tag = strings.TrimSpace(tag)
		if tag != "" {
			response_tags = append(response_tags, tag)
		}
	}
	return response_tags
}

func CreateTagRelationsForModel(db *gorm.DB, item RSSItem, model LLMModel, tags []string) {
	for _, tag := range tags {
		var itemTag ItemTag
		db.Where("name = ?", tag).First(&itemTag)
		if itemTag.ID == 0 {
			// If the tag does not exist, create a new one
			itemTag = ItemTag{Name: tag}
			db.Create(&itemTag)
		}

		ItemTagRSSItem := ItemTagRSSItem{
			ItemTagID: itemTag.ID,
			RSSItemID: item.ID,
			ModelID:   model.ID,
		}

		db.Create(&ItemTagRSSItem)
	}
}

func getRssItemTags(item RSSItem, db *gorm.DB) {
	model_name := "gemma:2b"
	fmt.Println("Waiting for it to go through")
	response, err := completeChatMessage("List the topics in this article. Topics include but are not limited to the company, industry, or sector. Limit the topic to between three and five words. Each point MUST be prefixed with a hyphen (-):\n\n "+item.Title+" "+item.Description+" "+*item.ArticleBody, model_name)
	check(err)

	tagsPhi := parseResponse(response)
	var model LLMModel
	db.Where("model_name = ?", model_name).First(&model)

	CreateTagRelationsForModel(db, item, model, tagsPhi)
}
