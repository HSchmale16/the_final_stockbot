package stocks

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	henry_groq "github.com/hschmale16/the_final_stockbot/pkg/groq"
	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"gorm.io/gorm"
)

func ReadDocumentUploads(rc io.ReadCloser) {
	decoder := xml.NewDecoder(rc)
	db, err := m.SetupDB()
	if err != nil {
		panic(err)
	}

	tx := db.Begin()
	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Member" {
				// Do something with the document
				var member Filing
				decoder.DecodeElement(&member, &se)

				var cm m.DB_CongressMember
				db.
					Where("name LIKE '%' || ? || '%'", member.Last).
					Where("name LIKE '%' || ? || '%'", member.First).
					First(&cm)

				var doc FinDisclosureDocument
				if cm.BioGuideId != "" {
					doc.MemberId.String = cm.BioGuideId
					doc.MemberId.Valid = true
				}
				doc.Filing = member

				var x int64
				tx.Model(&FinDisclosureDocument{}).Where("doc_id = ?", member.DocId).Count(&x)
				if x > 0 {
					continue
				}

				tx.Create(&doc)
				if tx.Error != nil {
					panic(tx.Error)
				}
			}
		}
	}
	tx.Commit()

}

type Filing struct {
	Prefix     string `xml:"Prefix"`
	Last       string `xml:"Last"`
	First      string `xml:"First"`
	Suffix     string `xml:"Suffix"`
	FilingType string `xml:"FilingType"`
	StateDst   string `xml:"StateDst"`
	Year       string `xml:"Year"`
	FilingDate string `xml:"FilingDate"`
	DocId      string `xml:"DocID"`
}

func ProcessTransactionsForTarget(t FinDisclosureDocument, db *gorm.DB) {
	url := fmt.Sprintf("https://disclosures-clerk.house.gov/public_disc/ptr-pdfs/%s/%s.pdf", t.Filing.Year, t.Filing.DocId)
	fmt.Println(url)

	// Get the transactions from the pdf
	bytes, err := senatelobbying.SendRequest(url)
	if err != nil {
		panic(err)
	}

	// Write the result to a file called test.pdf
	err = os.WriteFile("test.pdf", bytes, 0644)
	if err != nil {
		panic(err)
	}

	// Parse the pdf to a text file
	outputFile := "output.txt"
	cmd := exec.Command("pdftotext", "-layout", "test.pdf", outputFile)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	// Read the text file
	textBytes, err := os.ReadFile(outputFile)
	if err != nil {
		panic(err)
	}

	text := string(textBytes)
	// Do something with the parsed text
	fmt.Println(text)
	if len(text) < 200 {
		// Update status to hand-written
		t.Processed = "H"
		db.Save(&t)
		return
	}

	model := henry_groq.Llama3_1_Vers

	prompt := `Please identify all the transactions in the document below. Print the response as json keeping the structure the same.

	Ensure any json printed is in the code block backticks.

The most important items to extract are the transaction date, the stock, the description and the amount. There might be 2 dates in a single transaction. We only care about the earlier date. 

Keep descriptions relatively short. No more than 80 characters.

Parse it into the following structure:

{
	"IPO": true/false,
	"transactions": [
		{
			"date": "2021-01-01",
			"stock": "NWN",
			"company": "Northwest Natural Holding Company
			"description": "Bought 100 shares",
			"amount": "15 - 50k"
			"filing_status": "P",
			"cap_gains_greater_than_200": true/false
		}
	]
}

`

	resp, err := henry_groq.CallGroqChatApi(model, prompt, text)
	if err != nil {
		panic(err)
	}

	//fmt.Println(resp.Choices[0].Message.Content)

	// Grab all content between the triple backticks
	content := resp.Choices[0].Message.Content
	// Find first ```
	start := strings.Index(content, "```")
	// Find second ```
	end := strings.LastIndex(content, "```")
	// Grab the content between the two backticks
	if end <= start || (end-start) < 100 {
		// Update status to hand-written
		t.Processed = "H"
		db.Save(&t)
		return
	}

	jsonContent := strings.TrimLeft(content[start+3:end], "json")

	var output GroqTranactionOutput
	err = json.Unmarshal([]byte(jsonContent), &output)
	if err != nil {
		fmt.Println(jsonContent)
		panic(err)
	}

	for _, transaction := range output.Transactions {
		var trans FinTransaction
		// Parse to a time.Time
		tm, err := henry_groq.ParseDateTimeRssRobustly(transaction.Date)
		if err != nil {
			panic(err)
		}
		trans.Date = tm
		trans.Stock = transaction.Stock
		trans.Company = transaction.Company
		trans.Description = transaction.Description
		trans.AmountCategory = transaction.Amount
		trans.FilingStatus = transaction.FilingStatus
		trans.CapGainsGreater200 = transaction.CapGainsGreaterThan200
		trans.FinDisclosureDocumentID = t.ID

		db.Create(&trans)
		if db.Error != nil {
			panic(db.Error)
		}
	}

	t.Processed = "T"
	db.Save(&t)

}

func ProcessBatchOfDocuments(db *gorm.DB) {
	var docs []FinDisclosureDocument
	db.Debug().Where("processed = ?", "0").
		Where("filing_type = ?", "P").
		Where("member_id is not NULL").
		FindInBatches(&docs, 5, func(tx *gorm.DB, batch int) error {
			fmt.Println(batch)
			for _, doc := range docs {
				ProcessTransactionsForTarget(doc, tx)
			}
			return nil
		})
}

type GroqTranactionOutput struct {
	IPO          bool
	Transactions []Transaction
}

type Transaction struct {
	Date                   string
	Stock                  string
	Company                string
	Description            string
	Amount                 string
	FilingStatus           string `json:"filing_status"`
	CapGainsGreaterThan200 bool   `json:"cap_gains_greater_than_200"`
}
