package milter

import (
	"context"
	"log/slog"

	"github.com/d--j/go-milter/mailfilter"
)

const DefaultListenAddr string = "127.0.0.1:6785"

type Controller struct {
	listenAddr string
	ovooCli    OvooClient
	logger     *slog.Logger
}

func New(listenAddr string, logger *slog.Logger, ovooCli OvooClient) (*Controller, error) {
	ctrl := &Controller{
		listenAddr: listenAddr,
		ovooCli:    ovooCli,
		logger:     logger,
	}

	return ctrl, nil
}

func (m *Controller) Start(ctx context.Context) error {
	server, err := mailfilter.New(
		"tcp",
		m.listenAddr,
		AddressRewriter(m.ovooCli),
		mailfilter.WithDecisionAt(mailfilter.DecisionAtEndOfHeaders),
		mailfilter.WithoutBody(), mailfilter.WithErrorHandling(mailfilter.RejectWhenError),
	)

	if err != nil {
		return err
	}

	m.logger.Info("started Ovoo Milter on %s:%s", server.Addr().Network(), server.Addr().String())

	// quit when milter quits
	server.Wait()
	return nil
}
