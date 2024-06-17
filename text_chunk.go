package main

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

/**
 * Given a full text blob chunk the text into blocks of size maxTokens with an overlap.
 */
func ChunkTextIntoTokenBlocks(fullText string, maxTokens, overlap int) []string {
	tkm, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	if err != nil {
		fmt.Println("Failed to get tokenizer:", err)
	}

	tokens := tkm.Encode(fullText, nil, nil)
	if len(tokens) > maxTokens {
		// Chunk it
		chunks := make([]string, 10)
		for i := 0; i < len(tokens); i += maxTokens - overlap {
			end := i + maxTokens
			if end > len(tokens) {
				end = len(tokens)
			}
			chunk := tkm.Decode(tokens[i:end])
			chunks = append(chunks, chunk)
		}
		return chunks
	}

	return []string{fullText}
}
