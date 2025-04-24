package rest

import (
	"cmp"
	"context"
	"embed"
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Burmuley/ovoo/internal/applications"
	"github.com/Burmuley/ovoo/internal/applications/rest/middleware"
	"github.com/Burmuley/ovoo/internal/services"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	DefaultListenAddr string = "127.0.0.1:8808"
	loginURI                 = "/auth/login"
	callbackURI              = "/auth/callback"
)

//go:embed data
var staticData embed.FS

// Application represents the main structure for handling REST API requests.
// It contains references to various use cases, a listen address, context, and logger.
type Application struct {
	svcGw          *services.ServiceGateway
	listenAddr     string
	context        context.Context
	logger         *slog.Logger
	authSkipURIs   []string
	tls_cert       string
	tls_key        string
	providerConfig *middleware.OIDCProvider
}

// New creates and returns a new Controller instance.
// It initializes the controller with the provided listen address, logger, and use cases.
// If the listen address is empty, it uses the default address.
// It returns an error if any of the required use cases or the logger is nil.
func New(
	listenAddr string,
	logger *slog.Logger,
	svcGw *services.ServiceGateway,
	tls_key, tls_cert string,
	providerConfig map[string]any,
) (applications.Application, error) {
	ctrl := &Application{
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

	ctrl.authSkipURIs = []string{middleware.OIDCCallbackUri, middleware.OIDCLoginUri}

	{
		var err error
		if ctrl.providerConfig, err = parseProviderCfg(
			context.Background(), providerConfig,
		); err != nil {
			return nil, err
		}
		middleware.SetOIDCProvider(ctrl.providerConfig)
	}
	return ctrl, nil
}

// Start initializes and starts the HTTP server for the Ovoo API.
// It sets up routes for users, aliases, protected addresses, chains, and API tokens,
// registers authentication endpoints, applies middleware for security, logging, and authentication,
// and begins listening for incoming HTTPS requests with TLS.
// The server runs until the provided context is cancelled or an error occurs.
func (a *Application) Start(ctx context.Context) error {
	a.context = ctx
	mux := http.NewServeMux()

	// root
	mux.HandleFunc("/", a.handleRoot)

	// docs
	mux.HandleFunc("/api/docs/openapi.yaml", a.handleOpenAPI)
	mux.HandleFunc("/api/docs", a.handleDocs)

	// users routes
	mux.HandleFunc("GET /api/v1/users", a.GetUsers)
	mux.HandleFunc("GET /api/v1/users/{id}", a.GetUserById)
	mux.HandleFunc("POST /api/v1/users", a.CreateUser)
	mux.HandleFunc("PATCH /api/v1/users/{id}", a.UpdateUser)
	mux.HandleFunc("DELETE /api/v1/users", a.DeleteUser)

	// api tokens routes
	mux.HandleFunc("GET /api/v1/users/apitokens", a.GetApiTokens)
	mux.HandleFunc("GET /api/v1/users/apitokens/{id}", a.GetApiTokenById)
	mux.HandleFunc("POST /api/v1/users/apitokens", a.CreateApiToken)
	mux.HandleFunc("PATCH /api/v1/users/apitokens/{id}", a.UpdateApiToken)
	mux.HandleFunc("DELETE /api/v1/users/apitokens/{id}", a.DeleteApiToken)

	// aliases routes
	mux.HandleFunc("GET /api/v1/aliases", a.GetAliases)
	mux.HandleFunc("GET /api/v1/aliases/{id}", a.GetAliaseById)
	mux.HandleFunc("POST /api/v1/aliases", a.CreateAlias)
	mux.HandleFunc("PATCH /api/v1/alises/{id}", a.UpdateAlias)
	mux.HandleFunc("DELETE /api/v1/alias/{id}", a.DeleteAlias)

	// protected addresses routes
	mux.HandleFunc("GET /api/v1/praddrs", a.GetAllPrAddrs)
	mux.HandleFunc("GET /api/v1/praddrs/{id}", a.GetPrAddrById)
	mux.HandleFunc("POST /api/v1/praddrs", a.CreatePrAddr)
	mux.HandleFunc("PATCH /api/v1/praddrs/{id}", a.UpdatePrAddr)
	mux.HandleFunc("DELETE /api/v1/praddrs/{id}", a.DeletePrAddr)

	// chains routes
	mux.HandleFunc("GET /api/v1/chains/{hash}", a.getChainByHash)
	mux.HandleFunc("POST /api/v1/chains", a.CreateChain)
	mux.HandleFunc("DELETE /api/v1/chains/{hash}", a.DeleteChain)

	// authentication endpoints
	mux.HandleFunc(middleware.OIDCLoginUri, middleware.HandleOIDCLogin)
	mux.HandleFunc(middleware.OIDCCallbackUri, middleware.HandleOIDCCallback)

	handler := middleware.Adapt(mux,
		middleware.SecurityHeaders(),
		middleware.Logging(a.logger),
		middleware.Authentication(a.authSkipURIs, a.svcGw, a.logger),
	)
	a.logger.Info("started Ovoo API server", "addr", a.listenAddr)
	return http.ListenAndServeTLS(a.listenAddr, a.tls_cert, a.tls_key, handler)
}

// handleRoot serves the root page of the application.
// It parses and renders the login template, injecting the current user information if available.
// If there are any errors during template parsing or rendering, it will log the error
// and return an appropriate error response to the client.
//
// Parameters:
//   - w: The HTTP response writer to write the response to
//   - r: The HTTP request that triggered this handler
func (a *Application) handleRoot(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r)
	tmpl, err := template.New("index").ParseFS(staticData, "data/login/index.html")
	if err != nil {
		a.errorLogNResponse(w, "root page: parsing template", err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "index.html", user); err != nil {
		a.errorLogNResponse(w, "root page: rendering template", err)
	}
}

func (a *Application) handleDocs(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, staticData, "data/docs/index.html")
}

func (a *Application) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, staticData, "data/openapi.yaml")
}

// parseProviderCfg converts a generic configuration map into an OAuth2Provider struct.
// It extracts OAuth2 provider configuration details like client ID, client secret,
// authorization URL, token URL, user info URL, and issuer from the provided map.
// Returns a pointer to the configured OAuth2Provider and an error if the configuration
// format is invalid.
func parseProviderCfg(ctx context.Context, cfg map[string]any) (*middleware.OIDCProvider, error) {
	var err error

	p := &middleware.OIDCProvider{}
	p.Issuer = cfg["issuer"].(string)
	if p.OIDCProvider, err = oidc.NewProvider(ctx, p.Issuer); err != nil {
		return nil, err
	}

	p.OAuth2Config = &oauth2.Config{
		ClientID:     cfg["client_id"].(string),
		ClientSecret: cfg["client_secret"].(string),
		Endpoint:     p.OIDCProvider.Endpoint(),
		RedirectURL:  cfg["redirect_url"].(string),
		Scopes:       []string{"openid", "profile", "email"},
	}
	p.OIDCConfig = &oidc.Config{
		ClientID: cfg["client_id"].(string),
	}

	return p, nil
}
