package main

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/repositories/factory"
	"github.com/Burmuley/ovoo/internal/services"
)

func makeServices(repoFactory *factory.RepoFactory, domain string, dict []string) (*services.ServiceGateway, error) {
	var err error
	svcGw := services.ServiceGateway{}

	if svcGw.Aliases, err = services.NewAliasesService(domain, dict, repoFactory); err != nil {
		return nil, fmt.Errorf("initializing aliases service: %w", err)
	}

	if svcGw.PrAddrs, err = services.NewProtectedAddrService(repoFactory); err != nil {
		return nil, fmt.Errorf("initializing protected addresses service: %w", err)
	}

	if svcGw.Chains, err = services.NewChainsService(domain, repoFactory); err != nil {
		return nil, fmt.Errorf("initializing chains service: %w", err)
	}

	if svcGw.Users, err = services.NewUsersService(repoFactory); err != nil {
		return nil, fmt.Errorf("initializing users service: %w", err)
	}

	return &svcGw, nil
}
