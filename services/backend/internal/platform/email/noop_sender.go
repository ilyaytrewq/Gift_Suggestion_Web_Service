package email

import (
	"context"
	"log/slog"
)

type NoopSender struct {
	logger *slog.Logger
}

func (s NoopSender) Send(_ context.Context, msg Message) error {
	if err := msg.Validate(); err != nil {
		return err
	}
	if s.logger != nil {
		s.logger.Info(
			"email delivery skipped by noop sender",
			"subject",
			msg.Subject,
			"recipients",
			len(msg.To),
		)
	}

	return nil
}
