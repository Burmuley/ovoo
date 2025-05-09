package rest

import (
	"cmp"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Burmuley/ovoo/internal/applications"
	"github.com/Burmuley/ovoo/internal/applications/rest/middleware"
	"github.com/Burmuley/ovoo/internal/entities"
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
// It contains references to a service gateway for business logic, network configuration,
// application context, logging, authentication settings, and OIDC provider configurations.
type Application struct {
	svcGw           *services.ServiceGateway
	listenAddr      string
	context         context.Context
	logger          *slog.Logger
	authSkipURIs    []string
	tls_cert        string
	tls_key         string
	providerConfigs map[string]middleware.OIDCProvider
}

// New creates and returns a new Application instance configured for REST API handling.
// It validates and initializes all required components including services, logging, and auth providers.
//
// Parameters:
//   - listenAddr: Network address to listen on (uses DefaultListenAddr if empty)
//   - logger: Structured logger for application logging
//   - svcGw: Service gateway containing business logic implementations
//   - tls_key: Path to TLS private key file
//   - tls_cert: Path to TLS certificate file
//   - providersConfig: Map of OIDC provider configurations
//
// Returns:
//   - applications.Application: Configured application instance
//   - error: Non-nil if initialization fails
func New(
	listenAddr string,
	logger *slog.Logger,
	svcGw *services.ServiceGateway,
	tls_key, tls_cert string,
	providersConfig map[string]any,
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

	ctrl.authSkipURIs = []string{}

	{
		var err error
		if ctrl.providerConfigs, err = parseProvidersCfg(providersConfig); err != nil {
			return nil, err
		}
		middleware.SetOIDCConfigs(ctrl.providerConfigs)
	}

	if err := middleware.SetLogger(logger); err != nil {
		return nil, err
	}

	return ctrl, nil
}

// Start initializes and starts the HTTP server for the Ovoo API.
// It sets up all API routes, middleware chains, and starts the HTTPS server.
//
// The server provides endpoints for:
// - User management (/api/v1/users/*)
// - API token management (/api/v1/users/apitokens/*)
// - Alias management (/api/v1/aliases/*)
// - Protected address management (/api/v1/praddrs/*)
// - Chain management (/api/v1/chains/*)
// - Authentication flows
// - API documentation
//
// Parameters:
//   - ctx: Context for server lifecycle management
//
// Returns:
//   - error: Non-nil if server fails to start or encounters fatal error
func (a *Application) Start(ctx context.Context) error {
	a.context = ctx
	mux := http.NewServeMux()

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
	mux.HandleFunc("DELETE /api/v1/aliases/{id}", a.DeleteAlias)

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

	// root
	mux.HandleFunc("/{$}", a.handleRoot([]string(mapKeys(a.providerConfigs))))

	handler := middleware.Adapt(mux,
		middleware.SecurityHeaders(),
		middleware.Logging(a.logger),
		middleware.Authentication(a.authSkipURIs, a.svcGw),
	)
	a.logger.Info("started Ovoo API server", "addr", a.listenAddr)
	return http.ListenAndServeTLS(a.listenAddr, a.tls_cert, a.tls_key, handler)
}

// handleRoot serves the root page of the application.
// It parses and renders the login template with current user and provider information.
//
// Parameters:
//   - providers: List of configured authentication providers to display
//
// Returns:
//   - http.HandlerFunc: Handler that renders the root/login page
func (a *Application) handleRoot(providers []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := userFromContext(r)
		tmpl, err := template.New("index").ParseFS(staticData, "data/login/index.html")
		if err != nil {
			a.errorLogNResponse(w, "root page: parsing template", err)
			return
		}

		if err := tmpl.ExecuteTemplate(
			w,
			"index.html",
			struct {
				User      entities.User
				Providers []string
			}{User: user, Providers: providers},
		); err != nil {
			a.errorLogNResponse(w, "root page: rendering template", err)
		}
	}
}

// handleDocs serves the API documentation HTML page
func (a *Application) handleDocs(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, staticData, "data/docs/index.html")
}

// handleOpenAPI serves the OpenAPI specification file
func (a *Application) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, staticData, "data/openapi.yaml")
}

// parseProvidersCfg initializes OIDC providers from configuration.
// It creates provider instances with OAuth2 and OIDC configurations.
//
// Parameters:
//   - cfg: Map of provider names to their raw configurations
//
// Returns:
//   - map[string]middleware.OIDCProvider: Configured providers mapped by name
//   - error: Non-nil if provider initialization fails
func parseProvidersCfg(cfg map[string]any) (map[string]middleware.OIDCProvider, error) {
	providers := make(map[string]middleware.OIDCProvider)
	for name, config := range cfg {
		mapCfg := config.(map[string]any)
		var err error
		p := middleware.OIDCProvider{}
		p.Issuer = mapCfg["issuer"].(string)
		if p.OIDCProvider, err = oidc.NewProvider(context.Background(), p.Issuer); err != nil {
			return nil, err
		}

		p.OAuth2Config = &oauth2.Config{
			ClientID:     mapCfg["client_id"].(string),
			ClientSecret: mapCfg["client_secret"].(string),
			Endpoint:     p.OIDCProvider.Endpoint(),
			RedirectURL:  fmt.Sprintf("/auth/%s/callback", name),
			Scopes:       []string{"openid", "profile", "email"},
		}
		p.OIDCConfig = &oidc.Config{
			ClientID: mapCfg["client_id"].(string),
		}

		providers[name] = p
	}

	return providers, nil
}
