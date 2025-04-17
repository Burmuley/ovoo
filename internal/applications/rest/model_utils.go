package rest

import (
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/oapi-codegen/runtime/types"
)

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen.yaml ../../../openapi.yaml

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
// It maps fields from the internal user entity to the API response structure.
func userTResponse(u entities.User) UserData {
	return UserData{
		FirstName: u.FirstName,
		Id:        string(u.ID),
		LastName:  u.LastName,
		Login:     types.Email(u.Login),
		Type:      userTypeTStr(u.Type),
	}
}

// addressTAliasData converts an entities.Address to an AliasData response.
// This function is used for email alias representations in the API.
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

// addressTPrAddrData converts an entities.Address to a ProtectedAddressData response.
// This is used for protected email address representations in the API.
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

// chainTChainData converts an entities.Chain to a ChainData response.
// This function transforms the internal chain entity to the API response format.
func chainTChainData(chain entities.Chain) ChainData {
	return ChainData{
		FromEmail: string(chain.FromAddress.Email),
		Hash:      chain.Hash.String(),
		ToEmail:   string(chain.ToAddress.Email),
	}
}

// tokenTApiTokenData converts an entities.ApiToken to an ApiTokenData response.
// This is used for API token representations in standard responses.
func tokenTApiTokenData(token entities.ApiToken) ApiTokenData {
	return ApiTokenData{
		Active:      token.Active,
		Description: token.Description,
		Expiration:  token.Expiration,
		Name:        token.Name,
	}
}

// tokenTApiTokenDataOnCreate converts an entities.ApiToken to an ApiTokenDataOnCreate response.
// This is specifically used when creating a new API token to include the token value in the response.
func tokenTApiTokenDataOnCreate(token entities.ApiToken) ApiTokenDataOnCreate {
	return ApiTokenDataOnCreate{
		Active:      token.Active,
		Description: token.Description,
		Expiration:  token.Expiration,
		Name:        token.Name,
		ApiToken:    token.Token,
	}
}
