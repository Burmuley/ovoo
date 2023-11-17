package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route is the information for every URI.
type Route struct {
	Name        string          // Name is the name of this Route.
	Method      string          // Method is the string for the HTTP method. ex) GET, POST etc..
	Pattern     string          // Pattern is the pattern of the URI.
	HandlerFunc gin.HandlerFunc // HandlerFunc is the handler function of this route.
}

// Routes is the list of the generated Route.
type Routes []Route

// NewRouter returns a new router.
func NewRouter(a *Api) *gin.Engine {
	router := gin.Default()
	routes := routes(a)
	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			router.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			router.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			router.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			router.DELETE(route.Pattern, route.HandlerFunc)
		}
	}

	return router
}

// Index is the index handler.
func Index(c *gin.Context) {
	c.String(http.StatusOK, "Just and index page. See API documentation for other endpoints.")
}

func routes(a *Api) Routes {
	var routes = Routes{
		{
			"Index",
			http.MethodGet,
			"/",
			Index,
		},
		// Aliases prefix
		{
			"AliasesGet",
			http.MethodGet,
			"/aliases",
			a.AliasesGet,
		},

		{
			"AliasesPost",
			http.MethodPost,
			"/aliases",
			a.AliasesPost,
		},

		{
			"AliasesGetByEmail",
			http.MethodGet,
			"/aliases/email/:email",
			a.AliasesGetByEmail,
		},

		{
			"AliasesDeleteByEmail",
			http.MethodDelete,
			"/aliases/email/:email",
			a.AliasesDeleteByEmail,
		},

		{
			"AliasesGetById",
			http.MethodGet,
			"/aliases/:id",
			a.AliasesGetById,
		},

		{
			"AliasesDeleteById",
			http.MethodDelete,
			"/aliases/:id",
			a.AliasesDeleteById,
		},

		// ProtectedAddresses prefix
		{
			"ProtectedAddressesGet",
			http.MethodGet,
			"/protected-addresses",
			a.ProtectedAddressesGet,
		},

		{
			"ProtectedAddressesPost",
			http.MethodPost,
			"/protected-addresses",
			a.ProtectedAddressesPost,
		},

		{
			"ProtectedAddressesGetByEmail",
			http.MethodGet,
			"/protected-addresses/email/:email",
			a.ProtectedAddressesGetByEmail,
		},

		{
			"ProtectedAddressesDeleteByEmail",
			http.MethodDelete,
			"/protected-addresses/email/:email",
			a.ProtectedAddressesDeleteByEmail,
		},

		{
			"ProtectedAddressesGetById",
			http.MethodGet,
			"/protected-addresses/:id",
			a.ProtectedAddressesGetById,
		},

		{
			"ProtectedAddressesDeleteById",
			http.MethodDelete,
			"/protected-addresses/:id",
			a.ProtectedAddressesDeleteById,
		},
	}

	return routes
}
