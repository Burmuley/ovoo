package socketmap

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Burmuley/ovoo/internal/applications/ovooclient"
)

const (
	DefaultSocketmapNetwork = "unix"
	DefaultSocketmapAddr    = "/tmp/ovoo_socketmap.sock"
)

type Application struct {
	network string
	addr    string
	cli     ovooclient.Client
	logger  *slog.Logger
	handler Handler
}

func New(network, listenAddr string, logger *slog.Logger, ovooCli ovooclient.Client) (*Application, error) {
	ctrl := &Application{
		network: network,
		addr:    listenAddr,
		cli:     ovooCli,
		logger:  logger,
		handler: handler,
	}

	return ctrl, nil
}

func (m *Application) Start() error {
	// global context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv, err := newServer(m.network, m.addr, m.logger)
	if err != nil {
		return err
	}

	go func() {
		m.logger.Info("starting Ovoo Socketmap server", m.network, m.addr)
		srv.Wait(m.handler)
		stop()
	}()

	<-ctx.Done()
	m.logger.Info("shutting down Ovoo Socketmap server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		m.logger.Error("Ovoo Socketmap server shutdown failed", "err", err.Error())
		return err
	}

	return nil
}

func handler(ctx context.Context, lookup, key string) (result string, found bool, err error) {
	_ = ctx
	slog.Info("sockmap handler", "lookup", lookup, "key", key)
	return "YES", true, nil
}
