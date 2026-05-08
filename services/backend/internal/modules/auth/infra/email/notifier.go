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

func (n *Notifier) SendVerificationEmail(ctx context.Context, user *userdomain.User, rawToken, frontendBaseURL string) error {
	baseURL := n.resolveBaseURL(frontendBaseURL)
	message := verificationMessage(n.from, baseURL, user, rawToken)
	return n.sender.Send(ctx, message)
}

func (n *Notifier) SendPasswordResetEmail(ctx context.Context, user *userdomain.User, rawToken, frontendBaseURL string) error {
	baseURL := n.resolveBaseURL(frontendBaseURL)
	message := passwordResetMessage(n.from, baseURL, user, rawToken)
	return n.sender.Send(ctx, message)
}

func (n *Notifier) resolveBaseURL(perCallURL string) string {
	if strings.TrimSpace(perCallURL) != "" {
		return strings.TrimRight(perCallURL, "/")
	}
	return n.frontendBaseURL
}

func verificationMessage(
	from platformemail.Address,
	frontendBaseURL string,
	user *userdomain.User,
	rawToken string,
) platformemail.Message {
	displayName := recipientName(user)
	link := buildFrontendLink(frontendBaseURL, "/auth/email-verify", rawToken)

	textBody := fmt.Sprintf(
		"Hello %s,\n\nConfirm your email for Gift Suggestion Web Service.\n\n"+
			"Verification code (same as in the link): %s\n",
		displayName,
		rawToken,
	)
	if link != "" {
		textBody += fmt.Sprintf("\nOr open the verification link:\n%s\n", link)
	}
	textBody += "\nIf you did not create this account, you can ignore this email.\n"

	htmlBody := fmt.Sprintf(
		"<p>Hello %s,</p><p>Confirm your email for Gift Suggestion Web Service.</p>"+
			`<p>Verification code: <code style="word-break:break-all;">%s</code></p>`,
		displayName,
		rawToken,
	)
	if link != "" {
		htmlBody += fmt.Sprintf(`<p><a href="%s">Confirm email (open link)</a></p>`, link)
	}
	htmlBody += "<p>You can paste the code on the confirmation page if the link does not work.</p>"
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
	link := buildFrontendLink(frontendBaseURL, "/password-reset/confirm", rawToken)

	textBody := fmt.Sprintf(
		"Hello %s,\n\nWe received a password reset request for your Gift Suggestion Web Service account.\n",
		displayName,
	)
	if link != "" {
		textBody += fmt.Sprintf("Reset your password here: %s\n", link)
	}
	textBody += "\nIf you did not request a password reset, you can ignore this email.\n"

	htmlBody := fmt.Sprintf(
		"<p>Hello %s,</p><p>We received a password reset request for your Gift Suggestion Web Service account.</p>",
		displayName,
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
