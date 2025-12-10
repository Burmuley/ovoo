package rest

import (
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

// GetAliases retrieves a list of aliases based on optional filters for id and service name.
// It validates any provided IDs, fetches the aliases for the current user, and returns them
// in the HTTP response. Currently uses a temporary owner until authentication is implemented.
func (a *Application) GetAliases(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting aliases: identifying user", err)
		return
	}

	// filling filters
	filters := readFilters(r, []string{"owner", "id", "service_name", "email", "page_size", "page"})
	aliases, pgm, err := a.svcGw.Aliases.GetAll(r.Context(), cuser, filters)
	if err != nil {
		a.errorLogNResponse(w, "getting aliases", err)
		return
	}

	aliasData := make([]AliasData, 0, len(aliases))
	for _, alias := range aliases {
		aliasData = append(aliasData, addressTAliasData(alias))
	}

	resp := GetAliasesResponse{
		Aliases:            aliasData,
		PaginationMetadata: pgmTMetadata(pgm),
	}
	a.successResponse(w, resp, http.StatusOK)
}

// GetAliaseById retrieves an alias by its ID from the system. It validates the provided ID,
// fetches the corresponding alias details, and returns them in the HTTP response. If the alias
// is not found or there's an error, it returns an appropriate error response.
func (a *Application) GetAliaseById(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "geting alias: identifying user", err)
		return
	}

	aliasId := entities.Id(r.PathValue("id"))
	alias, err := a.svcGw.Aliases.GetById(r.Context(), cuser, entities.Id(aliasId))
	if err != nil {
		a.errorLogNResponse(w, "getting alias by id", err)
		return
	}

	resp := GetAliasDetailsResponse(addressTAliasData(alias))
	a.successResponse(w, resp, http.StatusOK)
}

// CreateAlias handles the creation of a new alias in the system. It processes the incoming HTTP
// request containing the protected address ID and metadata, creates a new alias associated with
// the current user (temporarily using first user until authentication is implemented), and returns
// the created alias details in the response.
func (a *Application) CreateAlias(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "creating alias: identifying user", err)
		return
	}

	req := CreateAliasRequest{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "parsing chain create request", err)
		return
	}

	alias, err := a.svcGw.Aliases.Create(r.Context(), cuser, services.AliasCreateCmd{
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment:     req.Metadata.Comment,
			ServiceName: req.Metadata.ServiceName,
		},
		ProtectedAddressId: req.ProtectedAddressId,
	})

	if err != nil {
		a.errorLogNResponse(w, "create alias", err)
		return
	}

	resp := CreateAliasResponse(addressTAliasData(alias))
	a.successResponse(w, resp, http.StatusCreated)
}

// UpdateAlias modifies an existing alias based on the provided update request.
// It can update the protected address ID and metadata of the alias.
// The function validates the alias ID, retrieves the current alias,
// applies the requested changes, and returns the updated alias in the response.
func (a *Application) UpdateAlias(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "creating alias: identifying user", err)
		return
	}

	aliasId := entities.Id(r.PathValue("id"))
	req := UpdateAliasRequest{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "parsing alias update request", err)
		return
	}

	alias, err := a.svcGw.Aliases.Update(r.Context(), cuser, services.AliasUpdateCmd{
		AliasId: aliasId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment:     req.Metadata.Comment,
			ServiceName: req.Metadata.ServiceName,
		},
	})
	if err != nil {
		a.errorLogNResponse(w, "updating alias", err)
		return
	}

	resp := UpdateAliasResponse(addressTAliasData(alias))
	a.successResponse(w, resp, http.StatusOK)
}

// DeleteAlias deletes an alias by its ID.
// It validates the alias ID, performs the deletion, and returns a no-content response on success.
func (a *Application) DeleteAlias(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "creating alias: identifying user", err)
		return
	}

	aliasId := entities.Id(r.PathValue("id"))
	if err := a.svcGw.Aliases.DeleteById(r.Context(), cuser, aliasId); err != nil {
		a.errorLogNResponse(w, "deleting alias by id", err)
		return
	}

	a.successResponse(w, "", http.StatusNoContent)
}
