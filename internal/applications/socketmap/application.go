package socketmap

import (
	"context"
	"log/slog"
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
}

func New(network, listenAddr string, logger *slog.Logger, ovooCli ovooclient.Client) (*Application, error) {
	ctrl := &Application{
		network: network,
		addr:    listenAddr,
		cli:     ovooCli,
		logger:  logger,
	}

	return ctrl, nil
}

func (m *Application) Start() error {
	// global context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	// defer stop()

	srv, err := newServer(m.network, m.addr, m.logger)
	if err != nil {
		return err
	}

	go func() {
		m.logger.Info("starting Ovoo Socketmap server", m.network, m.addr)
		srv.Wait(ovooHandler(m.cli))
		stop()
	}()

	<-ctx.Done()
	m.logger.Info("shutting down Ovoo Socketmap server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		m.logger.Error("server shutdown failed", "err", err.Error())
		return err
	}

	return nil
}

func ovooHandler(cli ovooclient.Client) func(ctx context.Context, lookup, key string) (result string, found bool, err error) {
	return func(ctx context.Context, lookup, key string) (result string, found bool, err error) {
		slog.Info("handler call: lookup=" + lookup + " key=" + key)
		switch lookup {
		case "relay_domain":
			hasDomain := cli.GetDomainByName(ctx, key)
			if hasDomain {
				return key, true, nil
			}

			return "", false, nil
		}

		return "", false, PermanentError{Reason: "unknown lookup " + lookup}
	}
}
