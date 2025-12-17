package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

func setupChainsService(t *testing.T) (*ChainsService, *MockChainRepo, *MockAddressRepo) {
	chainRepo := new(MockChainRepo)
	addressRepo := new(MockAddressRepo)

	repof := &factory.RepoFactory{
		Chain:   chainRepo,
		Address: addressRepo,
	}

	service, err := NewChainsService("test.com", repof)
	require.NoError(t, err)

	return service, chainRepo, addressRepo
}

func TestNewChainsService(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewChainsService("test.com", repof)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewChainsService_NilRepoFactory(t *testing.T) {
	service, err := NewChainsService("test.com", nil)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

func TestNewChainsService_EmptyDomain(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewChainsService("", repof)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.Contains(t, err.Error(), "domain should be defined")
}

func TestChainsService_GetByHash_Success_AdminUser(t *testing.T) {
	service, chainRepo, _ := setupChainsService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	hash := entities.NewHash("from@example.com", "to@test.com")

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ReplyAliasAddress,
		Email: "reply@test.com",
		Owner: owner,
	}

	toAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	expectedChain := entities.Chain{
		Hash:        hash,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		CreatedAt:   time.Now(),
	}

	chainRepo.On("GetByHash", ctx, hash).Return(expectedChain, nil)

	chain, err := service.GetByHash(ctx, admin, hash)

	assert.NoError(t, err)
	assert.Equal(t, expectedChain, chain)
	chainRepo.AssertExpectations(t)
}

func TestChainsService_GetByHash_Success_MilterUser(t *testing.T) {
	service, chainRepo, _ := setupChainsService(t)
	ctx := context.Background()

	milter := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	hash := entities.NewHash("from@example.com", "to@test.com")

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ReplyAliasAddress,
		Email: "reply@test.com",
		Owner: owner,
	}

	toAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	expectedChain := entities.Chain{
		Hash:        hash,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		CreatedAt:   time.Now(),
	}

	chainRepo.On("GetByHash", ctx, hash).Return(expectedChain, nil)

	chain, err := service.GetByHash(ctx, milter, hash)

	assert.NoError(t, err)
	assert.Equal(t, expectedChain, chain)
	chainRepo.AssertExpectations(t)
}

func TestChainsService_GetByHash_NotAuthorized_RegularUser(t *testing.T) {
	service, _, _ := setupChainsService(t)
	ctx := context.Background()

	regularUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	hash := entities.NewHash("from@example.com", "to@test.com")

	chain, err := service.GetByHash(ctx, regularUser, hash)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Chain{}, chain)
}

func TestChainsService_GetByHash_InvalidHash(t *testing.T) {
	service, _, _ := setupChainsService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	invalidHash := entities.Hash("invalid")

	chain, err := service.GetByHash(ctx, admin, invalidHash)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Chain{}, chain)
}

func TestChainsService_GetByHash_NotFound(t *testing.T) {
	service, chainRepo, _ := setupChainsService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	hash := entities.NewHash("from@example.com", "to@test.com")

	chainRepo.On("GetByHash", ctx, hash).Return(entities.Chain{}, entities.ErrNotFound)

	chain, err := service.GetByHash(ctx, admin, hash)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	assert.Equal(t, entities.Chain{}, chain)
}

func TestChainsService_DeleteByHash_Success_AdminUser(t *testing.T) {
	service, chainRepo, _ := setupChainsService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	hash := entities.NewHash("from@example.com", "to@test.com")

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ReplyAliasAddress,
		Email: "reply@test.com",
		Owner: owner,
	}

	toAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	deletedChain := entities.Chain{
		Hash:        hash,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		CreatedAt:   time.Now(),
	}

	chainRepo.On("Delete", ctx, hash).Return(deletedChain, nil)

	chain, err := service.DeleteByHash(ctx, admin, hash)

	assert.NoError(t, err)
	assert.Equal(t, deletedChain, chain)
	chainRepo.AssertExpectations(t)
}

func TestChainsService_DeleteByHash_Success_MilterUser(t *testing.T) {
	service, chainRepo, _ := setupChainsService(t)
	ctx := context.Background()

	milter := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	hash := entities.NewHash("from@example.com", "to@test.com")

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ReplyAliasAddress,
		Email: "reply@test.com",
		Owner: owner,
	}

	toAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	deletedChain := entities.Chain{
		Hash:        hash,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		CreatedAt:   time.Now(),
	}

	chainRepo.On("Delete", ctx, hash).Return(deletedChain, nil)

	chain, err := service.DeleteByHash(ctx, milter, hash)

	assert.NoError(t, err)
	assert.Equal(t, deletedChain, chain)
	chainRepo.AssertExpectations(t)
}

func TestChainsService_DeleteByHash_NotAuthorized_RegularUser(t *testing.T) {
	service, _, _ := setupChainsService(t)
	ctx := context.Background()

	regularUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	hash := entities.NewHash("from@example.com", "to@test.com")

	chain, err := service.DeleteByHash(ctx, regularUser, hash)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Chain{}, chain)
}

func TestChainsService_DeleteByHash_InvalidHash(t *testing.T) {
	service, _, _ := setupChainsService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	invalidHash := entities.Hash("invalid")

	chain, err := service.DeleteByHash(ctx, admin, invalidHash)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Chain{}, chain)
}

func TestChainsService_Create_NotAuthorized_RegularUser(t *testing.T) {
	service, _, _ := setupChainsService(t)
	ctx := context.Background()

	regularUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	chain, err := service.Create(ctx, regularUser, "from@example.com", "to@test.com", owner)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Chain{}, chain)
}

func TestChainsService_Create_ExistingChain(t *testing.T) {
	service, chainRepo, _ := setupChainsService(t)
	ctx := context.Background()

	milter := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromEmail := "from@example.com"
	toEmail := "to@test.com"
	hash := entities.NewHash(fromEmail, toEmail)

	existingChain := entities.Chain{
		Hash: hash,
		FromAddress: entities.Address{
			ID:    entities.NewId(),
			Type:  entities.ReplyAliasAddress,
			Email: "reply@test.com",
			Owner: owner,
		},
		ToAddress: entities.Address{
			ID:    entities.NewId(),
			Type:  entities.ProtectedAddress,
			Email: "protected@example.com",
			Owner: owner,
		},
		CreatedAt: time.Now(),
	}

	chainRepo.On("GetByHash", ctx, hash).Return(existingChain, nil)

	chain, err := service.Create(ctx, milter, fromEmail, toEmail, owner)

	assert.NoError(t, err)
	assert.Equal(t, existingChain, chain)
	chainRepo.AssertExpectations(t)
}

func TestChainsService_Create_AliasNotFound(t *testing.T) {
	service, chainRepo, addressRepo := setupChainsService(t)
	ctx := context.Background()

	milter := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromEmail := "from@example.com"
	toEmail := "to@test.com"
	hash := entities.NewHash(fromEmail, toEmail)

	// Chain doesn't exist
	chainRepo.On("GetByHash", ctx, hash).Return(entities.Chain{}, entities.ErrNotFound)

	// No alias address found
	addressRepo.On("GetByEmail", ctx, entities.Email(toEmail)).Return(nil, entities.ErrNotFound)

	chain, err := service.Create(ctx, milter, fromEmail, toEmail, owner)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	assert.Equal(t, entities.Chain{}, chain)
}

func TestChainsService_Create_DestinationNotAlias(t *testing.T) {
	service, chainRepo, addressRepo := setupChainsService(t)
	ctx := context.Background()

	milter := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromEmail := "from@example.com"
	toEmail := "to@test.com"
	hash := entities.NewHash(fromEmail, toEmail)

	// Chain doesn't exist
	chainRepo.On("GetByHash", ctx, hash).Return(entities.Chain{}, entities.ErrNotFound)

	// Return a protected address instead of an alias
	protectedAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: entities.Email(toEmail),
		Owner: owner,
	}
	addressRepo.On("GetByEmail", ctx, entities.Email(toEmail)).Return([]entities.Address{protectedAddr}, nil)

	chain, err := service.Create(ctx, milter, fromEmail, toEmail, owner)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Contains(t, err.Error(), "destination alias not found")
	assert.Equal(t, entities.Chain{}, chain)
}

func TestChainsService_Create_Success(t *testing.T) {
	service, chainRepo, addressRepo := setupChainsService(t)
	ctx := context.Background()

	milter := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromEmail := "sender@external.com"
	toEmail := "alias@test.com"
	hash := entities.NewHash(fromEmail, toEmail)

	protectedAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	aliasAddr := entities.Address{
		ID:             entities.NewId(),
		Type:           entities.AliasAddress,
		Email:          entities.Email(toEmail),
		ForwardAddress: &protectedAddr,
		Owner:          owner,
	}

	// Chain doesn't exist
	chainRepo.On("GetByHash", ctx, hash).Return(entities.Chain{}, entities.ErrNotFound)

	// Return alias address for toEmail
	addressRepo.On("GetByEmail", ctx, entities.Email(toEmail)).Return([]entities.Address{aliasAddr}, nil)

	// Check for existing external address (not found, will create new)
	addressRepo.On("GetByEmail", ctx, entities.Email(fromEmail)).Return(nil, entities.ErrNotFound)

	// Create external address
	addressRepo.On("Create", ctx, mock.MatchedBy(func(addr entities.Address) bool {
		return addr.Type == entities.ExternalAddress && addr.Email == entities.Email(fromEmail)
	})).Return(nil)

	// Create reply alias
	addressRepo.On("Create", ctx, mock.MatchedBy(func(addr entities.Address) bool {
		return addr.Type == entities.ReplyAliasAddress
	})).Return(nil)

	// Create both chains
	chainRepo.On("BatchCreate", ctx, mock.AnythingOfType("[]entities.Chain")).Return(nil)

	chain, err := service.Create(ctx, milter, fromEmail, toEmail, owner)

	assert.NoError(t, err)
	assert.Equal(t, hash, chain.Hash)
	chainRepo.AssertExpectations(t)
	addressRepo.AssertExpectations(t)
}

func TestChainsService_Create_ExistingExternalAddress(t *testing.T) {
	service, chainRepo, addressRepo := setupChainsService(t)
	ctx := context.Background()

	milter := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromEmail := "sender@external.com"
	toEmail := "alias@test.com"
	hash := entities.NewHash(fromEmail, toEmail)

	protectedAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	aliasAddr := entities.Address{
		ID:             entities.NewId(),
		Type:           entities.AliasAddress,
		Email:          entities.Email(toEmail),
		ForwardAddress: &protectedAddr,
		Owner:          owner,
	}

	existingExternalAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ExternalAddress,
		Email: entities.Email(fromEmail),
		Owner: owner,
	}

	// Chain doesn't exist
	chainRepo.On("GetByHash", ctx, hash).Return(entities.Chain{}, entities.ErrNotFound)

	// Return alias address for toEmail
	addressRepo.On("GetByEmail", ctx, entities.Email(toEmail)).Return([]entities.Address{aliasAddr}, nil)

	// Return existing external address for fromEmail
	addressRepo.On("GetByEmail", ctx, entities.Email(fromEmail)).Return([]entities.Address{existingExternalAddr}, nil)

	// Create reply alias
	addressRepo.On("Create", ctx, mock.MatchedBy(func(addr entities.Address) bool {
		return addr.Type == entities.ReplyAliasAddress
	})).Return(nil)

	// Create both chains
	chainRepo.On("BatchCreate", ctx, mock.AnythingOfType("[]entities.Chain")).Return(nil)

	chain, err := service.Create(ctx, milter, fromEmail, toEmail, owner)

	assert.NoError(t, err)
	assert.Equal(t, hash, chain.Hash)
	chainRepo.AssertExpectations(t)
	addressRepo.AssertExpectations(t)
}

func TestChainsService_Create_AdminUser(t *testing.T) {
	service, chainRepo, addressRepo := setupChainsService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	fromEmail := "sender@external.com"
	toEmail := "alias@test.com"
	hash := entities.NewHash(fromEmail, toEmail)

	protectedAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	aliasAddr := entities.Address{
		ID:             entities.NewId(),
		Type:           entities.AliasAddress,
		Email:          entities.Email(toEmail),
		ForwardAddress: &protectedAddr,
		Owner:          owner,
	}

	// Chain doesn't exist
	chainRepo.On("GetByHash", ctx, hash).Return(entities.Chain{}, entities.ErrNotFound)

	// Return alias address for toEmail
	addressRepo.On("GetByEmail", ctx, entities.Email(toEmail)).Return([]entities.Address{aliasAddr}, nil)

	// Check for existing external address (not found, will create new)
	addressRepo.On("GetByEmail", ctx, entities.Email(fromEmail)).Return(nil, entities.ErrNotFound)

	// Create external address
	addressRepo.On("Create", ctx, mock.MatchedBy(func(addr entities.Address) bool {
		return addr.Type == entities.ExternalAddress
	})).Return(nil)

	// Create reply alias
	addressRepo.On("Create", ctx, mock.MatchedBy(func(addr entities.Address) bool {
		return addr.Type == entities.ReplyAliasAddress
	})).Return(nil)

	// Create both chains
	chainRepo.On("BatchCreate", ctx, mock.AnythingOfType("[]entities.Chain")).Return(nil)

	chain, err := service.Create(ctx, admin, fromEmail, toEmail, owner)

	assert.NoError(t, err)
	assert.Equal(t, hash, chain.Hash)
	chainRepo.AssertExpectations(t)
	addressRepo.AssertExpectations(t)
}
