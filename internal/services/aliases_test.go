package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

func setupAliasesService(t *testing.T) (*AliasesService, *MockAddressRepo, *MockChainRepo) {
	addressRepo := new(MockAddressRepo)
	chainRepo := new(MockChainRepo)

	repof := &factory.RepoFactory{
		Address: addressRepo,
		Chain:   chainRepo,
	}

	service, err := NewAliasesService("test.com", []string{"word1", "word2", "word3"}, repof)
	require.NoError(t, err)

	return service, addressRepo, chainRepo
}

func TestNewAliasesService(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewAliasesService("test.com", []string{"word1", "word2"}, repof)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewAliasesService_InvalidDomain(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewAliasesService("x", []string{"word1", "word2"}, repof)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.Contains(t, err.Error(), "invalid domain")
}

func TestNewAliasesService_EmptyDomain(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewAliasesService("", []string{"word1", "word2"}, repof)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

func TestNewAliasesService_EmptyWordsDictionary(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewAliasesService("test.com", []string{}, repof)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.Contains(t, err.Error(), "words dictionary can not be empty")
}

func TestNewAliasesService_NilRepoFactory(t *testing.T) {
	service, err := NewAliasesService("test.com", []string{"word1", "word2"}, nil)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

func TestAliasesService_Create_Success(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	serviceName := "Test Service"
	comment := "Test comment"
	cmd := AliasCreateCmd{
		ProtectedAddressId: string(prAddrId),
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment:     &comment,
			ServiceName: &serviceName,
		},
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(protectedAddr, nil)
	addressRepo.On("Create", ctx, mock.AnythingOfType("entities.Address")).Return(nil)

	alias, err := service.Create(ctx, user, cmd)

	assert.NoError(t, err)
	assert.Equal(t, entities.AliasAddress, alias.Type)
	assert.Equal(t, serviceName, alias.Metadata.ServiceName)
	assert.Equal(t, comment, alias.Metadata.Comment)
	assert.NotEmpty(t, alias.ID)
	assert.NotEmpty(t, alias.Email)
	addressRepo.AssertExpectations(t)
}

func TestAliasesService_Create_NotAuthorized_MilterUser(t *testing.T) {
	service, _, _ := setupAliasesService(t)
	ctx := context.Background()

	milterUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	cmd := AliasCreateCmd{
		ProtectedAddressId: string(entities.NewId()),
	}

	alias, err := service.Create(ctx, milterUser, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Address{}, alias)
}

func TestAliasesService_Create_ProtectedAddressNotFound(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	cmd := AliasCreateCmd{
		ProtectedAddressId: string(prAddrId),
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(entities.Address{}, entities.ErrNotFound)

	alias, err := service.Create(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Address{}, alias)
}

func TestAliasesService_Update_Success(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	aliasId := entities.NewId()
	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}
	existingAlias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias123@test.com",
		ForwardAddress: &protectedAddr,
		Owner:          user,
		Metadata: entities.AddressMetadata{
			ServiceName: "Old Service",
			Comment:     "Old comment",
		},
	}

	newServiceName := "New Service"
	newComment := "New comment"
	cmd := AliasUpdateCmd{
		AliasId: aliasId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment:     &newComment,
			ServiceName: &newServiceName,
		},
	}

	addressRepo.On("GetById", ctx, aliasId).Return(existingAlias, nil)
	addressRepo.On("Update", ctx, mock.AnythingOfType("entities.Address")).Return(nil)

	updatedAlias, err := service.Update(ctx, user, cmd)

	assert.NoError(t, err)
	assert.Equal(t, newServiceName, updatedAlias.Metadata.ServiceName)
	assert.Equal(t, newComment, updatedAlias.Metadata.Comment)
	addressRepo.AssertExpectations(t)
}

func TestAliasesService_Update_NotAuthorized(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	nonOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "nonowner@test.com",
	}

	aliasId := entities.NewId()
	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}
	existingAlias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias123@test.com",
		ForwardAddress: &protectedAddr,
		Owner:          owner,
	}

	newComment := "New comment"
	cmd := AliasUpdateCmd{
		AliasId: aliasId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment: &newComment,
		},
	}

	addressRepo.On("GetById", ctx, aliasId).Return(existingAlias, nil)

	alias, err := service.Update(ctx, nonOwner, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Address{}, alias)
}

func TestAliasesService_Update_AliasNotFound(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	aliasId := entities.NewId()
	newComment := "New comment"
	cmd := AliasUpdateCmd{
		AliasId: aliasId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment: &newComment,
		},
	}

	addressRepo.On("GetById", ctx, aliasId).Return(entities.Address{}, entities.ErrNotFound)

	alias, err := service.Update(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Address{}, alias)
}

func TestAliasesService_GetAll_AdminUser(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	filters := map[string][]string{
		"page":      {"1"},
		"page_size": {"10"},
	}

	expectedAliases := []entities.Address{
		{ID: entities.NewId(), Email: "alias1@test.com"},
		{ID: entities.NewId(), Email: "alias2@test.com"},
	}
	expectedMetadata := entities.PaginationMetadata{
		CurrentPage:  1,
		PageSize:     10,
		TotalRecords: 2,
	}

	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(expectedAliases, expectedMetadata, nil)

	aliases, metadata, err := service.GetAll(ctx, admin, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedAliases, aliases)
	assert.Equal(t, expectedMetadata, metadata)
	addressRepo.AssertExpectations(t)
}

func TestAliasesService_GetAll_RegularUser(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	filters := map[string][]string{
		"page":      {"1"},
		"page_size": {"10"},
	}

	expectedAliases := []entities.Address{
		{ID: entities.NewId(), Email: "alias1@test.com", Owner: user},
	}
	expectedMetadata := entities.PaginationMetadata{
		CurrentPage:  1,
		PageSize:     10,
		TotalRecords: 1,
	}

	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(expectedAliases, expectedMetadata, nil)

	aliases, metadata, err := service.GetAll(ctx, user, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedAliases, aliases)
	assert.Equal(t, expectedMetadata, metadata)
	addressRepo.AssertExpectations(t)
}

func TestAliasesService_GetAll_MilterUserNotAuthorized(t *testing.T) {
	service, _, _ := setupAliasesService(t)
	ctx := context.Background()

	milterUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	filters := map[string][]string{}

	aliases, metadata, err := service.GetAll(ctx, milterUser, filters)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Empty(t, aliases)
	assert.Equal(t, entities.PaginationMetadata{}, metadata)
}

func TestAliasesService_GetById_Success_Owner(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	aliasId := entities.NewId()
	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}
	expectedAlias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias123@test.com",
		ForwardAddress: &protectedAddr,
		Owner:          user,
	}

	addressRepo.On("GetById", ctx, aliasId).Return(expectedAlias, nil)

	alias, err := service.GetById(ctx, user, aliasId)

	assert.NoError(t, err)
	assert.Equal(t, expectedAlias, alias)
	addressRepo.AssertExpectations(t)
}

func TestAliasesService_GetById_Success_Admin(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
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

	aliasId := entities.NewId()
	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}
	expectedAlias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias123@test.com",
		ForwardAddress: &protectedAddr,
		Owner:          owner,
	}

	addressRepo.On("GetById", ctx, aliasId).Return(expectedAlias, nil)

	alias, err := service.GetById(ctx, admin, aliasId)

	assert.NoError(t, err)
	assert.Equal(t, expectedAlias, alias)
	addressRepo.AssertExpectations(t)
}

func TestAliasesService_GetById_NotAuthorized(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	nonOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "nonowner@test.com",
	}

	aliasId := entities.NewId()
	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}
	existingAlias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias123@test.com",
		ForwardAddress: &protectedAddr,
		Owner:          owner,
	}

	addressRepo.On("GetById", ctx, aliasId).Return(existingAlias, nil)

	alias, err := service.GetById(ctx, nonOwner, aliasId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Address{}, alias)
}

func TestAliasesService_GetById_InvalidId(t *testing.T) {
	service, _, _ := setupAliasesService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	invalidId := entities.Id("invalid")

	alias, err := service.GetById(ctx, user, invalidId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Address{}, alias)
}

func TestAliasesService_GetById_NotFound(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	aliasId := entities.NewId()

	addressRepo.On("GetById", ctx, aliasId).Return(entities.Address{}, entities.ErrNotFound)

	alias, err := service.GetById(ctx, user, aliasId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	assert.Equal(t, entities.Address{}, alias)
}

func TestAliasesService_DeleteById_Success_Owner(t *testing.T) {
	service, addressRepo, chainRepo := setupAliasesService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	aliasId := entities.NewId()
	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}
	existingAlias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias123@test.com",
		ForwardAddress: &protectedAddr,
		Owner:          user,
	}

	addressRepo.On("GetById", ctx, aliasId).Return(existingAlias, nil)

	// Mocks for deleteAliasIds -> deleteChainsForAliasIds
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return([]entities.Chain{}, nil).Twice()
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil).Twice()

	err := service.DeleteById(ctx, user, aliasId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
	chainRepo.AssertExpectations(t)
}

func TestAliasesService_DeleteById_Success_Admin(t *testing.T) {
	service, addressRepo, chainRepo := setupAliasesService(t)
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

	aliasId := entities.NewId()
	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}
	existingAlias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias123@test.com",
		ForwardAddress: &protectedAddr,
		Owner:          owner,
	}

	addressRepo.On("GetById", ctx, aliasId).Return(existingAlias, nil)

	// Mocks for deleteAliasIds -> deleteChainsForAliasIds
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return([]entities.Chain{}, nil).Twice()
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil).Twice()

	err := service.DeleteById(ctx, admin, aliasId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
	chainRepo.AssertExpectations(t)
}

func TestAliasesService_DeleteById_NotAuthorized(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	owner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	nonOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "nonowner@test.com",
	}

	aliasId := entities.NewId()
	prAddrId := entities.NewId()
	protectedAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}
	existingAlias := entities.Address{
		ID:             aliasId,
		Type:           entities.AliasAddress,
		Email:          "alias123@test.com",
		ForwardAddress: &protectedAddr,
		Owner:          owner,
	}

	addressRepo.On("GetById", ctx, aliasId).Return(existingAlias, nil)

	err := service.DeleteById(ctx, nonOwner, aliasId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
}

func TestAliasesService_DeleteById_InvalidId(t *testing.T) {
	service, _, _ := setupAliasesService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	invalidId := entities.Id("invalid")

	err := service.DeleteById(ctx, user, invalidId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
}

func TestAliasesService_DeleteById_NotFound(t *testing.T) {
	service, addressRepo, _ := setupAliasesService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	aliasId := entities.NewId()

	addressRepo.On("GetById", ctx, aliasId).Return(entities.Address{}, entities.ErrNotFound)

	err := service.DeleteById(ctx, user, aliasId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}
