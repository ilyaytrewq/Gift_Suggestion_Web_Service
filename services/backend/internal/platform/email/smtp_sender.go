package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

type smtpClient interface {
	Extension(ext string) (bool, string)
	StartTLS(config *tls.Config) error
	Auth(auth smtp.Auth) error
	Mail(from string) error
	Rcpt(to string) error
	Data() (io.WriteCloser, error)
	Quit() error
	Close() error
}

type smtpFactory interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	NewClient(conn net.Conn, host string) (smtpClient, error)
}

type defaultSMTPFactory struct{}

func (defaultSMTPFactory) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, network, address)
}

func (defaultSMTPFactory) NewClient(conn net.Conn, host string) (smtpClient, error) {
	return smtp.NewClient(conn, host)
}

type SMTPSender struct {
	from     Address
	host     string
	port     int
	username string
	password string
	useTLS   bool
	factory  smtpFactory
}

func NewSMTPSender(cfg config.EmailConfig) (*SMTPSender, error) {
	sender := &SMTPSender{
		from: Address{
			Email: cfg.FromEmail,
			Name:  cfg.FromName,
		},
		host:     strings.TrimSpace(cfg.SMTP.Host),
		port:     cfg.SMTP.Port,
		username: strings.TrimSpace(cfg.SMTP.Username),
		password: cfg.SMTP.Password,
		useTLS:   cfg.SMTP.UseTLS,
		factory:  defaultSMTPFactory{},
	}
	if err := sender.from.Validate(); err != nil {
		return nil, err
	}

	return sender, nil
}

func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	if strings.TrimSpace(msg.From.Email) == "" {
		msg.From = s.from
	}
	if err := msg.Validate(); err != nil {
		return err
	}

	address := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))
	conn, err := s.factory.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}

	client, err := s.factory.NewClient(conn, s.host)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if s.useTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return ErrSMTPStartTLS
		}
		if err := client.StartTLS(&tls.Config{
			ServerName:         s.host,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}); err != nil {
			return err
		}
	}

	if s.username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.username, s.password, s.host)); err != nil {
			return err
		}
	}

	if err := client.Mail(msg.From.Email); err != nil {
		return err
	}
	for _, recipient := range msg.To {
		if err := client.Rcpt(recipient.Email); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	if _, err := writer.Write(renderMessage(msg)); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		return err
	}

	return nil
}

func renderMessage(msg Message) []byte {
	var body bytes.Buffer

	body.WriteString(fmt.Sprintf("From: %s\r\n", msg.From.String()))
	body.WriteString(fmt.Sprintf("To: %s\r\n", recipientsHeader(msg.To)))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	body.WriteString("MIME-Version: 1.0\r\n")

	if strings.TrimSpace(msg.HTMLBody) != "" {
		boundary := fmt.Sprintf("gift-suggestion-%d", time.Now().UnixNano())
		body.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n", boundary))
		body.WriteString("\r\n")
		body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		body.WriteString(msg.TextBody)
		body.WriteString("\r\n")
		body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		body.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		body.WriteString(msg.HTMLBody)
		body.WriteString("\r\n")
		body.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

		return body.Bytes()
	}

	body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	body.WriteString(msg.TextBody)

	return body.Bytes()
}

func recipientsHeader(recipients []Address) string {
	values := make([]string, 0, len(recipients))
	for _, recipient := range recipients {
		values = append(values, recipient.String())
	}

	return strings.Join(values, ", ")
}
