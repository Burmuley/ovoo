package rest

import "net/http"

func (a *Application) GetDomains(w http.ResponseWriter, r *http.Request) {
	resp := GetDomainsResponse{Domains: a.svcGw.Aliases.Domains()}
	a.successResponse(w, resp, http.StatusOK)
}
