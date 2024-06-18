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
	GROQ_TOKEN string = ""
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
				"role":    "user",
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", GROQ_TOKEN))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return chatCompletion, err
	}

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Handle Rate Limit
		if resp.StatusCode == http.StatusTooManyRequests {
			fmt.Println("Rate limit exceeded. Please wait and try again later.")
			return chatCompletion, generateRateLimitErrorMessage(resp)
		}

		fmt.Println("Received non-OK status code:", resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Failed to read response body:", err)
			return chatCompletion, err
		}
		fmt.Println("Response Body:", string(body))
		return chatCompletion, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	// Verify the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		return chatCompletion, err
	}
	fmt.Println("Response Body:", string(body))

	err = json.Unmarshal(body, &chatCompletion)
	if err != nil {
		fmt.Println("Failed to unmarshal response:", err)
		return chatCompletion, err
	}

	return chatCompletion, nil
}

/*
*

Header	Value	Notes
retry-after	2	In seconds
x-ratelimit-limit-requests	14400	Always refers to Requests Per Day (RPD)
x-ratelimit-limit-tokens	18000	Always refers to Tokens Per Minute (TPM)
x-ratelimit-remaining-requests	14370	Always refers to Requests Per Day (RPD)
x-ratelimit-remaining-tokens	17997	Always refers to Tokens Per Minute (TPM)
x-ratelimit-reset-requests	2m59.56s	Always refers to Requests Per Day (RPD)
x-ratelimit-reset-tokens	7.66s	Always refers to Tokens Per Minute (TPM)
*/
func generateRateLimitErrorMessage(resp *http.Response) error {
	// Parse the rate limit headers
	retryAfter := resp.Header.Get("retry-after")
	rateLimitLimitRequests := resp.Header.Get("x-ratelimit-limit-requests")
	rateLimitLimitTokens := resp.Header.Get("x-ratelimit-limit-tokens")
	rateLimitRemainingRequests := resp.Header.Get("x-ratelimit-remaining-requests")
	rateLimitRemainingTokens := resp.Header.Get("x-ratelimit-remaining-tokens")
	rateLimitResetRequests := resp.Header.Get("x-ratelimit-reset-requests")
	rateLimitResetTokens := resp.Header.Get("x-ratelimit-reset-tokens")

	fmt.Println("Rate limit details", retryAfter, rateLimitLimitRequests, rateLimitLimitTokens, rateLimitRemainingRequests, rateLimitRemainingTokens, rateLimitResetRequests, rateLimitResetTokens)
	return fmt.Errorf("unknown rate limit error")
}

type LlmTool struct {
	Type     string          `json:"type"`
	Function LlmToolFunction `json:"function"`
}

type ToolParameter struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ParameterMap struct {
	Type       string                   `json:"type"`
	Properties map[string]ToolParameter `json:"properties"`
	Required   []string                 `json:"required"`
}

type LlmToolFunction struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Parameters  ParameterMap `json:"parameters"`
}

func CallGroqTools(model Model, systemPrompt, userData string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"
	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a function calling LLM that uses the get_mtg_card tool to look up magic the gathering cards and analyze interactions between them",
			},
			{
				"role":    "user",
				"content": userData,
			},
		},
		"tool_choice": "auto",
		"tools": []LlmTool{
			{
				Type: "function",
				Function: LlmToolFunction{
					Name:        "get_mtg_card",
					Description: "Get a Magic: The Gathering card by name",
					Parameters: ParameterMap{
						Type: "object",
						Properties: map[string]ToolParameter{
							"card_name": {
								Type:        "string",
								Description: "The name of the Magic: The Gathering card",
							},
						},
					},
				},
			},
		},
	}

	var chatCompletion string = ""

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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", GROQ_TOKEN))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return chatCompletion, err
	}

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Handle Rate Limit
		if resp.StatusCode == http.StatusTooManyRequests {
			fmt.Println("Rate limit exceeded. Please wait and try again later.")
			return chatCompletion, generateRateLimitErrorMessage(resp)
		}

		fmt.Println("Received non-OK status code:", resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Failed to read response body:", err)
			return chatCompletion, err
		}
		fmt.Println("Response Body:", string(body))
		return string(body), fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	// Verify the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		return "", err
	}
	fmt.Println("Response Body:", string(body))

	return string(body), nil
}
