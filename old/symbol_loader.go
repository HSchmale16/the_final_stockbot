package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
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

func loadOtherSecurities(db *gorm.DB, filePath string) {
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
		if strings.HasPrefix(line, "ACT Symbol") || strings.HasPrefix(line, "File Creation Time") {
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
			Exchange: "OTHER",
		}
		var existingSecurity MarketSecurity
		db.Where("symbol = ?", symbol).First(&existingSecurity)
		if existingSecurity.ID == 0 {
			db.Create(&marketSecurity)
		}
	}
}

func downloadSecurities(db *gorm.DB) {
	url := "ftp://ftp.nasdaqtrader.com/SymbolDirectory/nasdaqlisted.txt"
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, "nasdaqlisted.txt")

	log.Println("Downloading file from", url)
	err := runCurlCommand(url, filePath)
	if err != nil {
		fmt.Println("File download failed:", err)
		return
	}
	loadNasdaqSecurities(db, filePath)

	log.Println("Downloaded NASDAQ securities successfully!")
	log.Println("Downloading other listed securities...")
	url = "ftp://ftp.nasdaqtrader.com/SymbolDirectory/otherlisted.txt"
	filePath = filepath.Join(tempDir, "otherlisted.txt")
	err = runCurlCommand(url, filePath)
	check(err)
	loadOtherSecurities(db, filePath)

	fmt.Println("File downloaded successfully!")
}
