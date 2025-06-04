package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
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

	aliases, pgm, err := a.svcGw.Aliases.GetAll(a.context, cuser, filters)
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
	alias, err := a.svcGw.Aliases.GetById(a.context, cuser, entities.Id(aliasId))
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

	rb := CreateAliasRequest{}
	if err := readBody(r.Body, &rb); err != nil {
		a.errorLogNResponse(w, "parsing chain create request", err)
		return
	}

	prot_addr, err := a.svcGw.PrAddrs.GetById(a.context, cuser, entities.Id(rb.ProtectedAddressId))
	if err != nil {
		a.errorLogNResponse(w, "getting new alias forward address", err)
		return
	}

	alias, err := a.svcGw.Aliases.Create(a.context, cuser, prot_addr,
		entities.AddressMetadata{
			Comment:     rb.Metadata.Comment,
			ServiceName: rb.Metadata.ServiceName,
		},
	)

	if err != nil {
		a.errorLogNResponse(w, "creating alias", err)
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
	alias, err := a.svcGw.Aliases.GetById(a.context, cuser, aliasId)
	if err != nil {
		a.errorLogNResponse(w, "updating alias: retrieving alias by id", err)
		return
	}

	rraw, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorLogNResponse(w, "reading alias update request", err)
		return
	}

	rb := UpdateAliasRequest{}
	if err := json.Unmarshal(rraw, &rb); err != nil {
		a.errorLogNResponse(w, "parsing alias update request", err)
		return
	}

	if rb.ProtectedAddressId != nil {
		newFwd, err := a.svcGw.PrAddrs.GetById(a.context, cuser, entities.Id(*rb.ProtectedAddressId))
		if err != nil {
			a.errorLogNResponse(w, "updating alias: retrieving protected address", err)
			return
		}
		alias.ForwardAddress = &newFwd
	}

	if rb.Metadata != nil {
		alias.Metadata = entities.AddressMetadata(*rb.Metadata)
	}

	alias, err = a.svcGw.Aliases.Update(a.context, cuser, alias)
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
	if err := a.svcGw.Aliases.DeleteById(a.context, cuser, aliasId); err != nil {
		a.errorLogNResponse(w, "deleting alias by id", err)
		return
	}

	a.successResponse(w, "", http.StatusNoContent)
}
