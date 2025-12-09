package rest

import (
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

// GetAllPrAddrs retrieves all protected addresses for the current user.
// Supports filtering by ID and email through query parameters.
func (a *Application) GetAllPrAddrs(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting protected addresses: identifying user", err)
		return
	}

	// filling filters
	filters := readFilters(r, []string{"owner", "id", "email", "page_size", "page"})
	praddrs, pgm, err := a.svcGw.PrAddrs.GetAll(a.context, user, filters)
	if err != nil {
		a.errorLogNResponse(w, "getting protected addresses", err)
		return
	}

	prAddrsData := make([]ProtectedAddressData, 0, len(praddrs))
	for _, praddr := range praddrs {
		prAddrsData = append(prAddrsData, addressTPrAddrData(praddr))
	}

	resp := GetPrAddrsResponse{
		ProtectedAddresses: prAddrsData,
		PaginationMetadata: pgmTMetadata(pgm),
	}
	a.successResponse(w, resp, http.StatusOK)
}

// GetPrAddrById retrieves a single protected address by its ID.
func (a *Application) GetPrAddrById(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting protected addresses: identifying user", err)
		return
	}

	prAddrId := entities.Id(r.PathValue("id"))
	prAddr, err := a.svcGw.Aliases.GetById(a.context, user, entities.Id(prAddrId))
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
		a.errorLogNResponse(w, "getting protected addresses: identifying user", err)
		return
	}

	req := CreateProtectedAddressRequest{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "parsing protected address create request", err)
		return
	}

	praddr, err := a.svcGw.PrAddrs.Create(
		a.context, user, services.PrAddrCreateCmd{
			Email: entities.Email(req.Email),
			Metadata: struct {
				Comment     *string
				ServiceName *string
			}{
				Comment:     req.Metadata.Comment,
				ServiceName: req.Metadata.ServiceName,
			},
		},
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
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting protected addresses: identifying user", err)
		return
	}

	praddrId := entities.Id(r.PathValue("id"))
	req := UpdateProtectedAddressRequest{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "parsing protected address update request", err)
		return
	}

	praddr, err := a.svcGw.PrAddrs.Update(a.context, user, services.PrAddrUpdateCmd{
		PrAddrId: praddrId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment:     req.Metadata.Comment,
			ServiceName: req.Metadata.ServiceName,
		},
	})
	if err != nil {
		a.errorLogNResponse(w, "updating protected address", err)
		return
	}

	resp := UpdatePrAddrResponse(addressTPrAddrData(praddr))
	a.successResponse(w, resp, http.StatusOK)
}

// DeletePrAddr deletes a protected address by its ID.
func (a *Application) DeletePrAddr(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting protected addresses: identifying user", err)
		return
	}

	prAddrId := entities.Id(r.PathValue("id"))
	if err := a.svcGw.PrAddrs.DeleteById(a.context, user, prAddrId); err != nil {
		a.errorLogNResponse(w, "getting alias by id: parsing id", err)
		return
	}

	a.successResponse(w, "", http.StatusNoContent)
}
