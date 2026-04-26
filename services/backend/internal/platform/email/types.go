package email

import (
	"context"
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidAddress    = errors.New("email address is invalid")
	ErrMissingRecipients = errors.New("email recipients are required")
	ErrMissingSubject    = errors.New("email subject is required")
	ErrMissingBody       = errors.New("email body is required")
	ErrUnsupportedSender = errors.New("email provider is unsupported")
	ErrSMTPStartTLS      = errors.New("smtp server does not support STARTTLS")
)

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type Address struct {
	Email string
	Name  string
}

func (a Address) Validate() error {
	if _, err := mail.ParseAddress(a.String()); err != nil {
		return ErrInvalidAddress
	}

	return nil
}

func (a Address) String() string {
	if strings.TrimSpace(a.Name) == "" {
		return strings.TrimSpace(a.Email)
	}

	return (&mail.Address{
		Name:    strings.TrimSpace(a.Name),
		Address: strings.TrimSpace(a.Email),
	}).String()
}

type Message struct {
	From     Address
	To       []Address
	Subject  string
	TextBody string
	HTMLBody string
}

func (m Message) Validate() error {
	if err := m.From.Validate(); err != nil {
		return err
	}
	if len(m.To) == 0 {
		return ErrMissingRecipients
	}
	for _, recipient := range m.To {
		if err := recipient.Validate(); err != nil {
			return err
		}
	}
	if strings.TrimSpace(m.Subject) == "" {
		return ErrMissingSubject
	}
	if strings.TrimSpace(m.TextBody) == "" && strings.TrimSpace(m.HTMLBody) == "" {
		return ErrMissingBody
	}

	return nil
}
