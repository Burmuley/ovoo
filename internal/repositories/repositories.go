package repositories

import (
	"context"

	"github.com/Burmuley/ovoo/internal/entities"
)

// AddressReader defines methods for reading address data.
type AddressReader interface {
	GetById(ctx context.Context, id entities.Id) (entities.Address, error)
	GetByEmail(ctx context.Context, email entities.Email) ([]entities.Address, error)
	GetAll(ctx context.Context, filter entities.AddressFilter) ([]entities.Address, entities.PaginationMetadata, error)
}

// AddressWriter defines methods for writing address data.
type AddressWriter interface {
	Create(ctx context.Context, address entities.Address) error
	BatchCreate(ctx context.Context, addresses []entities.Address) error
	Update(ctx context.Context, address entities.Address) error
	DeleteById(ctx context.Context, cuser entities.User, id entities.Id) error
	BatchDeleteById(ctx context.Context, cuser entities.User, ids []entities.Id) error
	BatchUpdate(ctx context.Context, filter entities.AddressFilter, values entities.AddressBulkUpdateFields) error
}

// AddressReadWriter combines AddressReader and AddressWriter interfaces.
type AddressReadWriter interface {
	AddressReader
	AddressWriter
}

// ChainReader defines methods for reading chain data.
type ChainReader interface {
	GetByHash(ctx context.Context, hash entities.Hash) (entities.Chain, error)
	GetByFilters(ctx context.Context, filter entities.ChainFilter) ([]entities.Chain, error)
}

// ChainWriter defines methods for writing chain data.
type ChainWriter interface {
	Create(ctx context.Context, chain entities.Chain) error
	BatchCreate(ctx context.Context, chains []entities.Chain) error
	Delete(ctx context.Context, cuser entities.User, hash entities.Hash) (entities.Chain, error)
	BatchDelete(ctx context.Context, cuser entities.User, hashes []entities.Hash) error
}

// ChainReadWriter combines ChainReader and ChainWriter interfaces.
type ChainReadWriter interface {
	ChainReader
	ChainWriter
}

// UsersReader defines methods for reading user data.
type UsersReader interface {
	GetAll(ctx context.Context, filter entities.UserFilter) ([]entities.User, entities.PaginationMetadata, error)
	GetById(ctx context.Context, id entities.Id) (entities.User, error)
	GetByLogin(ctx context.Context, login string) (entities.User, error)
}

// UsersWriter defines methods for writing user data.
type UsersWriter interface {
	Create(ctx context.Context, user entities.User) error
	BatchCreate(ctx context.Context, users []entities.User) error
	Update(ctx context.Context, user entities.User) error
	Delete(ctx context.Context, cuser entities.User, id entities.Id) error
}

// UsersReadWriter combines UsersReader and UsersWriter interfaces.
type UsersReadWriter interface {
	UsersReader
	UsersWriter
}

// TokensReader defines methods for reading token data.
type TokensReader interface {
	GetById(ctx context.Context, tokenId entities.Id) (entities.ApiToken, error)
	GetAll(ctx context.Context, filter entities.ApiTokenFilter) ([]entities.ApiToken, error)
}

// TokensWriter defines methods for writing token data.
type TokensWriter interface {
	Create(ctx context.Context, token entities.ApiToken) error
	Update(ctx context.Context, token entities.ApiToken) (entities.ApiToken, error)
	BatchCreate(ctx context.Context, tokens []entities.ApiToken) error
	Delete(ctx context.Context, cuser entities.User, tokenId entities.Id) error
	BatchDeleteById(ctx context.Context, cuser entities.User, ids []entities.Id) error
	BatchDeleteForUser(ctx context.Context, cuser entities.User, id entities.Id) error
	// BatchUpdate()
}

// TokensReadWriter combines TokensReader and TokensWriter interfaces.
type TokensReadWriter interface {
	TokensReader
	TokensWriter
}
