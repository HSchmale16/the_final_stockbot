package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
)

func runCurlCommand(url string, filePath string) error {
	cmd := exec.Command("curl", "-o", filePath, url)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func loadNasdaqSecurities(db *gorm.DB, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Process the fields as needed
		if strings.HasPrefix(line, "Symbol") || strings.HasPrefix(line, "File Creation Time") {
			// Skip the header line
			continue
		}

		fields := strings.Split(line, "|")

		symbol := fields[0]
		name := fields[1]

		marketSecurity := MarketSecurity{
			Symbol:   symbol,
			Name:     name,
			IsEtf:    false,
			Exchange: "NASDAQ",
		}
		var existingSecurity MarketSecurity
		db.Where("symbol = ?", symbol).First(&existingSecurity)
		if existingSecurity.ID == 0 {
			db.Create(&marketSecurity)
		}
		// TODO: Add marketSecurity to a collection or perform any other desired operation
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

}

func downloadSecurities(db *gorm.DB) {
	url := "ftp://ftp.nasdaqtrader.com/SymbolDirectory/nasdaqlisted.txt"
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, "nasdaqlisted.txt")

	err := runCurlCommand(url, filePath)
	if err != nil {
		fmt.Println("File download failed:", err)
		return
	}
	loadNasdaqSecurities(db, filePath)

	fmt.Println("File downloaded successfully!")
}
