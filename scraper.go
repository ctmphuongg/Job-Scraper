package main

import (
	// "encoding/csv"

	"context"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
) 
 
type Job struct {
	Title    string `bson:"title"`
	Company  string `bson:"company"`
	Location string `bson:"location"`
}

func main() { 

	var newJobPostings []Job;

// debug 
// newJob := Job{
// 	Title:    "testA",
// 	Company:  "companyA",
// 	Location: "locA",
// }
// newJobPostings = append(newJobPostings, newJob)
// newJob2 := Job{
// 	Title:    "testB",
// 	Company:  "companyB",
// 	Location: "locB",
// }
// newJobPostings = append(newJobPostings, newJob2)
	
	// Set up MongoDB 
		// Get keys from dotenv
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}
		uri := os.Getenv("MONGO_URI")
	
		// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	
		// Create a new client and connect to the server
		client, err := mongo.Connect(context.TODO(), opts)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()
			// Send a ping to confirm a successful connection
		if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
			panic(err)
		}
		fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	// Create collection 
	var dbName = "jobDatabase"
	var collectionName = "jobs"
	collection := client.Database(dbName).Collection(collectionName)

	// Set up Go colly for Web scraping
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {fmt.Println("Scraping:", r.URL)})
	c.OnResponse(func(r *colly.Response) {fmt.Println("Status:", r.StatusCode)})


	c.OnHTML("table > tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			// newJob := Job{}
			company := el.ChildText("td:nth-child(1)")
			title := el.ChildText("td:nth-child(2)")
			location := el.ChildText("td:nth-child(3)")

			newJob := Job{
				Title:    title,
				Company:  company,
				Location: location,
			}

				// Check if the job already exists in the database
			exists, err := jobExists(client, collection, newJob)
			if err != nil {
				log.Fatal("Error checking job existence:", err)
			}
			
			if exists {
				fmt.Println("Job already exist, skipping")
				
			} else {
				// JobPostings = append(JobPostings, newJob)
				_, err := collection.InsertOne(context.Background(), newJob)
				if err != nil {
					log.Printf("Failed to insert job into MongoDB: %v", err)
				} else {
					fmt.Println("Job added to the database successfully!")
				}
				newJobPostings = append(newJobPostings, newJob)
			}

			
	})})
	
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
})
	c.Visit("https://github.com/SimplifyJobs/Summer2024-Internships")
	emailSending(newJobPostings)
}

func jobExists(client *mongo.Client, collection *mongo.Collection, job Job) (bool, error) {
	// Define filter criteria to find a job by company and title
	filter := bson.M{"company": job.Company, "title": job.Title}

	// Execute a find one operation
	var result Job
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Job not found
			return false, nil
		}
		// Other error occurred
		return false, err
	}

	// Job found
	return true, nil
}

