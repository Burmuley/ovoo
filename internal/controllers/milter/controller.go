package milter

import (
	"context"
	"log"
	"log/slog"

	"github.com/d--j/go-milter/mailfilter"
)

type Controller struct {
	listenAddr string
	ovooCli    OvooClient
}

func New(listenAddr string, logger *slog.Logger, ovooCli OvooClient) (*Controller, error) {
	ctrl := &Controller{
		listenAddr: listenAddr,
		ovooCli:    ovooCli,
	}

	return ctrl, nil
}

func (m *Controller) Start(ctx context.Context) error {
	server, err := mailfilter.New(
		"tcp",
		"127.0.0.1:6785",
		AddressRewriter(m.ovooCli),
		mailfilter.WithDecisionAt(mailfilter.DecisionAtEndOfHeaders),
	)

	if err != nil {
		return err
	}

	log.Printf("Started Ovoo milter on %s:%s", server.Addr().Network(), server.Addr().String())

	// quit when milter quits
	server.Wait()
	return nil
}
