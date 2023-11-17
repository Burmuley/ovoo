package api

import (
	"github.com/Burmuley/domovoi/pkg/state"
)

func aliasFromState(stAlias state.Alias) Alias {
	return Alias{
		Active:  stAlias.Active,
		Comment: stAlias.Comment,
		Email:   stAlias.Email,
		Id:      stAlias.ID,
		ProtectedAddress: ProtectedAddress{
			Active: stAlias.ProtectedAddress.Active,
			Email:  stAlias.ProtectedAddress.Email,
			Id:     stAlias.ProtectedAddress.ID,
		},
		ServiceName: stAlias.ServiceName,
	}
}

func aliasToState(apiAlias Alias) state.Alias {
	return state.Alias{
		ProtectedAddressID: apiAlias.ProtectedAddress.Id,
		Comment:            apiAlias.Comment,
		ServiceName:        apiAlias.ServiceName,
		Email:              apiAlias.Email,
		Active:             apiAlias.Active,
	}
}
