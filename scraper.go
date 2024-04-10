package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
) 
 
func main() { 

	var JobPostings []Job;
	c := colly.NewCollector()


	c.OnRequest(func(r *colly.Request) {fmt.Println("Scraping:", r.URL)})
	c.OnResponse(func(r *colly.Response) {fmt.Println("Status:", r.StatusCode)})


	c.OnHTML("table > tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			newJob := Job{}
			newJob.company = el.ChildText("td:nth-child(1)")
			newJob.title = el.ChildText("td:nth-child(2)")
			newJob.location = el.ChildText("td:nth-child(3)")
			
			JobPostings = append(JobPostings, newJob)
			// fmt.Println(el.ChildAttr("td:nth-of-type(4)", "a"))
	})})
	
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
})
	c.Visit("https://github.com/SimplifyJobs/Summer2024-Internships")
	 
		// opening the CSV file 
		file, err := os.Create("job-infos.csv") 
		if err != nil { 
			log.Fatalln("Failed to create output CSV file", err) 
		} 
		defer file.Close() 
	 
		// initializing a file writer 
		writer := csv.NewWriter(file) 
		headers := []string{ 
			"title", 
			"company", 
			"location", 
			// "url", 
		} 
		writer.Write(headers) 

	for _, JobPosting := range JobPostings { 
		// converting a JobPosting to an array of strings 
		record := []string{ 
			// JobPosting.url, 
			JobPosting.company, 
			JobPosting.title, 
			JobPosting.location, 
		} 
 
		// adding a CSV record to the output file 
		writer.Write(record) 
	} 
	defer writer.Flush() 
		}

type Job struct {
	title, company, location string
}
