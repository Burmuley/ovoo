package rest

import (
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
)

func (a *Application) getChainByHash(w http.ResponseWriter, r *http.Request) {
	chainHash := entities.Hash(r.PathValue("hash"))
	if err := chainHash.Validate(); err != nil {
		a.errorLogNResponse(w, "getting chain by hash: parsing hash", err)
		return
	}

	chain, err := a.svcGw.Chains.GetByHash(a.context, chainHash)
	if err != nil {
		a.errorLogNResponse(w, "getting chain by hash", err)
		return
	}

	resp := GetEmailChainDetailsResponse(chainTChainData(chain))
	a.successResponse(w, resp, http.StatusOK)
}

func (a *Application) CreateChain(w http.ResponseWriter, r *http.Request) {
	rb := CreateEmailChain{}
	if err := readBody(r.Body, &rb); err != nil {
		a.errorLogNResponse(w, "parsing chain create request", err)
		return
	}

	user, err := userFromContext(r)
	if err != nil {
		a.errorLogNResponse(w, "getting aliases: identifying user", err)
	}

	chain, err := a.svcGw.Chains.Create(
		a.context,
		string(rb.FromEmail),
		string(rb.ToEmail),
		user,
	)

	if err != nil {
		a.errorLogNResponse(w, "creating chain", err)
		return
	}

	resp := CreateEmailChainResponse(chainTChainData(chain))
	a.successResponse(w, resp, http.StatusCreated)
}

func (c *Application) DeleteChain(w http.ResponseWriter, r *http.Request) {
	panic("implement me!")
}
