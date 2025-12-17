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

func setupProtectedAddrService(t *testing.T) (*ProtectedAddrService, *MockAddressRepo, *MockChainRepo) {
	addressRepo := new(MockAddressRepo)
	chainRepo := new(MockChainRepo)

	repof := &factory.RepoFactory{
		Address: addressRepo,
		Chain:   chainRepo,
	}

	service, err := NewProtectedAddrService(repof)
	require.NoError(t, err)

	return service, addressRepo, chainRepo
}

func TestNewProtectedAddrService(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewProtectedAddrService(repof)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewProtectedAddrService_NilRepoFactory(t *testing.T) {
	service, err := NewProtectedAddrService(nil)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

func TestProtectedAddrService_Create_Success(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	serviceName := "Test Service"
	comment := "Test comment"
	cmd := PrAddrCreateCmd{
		Email: "protected@example.com",
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment:     &comment,
			ServiceName: &serviceName,
		},
	}

	// No existing protected address with this email
	addressRepo.On("GetByEmail", ctx, entities.Email("protected@example.com")).Return(nil, entities.ErrNotFound)
	addressRepo.On("Create", ctx, mock.AnythingOfType("entities.Address")).Return(nil)

	praddr, err := service.Create(ctx, user, cmd)

	assert.NoError(t, err)
	assert.Equal(t, entities.ProtectedAddress, praddr.Type)
	assert.Equal(t, entities.Email("protected@example.com"), praddr.Email)
	assert.Equal(t, serviceName, praddr.Metadata.ServiceName)
	assert.Equal(t, comment, praddr.Metadata.Comment)
	assert.NotEmpty(t, praddr.ID)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_Create_AdminUser(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	cmd := PrAddrCreateCmd{
		Email: "protected@example.com",
	}

	addressRepo.On("GetByEmail", ctx, entities.Email("protected@example.com")).Return(nil, entities.ErrNotFound)
	addressRepo.On("Create", ctx, mock.AnythingOfType("entities.Address")).Return(nil)

	praddr, err := service.Create(ctx, admin, cmd)

	assert.NoError(t, err)
	assert.Equal(t, entities.ProtectedAddress, praddr.Type)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_Create_NotAuthorized_MilterUser(t *testing.T) {
	service, _, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	milterUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.MilterUser,
		Login: "milter@test.com",
	}

	cmd := PrAddrCreateCmd{
		Email: "protected@example.com",
	}

	praddr, err := service.Create(ctx, milterUser, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_Create_InvalidEmail(t *testing.T) {
	service, _, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	cmd := PrAddrCreateCmd{
		Email: "invalid-email",
	}

	praddr, err := service.Create(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_Create_DuplicateEmail(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	cmd := PrAddrCreateCmd{
		Email: "protected@example.com",
	}

	existingAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	addressRepo.On("GetByEmail", ctx, entities.Email("protected@example.com")).Return([]entities.Address{existingAddr}, nil)

	praddr, err := service.Create(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrDuplicateEntry)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_Update_Success(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
		Metadata: entities.AddressMetadata{
			ServiceName: "Old Service",
			Comment:     "Old comment",
		},
	}

	newServiceName := "New Service"
	newComment := "New comment"
	cmd := PrAddrUpdateCmd{
		PrAddrId: prAddrId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment:     &newComment,
			ServiceName: &newServiceName,
		},
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)
	addressRepo.On("Update", ctx, mock.AnythingOfType("entities.Address")).Return(nil)

	updatedPrAddr, err := service.Update(ctx, user, cmd)

	assert.NoError(t, err)
	assert.Equal(t, newServiceName, updatedPrAddr.Metadata.ServiceName)
	assert.Equal(t, newComment, updatedPrAddr.Metadata.Comment)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_Update_AdminUser(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
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

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	newComment := "Admin update"
	cmd := PrAddrUpdateCmd{
		PrAddrId: prAddrId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment: &newComment,
		},
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)
	addressRepo.On("Update", ctx, mock.AnythingOfType("entities.Address")).Return(nil)

	updatedPrAddr, err := service.Update(ctx, admin, cmd)

	assert.NoError(t, err)
	assert.Equal(t, newComment, updatedPrAddr.Metadata.Comment)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_Update_NotAuthorized(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
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

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	newComment := "Unauthorized update"
	cmd := PrAddrUpdateCmd{
		PrAddrId: prAddrId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment: &newComment,
		},
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)

	praddr, err := service.Update(ctx, nonOwner, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_Update_NotFound(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	newComment := "Update"
	cmd := PrAddrUpdateCmd{
		PrAddrId: prAddrId,
		Metadata: struct {
			Comment     *string
			ServiceName *string
		}{
			Comment: &newComment,
		},
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(entities.Address{}, entities.ErrNotFound)

	praddr, err := service.Update(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_GetAll_AdminUser(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
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

	expectedPrAddrs := []entities.Address{
		{ID: entities.NewId(), Email: "protected1@example.com", Type: entities.ProtectedAddress},
		{ID: entities.NewId(), Email: "protected2@example.com", Type: entities.ProtectedAddress},
	}
	expectedMetadata := entities.PaginationMetadata{
		CurrentPage:  1,
		PageSize:     10,
		TotalRecords: 2,
	}

	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(expectedPrAddrs, expectedMetadata, nil)

	praddrs, metadata, err := service.GetAll(ctx, admin, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedPrAddrs, praddrs)
	assert.Equal(t, expectedMetadata, metadata)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_GetAll_RegularUser(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
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

	expectedPrAddrs := []entities.Address{
		{ID: entities.NewId(), Email: "protected1@example.com", Type: entities.ProtectedAddress, Owner: user},
	}
	expectedMetadata := entities.PaginationMetadata{
		CurrentPage:  1,
		PageSize:     10,
		TotalRecords: 1,
	}

	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(expectedPrAddrs, expectedMetadata, nil)

	praddrs, metadata, err := service.GetAll(ctx, user, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedPrAddrs, praddrs)
	assert.Equal(t, expectedMetadata, metadata)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_GetById_Success_Owner(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	expectedPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(expectedPrAddr, nil)

	praddr, err := service.GetById(ctx, user, prAddrId)

	assert.NoError(t, err)
	assert.Equal(t, expectedPrAddr, praddr)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_GetById_Success_Admin(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
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

	prAddrId := entities.NewId()
	expectedPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(expectedPrAddr, nil)

	praddr, err := service.GetById(ctx, admin, prAddrId)

	assert.NoError(t, err)
	assert.Equal(t, expectedPrAddr, praddr)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_GetById_NotAuthorized(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
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

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)

	praddr, err := service.GetById(ctx, nonOwner, prAddrId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_GetById_InvalidId(t *testing.T) {
	service, _, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	invalidId := entities.Id("invalid")

	praddr, err := service.GetById(ctx, user, invalidId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_GetById_NotFound(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()

	addressRepo.On("GetById", ctx, prAddrId).Return(entities.Address{}, entities.ErrNotFound)

	praddr, err := service.GetById(ctx, user, prAddrId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_GetByEmail_Success(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	email := entities.Email("protected@example.com")
	expectedPrAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: email,
		Owner: user,
	}

	addressRepo.On("GetByEmail", ctx, email).Return([]entities.Address{expectedPrAddr}, nil)

	praddr, err := service.GetByEmail(ctx, user, email)

	assert.NoError(t, err)
	assert.Equal(t, expectedPrAddr, praddr)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_GetByEmail_NotAuthorized(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
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

	email := entities.Email("protected@example.com")
	existingPrAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: email,
		Owner: owner,
	}

	addressRepo.On("GetByEmail", ctx, email).Return([]entities.Address{existingPrAddr}, nil)

	praddr, err := service.GetByEmail(ctx, nonOwner, email)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_GetByEmail_NotFound(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	email := entities.Email("nonexistent@example.com")

	// Return an alias address, not a protected one
	aliasAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.AliasAddress,
		Email: email,
		Owner: user,
	}

	addressRepo.On("GetByEmail", ctx, email).Return([]entities.Address{aliasAddr}, nil)

	praddr, err := service.GetByEmail(ctx, user, email)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_GetByEmail_InvalidEmail(t *testing.T) {
	service, _, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	invalidEmail := entities.Email("invalid-email")

	praddr, err := service.GetByEmail(ctx, user, invalidEmail)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Address{}, praddr)
}

func TestProtectedAddrService_DeleteById_Success_Owner(t *testing.T) {
	service, addressRepo, chainRepo := setupProtectedAddrService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: user,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)

	// Mocks for deleteAliasesForPrAddr (no aliases)
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{},
		entities.PaginationMetadata{},
		nil,
	)

	// Mock for deleting the protected address
	addressRepo.On("DeleteById", ctx, prAddrId).Return(nil)

	// Prevent unused mock warning
	_ = chainRepo

	err := service.DeleteById(ctx, user, prAddrId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_DeleteById_Success_Admin(t *testing.T) {
	service, addressRepo, chainRepo := setupProtectedAddrService(t)
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

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)

	// Mocks for deleteAliasesForPrAddr (no aliases)
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{},
		entities.PaginationMetadata{},
		nil,
	)

	// Mock for deleting the protected address
	addressRepo.On("DeleteById", ctx, prAddrId).Return(nil)

	_ = chainRepo

	err := service.DeleteById(ctx, admin, prAddrId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_DeleteById_WithAliases(t *testing.T) {
	service, addressRepo, chainRepo := setupProtectedAddrService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
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
		ForwardAddress: &existingPrAddr,
		Owner:          user,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)

	// Mocks for deleteAliasesForPrAddr
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{alias},
		entities.PaginationMetadata{},
		nil,
	)

	// Mocks for deleteChainsForAliasIds
	chainRepo.On("GetByFilters", ctx, mock.AnythingOfType("entities.ChainFilter")).Return([]entities.Chain{}, nil).Twice()
	chainRepo.On("BatchDelete", ctx, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil).Twice()

	// Mock for deleting the protected address
	addressRepo.On("DeleteById", ctx, prAddrId).Return(nil)

	err := service.DeleteById(ctx, user, prAddrId)

	assert.NoError(t, err)
	addressRepo.AssertExpectations(t)
	chainRepo.AssertExpectations(t)
}

func TestProtectedAddrService_DeleteById_NotAuthorized(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
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

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:    prAddrId,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)

	err := service.DeleteById(ctx, nonOwner, prAddrId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
}

func TestProtectedAddrService_DeleteById_InvalidId(t *testing.T) {
	service, _, _ := setupProtectedAddrService(t)
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

func TestProtectedAddrService_DeleteById_NotFound(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()

	addressRepo.On("GetById", ctx, prAddrId).Return(entities.Address{}, entities.ErrNotFound)

	err := service.DeleteById(ctx, user, prAddrId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}
