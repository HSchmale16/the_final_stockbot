package main

import "testing"

func TestChunkTextIntoTokenBlocks(t *testing.T) {
	fullText := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	maxTokens := 10
	overlap := 5

	expectedChunks := []string{
		"Lorem ipsum",
		"ipsum dolor",
		"dolor sit",
		"sit amet,",
		"amet, consectetur",
		"consectetur adipiscing",
		"adipiscing elit.",
	}

	chunks := ChunkTextIntoTokenBlocks(fullText, maxTokens, overlap)

	if len(chunks) != len(expectedChunks) {
		t.Errorf("Expected %d chunks, but got %d", len(expectedChunks), len(chunks))
	}

	for i, chunk := range chunks {
		if chunk != expectedChunks[i] {
			t.Errorf("Expected chunk '%s', but got '%s'", expectedChunks[i], chunk)
		}
	}
}
