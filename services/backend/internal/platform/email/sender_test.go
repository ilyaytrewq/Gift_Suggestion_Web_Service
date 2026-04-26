package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/smtp"
	"testing"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

func TestMessageValidateRequiresRecipients(t *testing.T) {
	t.Parallel()

	message := Message{
		From:     Address{Email: "noreply@example.com"},
		Subject:  "Subject",
		TextBody: "Body",
	}

	if err := message.Validate(); !errors.Is(err, ErrMissingRecipients) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrMissingRecipients)
	}
}

func TestNoopSenderValidatesMessage(t *testing.T) {
	t.Parallel()

	sender := NoopSender{}
	if err := sender.Send(context.Background(), Message{
		From:     Address{Email: "noreply@example.com"},
		To:       []Address{{Email: "user@example.com"}},
		Subject:  "Hello",
		TextBody: "World",
	}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSMTPSenderSendUsesStartTLSAndAuth(t *testing.T) {
	t.Parallel()

	client := &fakeSMTPClient{
		writer:    &fakeWriteCloser{},
		startTLS:  true,
		extension: true,
	}
	sender, err := NewSMTPSender(config.EmailConfig{
		FromEmail: "noreply@example.com",
		FromName:  "Gift Suggestion",
		SMTP: config.SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "mailer",
			Password: "secret",
			UseTLS:   true,
		},
	})
	if err != nil {
		t.Fatalf("NewSMTPSender() error = %v", err)
	}
	sender.factory = fakeSMTPFactory{client: client}

	err = sender.Send(context.Background(), Message{
		From:     Address{Email: "noreply@example.com", Name: "Gift Suggestion"},
		To:       []Address{{Email: "user@example.com", Name: "Alice"}},
		Subject:  "Confirm your email",
		TextBody: "Verification code: abc123",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if !client.startTLSCalled {
		t.Fatal("expected STARTTLS to be used")
	}
	if !client.authCalled {
		t.Fatal("expected AUTH to be used")
	}
	if client.mailFrom != "noreply@example.com" {
		t.Fatalf("mail from = %q, want %q", client.mailFrom, "noreply@example.com")
	}
	if len(client.recipients) != 1 || client.recipients[0] != "user@example.com" {
		t.Fatalf("unexpected recipients: %+v", client.recipients)
	}
	if !bytes.Contains(client.writer.buffer.Bytes(), []byte("Confirm your email")) {
		t.Fatalf("rendered message does not contain subject: %q", client.writer.buffer.String())
	}
}

type fakeSMTPFactory struct {
	client smtpClient
}

func (f fakeSMTPFactory) DialContext(context.Context, string, string) (net.Conn, error) {
	return fakeConn{}, nil
}

func (f fakeSMTPFactory) NewClient(net.Conn, string) (smtpClient, error) {
	return f.client, nil
}

type fakeSMTPClient struct {
	writer         *fakeWriteCloser
	startTLS       bool
	extension      bool
	startTLSCalled bool
	authCalled     bool
	mailFrom       string
	recipients     []string
}

func (c *fakeSMTPClient) Extension(string) (bool, string) {
	return c.extension, ""
}

func (c *fakeSMTPClient) StartTLS(*tls.Config) error {
	c.startTLSCalled = true
	return nil
}

func (c *fakeSMTPClient) Auth(smtp.Auth) error {
	c.authCalled = true
	return nil
}

func (c *fakeSMTPClient) Mail(from string) error {
	c.mailFrom = from
	return nil
}

func (c *fakeSMTPClient) Rcpt(to string) error {
	c.recipients = append(c.recipients, to)
	return nil
}

func (c *fakeSMTPClient) Data() (io.WriteCloser, error) {
	return c.writer, nil
}

func (c *fakeSMTPClient) Quit() error {
	return nil
}

func (c *fakeSMTPClient) Close() error {
	return nil
}

type fakeWriteCloser struct {
	buffer bytes.Buffer
}

func (w *fakeWriteCloser) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}

func (w *fakeWriteCloser) Close() error {
	return nil
}

type fakeConn struct{}

func (fakeConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (fakeConn) Write(p []byte) (int, error)      { return len(p), nil }
func (fakeConn) Close() error                     { return nil }
func (fakeConn) LocalAddr() net.Addr              { return fakeAddr("local") }
func (fakeConn) RemoteAddr() net.Addr             { return fakeAddr("remote") }
func (fakeConn) SetDeadline(time.Time) error      { return nil }
func (fakeConn) SetReadDeadline(time.Time) error  { return nil }
func (fakeConn) SetWriteDeadline(time.Time) error { return nil }

type fakeAddr string

func (a fakeAddr) Network() string { return string(a) }
func (a fakeAddr) String() string  { return string(a) }
