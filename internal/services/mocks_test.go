package services

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/Burmuley/ovoo/internal/entities"
)

// MockUsersRepo is a mock implementation of repositories.UsersReadWriter
type MockUsersRepo struct {
	mock.Mock
}

func (m *MockUsersRepo) GetAll(ctx context.Context, filter entities.UserFilter) ([]entities.User, entities.PaginationMetadata, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.User), args.Get(1).(entities.PaginationMetadata), args.Error(2)
}

func (m *MockUsersRepo) GetById(ctx context.Context, id entities.Id) (entities.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entities.User), args.Error(1)
}

func (m *MockUsersRepo) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(entities.User), args.Error(1)
}

func (m *MockUsersRepo) Create(ctx context.Context, user entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUsersRepo) BatchCreate(ctx context.Context, users []entities.User) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *MockUsersRepo) Update(ctx context.Context, user entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUsersRepo) Delete(ctx context.Context, id entities.Id) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockAddressRepo is a mock implementation of repositories.AddressReadWriter
type MockAddressRepo struct {
	mock.Mock
}

func (m *MockAddressRepo) GetById(ctx context.Context, id entities.Id) (entities.Address, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entities.Address), args.Error(1)
}

func (m *MockAddressRepo) GetByEmail(ctx context.Context, email entities.Email) ([]entities.Address, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Address), args.Error(1)
}

func (m *MockAddressRepo) GetAll(ctx context.Context, filter entities.AddressFilter) ([]entities.Address, entities.PaginationMetadata, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Address), args.Get(1).(entities.PaginationMetadata), args.Error(2)
}

func (m *MockAddressRepo) Create(ctx context.Context, address entities.Address) error {
	args := m.Called(ctx, address)
	return args.Error(0)
}

func (m *MockAddressRepo) BatchCreate(ctx context.Context, addresses []entities.Address) error {
	args := m.Called(ctx, addresses)
	return args.Error(0)
}

func (m *MockAddressRepo) Update(ctx context.Context, address entities.Address) error {
	args := m.Called(ctx, address)
	return args.Error(0)
}

func (m *MockAddressRepo) DeleteById(ctx context.Context, id entities.Id) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAddressRepo) BatchDeleteById(ctx context.Context, ids []entities.Id) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// MockApiTokensRepo is a mock implementation of repositories.TokensReadWriter
type MockApiTokensRepo struct {
	mock.Mock
}

func (m *MockApiTokensRepo) GetById(ctx context.Context, tokenId entities.Id) (entities.ApiToken, error) {
	args := m.Called(ctx, tokenId)
	return args.Get(0).(entities.ApiToken), args.Error(1)
}

func (m *MockApiTokensRepo) GetAllForUser(ctx context.Context, userId entities.Id, filter entities.ApiTokenFilter) ([]entities.ApiToken, error) {
	args := m.Called(ctx, userId, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.ApiToken), args.Error(1)
}

func (m *MockApiTokensRepo) Create(ctx context.Context, token entities.ApiToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockApiTokensRepo) Update(ctx context.Context, token entities.ApiToken) (entities.ApiToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(entities.ApiToken), args.Error(1)
}

func (m *MockApiTokensRepo) BatchCreate(ctx context.Context, tokens []entities.ApiToken) error {
	args := m.Called(ctx, tokens)
	return args.Error(0)
}

func (m *MockApiTokensRepo) Delete(ctx context.Context, tokenId entities.Id) error {
	args := m.Called(ctx, tokenId)
	return args.Error(0)
}

func (m *MockApiTokensRepo) BatchDeleteById(ctx context.Context, ids []entities.Id) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockApiTokensRepo) BatchDeleteForUser(ctx context.Context, id entities.Id) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockChainRepo is a mock implementation of repositories.ChainReadWriter
type MockChainRepo struct {
	mock.Mock
}

func (m *MockChainRepo) GetByHash(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(entities.Chain), args.Error(1)
}

func (m *MockChainRepo) GetByFilters(ctx context.Context, filter entities.ChainFilter) ([]entities.Chain, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Chain), args.Error(1)
}

func (m *MockChainRepo) Create(ctx context.Context, chain entities.Chain) error {
	args := m.Called(ctx, chain)
	return args.Error(0)
}

func (m *MockChainRepo) BatchCreate(ctx context.Context, chains []entities.Chain) error {
	args := m.Called(ctx, chains)
	return args.Error(0)
}

func (m *MockChainRepo) Delete(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(entities.Chain), args.Error(1)
}

func (m *MockChainRepo) BatchDelete(ctx context.Context, hashes []entities.Hash) error {
	args := m.Called(ctx, hashes)
	return args.Error(0)
}
