package mailer

import "github.com/georgifotev1/nuvelaone-api/internal/domain"

type Mailer interface {
	Send(email domain.EmailData) error
}
