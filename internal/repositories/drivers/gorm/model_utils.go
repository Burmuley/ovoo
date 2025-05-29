package gorm

import "github.com/Burmuley/ovoo/internal/entities"

// userFromEntity converts an entities.User to a User
func userFromEntity(e entities.User) User {
	updatedBy := User{}
	if e.UpdatedBy != nil {
		updatedBy = userFromEntity(*e.UpdatedBy)
	}

	u := User{
		Model: Model{
			ID:        string(e.ID),
			UpdatedAt: e.UpdatedAt,
			CreatedAt: e.UpdatedAt,
		},
		FirstName:      e.FirstName,
		LastName:       e.LastName,
		Login:          e.Login,
		Type:           int(e.Type),
		PwdHash:        e.PasswordHash,
		FailedAttempts: e.FailedAttempts,
		LockoutUntil:   e.LockoutUntil,
		UpdatedBy:      &updatedBy,
	}

	return u
}

// userFromEntityList converts a list of entities.User to list of User
func userFromEntityList(eusers []entities.User) []User {
	gusers := make([]User, 0, len(eusers))
	for _, euser := range eusers {
		gusers = append(gusers, userFromEntity(euser))
	}

	return gusers
}

// userToEntity converts a User to an entities.User
func userToEntity(u User) entities.User {
	updatedBy := entities.User{}
	if u.UpdatedBy != nil {
		updatedBy = userToEntity(*u.UpdatedBy)
	}

	return entities.User{
		ID:             entities.Id(u.ID),
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Login:          u.Login,
		Type:           entities.UserType(u.Type),
		PasswordHash:   u.PwdHash,
		LockoutUntil:   u.LockoutUntil,
		FailedAttempts: u.FailedAttempts,
		UpdatedAt:      u.UpdatedAt,
		CreatedAt:      u.UpdatedAt,
		UpdatedBy:      &updatedBy,
	}
}

// addressFromEntity converts an entities.Address to an Address
func addressFromEntity(e entities.Address) Address {
	addr := Address{
		Model: Model{
			ID:        string(e.ID),
			UpdatedAt: e.UpdatedAt,
			CreatedAt: e.UpdatedAt,
		},
		Email:          e.Email.String(),
		ForwardAddress: nil,
		Owner:          userFromEntity(e.Owner),
		Type:           int(e.Type),
		Metadata: AddressMetadata{
			Comment:     e.Metadata.Comment,
			ServiceName: e.Metadata.ServiceName,
		},
		UpdatedBy: userFromEntity(e.UpdatedBy),
	}
	if e.ForwardAddress != nil {
		fa := addressFromEntity(*e.ForwardAddress)
		addr.ForwardAddress = &fa
	}

	return addr
}

// addressFromEntityList converts a list of entities.Address to list of Address
func addressFromEntityList(eaddrs []entities.Address) []Address {
	gaddrs := make([]Address, 0, len(eaddrs))
	for _, eaddr := range eaddrs {
		gaddrs = append(gaddrs, addressFromEntity(eaddr))
	}

	return gaddrs
}

// addressToEntity converts an Address to an entities.Address
func addressToEntity(a Address) entities.Address {
	addr := entities.Address{
		ID:             entities.Id(a.ID),
		Type:           entities.AddressType(a.Type),
		Email:          entities.Email(a.Email),
		ForwardAddress: nil,
		Owner:          userToEntity(a.Owner),
		Metadata: entities.AddressMetadata{
			Comment:     a.Metadata.Comment,
			ServiceName: a.Metadata.ServiceName,
		},
		UpdatedAt: a.UpdatedAt,
		CreatedAt: a.UpdatedAt,
		UpdatedBy: userToEntity(a.UpdatedBy),
	}

	if a.ForwardAddress != nil {
		fa := addressToEntity(*a.ForwardAddress)
		addr.ForwardAddress = &fa
	}

	return addr
}

func addressToEntityList(a []Address) []entities.Address {
	ea := make([]entities.Address, 0, len(a))
	for _, addr := range a {
		ea = append(ea, addressToEntity(addr))
	}

	return ea
}

// chainFromEntity converts an entities.Chain to a Chain
func chainFromEntity(e entities.Chain) Chain {
	return Chain{
		Hash:            string(e.Hash),
		CreatedAt:       e.CreatedAt,
		FromAddress:     addressFromEntity(e.FromAddress),
		ToAddress:       addressFromEntity(e.ToAddress),
		OrigFromAddress: addressFromEntity(e.OrigFromAddress),
		OrigToAddress:   addressFromEntity(e.OrigToAddress),
		UpdatedAt:       e.UpdatedAt,
		UpdatedBy:       userFromEntity(e.UpdatedBy),
	}
}

// ChainFromEntity converts an entities.Chain to a Chain
func chainFromEntityList(echains []entities.Chain) []Chain {
	gchains := make([]Chain, 0, len(echains))
	for _, chain := range echains {
		gchains = append(gchains, chainFromEntity(chain))
	}
	return gchains
}

// chainToEntity converts a Chain to an entities.Chain
func chainToEntity(e Chain) entities.Chain {
	return entities.Chain{
		Hash:            entities.Hash(e.Hash),
		FromAddress:     addressToEntity(e.FromAddress),
		ToAddress:       addressToEntity(e.ToAddress),
		OrigFromAddress: addressToEntity(e.OrigFromAddress),
		OrigToAddress:   addressToEntity(e.OrigToAddress),
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
		UpdatedBy:       userToEntity(e.UpdatedBy),
	}
}

// apiTokenFromEntity converts an entities.ApiToken to an ApiToken
func apiTokenFromEntity(e entities.ApiToken) ApiToken {
	return ApiToken{
		Model: Model{
			ID:        e.ID.String(),
			CreatedAt: e.CreatedAt,
			UpdatedAt: e.UpdatedAt,
		},
		Name:        e.Name,
		TokenHash:   e.TokenHash,
		Salt:        e.Salt,
		Description: e.Description,
		Owner:       userFromEntity(e.Owner),
		Expiration:  e.Expiration,
		Active:      e.Active,
		UpdatedBy:   userFromEntity(e.UpdatedBy),
	}
}

func apiTokenFromEntityList(etokens []entities.ApiToken) []ApiToken {
	gtokens := make([]ApiToken, 0, len(etokens))
	for _, etoken := range etokens {
		gtokens = append(gtokens, apiTokenFromEntity(etoken))
	}

	return gtokens
}

// apiTokenToEntity converts an ApiToken to an entities.ApiToken
func apiTokenToEntity(t ApiToken) entities.ApiToken {
	return entities.ApiToken{
		ID:          entities.Id(t.ID),
		Name:        t.Name,
		TokenHash:   t.TokenHash,
		Salt:        t.Salt,
		Description: t.Description,
		Owner:       userToEntity(t.Owner),
		Expiration:  t.Expiration,
		Active:      t.Active,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		UpdatedBy:   userToEntity(t.UpdatedBy),
	}
}
