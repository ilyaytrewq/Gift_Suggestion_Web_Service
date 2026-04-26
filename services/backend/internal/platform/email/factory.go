package email

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

func NewSender(cfg config.EmailConfig, logger *slog.Logger) (Sender, error) {
	if !cfg.Enabled || strings.EqualFold(cfg.Provider, "noop") {
		return NoopSender{logger: logger}, nil
	}

	switch strings.ToLower(cfg.Provider) {
	case "smtp":
		sender, err := NewSMTPSender(cfg)
		if err != nil {
			return nil, err
		}

		return timeoutSender{
			sender:  sender,
			timeout: cfg.SendTimeout,
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedSender, cfg.Provider)
	}
}

type timeoutSender struct {
	sender  Sender
	timeout time.Duration
}

func (s timeoutSender) Send(ctx context.Context, msg Message) error {
	if s.timeout <= 0 {
		return s.sender.Send(ctx, msg)
	}

	sendCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.sender.Send(sendCtx, msg)
}
