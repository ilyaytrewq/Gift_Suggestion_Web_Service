package email

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	platformemail "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/email"
)

type Notifier struct {
	sender          platformemail.Sender
	logger          *slog.Logger
	from            platformemail.Address
	frontendBaseURL string
}

func NewNotifier(
	sender platformemail.Sender,
	logger *slog.Logger,
	from platformemail.Address,
	frontendBaseURL string,
) (*Notifier, error) {
	if sender == nil {
		return nil, fmt.Errorf("email sender is required")
	}
	if err := from.Validate(); err != nil {
		return nil, err
	}

	return &Notifier{
		sender:          sender,
		logger:          logger,
		from:            from,
		frontendBaseURL: strings.TrimRight(frontendBaseURL, "/"),
	}, nil
}

func (n *Notifier) SendVerificationEmail(ctx context.Context, user *userdomain.User, rawToken string) error {
	message := verificationMessage(n.from, n.frontendBaseURL, user, rawToken)
	return n.sender.Send(ctx, message)
}

func (n *Notifier) SendPasswordResetEmail(ctx context.Context, user *userdomain.User, rawToken string) error {
	message := passwordResetMessage(n.from, n.frontendBaseURL, user, rawToken)
	return n.sender.Send(ctx, message)
}

func verificationMessage(
	from platformemail.Address,
	frontendBaseURL string,
	user *userdomain.User,
	rawToken string,
) platformemail.Message {
	displayName := recipientName(user)
	link := buildFrontendLink(frontendBaseURL, "/auth/verify-email", rawToken)

	textBody := fmt.Sprintf(
		"Hello %s,\n\nConfirm your email for Gift Suggestion Web Service.\nVerification code: %s\n",
		displayName,
		rawToken,
	)
	if link != "" {
		textBody += fmt.Sprintf("Verification link: %s\n", link)
	}
	textBody += "\nIf you did not create this account, you can ignore this email.\n"

	htmlBody := fmt.Sprintf(
		"<p>Hello %s,</p><p>Confirm your email for Gift Suggestion Web Service.</p><p><strong>Verification code:</strong> %s</p>",
		displayName,
		rawToken,
	)
	if link != "" {
		htmlBody += fmt.Sprintf(`<p><a href="%s">Confirm email</a></p>`, link)
	}
	htmlBody += "<p>If you did not create this account, you can ignore this email.</p>"

	return platformemail.Message{
		From:     from,
		To:       []platformemail.Address{{Email: user.Email().String(), Name: displayName}},
		Subject:  "Confirm your email",
		TextBody: textBody,
		HTMLBody: htmlBody,
	}
}

func passwordResetMessage(
	from platformemail.Address,
	frontendBaseURL string,
	user *userdomain.User,
	rawToken string,
) platformemail.Message {
	displayName := recipientName(user)
	link := buildFrontendLink(frontendBaseURL, "/auth/reset-password", rawToken)

	textBody := fmt.Sprintf(
		"Hello %s,\n\nWe received a password reset request for your Gift Suggestion Web Service account.\nReset code: %s\n",
		displayName,
		rawToken,
	)
	if link != "" {
		textBody += fmt.Sprintf("Reset link: %s\n", link)
	}
	textBody += "\nIf you did not request a password reset, you can ignore this email.\n"

	htmlBody := fmt.Sprintf(
		"<p>Hello %s,</p><p>We received a password reset request for your Gift Suggestion Web Service account.</p><p><strong>Reset code:</strong> %s</p>",
		displayName,
		rawToken,
	)
	if link != "" {
		htmlBody += fmt.Sprintf(`<p><a href="%s">Reset password</a></p>`, link)
	}
	htmlBody += "<p>If you did not request a password reset, you can ignore this email.</p>"

	return platformemail.Message{
		From:     from,
		To:       []platformemail.Address{{Email: user.Email().String(), Name: displayName}},
		Subject:  "Reset your password",
		TextBody: textBody,
		HTMLBody: htmlBody,
	}
}

func buildFrontendLink(baseURL, path, token string) string {
	if strings.TrimSpace(baseURL) == "" {
		return ""
	}

	values := url.Values{}
	values.Set("token", token)

	return fmt.Sprintf("%s%s?%s", baseURL, path, values.Encode())
}

func recipientName(user *userdomain.User) string {
	if user == nil {
		return "there"
	}
	if strings.TrimSpace(user.DisplayName()) != "" {
		return user.DisplayName()
	}

	return user.Email().String()
}
