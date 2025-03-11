package rest

import (
	"cmp"
	"context"
	"embed"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Burmuley/ovoo/internal/controllers"
	"github.com/Burmuley/ovoo/internal/controllers/rest/middleware"
	"github.com/Burmuley/ovoo/internal/services"
)

const DefaultListenAddr string = "127.0.0.1:8808"

//go:embed data/login/index.html
var loginStatic embed.FS

// Controller represents the main structure for handling REST API requests.
// It contains references to various use cases, a listen address, context, and logger.
type Controller struct {
	svcGw        *services.ServiceGateway
	listenAddr   string
	context      context.Context
	logger       *slog.Logger
	authSkipURIs []string
	tls_cert     string
	tls_key      string
}

// New creates and returns a new Controller instance.
// It initializes the controller with the provided listen address, logger, and use cases.
// If the listen address is empty, it uses the default address.
// It returns an error if any of the required use cases or the logger is nil.
func New(listenAddr string, logger *slog.Logger, svcGw *services.ServiceGateway, tls_key, tls_cert string) (controllers.Controller, error) {
	ctrl := &Controller{
		svcGw:      svcGw,
		listenAddr: listenAddr,
		logger:     logger,
		tls_key:    tls_key,
		tls_cert:   tls_cert,
	}

	if len(listenAddr) < 1 {
		ctrl.listenAddr = DefaultListenAddr
	}

	// if any of "validation" values above true -> one of usecases is nil -> error
	if cmp.Or([]bool{
		ctrl.svcGw.Aliases == nil,
		ctrl.svcGw.Users == nil,
		ctrl.svcGw.PrAddrs == nil,
		ctrl.svcGw.Chains == nil,
	}...) {
		return nil, errors.New("all services should be set in service gateway")
	}

	if ctrl.logger == nil {
		return nil, errors.New("logger must be set")
	}

	ctrl.authSkipURIs = []string{
		"/api/v1/users/login",
	}

	return ctrl, nil
}

// Start initializes and starts the HTTP server for the Ovoo API.
// It sets up routes for users, aliases, and protected addresses,
// applies logging middleware, and begins listening for incoming requests.
// The server runs until the provided context is cancelled or an error occurs.
func (c *Controller) Start(ctx context.Context) error {
	c.context = ctx
	mux := http.NewServeMux()
	// users routes
	mux.HandleFunc("GET /api/v1/users", c.GetUsers)
	mux.HandleFunc("GET /api/v1/users/{id}", c.GetUserById)
	mux.HandleFunc("POST /api/v1/users", c.CreateUser)
	mux.HandleFunc("PUT /api/v1/users/{id}", c.UpdateUser)
	mux.HandleFunc("DELETE /api/v1/users", c.DeleteUser)

	// aliases routes
	mux.HandleFunc("GET /api/v1/aliases", c.GetAliases)
	mux.HandleFunc("GET /api/v1/aliases/{id}", c.GetAliaseById)
	mux.HandleFunc("POST /api/v1/aliases", c.CreateAlias)
	mux.HandleFunc("PUT /api/v1/alises/{id}", c.UpdateAlias)
	mux.HandleFunc("DELETE /api/v1/alias/{id}", c.DeleteAlias)

	// protected addresses routes
	mux.HandleFunc("GET /api/v1/praddrs", c.GetAllPrAddrs)
	mux.HandleFunc("GET /api/v1/praddrs/{id}", c.GetPrAddrById)
	mux.HandleFunc("POST /api/v1/praddrs", c.CreatePrAddr)
	mux.HandleFunc("PUT /api/v1/praddrs/{id}", c.UpdatePrAddr)
	mux.HandleFunc("DELETE /api/v1/praddrs/{id}", c.DeletePrAddr)

	// chains routes
	mux.HandleFunc("GET /api/v1/chains/{hash}", c.getChainByHash)
	mux.HandleFunc("POST /api/v1/chains", c.CreateChain)
	mux.HandleFunc("DELETE /api/v1/chains/{hash}", c.DeleteChain)

	// authentication endpoints
	mux.HandleFunc("GET /api/v1/users/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, loginStatic, "data/login/index.html")
	}))

	handler := middleware.Adapt(mux,
		middleware.SecurityHeaders(),
		middleware.Logging(c.logger),
		middleware.Authentication([]string{"/api/v1/users/login"}, c.svcGw, c.logger),
	)
	c.logger.Info("started Ovoo API server", "addr", c.listenAddr)
	return http.ListenAndServeTLS(c.listenAddr, c.tls_cert, c.tls_key, handler)
}
