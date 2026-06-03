package socketmap

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"sync"
)

type server struct {
	ctx      context.Context
	cancel   context.CancelFunc
	listener net.Listener
	wg       sync.WaitGroup
	logger   *slog.Logger
}

func newServer(network, addr string, logger *slog.Logger) (*server, error) {
	lnr, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &server{
		ctx:      ctx,
		cancel:   cancel,
		listener: lnr,
	}, nil
}

func (s *server) Wait(handler Handler) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// check if listener was closed intentionally
			if errors.Is(err, net.ErrClosed) {
				slog.Info("listener closed")
				return
			}
			slog.Error("accept error", "err", err.Error())
			continue
		}

		// track active connection
		s.wg.Add(1)
		go handle(s.ctx, &s.wg, conn, handler)
	}
}

func (s *server) Shutdown(ctx context.Context) error {
	err := s.listener.Close()
	s.cancel()

	// create a channel to wait for all connections to finish processing
	waitCh := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(waitCh)
	}()

	// enforce shutdown timeout limit
	select {
	case <-waitCh:
		log.Println("all connections closed cleanly")
		return err
	case <-ctx.Done():
		return errors.New("shutdown timed out; forcing exit")
	}
}
