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
var templatesFS embed.FS

type ResendMailer struct {
	client    *resend.Client
	fromEmail string
}

func NewResendMailer(apiKey, fromEmail string) *ResendMailer {
	client := resend.NewClient(apiKey)
	return &ResendMailer{
		client:    client,
		fromEmail: fromEmail,
	}
}

func (r *ResendMailer) Send(email domain.EmailData) error {
	templatePath := "templates/" + email.Template

	templateContent, err := templatesFS.ReadFile(templatePath)
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
