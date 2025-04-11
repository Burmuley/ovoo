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
func (c *Application) GetAliases(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		c.errorLogNResponse(w, "getting aliases: identifying user", err)
	}

	filters := map[string][]string{"owner": []string{user.ID.String()}}
	if ids, ok := r.URL.Query()["id"]; ok {
		filters["id"] = ids
		for _, id := range ids {
			if err := entities.Id(id).Validate(); err != nil {
				c.errorLogNResponse(w, "getting aliases: parsing id", err)
				return
			}
		}
	}

	if service_names, ok := r.URL.Query()["service"]; ok {
		filters["service"] = service_names
	}

	aliases, err := c.svcGw.Aliases.GetAll(c.context, filters)
	if err != nil {
		c.errorLogNResponse(w, "getting aliases", err)
		return
	}
	aliasData := make([]AliasData, 0, len(aliases))
	for _, alias := range aliases {
		aliasData = append(aliasData, addressTAliasData(alias))
	}
	resp := GetAliasesResponse(aliasData)
	c.successResponse(w, resp, http.StatusOK)
}

// GetAliaseById retrieves an alias by its ID from the system. It validates the provided ID,
// fetches the corresponding alias details, and returns them in the HTTP response. If the alias
// is not found or there's an error, it returns an appropriate error response.
func (c *Application) GetAliaseById(w http.ResponseWriter, r *http.Request) {
	aliasId := entities.Id(r.PathValue("id"))
	if err := aliasId.Validate(); err != nil {
		c.errorLogNResponse(w, "getting alias by id: parsing id", err)
		return
	}

	alias, err := c.svcGw.Aliases.GetById(c.context, entities.Id(aliasId))
	if err != nil {
		c.errorLogNResponse(w, "getting alias by id", err)
		return
	}

	resp := GetAliasDetailsResponse(addressTAliasData(alias))
	c.successResponse(w, resp, http.StatusOK)
}

// CreateAlias handles the creation of a new alias in the system. It processes the incoming HTTP
// request containing the protected address ID and metadata, creates a new alias associated with
// the current user (temporarily using first user until authentication is implemented), and returns
// the created alias details in the response.
func (c *Application) CreateAlias(w http.ResponseWriter, r *http.Request) {
	rb := CreateAliasRequest{}
	if err := readBody(r.Body, &rb); err != nil {
		c.errorLogNResponse(w, "parsing chain create request", err)
		return
	}

	prot_addr, err := c.svcGw.PrAddrs.GetById(c.context, entities.Id(rb.ProtectedAddressId))
	if err != nil {
		c.errorLogNResponse(w, "getting new alias forward address", err)
		return
	}

	user, err := userFromContext(r)
	if err != nil {
		c.errorLogNResponse(w, "getting aliases: identifying user", err)
	}

	// TODO: get real user from auth info
	alias, err := c.svcGw.Aliases.Create(c.context, prot_addr,
		entities.AddressMetadata{
			Comment:     rb.Metadata.Comment,
			ServiceName: rb.Metadata.ServiceName,
		},
		user,
	)

	if err != nil {
		c.errorLogNResponse(w, "creating alias", err)
		return
	}

	resp := CreateAliasResponse(addressTAliasData(alias))
	c.successResponse(w, resp, http.StatusCreated)
}

// UpdateAlias modifies an existing alias based on the provided update request.
// It can update the protected address ID and metadata of the alias.
// The function validates the alias ID, retrieves the current alias,
// applies the requested changes, and returns the updated alias in the response.
func (c *Application) UpdateAlias(w http.ResponseWriter, r *http.Request) {
	aliasId := entities.Id(r.PathValue("id"))
	if err := aliasId.Validate(); err != nil {
		c.errorLogNResponse(w, "getting alias by id: parsing id", err)
		return
	}

	alias, err := c.svcGw.Aliases.GetById(c.context, aliasId)
	if err != nil {
		c.errorLogNResponse(w, "updating alias: retrieving alias by id", err)
		return
	}

	rraw, err := io.ReadAll(r.Body)
	if err != nil {
		c.errorLogNResponse(w, "reading alias update request", err)
		return
	}

	rb := UpdateAliasRequest{}
	if err := json.Unmarshal(rraw, &rb); err != nil {
		c.errorLogNResponse(w, "parsing alias update request", err)
		return
	}

	if rb.ProtectedAddressId != nil {
		newFwd, err := c.svcGw.PrAddrs.GetById(c.context, entities.Id(*rb.ProtectedAddressId))
		if err != nil {
			c.errorLogNResponse(w, "updating alias: retrieving protected address", err)
			return
		}
		alias.ForwardAddress = &newFwd
	}

	if rb.Metadata != nil {
		alias.Metadata = entities.AddressMetadata(*rb.Metadata)
	}

	alias, err = c.svcGw.Aliases.Update(c.context, alias)
	if err != nil {
		c.errorLogNResponse(w, "updating alias", err)
		return
	}

	resp := UpdateAliasResponse(addressTAliasData(alias))
	c.successResponse(w, resp, http.StatusOK)
}

// DeleteAlias deletes an alias by its ID.
// It validates the alias ID, performs the deletion, and returns a no-content response on success.
func (c *Application) DeleteAlias(w http.ResponseWriter, r *http.Request) {
	aliasId := entities.Id(r.PathValue("id"))
	if err := aliasId.Validate(); err != nil {
		c.errorLogNResponse(w, "getting alias by id: parsing id", err)
		return
	}

	if err := c.svcGw.Aliases.DeleteById(c.context, aliasId); err != nil {
		c.errorLogNResponse(w, "getting alias by id: parsing id", err)
		return
	}

	c.successResponse(w, "", http.StatusNoContent)
}
