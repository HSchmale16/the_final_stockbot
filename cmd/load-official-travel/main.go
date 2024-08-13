package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

// var filename string

// func init() {
// 	flag.StringVar(&filename, "filename", "official-travel.csv", "The filename of the official travel data")
// }

func main() {
	promptFile, err := os.Open("travel-prompt.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer promptFile.Close()
	promptText, err := io.ReadAll(promptFile)
	if err != nil {
		log.Fatal(err)
	}

	outputFile, err := os.Create("batch.jsonl")
	if err != nil {
		log.Fatal(err)
	}

	// foreach arg process file
	for i := 1; i < min(10, len(os.Args)); i++ {
		fmt.Println(os.Args[i])
		blob := ProcessFileForBatch(os.Args[i], string(promptText))

		jsonBlob, err := json.Marshal(blob)
		if err != nil {
			log.Fatal(err)
		}

		_, err = outputFile.Write(jsonBlob)
		if err != nil {
			log.Fatal(err)
		}
		outputFile.Write([]byte("\n"))
	}
}

func ProcessFileForBatch(filename, prompt string) BatchMessage {
	// Slurp the file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	text, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return BatchMessage{
		CustomId: filename,
		Method:   "POST",
		Url:      "/v1/chat/completions",
		Body: MsgBody{
			Model: "gpt-4o-mini",
			Messages: []Message{
				{Role: "system", Content: string(prompt)},
				{Role: "user", Content: string(text)},
			},
			MaxTokens: 8192,
		},
	}
}

type BatchMessage struct {
	CustomId string  `json:"custom_id"`
	Method   string  `json:"method"`
	Url      string  `json:"url"`
	Body     MsgBody `json:"body"`
}

type MsgBody struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Old code below

func ProcessFile(filename string) {
	// Open the filename
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the file
	// Create a line reader
	scanner := bufio.NewScanner(file)

	lines := make([]string, 0)
	// Read each line
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Parse the file
	tables := parseTables(lines)

	fmt.Println(len(tables))
}

func parseTables(lines []string) []TravelReportTable {
	tables := make([]TravelReportTable, 0)

	// Iterate over the lines and extract the tables
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check if the line starts a new table
		if isTableStart(line) {
			table := TravelReportTable{}
			table.CommitteeName = extractCommitteeName(line)
			table.StartMonth, table.EndMonth, table.Year = extractDate(line)

			i++
			i++
			// until the string contains a lot of '-' we are still in the header
			for strings.Count(lines[i], "-") < 130 {
				i++
			}
			i++

			// Read the lines until the end of the table
			for i < len(lines) && !isTableEnd(lines[i]) {
				// Parse the line and add it to the table
				reportLine := parseReportLine(lines[i])
				table.Lines = append(table.Lines, reportLine)

				i++
			}

			// Add the table to the list of tables
			tables = append(tables, table)
		}
	}

	return tables
}

func isTableStart(line string) bool {
	// Implement the logic to check if the line starts a new table
	// Return true if it does, false otherwise
	return strings.Contains(line, "REPORT OF EXPEND")
}

func extractCommitteeName(line string) string {
	// Implement the logic to extract the committee name from the line
	// Return the extracted committee name
	return strings.Trim(line, " ")
}

func extractDate(line string) (string, string, string) {
	// Implement the logic to extract the start month, end month, and year from the line
	// Return the extracted start month, end month, and year
	return "", "", ""
}

func isTableEnd(line string) bool {
	// Implement the logic to check if the line is the end of a table
	// Return true if it is, false otherwise

	// 2nd scenario covers empty tables
	return strings.HasPrefix(line, "\\1\\ Per diem") || strings.Contains(line, "HOUSE COMMITTEES") || strings.HasPrefix(line, "            ")
}

func parseReportLine(line string) ReportLine {
	// Hon. Tom Price.........................    11/27       11/29   Pakistan.................  ...........        76.00  ...........       438.60  ...........  ...........  ...........  ...........
	// Split the line into fields

	fields := regexp.MustCompile(`(  )+`).Split(line, -1)

	fmt.Println(len(fields), fields)

	// Create a new ReportLine struct and populate its fields
	reportLine := ReportLine{
		Name:          fields[0],
		ArrivalDate:   fields[1],
		DepartureDate: fields[2],
		Country:       fields[3],
		// PerDiem_USD:         fields[4],
		// PerDiem_Forex:       fields[5],
		// Transport_USD:       fields[6],
		// Transport_Forex:     fields[7],
		// OtherPurposes_USD:   fields[8],
		// OtherPurposes_Forex: fields[9],
		// Total_USD:           fields[10],
		// Totel_Forex:         fields[11],
	}

	return reportLine
}

type TravelReportTable struct {
	CommitteeName string
	StartMonth    string
	EndMonth      string
	Year          string

	Lines []ReportLine
}

type ReportLine struct {
	Name          string
	ArrivalDate   string
	DepartureDate string
	Country       string
	// PerDiem_USD         string
	// PerDiem_Forex       string
	// Transport_USD       string
	// Transport_Forex     string
	// OtherPurposes_USD   string
	// OtherPurposes_Forex string
	// Total_USD           string
	// Totel_Forex         string
}
