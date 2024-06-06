package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runCurlCommand(url string, filePath string) error {
	cmd := exec.Command("curl", "-o", filePath, url)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func downloadNasdaqListed() {
	url := "ftp://ftp.nasdaqtrader.com/SymbolDirectory/nasdaqlisted.txt"
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, "nasdaqlisted.txt")

	err := runCurlCommand(url, filePath)
	if err != nil {
		fmt.Println("File download failed:", err)
		return
	}

	fmt.Println("File downloaded successfully!")
}
