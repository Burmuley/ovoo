package services

import (
	"fmt"

	"github.com/Burmuley/ovoo/internal/entities"
)

type ServiceGateway struct {
	Aliases *AliasesService
	Users   *UsersService
	PrAddrs *ProtectedAddrService
	Chains  *ChainsService
}

func New(usecases ...any) (*ServiceGateway, error) {
	f := &ServiceGateway{}
	for _, uc := range usecases {
		switch t := uc.(type) {
		case *AliasesService:
			f.Aliases = t
		case *UsersService:
			f.Users = t
		case *ProtectedAddrService:
			f.PrAddrs = t
		case *ChainsService:
			f.Chains = t
		default:
			return nil, fmt.Errorf("%w: unknown service type %T", entities.ErrConfiguration, t)
		}
	}

	return f, nil
}
