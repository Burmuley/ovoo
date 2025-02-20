package gorm

import "github.com/Burmuley/ovoo/internal/entities"

// UserFEntity converts an entities.User to a User
func UserFEntity(e entities.User) User {
	u := User{}
	u.ID = e.ID.String()
	u.FirstName = e.FirstName
	u.LastName = e.LastName
	u.Login = e.Login
	return u
}

// UserTEntity converts a User to an entities.User
func UserTEntity(u User) entities.User {
	return entities.User{
		ID:        entities.Id(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Login:     u.Login,
	}
}

// AddressFEntity converts an entities.Address to an Address
func AddressFEntity(e entities.Address) Address {
	addr := Address{
		Model:          Model{ID: string(e.ID)},
		Email:          e.Email.String(),
		ForwardAddress: nil,
		Owner:          UserFEntity(e.Owner),
		Type:           int(e.Type),
		Metadata: AddressMetadata{
			Comment:     e.Metadata.Comment,
			ServiceName: e.Metadata.ServiceName,
		},
	}
	if e.ForwardAddress != nil {
		fa := AddressFEntity(*e.ForwardAddress)
		addr.ForwardAddress = &fa
	}

	return addr
}

// AddressTEntity converts an Address to an entities.Address
func AddressTEntity(a Address) entities.Address {
	addr := entities.Address{
		ID:             entities.Id(a.ID),
		Type:           entities.AddressType(a.Type),
		Email:          entities.Email(a.Email),
		ForwardAddress: nil,
		Owner:          UserTEntity(a.Owner),
		Metadata: entities.AddressMetadata{
			Comment:     a.Metadata.Comment,
			ServiceName: a.Metadata.ServiceName,
		},
	}

	if a.ForwardAddress != nil {
		fa := AddressTEntity(*a.ForwardAddress)
		addr.ForwardAddress = &fa
	}

	return addr
}

// ChainFEntity converts an entities.Chain to a Chain
func ChainFEntity(e entities.Chain) Chain {
	return Chain{
		Hash:        string(e.Hash),
		CreatedAt:   e.CreatedAt,
		FromAddress: AddressFEntity(e.FromAddress),
		ToAddress:   AddressFEntity(e.FromAddress),
	}
}

// ChainTEntity converts a Chain to an entities.Chain
func ChainTEntity(e Chain) entities.Chain {
	return entities.Chain{
		Hash:        entities.Hash(e.Hash),
		FromAddress: AddressTEntity(e.FromAddress),
		ToAddress:   AddressTEntity(e.ToAddress),
		CreatedAt:   e.CreatedAt,
	}
}

// ApiTokenFEntity converts an entities.ApiToken to an ApiToken
func ApiTokenFEntity(e entities.ApiToken) ApiToken {
	return ApiToken{
		Model:       Model{ID: e.ID.String()},
		Token:       e.Token,
		Description: e.Description,
		Owner:       UserFEntity(e.Owner),
		Expiration:  e.Expiration,
	}
}

// ApiTokenTEntity converts an ApiToken to an entities.ApiToken
func ApiTokenTEntity(t ApiToken) entities.ApiToken {
	return entities.ApiToken{
		ID:          entities.Id(t.ID),
		Description: t.Description,
		Owner:       UserTEntity(t.Owner),
		Expiration:  t.Expiration,
	}
}
