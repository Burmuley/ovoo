package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Burmuley/ovoo/internal/controllers/milter"
)

func main() {
	// logger configuration
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	))
	slog.SetDefault(logger)
	ovooClient := milter.NewOvooClient("http://127.0.0.1:8808")
	ctrl, _ := milter.New("127.0.0.1:8825", logger, ovooClient)
	ctrl.Start(context.Background())
}
