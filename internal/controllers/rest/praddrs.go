package rest

import (
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
)

// GetAllPrAddrs retrieves all protected addresses for the current user.
// Supports filtering by ID and email through query parameters.
func (c *Controller) GetAllPrAddrs(w http.ResponseWriter, r *http.Request) {
	// TODO: get real user when authentication is enabled
	user, err := userFromContext(r)
	if err != nil {
		c.errorLogNResponse(w, "getting aliases: identifying user", err)
	}
	filters := map[string][]string{"owner": []string{user.ID.String()}}
	if ids, ok := r.URL.Query()["id"]; ok {
		filters["id"] = ids
		for _, id := range ids {
			if err := entities.Id(id).Validate(); err != nil {
				c.errorLogNResponse(w, "getting protected addresses: parsing id", err)
				return
			}
		}
	}

	if emails, ok := r.URL.Query()["email"]; ok {
		filters["email"] = emails
		for _, email := range emails {
			if err := entities.Email(email).Validate(); err != nil {
				c.errorLogNResponse(w, "getting protected addresses: parsing email", err)
				return
			}
		}
	}

	praddrs, err := c.svcGw.PrAddrs.GetAll(c.context, filters)
	if err != nil {
		c.errorLogNResponse(w, "getting protected addresses", err)
		return
	}

	prAddrsData := make([]ProtectedAddressData, 0, len(praddrs))
	for _, praddr := range praddrs {
		prAddrsData = append(prAddrsData, addressTPrAddrData(praddr))
	}

	resp := GetPrAddrsResponse(prAddrsData)
	c.successResponse(w, resp, http.StatusOK)
}

// GetPrAddrById retrieves a single protected address by its ID.
func (c *Controller) GetPrAddrById(w http.ResponseWriter, r *http.Request) {
	prAddrId := entities.Id(r.PathValue("id"))
	if err := prAddrId.Validate(); err != nil {
		c.errorLogNResponse(w, "getting protected address by id: parsing id", err)
		return
	}

	prAddr, err := c.svcGw.Aliases.GetById(c.context, entities.Id(prAddrId))
	if err != nil {
		c.errorLogNResponse(w, "getting protected address by id", err)
		return
	}

	resp := GetPrAddrDetailsResponse(addressTPrAddrData(prAddr))
	c.successResponse(w, resp, http.StatusOK)
}

// CreatePrAddr creates a new protected address for the current user.
func (c *Controller) CreatePrAddr(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		c.errorLogNResponse(w, "getting aliases: identifying user", err)
	}

	rb := CreateProtectedAddressRequest{}
	if err := readBody(r.Body, &rb); err != nil {
		c.errorLogNResponse(w, "parsing protected address create request", err)
		return
	}

	praddr, err := c.svcGw.PrAddrs.Create(
		c.context, entities.Email(rb.Email), entities.AddressMetadata(rb.Metadata), user,
	)
	if err != nil {
		c.errorLogNResponse(w, "creating protected address", err)
		return
	}

	resp := CreatePrAddrResponse(addressTPrAddrData(praddr))
	c.successResponse(w, resp, http.StatusCreated)
}

// UpdatePrAddr updates an existing protected address by its ID.
func (c *Controller) UpdatePrAddr(w http.ResponseWriter, r *http.Request) {
	prAddrId := entities.Id(r.PathValue("id"))
	if err := prAddrId.Validate(); err != nil {
		c.errorLogNResponse(w, "getting protected address by id: parsing id", err)
		return
	}

	prAddr, err := c.svcGw.PrAddrs.GetById(c.context, prAddrId)
	if err != nil {
		c.errorLogNResponse(w, "updating protected address: retrieving alias by id", err)
		return
	}

	rb := UpdateProtectedAddressRequest{}
	if err := readBody(r.Body, &rb); err != nil {
		c.errorLogNResponse(w, "parsing protected address update request", err)
		return
	}

	if rb.Metadata != nil {
		prAddr.Metadata = entities.AddressMetadata(*rb.Metadata)
	}

	prAddr, err = c.svcGw.PrAddrs.Update(c.context, prAddr)
	if err != nil {
		c.errorLogNResponse(w, "updating protected address", err)
		return
	}

	resp := UpdatePrAddrResponse(addressTPrAddrData(prAddr))
	c.successResponse(w, resp, http.StatusOK)
}

// DeletePrAddr deletes a protected address by its ID.
func (c *Controller) DeletePrAddr(w http.ResponseWriter, r *http.Request) {
	prAddrId := entities.Id(r.PathValue("id"))
	if err := prAddrId.Validate(); err != nil {
		c.errorLogNResponse(w, "getting protected address by id: parsing id", err)
		return
	}

	if err := c.svcGw.PrAddrs.DeleteById(c.context, prAddrId); err != nil {
		c.errorLogNResponse(w, "getting alias by id: parsing id", err)
		return
	}

	c.successResponse(w, "", http.StatusNoContent)
}
