package rest

import (
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
)

// GetAllPrAddrs retrieves all protected addresses for the current user.
// Supports filtering by ID and email through query parameters.
func (a *Application) GetAllPrAddrs(w http.ResponseWriter, r *http.Request) {
	// TODO: get real user when authentication is enabled
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting aliases: identifying user", err)
	}
	filters := map[string][]string{"owner": []string{user.ID.String()}}
	if ids, ok := r.URL.Query()["id"]; ok {
		filters["id"] = ids
		for _, id := range ids {
			if err := entities.Id(id).Validate(); err != nil {
				a.errorLogNResponse(w, "getting protected addresses: parsing id", err)
				return
			}
		}
	}

	if emails, ok := r.URL.Query()["email"]; ok {
		filters["email"] = emails
		for _, email := range emails {
			if err := entities.Email(email).Validate(); err != nil {
				a.errorLogNResponse(w, "getting protected addresses: parsing email", err)
				return
			}
		}
	}

	praddrs, err := a.svcGw.PrAddrs.GetAll(a.context, filters)
	if err != nil {
		a.errorLogNResponse(w, "getting protected addresses", err)
		return
	}

	prAddrsData := make([]ProtectedAddressData, 0, len(praddrs))
	for _, praddr := range praddrs {
		prAddrsData = append(prAddrsData, addressTPrAddrData(praddr))
	}

	resp := GetPrAddrsResponse(prAddrsData)
	a.successResponse(w, resp, http.StatusOK)
}

// GetPrAddrById retrieves a single protected address by its ID.
func (a *Application) GetPrAddrById(w http.ResponseWriter, r *http.Request) {
	prAddrId := entities.Id(r.PathValue("id"))
	if err := prAddrId.Validate(); err != nil {
		a.errorLogNResponse(w, "getting protected address by id: parsing id", err)
		return
	}

	prAddr, err := a.svcGw.Aliases.GetById(a.context, entities.Id(prAddrId))
	if err != nil {
		a.errorLogNResponse(w, "getting protected address by id", err)
		return
	}

	resp := GetPrAddrDetailsResponse(addressTPrAddrData(prAddr))
	a.successResponse(w, resp, http.StatusOK)
}

// CreatePrAddr creates a new protected address for the current user.
func (a *Application) CreatePrAddr(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting aliases: identifying user", err)
	}

	rb := CreateProtectedAddressRequest{}
	if err := readBody(r.Body, &rb); err != nil {
		a.errorLogNResponse(w, "parsing protected address create request", err)
		return
	}

	praddr, err := a.svcGw.PrAddrs.Create(
		a.context, entities.Email(rb.Email), entities.AddressMetadata(rb.Metadata), user,
	)
	if err != nil {
		a.errorLogNResponse(w, "creating protected address", err)
		return
	}

	resp := CreatePrAddrResponse(addressTPrAddrData(praddr))
	a.successResponse(w, resp, http.StatusCreated)
}

// UpdatePrAddr updates an existing protected address by its ID.
func (a *Application) UpdatePrAddr(w http.ResponseWriter, r *http.Request) {
	prAddrId := entities.Id(r.PathValue("id"))
	if err := prAddrId.Validate(); err != nil {
		a.errorLogNResponse(w, "getting protected address by id: parsing id", err)
		return
	}

	prAddr, err := a.svcGw.PrAddrs.GetById(a.context, prAddrId)
	if err != nil {
		a.errorLogNResponse(w, "updating protected address: retrieving alias by id", err)
		return
	}

	rb := UpdateProtectedAddressRequest{}
	if err := readBody(r.Body, &rb); err != nil {
		a.errorLogNResponse(w, "parsing protected address update request", err)
		return
	}

	if rb.Metadata != nil {
		prAddr.Metadata = entities.AddressMetadata(*rb.Metadata)
	}

	prAddr, err = a.svcGw.PrAddrs.Update(a.context, prAddr)
	if err != nil {
		a.errorLogNResponse(w, "updating protected address", err)
		return
	}

	resp := UpdatePrAddrResponse(addressTPrAddrData(prAddr))
	a.successResponse(w, resp, http.StatusOK)
}

// DeletePrAddr deletes a protected address by its ID.
func (a *Application) DeletePrAddr(w http.ResponseWriter, r *http.Request) {
	prAddrId := entities.Id(r.PathValue("id"))
	if err := prAddrId.Validate(); err != nil {
		a.errorLogNResponse(w, "getting protected address by id: parsing id", err)
		return
	}

	if err := a.svcGw.PrAddrs.DeleteById(a.context, prAddrId); err != nil {
		a.errorLogNResponse(w, "getting alias by id: parsing id", err)
		return
	}

	a.successResponse(w, "", http.StatusNoContent)
}
