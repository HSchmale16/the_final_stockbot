/**
 * Scrape the official travel from the congress site. Downloads all the text files
 */

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func findElementWithRetry(driver selenium.WebDriver, by string, value string) (selenium.WebElement, error) {
	var element selenium.WebElement
	var err error
	for i := 0; i < 5; i++ {
		element, err = driver.FindElement(by, value)
		if err == nil {
			return element, nil
		}
		time.Sleep(400 * time.Millisecond)
	}
	return element, err
}

func updateTheYear(driver selenium.WebDriver, year string) {
	// Select the year
	yearElement, err := findElementWithRetry(driver, selenium.ByXPATH, "//select[@id='Year']")
	if err != nil {
		log.Fatal("Error:", err)
	}
	yearElement.Click()

	// Select the year
	yearElement, err = driver.FindElement(selenium.ByXPATH, "//select[@id='Year']/option[@value='"+year+"']")
	if err != nil {
		log.Fatal("Error:", err)
	}
	yearElement.Click()
}

func updateTheQuarter(driver selenium.WebDriver, quarter string) {
	// Select the quarter
	quarterElement, err := findElementWithRetry(driver, selenium.ByXPATH, "//select[@id='Quarter']")
	if err != nil {
		log.Fatal("Error:", err)
	}
	quarterElement.Click()

	// Select the quarter
	quarterElement, err = driver.FindElement(selenium.ByXPATH, "//select[@id='Quarter']/option[@value='"+quarter+"']")
	if err != nil {
		log.Fatal("Error:", err)
	}
	quarterElement.Click()
}

func dumpStringList(fileName string, list []string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Error:", err)
	}
	defer file.Close()

	for _, item := range list {
		file.WriteString(item + "\n")
	}
}

func main() {
	// Scrape the official travel from the congress site
	// Downloads all the text files

	urlListFile := make([]string, 0)

	service, err := selenium.NewChromeDriverService("chromedriver", 4444)
	if err != nil {
		fmt.Println("Error starting the ChromeDriver server: ", err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"--headless-new", // comment out this line for testing
	}})

	// create a new remote client with the specified options
	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		log.Fatal("Error:", err)
	}
	// defer driver.Quit()

	// maximize the current window to avoid responsive rendering
	err = driver.MaximizeWindow("")
	if err != nil {
		log.Fatal("Error:", err)
	}

	driver.Get("https://disclosures-clerk.house.gov/ForeignTravel")

	for i := 2016; i <= 2024; i++ {
		year := fmt.Sprintf("%d", i)
		updateTheYear(driver, year)
		quarters := []string{"q1", "q2", "q3", "q4"}
		for _, quarter := range quarters {
			updateTheQuarter(driver, quarter)

			// Find the search button
			xpath := "//*[@id=\"electronic-reports\"]/div[1]/div/form/div[2]/button"
			searchButton, err := findElementWithRetry(driver, selenium.ByXPATH, xpath)
			if err != nil {
				log.Fatal("Error:", err)
			}
			searchButton.Click()

			time.Sleep(2 * time.Second)

			// Find the search results
			findElementWithRetry(driver, selenium.ByID, "search-result")

			// Find the file download buttons
			links, err := driver.FindElements(selenium.ByXPATH, "//a[contains(@href, 'txt')]")
			if err != nil {
				log.Fatal("Error:", err)
			}

			for _, link := range links {
				href, err := link.GetAttribute("href")
				if err != nil {
					log.Fatal("Error:", err)
				}

				// clobber
				urlListFile = append(urlListFile, href)
			}

			dumpStringList("urlList2.txt", urlListFile)

			fmt.Println("Year:", year, "Quarter:", quarter, "Links:", len(links))
			// Wait for a bit
			time.Sleep(5 * time.Second)
		}

	}

	fmt.Println("Scrape the official travel from the congress site")
}
