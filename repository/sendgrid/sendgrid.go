package sendgrid

import (
	"os"

	"github.com/g-graziano/userland/models"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGrid interface {
	SendEmail(email *models.Email) error
}

func SendEmail(email *models.Email) error {
	go func() {
		from := mail.NewEmail("User Land", "verifier@userland.com")
		subject := email.Subject
		to := mail.NewEmail(email.RecipientName, email.RecipientEmail)
		plainTextContent := email.PlainContent
		htmlContent := email.HTMLContent

		message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
		client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
		client.Send(message)
	}()

	return nil
}
