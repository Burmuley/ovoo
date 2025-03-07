package rest

import (
	"net/http"

	"github.com/Burmuley/ovoo/internal/entities"
)

func (c *Controller) getChainByHash(w http.ResponseWriter, r *http.Request) {
	chainHash := entities.Hash(r.PathValue("hash"))
	if err := chainHash.Validate(); err != nil {
		c.errorLogNResponse(w, "getting chain by hash: parsing hash", err)
		return
	}

	chain, err := c.svcGw.Chains.GetByHash(c.context, chainHash)
	if err != nil {
		c.errorLogNResponse(w, "getting chain by hash", err)
		return
	}

	resp := GetEmailChainDetailsResponse(chainTChainData(chain))
	c.successResponse(w, resp, http.StatusOK)
}

func (c *Controller) CreateChain(w http.ResponseWriter, r *http.Request) {
	rb := CreateEmailChain{}
	if err := readBody(r.Body, &rb); err != nil {
		c.errorLogNResponse(w, "parsing chain create request", err)
		return
	}

	chain, err := c.svcGw.Chains.Create(
		c.context,
		string(rb.FromEmail),
		string(rb.ToEmail),
	)

	if err != nil {
		c.errorLogNResponse(w, "creating chain", err)
		return
	}

	resp := CreateEmailChainResponse(chainTChainData(chain))
	c.successResponse(w, resp, http.StatusCreated)
}

func (c *Controller) DeleteChain(w http.ResponseWriter, r *http.Request) {
	panic("implement me!")
}
