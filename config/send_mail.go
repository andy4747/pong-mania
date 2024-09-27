package config

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

const (
	CharSet = "UTF-8"
	Timeout = 10 * time.Second
)

func SendEmail(receiverEmail, senderEmail, subject, htmlTemplate string, data interface{}, sesClient *ses.Client) error {
	// Parse and execute the template with the provided data
	tmpl, err := template.New("emailTemplate").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	textBody := "Please view this email in an HTML-compatible client."

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	// Send the email
	_, err = sesClient.SendEmail(ctx, &ses.SendEmailInput{
		Source: aws.String(senderEmail),
		Destination: &types.Destination{
			ToAddresses: []string{receiverEmail},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(htmlBody.String()),
				},
				Text: &types.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(textBody),
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	log.Printf("Email sent successfully to %s", receiverEmail)
	return nil
}

func SendMagicLink(receiverEmail string, senderEmail string, token string, client *ses.Client) error {
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Pong Arena Account</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background-color: #000;
            color: #39FF14;
            line-height: 1.6;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #0a0a0a;
            border: 2px solid #00f3ff;
            box-shadow: 0 0 10px #00f3ff;
        }
        h1 {
            color: #9d00ff;
            text-align: center;
        }
				p{
					color: #00F3FF;
				}
        .btn {
            display: inline-block;
            padding: 10px 20px;
            background-color: #9d00ff;
            color: #000;
            text-decoration: none;
            border-radius: 5px;
            font-weight: bold;
            text-align: center;
        }
        .btn:hover {
            background-color: #ff00f7;
        }
    </style>
</head>
<body>
    <div class="container">
			<h1>Welcome to Pong Arena!</h1>
			<p>Greetings, {{.Username}}!</p>
			<p>You're just one step away from entering the arena. Click the button below to verify your account and begin your journey:</p>
			<p style="text-align: center;">
					<a href="{{.Link}}" class="btn">Login TO Your Account</a>
			</p>
			<p>If you didn't request this verification, please ignore this email.</p>
			<p>Good luck, and may your reflexes be swift!</p>
			<p>The Pong Arena Team</p>
    </div>
</body>
</html>
`
	link := fmt.Sprintf("https://pongmania.com/auth/magic/verify?token=%s", url.QueryEscape(token))
	data := struct {
		Username string
		Link     string
	}{
		Username: strings.Split(receiverEmail, "@")[0],
		Link:     link,
	}

	err := SendEmail(receiverEmail, senderEmail, "Verify Your Pong Arena Session", htmlTemplate, data, client)
	if err != nil {
		return err
	}
	return nil
}
