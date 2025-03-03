package rest

import (
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/oapi-codegen/runtime/types"
)

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen.yaml -initialism-overrides ../../../openapi.yaml

// userTypeFStr converts a string representation of user type to its corresponding UserType.
// It returns 99 if the provided user type is not recognized.
func userTypeFStr(st string) entities.UserType {
	umap := map[string]entities.UserType{
		"regular": entities.RegularUser,
		"admin":   entities.AdminUser,
		"milter":  entities.MilterUser,
	}

	if ut, ok := umap[st]; !ok {
		return ut
	}

	return 99
}

// userTypeTStr converts a UserType to its string representation.
// It returns "unknown" if the provided UserType is not recognized.
func userTypeTStr(ut entities.UserType) string {
	umap := map[entities.UserType]string{
		entities.RegularUser: "regular",
		entities.AdminUser:   "admin",
		entities.MilterUser:  "milter",
	}

	if st, ok := umap[ut]; ok {
		return st
	}

	return "unknown"
}

// userTResponse converts an entities.User to a UserData response.
func userTResponse(u entities.User) UserData {
	return UserData{
		FirstName: u.FirstName,
		Id:        string(u.ID),
		LastName:  u.LastName,
		Login:     types.Email(u.Login),
		Type:      userTypeTStr(u.Type),
	}
}

func addressTAliasData(alias entities.Address) AliasData {
	return AliasData{
		Email:        types.Email(alias.Email),
		ForwardEmail: types.Email(alias.ForwardAddress.Email),
		Id:           alias.ID.String(),
		Metadata: AddressMetadata{
			Comment:     alias.Metadata.Comment,
			ServiceName: alias.Metadata.ServiceName,
		},
		Owner: userTResponse(alias.Owner),
	}
}

func addressTPrAddrData(praddr entities.Address) ProtectedAddressData {
	return ProtectedAddressData{
		Email: types.Email(praddr.Email),
		Id:    praddr.ID.String(),
		Metadata: AddressMetadata{
			Comment:     praddr.Metadata.Comment,
			ServiceName: praddr.Metadata.ServiceName,
		},
		Owner: userTResponse(praddr.Owner),
	}
}

func chainTChainData(chain entities.Chain) ChainData {
	return ChainData{
		FromEmail: string(chain.FromAddress.Email),
		Hash:      chain.Hash.String(),
		ToEmail:   string(chain.ToAddress.Email),
	}
}
