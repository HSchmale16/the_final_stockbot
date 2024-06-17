package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Model string

const (
	Llama3_70B   Model = "llama3-70b-8192"
	Llama3_8B    Model = "llama3-8b-8192"
	Mixtral_8x7b Model = "mixtral-8x7b-32768"
	Gemma_7b     Model = "gemma-7b-it"
	// Add more models here
)

type GroqStreamingResponse struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             Model  `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index        int     `json:"index"`
		Delta        Message `json:"delta"`
		Logprobs     string  `json:"logprobs"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	XGroq struct {
		ID    string    `json:"id"`
		Usage GroqUsage `json:"usage"`
	} `json:"x_groq"`
}

type GroqUsage struct {
	QueueTime        float64 `json:"queue_time"`
	PromptTokens     int     `json:"prompt_tokens"`
	PromptTime       float64 `json:"prompt_time"`
	CompletionTokens int     `json:"completion_tokens"`
	CompletionTime   float64 `json:"completion_time"`
	TotalTokens      int     `json:"total_tokens"`
	TotalTime        float64 `json:"total_time"`
}

type GroqChatCompletion struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             Model  `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		Logprobs     string  `json:"logprobs"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage GroqUsage `json:"usage"`
	XGroq struct {
		ID    string    `json:"id"`
		Usage GroqUsage `json:"usage"`
	} `json:"x_groq"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func CallGroqChatApi(model Model, systemPrompt, userData string) (GroqChatCompletion, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"
	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userData,
			},
		},
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
	}

	var chatCompletion GroqChatCompletion

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to marshal JSON payload:", err)
		return chatCompletion, err
	}

	// Send the HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return chatCompletion, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return chatCompletion, err
	}

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received non-OK status code:", resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Failed to read response body:", err)
			return chatCompletion, err
		}
		fmt.Println("Response Body:", string(body))
		return chatCompletion, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		return chatCompletion, err
	}
	fmt.Println("Response Body:", string(body))

	// // Read the response body line by line
	// var responseBody bytes.Buffer
	// scanner := bufio.NewScanner(resp.Body)
	// for scanner.Scan() {
	// 	line := scanner.Text()

	// 	// fmt.Print(line)
	// 	if line == "" {
	// 		continue
	// 	}

	// 	if strings.Contains(line, "[DONE]") {
	// 		break
	// 	}

	// 	data := []byte(line)[5:]

	// 	var response ChatResponse
	// 	err := json.Unmarshal(data, &response)
	// 	if err != nil {
	// 		fmt.Println("Failed to unmarshal response:", err)
	// 		return "", err
	// 	}

	// 	responseBody.WriteString(response.Choices[0].Delta.Content)

	// }

	// if err := scanner.Err(); err != nil {
	// 	fmt.Println("Failed to read response body:", err)
	// 	return "", err
	// }

	err = json.Unmarshal(body, &chatCompletion)
	if err != nil {
		fmt.Println("Failed to unmarshal response:", err)
		return chatCompletion, err
	}

	return chatCompletion, nil
}
