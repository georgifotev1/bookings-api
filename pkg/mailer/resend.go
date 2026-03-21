package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/resend/resend-go/v3"
)

//go:embed templates/*.html
var templates embed.FS

type resendClient struct {
	client    *resend.Client
	fromEmail string
}

func NewResendClient(apiKey, fromEmail string) *resendClient {
	client := resend.NewClient(apiKey)
	return &resendClient{
		client:    client,
		fromEmail: fromEmail,
	}
}

func (r *resendClient) Send(email domain.EmailData) error {
	templatePath := "templates/" + email.Template

	templateContent, err := templates.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", email.Template, err)
	}

	tmpl, err := template.New(email.Template).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", email.Template, err)
	}

	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, email.Data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", email.Template, err)
	}

	params := &resend.SendEmailRequest{
		From:    r.fromEmail,
		To:      email.To,
		Subject: email.Subject,
		Html:    htmlBody.String(),
	}

	_, err = r.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email via resend: %w", err)
	}

	return nil
}
