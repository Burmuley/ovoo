package gorm

import "github.com/Burmuley/ovoo/internal/entities"

// UserFromEntity converts an entities.User to a User
func UserFromEntity(e entities.User) User {
	u := User{}
	u.ID = e.ID.String()
	u.FirstName = e.FirstName
	u.LastName = e.LastName
	u.Login = e.Login
	return u
}

func UserFromEntityList(eusers []entities.User) []User {
	gusers := make([]User, 0, len(eusers))
	for _, euser := range eusers {
		gusers = append(gusers, UserFromEntity(euser))
	}

	return gusers
}

// UserToEntity converts a User to an entities.User
func UserToEntity(u User) entities.User {
	return entities.User{
		ID:        entities.Id(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Login:     u.Login,
	}
}

// AddressFromEntity converts an entities.Address to an Address
func AddressFromEntity(e entities.Address) Address {
	addr := Address{
		Model:          Model{ID: string(e.ID)},
		Email:          e.Email.String(),
		ForwardAddress: nil,
		Owner:          UserFromEntity(e.Owner),
		Type:           int(e.Type),
		Metadata: AddressMetadata{
			Comment:     e.Metadata.Comment,
			ServiceName: e.Metadata.ServiceName,
		},
	}
	if e.ForwardAddress != nil {
		fa := AddressFromEntity(*e.ForwardAddress)
		addr.ForwardAddress = &fa
	}

	return addr
}

// AddressFromEntityList converts a list of entities.Address to list of Address
func AddressFromEntityList(eaddrs []entities.Address) []Address {
	gaddrs := make([]Address, 0, len(eaddrs))
	for _, eaddr := range eaddrs {
		gaddrs = append(gaddrs, AddressFromEntity(eaddr))
	}

	return gaddrs
}

// AddressToEntity converts an Address to an entities.Address
func AddressToEntity(a Address) entities.Address {
	addr := entities.Address{
		ID:             entities.Id(a.ID),
		Type:           entities.AddressType(a.Type),
		Email:          entities.Email(a.Email),
		ForwardAddress: nil,
		Owner:          UserToEntity(a.Owner),
		Metadata: entities.AddressMetadata{
			Comment:     a.Metadata.Comment,
			ServiceName: a.Metadata.ServiceName,
		},
	}

	if a.ForwardAddress != nil {
		fa := AddressToEntity(*a.ForwardAddress)
		addr.ForwardAddress = &fa
	}

	return addr
}

// ChainFromEntity converts an entities.Chain to a Chain
func ChainFromEntity(e entities.Chain) Chain {
	return Chain{
		Hash:        string(e.Hash),
		CreatedAt:   e.CreatedAt,
		FromAddress: AddressFromEntity(e.FromAddress),
		ToAddress:   AddressFromEntity(e.ToAddress),
	}
}

// ChainFromEntity converts an entities.Chain to a Chain
func ChainFromEntityList(echains []entities.Chain) []Chain {
	gchains := make([]Chain, 0, len(echains))
	for _, chain := range echains {
		gchains = append(gchains, ChainFromEntity(chain))
	}
	return gchains
}

// ChainToEntity converts a Chain to an entities.Chain
func ChainToEntity(e Chain) entities.Chain {
	return entities.Chain{
		Hash:        entities.Hash(e.Hash),
		FromAddress: AddressToEntity(e.FromAddress),
		ToAddress:   AddressToEntity(e.ToAddress),
		CreatedAt:   e.CreatedAt,
	}
}

// ApiTokenFromEntity converts an entities.ApiToken to an ApiToken
func ApiTokenFromEntity(e entities.ApiToken) ApiToken {
	return ApiToken{
		Model:       Model{ID: e.ID.String()},
		Token:       e.Token,
		Description: e.Description,
		Owner:       UserFromEntity(e.Owner),
		Expiration:  e.Expiration,
	}
}

func ApiTokenFromEntityList(etokens []entities.ApiToken) []ApiToken {
	gtokens := make([]ApiToken, 0, len(etokens))
	for _, etoken := range etokens {
		gtokens = append(gtokens, ApiTokenFromEntity(etoken))
	}

	return gtokens
}

// ApiTokenToEntity converts an ApiToken to an entities.ApiToken
func ApiTokenToEntity(t ApiToken) entities.ApiToken {
	return entities.ApiToken{
		ID:          entities.Id(t.ID),
		Description: t.Description,
		Owner:       UserToEntity(t.Owner),
		Expiration:  t.Expiration,
	}
}
