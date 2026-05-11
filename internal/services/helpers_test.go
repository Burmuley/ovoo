package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

func setupHelpersTest() (*factory.RepoFactory, *MockAddressRepo, *MockChainRepo) {
	addressRepo := new(MockAddressRepo)
	chainRepo := new(MockChainRepo)

	repof := &factory.RepoFactory{
		Address: addressRepo,
		Chain:   chainRepo,
	}

	return repof, addressRepo, chainRepo
}

func setupDeactivateHelpersTest() (*factory.RepoFactory, *MockAddressRepo, *MockApiTokensRepo) {
	addressRepo := new(MockAddressRepo)
	tokensRepo := new(MockApiTokensRepo)

	repof := &factory.RepoFactory{
		Address:   addressRepo,
		ApiTokens: tokensRepo,
	}

	return repof, addressRepo, tokensRepo
}

func TestDeletePrAddrsForUser_NoAddresses(t *testing.T) {
	repof, addressRepo, _ := setupHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()

	// No protected addresses for user
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{},
		entities.PaginationMetadata{},
		nil,
	)

	// Batch delete with empty list
	addressRepo.On("BatchDeleteById", ctx, []entities.Id{}).Return(nil)

	err := deletePrAddrsForUser(ctx, repof, userId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestDeletePrAddrsForUser_WithAddresses(t *testing.T) {
	repof, addressRepo, chainRepo := setupHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	prAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	// Return protected address
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(filter entities.AddressFilter) bool {
		return len(filter.Types) == 1 && filter.Types[0] == entities.ProtectedAddress
	})).Return(
		[]entities.Address{prAddr},
		entities.PaginationMetadata{},
		nil,
	).Once()

	// For deleteAliasesForPrAddr - no aliases
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(filter entities.AddressFilter) bool {
		return len(filter.Types) == 1 && filter.Types[0] == entities.AliasAddress
	})).Return(
		[]entities.Address{},
		entities.PaginationMetadata{},
		nil,
	).Once()

	// Batch delete protected addresses
	addressRepo.On("BatchDeleteById", ctx, []entities.Id{prAddrId}).Return(nil)

	_ = chainRepo

	err := deletePrAddrsForUser(ctx, repof, userId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestDeletePrAddrsForUser_WithAddressesAndAliases(t *testing.T) {
	repof, addressRepo, chainRepo := setupHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	prAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	aliasId := entities.NewId()
	alias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias@test.com",
		ForwardAddress: &prAddr,
		Owner:          user,
	}

	// Return protected address
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(filter entities.AddressFilter) bool {
		return len(filter.Types) == 1 && filter.Types[0] == entities.ProtectedAddress
	})).Return(
		[]entities.Address{prAddr},
		entities.PaginationMetadata{},
		nil,
	).Once()

	// For deleteAliasesForPrAddr - return alias
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(filter entities.AddressFilter) bool {
		return len(filter.Types) == 1 && filter.Types[0] == entities.AliasAddress
	})).Return(
		[]entities.Address{alias},
		entities.PaginationMetadata{},
		nil,
	).Once()

	// Mocks for deleteChainsForAliasIds
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return([]entities.Chain{}, nil).Twice()
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil).Times(3)

	err := deletePrAddrsForUser(ctx, repof, userId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
	chainRepo.AssertExpectations(t)
}

func TestDeleteAliasesForPrAddr_NoAliases(t *testing.T) {
	repof, addressRepo, _ := setupHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	// No aliases for protected address
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{},
		entities.PaginationMetadata{},
		nil,
	)

	err := deleteAliasesForPrAddr(ctx, repof, prAddr)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestDeleteAliasesForPrAddr_WithAliases(t *testing.T) {
	repof, addressRepo, chainRepo := setupHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	aliasId := entities.NewId()
	alias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias@test.com",
		ForwardAddress: &prAddr,
		Owner:          user,
	}

	// Return aliases for protected address
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{alias},
		entities.PaginationMetadata{},
		nil,
	)

	// Mocks for deleteChainsForAliasIds
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return([]entities.Chain{}, nil).Twice()
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil).Twice()

	err := deleteAliasesForPrAddr(ctx, repof, prAddr)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
	chainRepo.AssertExpectations(t)
}

func TestDeleteAliasIds_EmptyList(t *testing.T) {
	repof, addressRepo, chainRepo := setupHelpersTest()
	ctx := context.Background()

	// Empty chains
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return([]entities.Chain{}, nil).Twice()
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil).Twice()

	err := deleteAliasIds(ctx, repof, []entities.Id{})

	assert.NoError(t, err)
	chainRepo.AssertExpectations(t)
	addressRepo.AssertExpectations(t)
}

func TestDeleteAliasIds_WithAliases(t *testing.T) {
	repof, addressRepo, chainRepo := setupHelpersTest()
	ctx := context.Background()

	aliasId := entities.NewId()

	// Empty chains
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return([]entities.Chain{}, nil).Twice()
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil).Twice()

	err := deleteAliasIds(ctx, repof, []entities.Id{aliasId})

	assert.NoError(t, err)
	chainRepo.AssertExpectations(t)
	addressRepo.AssertExpectations(t)
}

func TestDeleteChainsForAliasIds_NoChains(t *testing.T) {
	repof, addressRepo, chainRepo := setupHelpersTest()
	ctx := context.Background()

	aliasId := entities.NewId()

	// No chains found
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return([]entities.Chain{}, nil).Twice()
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil)

	err := deleteChainsForAliasIds(ctx, repof, []entities.Id{aliasId})

	assert.NoError(t, err)
	chainRepo.AssertExpectations(t)
	addressRepo.AssertExpectations(t)
}

func TestDeleteChainsForAliasIds_WithChains(t *testing.T) {
	repof, addressRepo, chainRepo := setupHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	aliasId := entities.NewId()
	alias := entities.Address{
		ID:    aliasId,
		Type:  entities.AliasAddress,
		Email: "alias@test.com",
		Owner: user,
	}

	replyAliasId := entities.NewId()
	replyAlias := entities.Address{
		ID:    replyAliasId,
		Type:  entities.ReplyAliasAddress,
		Email: "reply@test.com",
		Owner: user,
	}

	externalAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ExternalAddress,
		Email: "external@example.com",
		Owner: user,
	}

	protectedAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	fwdChain := entities.Chain{
		Hash:            entities.NewHash("external@example.com", "alias@test.com"),
		FromAddress:     replyAlias,
		ToAddress:       protectedAddr,
		OrigFromAddress: externalAddr,
		OrigToAddress:   alias,
	}

	replyChain := entities.Chain{
		Hash:            entities.NewHash("protected@example.com", "reply@test.com"),
		FromAddress:     alias,
		ToAddress:       externalAddr,
		OrigFromAddress: protectedAddr,
		OrigToAddress:   replyAlias,
	}

	// Forward chains
	chainRepo.On("GetByFilters", ctx, mock.MatchedBy(func(filter entities.ChainFilter) bool {
		return len(filter.OrigToAddrIds) > 0
	})).Return([]entities.Chain{fwdChain}, nil).Once()

	// Reply chains
	chainRepo.On("GetByFilters", ctx, mock.MatchedBy(func(filter entities.ChainFilter) bool {
		return len(filter.FromAddrsIds) > 0
	})).Return([]entities.Chain{replyChain}, nil).Once()

	// Batch delete chains
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)

	// Batch delete reply alias addresses
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil)

	err := deleteChainsForAliasIds(ctx, repof, []entities.Id{aliasId})

	assert.NoError(t, err)
	chainRepo.AssertExpectations(t)
	addressRepo.AssertExpectations(t)
}

func TestDeleteChainsForAliasIds_ErrorGettingForwardChains(t *testing.T) {
	repof, _, chainRepo := setupHelpersTest()
	ctx := context.Background()

	aliasId := entities.NewId()

	// Error getting forward chains
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return(nil, entities.ErrDatabase).Once()

	err := deleteChainsForAliasIds(ctx, repof, []entities.Id{aliasId})

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrDatabase)
}

func TestDeleteAliasesForPrAddr_ErrorGettingAliases(t *testing.T) {
	repof, addressRepo, _ := setupHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	// Error getting aliases
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{},
		entities.PaginationMetadata{},
		entities.ErrDatabase,
	)

	err := deleteAliasesForPrAddr(ctx, repof, prAddr)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrDatabase)
}

func TestDeletePrAddrsForUser_ErrorGettingAddresses(t *testing.T) {
	repof, addressRepo, _ := setupHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()

	// Error getting protected addresses
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{},
		entities.PaginationMetadata{},
		entities.ErrDatabase,
	)

	err := deletePrAddrsForUser(ctx, repof, userId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrDatabase)
}

// Tests for deactivateAliasesForPrAddr

func TestDeactivateAliasesForPrAddr_NoAliases(t *testing.T) {
	repof, addressRepo, _ := setupDeactivateHelpersTest()
	ctx := context.Background()

	praddrId := entities.NewId()

	active := true
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.AddressFilter) bool {
		return f.Active != nil && *f.Active == active
	})).Return([]entities.Address{}, entities.PaginationMetadata{}, nil)

	err := deactivateAliasesForPrAddr(ctx, repof, praddrId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestDeactivateAliasesForPrAddr_WithAliases(t *testing.T) {
	repof, addressRepo, _ := setupDeactivateHelpersTest()
	ctx := context.Background()

	praddrId := entities.NewId()
	owner := entities.User{ID: entities.NewId(), Type: entities.RegularUser}

	alias1 := entities.Address{ID: entities.NewId(), Type: entities.AliasAddress, Email: "alias1@test.com", Owner: owner, Active: true}
	alias2 := entities.Address{ID: entities.NewId(), Type: entities.AliasAddress, Email: "alias2@test.com", Owner: owner, Active: true}

	active := true
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.AddressFilter) bool {
		return f.Active != nil && *f.Active == active
	})).Return([]entities.Address{alias1, alias2}, entities.PaginationMetadata{}, nil)

	addressRepo.On("Update", ctx, mock.MatchedBy(func(a entities.Address) bool {
		return !a.Active
	})).Return(nil).Times(2)

	err := deactivateAliasesForPrAddr(ctx, repof, praddrId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestDeactivateAliasesForPrAddr_ErrorGettingAliases(t *testing.T) {
	repof, addressRepo, _ := setupDeactivateHelpersTest()
	ctx := context.Background()

	praddrId := entities.NewId()

	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{}, entities.PaginationMetadata{}, entities.ErrDatabase,
	)

	err := deactivateAliasesForPrAddr(ctx, repof, praddrId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrDatabase)
}

// Tests for deactivatePrAddrsForUser

func TestDeactivatePrAddrsForUser_NoAddresses(t *testing.T) {
	repof, addressRepo, _ := setupDeactivateHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()

	active := true
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.AddressFilter) bool {
		return f.Active != nil && *f.Active == active
	})).Return([]entities.Address{}, entities.PaginationMetadata{}, nil)

	err := deactivatePrAddrsForUser(ctx, repof, userId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestDeactivatePrAddrsForUser_WithAddresses(t *testing.T) {
	repof, addressRepo, _ := setupDeactivateHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	owner := entities.User{ID: userId, Type: entities.RegularUser}
	praddrId := entities.NewId()
	praddr := entities.Address{ID: praddrId, Type: entities.ProtectedAddress, Email: "protected@example.com", Owner: owner, Active: true}

	active := true
	// First GetAll: returns the active praddr
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.AddressFilter) bool {
		return f.Active != nil && *f.Active == active && len(f.Owners) > 0
	})).Return([]entities.Address{praddr}, entities.PaginationMetadata{}, nil).Once()

	// Second GetAll: deactivateAliasesForPrAddr looks up active aliases for the praddr → none
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.AddressFilter) bool {
		return f.Active != nil && *f.Active == active && len(f.ForwardAddressIds) > 0
	})).Return([]entities.Address{}, entities.PaginationMetadata{}, nil).Once()

	// Update praddr to inactive
	addressRepo.On("Update", ctx, mock.MatchedBy(func(a entities.Address) bool {
		return a.ID == praddrId && !a.Active
	})).Return(nil)

	err := deactivatePrAddrsForUser(ctx, repof, userId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestDeactivatePrAddrsForUser_WithAddressesAndAliases(t *testing.T) {
	repof, addressRepo, _ := setupDeactivateHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	owner := entities.User{ID: userId, Type: entities.RegularUser}
	praddrId := entities.NewId()
	praddr := entities.Address{ID: praddrId, Type: entities.ProtectedAddress, Email: "protected@example.com", Owner: owner, Active: true}
	alias := entities.Address{ID: entities.NewId(), Type: entities.AliasAddress, Email: "alias@test.com", Owner: owner, Active: true}

	active := true
	// GetAll for active praddrs owned by user
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.AddressFilter) bool {
		return f.Active != nil && *f.Active == active && len(f.Owners) > 0
	})).Return([]entities.Address{praddr}, entities.PaginationMetadata{}, nil).Once()

	// GetAll for active aliases forwarding to praddr
	addressRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.AddressFilter) bool {
		return f.Active != nil && *f.Active == active && len(f.ForwardAddressIds) > 0
	})).Return([]entities.Address{alias}, entities.PaginationMetadata{}, nil).Once()

	// Update alias to inactive, then praddr to inactive
	addressRepo.On("Update", ctx, mock.MatchedBy(func(a entities.Address) bool {
		return !a.Active
	})).Return(nil).Times(2)

	err := deactivatePrAddrsForUser(ctx, repof, userId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestDeactivatePrAddrsForUser_ErrorGettingAddresses(t *testing.T) {
	repof, addressRepo, _ := setupDeactivateHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()

	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{}, entities.PaginationMetadata{}, entities.ErrDatabase,
	)

	err := deactivatePrAddrsForUser(ctx, repof, userId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrDatabase)
}

// Tests for deactivateTokensForUser

func TestDeactivateTokensForUser_NoTokens(t *testing.T) {
	repof, _, tokensRepo := setupDeactivateHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()

	active := true
	tokensRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.ApiTokenFilter) bool {
		return f.Active != nil && *f.Active == active
	})).Return([]entities.ApiToken{}, nil)

	err := deactivateTokensForUser(ctx, repof, userId)

	assert.NoError(t, err)
	tokensRepo.AssertExpectations(t)
}

func TestDeactivateTokensForUser_WithTokens(t *testing.T) {
	repof, _, tokensRepo := setupDeactivateHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()
	owner := entities.User{ID: userId, Type: entities.RegularUser}
	token1 := entities.ApiToken{ID: entities.NewId(), Name: "Token 1", Owner: owner, Active: true}
	token2 := entities.ApiToken{ID: entities.NewId(), Name: "Token 2", Owner: owner, Active: true}

	active := true
	tokensRepo.On("GetAll", ctx, mock.MatchedBy(func(f entities.ApiTokenFilter) bool {
		return f.Active != nil && *f.Active == active
	})).Return([]entities.ApiToken{token1, token2}, nil)

	tokensRepo.On("Update", ctx, mock.MatchedBy(func(tk entities.ApiToken) bool {
		return !tk.Active
	})).Return(entities.ApiToken{}, nil).Times(2)

	err := deactivateTokensForUser(ctx, repof, userId)

	assert.NoError(t, err)
	tokensRepo.AssertExpectations(t)
}

func TestDeactivateTokensForUser_ErrorGettingTokens(t *testing.T) {
	repof, _, tokensRepo := setupDeactivateHelpersTest()
	ctx := context.Background()

	userId := entities.NewId()

	tokensRepo.On("GetAll", ctx, mock.AnythingOfType("entities.ApiTokenFilter")).Return(nil, entities.ErrDatabase)

	err := deactivateTokensForUser(ctx, repof, userId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrDatabase)
}
