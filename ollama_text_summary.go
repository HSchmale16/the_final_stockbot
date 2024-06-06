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

	"github.com/jinzhu/gorm"
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

func completeChatMessage(chatMessage string) (string, error) {
	// Define the request payload
	payload := map[string]interface{}{
		"model": "gemma:2b",
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
		fmt.Println(line)

		var response ChatResponse
		err := json.Unmarshal([]byte(line), &response)
		if err != nil {
			fmt.Println("Failed to unmarshal response:", err)
			return "", err
		}

		responseBody.WriteString(response.Message.Content)

		fmt.Println(responseBody.String())

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

func getRssItemTags(item RSSItem, db *gorm.DB) {
	fmt.Println("Waiting for it to go through")
	response, err := completeChatMessage("List the topics in this article. Each point should be prefixed with a hyphen (-): " + item.Title + " " + item.Description)
	check(err)

	// Split the response into individual tags
	tags := strings.Split(response, "\n")
	for _, tag := range tags {
		if !strings.HasPrefix(tag, "-") {
			continue
		}
		tag = strings.Trim(tag, "-")
		tag = strings.TrimSpace(tag)
		if tag != "" {
			fmt.Println("Tag:", tag)
			// Check if the tag already exists in the database
			var itemTag ItemTag
			db.Where("name = ?", tag).First(&itemTag)
			if itemTag.ID == 0 {
				// If the tag does not exist, create a new one
				itemTag = ItemTag{Name: tag}
				db.Create(&itemTag)
			}

			// Associate the tag with the RSS item
			db.Model(&itemTag).Association("RSSItems").Append(&item)
			db.Save(&itemTag)
		}
	}
}
