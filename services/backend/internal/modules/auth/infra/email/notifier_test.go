package email

import (
	"context"
	"strings"
	"testing"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	platformemail "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/email"
)

func TestNotifierSendVerificationEmailBuildsFrontendLink(t *testing.T) {
	t.Parallel()

	sender := &captureSender{}
	notifier, err := NewNotifier(
		sender,
		nil,
		platformemail.Address{Email: "noreply@example.com", Name: "Gift Suggestion"},
		"http://localhost:5173",
	)
	if err != nil {
		t.Fatalf("NewNotifier() error = %v", err)
	}

	user := mustNotifierUser(t)
	if err := notifier.SendVerificationEmail(context.Background(), user, "verify-token", ""); err != nil {
		t.Fatalf("SendVerificationEmail() error = %v", err)
	}

	if sender.message.Subject != "Confirm your email" {
		t.Fatalf("subject = %q, want %q", sender.message.Subject, "Confirm your email")
	}
	if !strings.Contains(sender.message.TextBody, "verify-token") {
		t.Fatalf("text body does not contain token: %q", sender.message.TextBody)
	}
	if !strings.Contains(sender.message.TextBody, "http://localhost:5173/auth/email-verify?token=verify-token") {
		t.Fatalf("text body does not contain verification link: %q", sender.message.TextBody)
	}
}

func TestNotifierSendPasswordResetEmailWithoutFrontendLink(t *testing.T) {
	t.Parallel()

	sender := &captureSender{}
	notifier, err := NewNotifier(
		sender,
		nil,
		platformemail.Address{Email: "noreply@example.com"},
		"",
	)
	if err != nil {
		t.Fatalf("NewNotifier() error = %v", err)
	}

	user := mustNotifierUser(t)
	if err := notifier.SendPasswordResetEmail(context.Background(), user, "reset-token", ""); err != nil {
		t.Fatalf("SendPasswordResetEmail() error = %v", err)
	}

	if strings.Contains(sender.message.TextBody, "/password-reset/confirm") {
		t.Fatalf("unexpected frontend link in text body: %q", sender.message.TextBody)
	}
	if strings.Contains(sender.message.TextBody, "reset-token") {
		t.Fatalf("raw token must not appear in email body: %q", sender.message.TextBody)
	}
}

func TestNotifierSendPasswordResetEmailContainsLink(t *testing.T) {
	t.Parallel()

	sender := &captureSender{}
	notifier, err := NewNotifier(
		sender,
		nil,
		platformemail.Address{Email: "noreply@example.com", Name: "Gift Suggestion"},
		"http://localhost:5173",
	)
	if err != nil {
		t.Fatalf("NewNotifier() error = %v", err)
	}

	user := mustNotifierUser(t)
	if err := notifier.SendPasswordResetEmail(context.Background(), user, "reset-token", ""); err != nil {
		t.Fatalf("SendPasswordResetEmail() error = %v", err)
	}

	if !strings.Contains(sender.message.TextBody, "http://localhost:5173/password-reset/confirm?token=reset-token") {
		t.Fatalf("text body does not contain reset link: %q", sender.message.TextBody)
	}
	if strings.Contains(sender.message.TextBody, "Reset code") {
		t.Fatalf("raw token code must not appear in text body: %q", sender.message.TextBody)
	}
}

type captureSender struct {
	message platformemail.Message
}

func (s *captureSender) Send(_ context.Context, msg platformemail.Message) error {
	s.message = msg
	return nil
}

func mustNotifierUser(t *testing.T) *userdomain.User {
	t.Helper()

	user, err := userdomain.NewUser(
		"550e8400-e29b-41d4-a716-446655440000",
		"user@example.com",
		"ValidPass1!",
		string(userdomain.UserRoleUser),
	)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}
	if err := user.UpdateDisplayName("Alice", user.CreatedAt()); err != nil {
		t.Fatalf("UpdateDisplayName() error = %v", err)
	}

	return &user
}
