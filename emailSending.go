package main

import (
	// "encoding/csv"
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
) 


func emailSending(jobs []Job) error {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	
	from := os.Getenv("EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")

	  // Receiver email address.
		to := []string{
			"notijob2@gmail.com",
		}

		// smtp server configuration.
		smtpHost := "smtp.gmail.com"
		smtpPort := "587"

		// Authentication.
		auth := smtp.PlainAuth("", from, password, smtpHost)

		t, err := template.New("emailTemplate").Parse(emailTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse email template: %v", err)
		}
	
		var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
  body.Write([]byte(fmt.Sprintf("Subject: New Job Posting Today! \n%s\n\n", mimeHeaders)))

	t.Execute(&body, jobs)
	if err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	// Sending email.
	e := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if e != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to send email: %v", err)
	}
	fmt.Println("Email Sent!")
	return nil
}

// HTML email template with job list.
var emailTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>Job Listings</title>
</head>
<body>
	<h1>Job Listings</h1>
	<table border="1">
		<tr>
			<th>Company</th>
			<th>Title</th>
			<th>Location</th>
		</tr>
		{{range .}}
		<tr>
			<td>{{.Company}}</td>
			<td>{{.Title}}</td>
			<td>{{.Location}}</td>
		</tr>
		{{end}}
	</table>
</body>
</html>
`