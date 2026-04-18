package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err := app.Run(ctx); err != nil {
		stop()
		logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		logger.Error("backend stopped with error", "error", err)
		os.Exit(1)
	}

	stop()
}
