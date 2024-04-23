# Job Scraper

Job Scraper is a Go application that scrapes job postings from GitHub and Discord servers. It collects new job postings and sends email notifications to subscribed users.

## Features

- Scrape job postings from GitHub repositories and Discord servers
- Store collected job postings in a database
- Send email notifications to subscribed users with new job postings

## Installation

1. Clone the repository
2. Install dependencies: Golang, MongoDB, chromedp, colly
3. Create a .env file that store the following parameters: MONGO_URI, EMAIL, EMAIL_PASSWORD, GMAIL_AUTH,DISCORD_USERNAME, DISCORD_PASS
4. Build and run the application
