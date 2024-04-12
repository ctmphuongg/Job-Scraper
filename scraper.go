package main

import (
	// "encoding/csv"

	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/chromedp/chromedp"

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
	
		// ------ SCRAPE FROM DISCORD ---------
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}

	  fmt.Println("Scrape from discord servers")
		opt := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
		)
	 
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opt...)
		defer cancel()
 
		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()
 
		ctx, cancel = context.WithTimeout(ctx, time.Duration(300)*time.Second)
		defer cancel()

		discordSource := getDiscord(ctx)
		readDiscordHTML(discordSource)
		listJobs := readDiscordHTML(discordSource)

		defer chromedp.Cancel(ctx)


	// ------- SET UP MONGODB --------
		// Get keys from dotenv
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
		fmt.Println("Successfully connected to MongoDB!")

	// Create collection 
	var dbName = "jobDatabase"
	var collectionName = "jobs"
	collection := client.Database(dbName).Collection(collectionName)

	for _, s := range listJobs {
		if len(s) <= 2 {
			continue
		}

		newJob := Job {
			Company: s,
			Location: "US",
			Title: "Software Engineer Intern",
		}

		exists, err := jobExists(client, collection, newJob)
		if err != nil {
			log.Fatal("Error in checking job existence:", err)
		}

		if exists {
			fmt.Println("Job already exist, skipping")
		} else {
			_, err := collection.InsertOne(context.Background(), newJob)
			if err != nil {
				log.Printf("Failed to insert job into MongoDB: %v", err)
			} else {
				fmt.Println("Job added to the database successfully!")
			}
			newJobPostings = append(newJobPostings, newJob)
		}
	}
		

	// ----- SCRAPE FROM GITHUB  --------
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {fmt.Println("Scraping:", r.URL)})
	c.OnResponse(func(r *colly.Response) {fmt.Println("Status:", r.StatusCode)})


	c.OnHTML("table > tbody", func(h *colly.HTMLElement) {
		count := 0
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			if count > 20 {
				return
			}

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

			count ++

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

func getDiscord(ctx context.Context) string {
	var discordSource string
	discordChannelUrl := "https://discord.com/channels/698366411864670250/1118750128266825768"
	discordUsername := os.Getenv("DISCORD_USERNAME")
	discordPass := os.Getenv("DISCORD_PASS")
	discordScrollUp := 10

	err := chromedp.Run(ctx, 
		chromedp.Navigate("https://discord.com/channels/@me"),
	)

	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(3 * time.Second)
	
	err = chromedp.Run(ctx, 
		chromedp.SendKeys(`//input[@name="email"]`, discordUsername, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="password"]`, discordPass, chromedp.BySearch),
		chromedp.Click(`//button[@type="submit"]`, chromedp.BySearch),
	)

	if err != nil {
		log.Fatal((err))
	}

	fmt.Println("Log in successfully")

	if err != nil {
		log.Fatal("Error logging in:", err)
	}

	time.Sleep(3 * time.Second)

	err = chromedp.Run(ctx,
		chromedp.Navigate(discordChannelUrl),   
		chromedp.WaitVisible(`body`),     
		                                              
	)

	if err != nil {
		log.Fatal("Error entering channel:", err)
		
	}

	time.Sleep(4 * time.Second)

	fmt.Println("Enter channel successfully ")

	for i :=0; i < discordScrollUp; i++ {
		err := chromedp.Run(ctx, 
			chromedp.ScrollIntoView(`//ol/li[1]`, chromedp.BySearch),
		)

		if err != nil {
			log.Fatal("Error scrolling:", err)
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("Take tags successfully")

	err = chromedp.Run(ctx,
		chromedp.TextContent(`//ol`, &discordSource, chromedp.BySearch),
	)

	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(3 * time.Second)
	// fmt.Println(discordSource)
	return discordSource	
}


func readDiscordHTML(discordSource string) []string {
	listJobs := []string{}

		pattern := `\!process ([a-zA-Z]*)`
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(discordSource, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "Company" {
				listJobs = append(listJobs, match[1])
			}
		}
	// fmt.Println(listJobs)
	return listJobs
}



