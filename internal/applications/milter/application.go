package milter

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/d--j/go-milter/mailfilter"
)

const DefaultListenAddr string = "127.0.0.1:6785"

type Application struct {
	listenAddr string
	ovooCli    OvooClient
	logger     *slog.Logger
}

func New(listenAddr string, logger *slog.Logger, ovooCli OvooClient) (*Application, error) {
	ctrl := &Application{
		listenAddr: listenAddr,
		ovooCli:    ovooCli,
		logger:     logger,
	}

	return ctrl, nil
}

func (m *Application) Start() error {
	// global context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server, err := mailfilter.New(
		"tcp",
		m.listenAddr,
		AddressRewriter(m.ovooCli),
		mailfilter.WithDecisionAt(mailfilter.DecisionAtEndOfMessage),
		mailfilter.WithErrorHandling(mailfilter.RejectWhenError),
	)

	if err != nil {
		return err
	}

	go func() {
		m.logger.Info("starting Ovoo Milter server", server.Addr().Network(), server.Addr().String())
		server.Wait()
		stop()
	}()

	<-ctx.Done()
	m.logger.Info("shutting down Ovoo Milter server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		m.logger.Error("Ovoo Milter server shutdown failed", "err", err.Error())
		return err
	}
	server.Close()
	return nil
}
