package rest

import (
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
)

func (a *Application) getChainByHash(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting chain: identifying user", err)
	}

	chainHash := entities.Hash(r.PathValue("hash"))
	chain, err := a.svcGw.Chains.GetByHash(a.context, cuser, chainHash)
	if err != nil {
		a.errorLogNResponse(w, "getting chain by hash", err)
		return
	}

	resp := GetEmailChainDetailsResponse(chainTChainData(chain))
	a.successResponse(w, resp, http.StatusOK)
}

func (a *Application) CreateChain(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting chain: identifying user", err)
	}

	rb := CreateEmailChain{}
	if err := readBody(r.Body, &rb); err != nil {
		a.errorLogNResponse(w, "parsing chain create request", err)
		return
	}

	chain, err := a.svcGw.Chains.Create(
		a.context,
		cuser,
		string(rb.FromEmail),
		string(rb.ToEmail),
		cuser,
	)

	if err != nil {
		a.errorLogNResponse(w, "creating chain", err)
		return
	}

	resp := CreateEmailChainResponse(chainTChainData(chain))
	a.successResponse(w, resp, http.StatusCreated)
}

func (a *Application) DeleteChain(w http.ResponseWriter, r *http.Request) {
	cuser, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting chain: identifying user", err)
	}

	hash := entities.Hash(r.PathValue("hash"))
	if _, err := a.svcGw.Chains.DeleteByHash(r.Context(), cuser, hash); err != nil {
		a.errorLogNResponse(w, "deleting chain", err)
		return
	}

	a.successResponse(w, "", http.StatusNoContent)
}
