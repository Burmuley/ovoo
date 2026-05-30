package rest

import (
	"fmt"
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/services"
)

func (a *Application) GetDomains(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting domains: identifying user", err)
		return
	}

	domains, err := a.svcGw.Domains.GetAll(r.Context(), cuser)
	if err != nil {
		a.errorLogNResponse(w, "getting domains", err)
		return
	}

	resp := GetDomainsResponse{Domains: make([]DomainData, 0, len(domains))}
	for _, d := range domains {
		resp.Domains = append(resp.Domains, customDomainTDomainData(d))
	}

	a.successResponse(w, resp, http.StatusOK)
}

func (a *Application) CreateDomain(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "creating domain: identifying user", err)
		return
	}

	req := CreateDomainRequest{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "creating domain: parsing request", fmt.Errorf("%w: %w", entities.ErrValidation, err))
		return
	}

	cmd := services.DomainCreateCmd{Name: req.Name}
	if req.Source != nil && *req.Source == Global {
		cmd.Global = true
	}

	domain, err := a.svcGw.Domains.Create(r.Context(), cuser, cmd)
	if err != nil {
		a.errorLogNResponse(w, "creating domain", err)
		return
	}

	resp := CreateDomainResponse(customDomainTDomainData(domain))
	a.successResponse(w, resp, http.StatusCreated)
}

func (a *Application) UpdateDomain(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "updating domain: identifying user", err)
		return
	}

	req := UpdateDomainRequest{}
	if err := readBody(r.Body, &req); err != nil {
		a.errorLogNResponse(w, "updating domain: parsing request", fmt.Errorf("%w: %w", entities.ErrValidation, err))
		return
	}

	domain, err := a.svcGw.Domains.Update(r.Context(), cuser, services.DomainUpdateCmd{
		DomainId: entities.Id(r.PathValue("id")),
		Active:   req.Active,
	})
	if err != nil {
		a.errorLogNResponse(w, "updating domain", err)
		return
	}

	resp := UpdateDomainResponse(customDomainTDomainData(domain))
	a.successResponse(w, resp, http.StatusOK)
}

func (a *Application) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "deleting domain: identifying user", err)
		return
	}

	if err := a.svcGw.Domains.Delete(r.Context(), cuser, entities.Id(r.PathValue("id"))); err != nil {
		a.errorLogNResponse(w, "deleting domain", err)
		return
	}

	a.successResponse(w, struct{}{}, http.StatusNoContent)
}

func (a *Application) VerifyDomain(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
