package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
	gomail "github.com/wneessen/go-mail"
)

func testSMTPCfg() config.MailNotificationConfig {
	return config.MailNotificationConfig{
		SMTPHost:     "localhost",
		SMTPPort:     25,
		FromAddress:  "noreply@example.com",
		FromName:     "Ovoo",
		OvooHostname: "example.com",
	}
}

func setupProtectedAddrService(t *testing.T) (*ProtectedAddrService, *MockAddressRepo, *MockChainRepo) {
	addressRepo := new(MockAddressRepo)
	chainRepo := new(MockChainRepo)

	repof := &factory.RepoFactory{
		Address: addressRepo,
		Chain:   chainRepo,
	}

	service, err := NewProtectedAddrService(repof, "", testSMTPCfg(), nil)
	require.NoError(t, err)

	return service, addressRepo, chainRepo
}

func TestNewProtectedAddrService(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewProtectedAddrService(repof, "", testSMTPCfg(), nil)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewProtectedAddrService_NilRepoFactory(t *testing.T) {
	service, err := NewProtectedAddrService(nil, "", testSMTPCfg(), nil)

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

	filter := entities.AddressFilter{
		Filter: entities.Filter{Page: 1, PageSize: 10},
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

	praddrs, metadata, err := service.GetAll(ctx, admin, filter)

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

	filter := entities.AddressFilter{
		Filter: entities.Filter{Page: 1, PageSize: 10},
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

	praddrs, metadata, err := service.GetAll(ctx, user, filter)

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
	addressRepo.On("DeleteById", ctx, mock.Anything, prAddrId).Return(nil)

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
	addressRepo.On("DeleteById", ctx, mock.Anything, prAddrId).Return(nil)

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
	chainRepo.On("BatchDelete", ctx, mock.Anything, mock.AnythingOfType("[]entities.Hash")).Return(nil)
	addressRepo.On("BatchDeleteById", ctx, mock.Anything, mock.AnythingOfType("[]entities.Id")).Return(nil).Twice()

	// Mock for deleting the protected address
	addressRepo.On("DeleteById", ctx, mock.Anything, prAddrId).Return(nil)

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

func TestProtectedAddrService_Update_Activate(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:     prAddrId,
		Type:   entities.ProtectedAddress,
		Email:  "protected@example.com",
		Owner:  user,
		Active: false,
	}

	activate := true
	cmd := PrAddrUpdateCmd{
		PrAddrId: prAddrId,
		Active:   &activate,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)
	addressRepo.On("Update", ctx, mock.MatchedBy(func(a entities.Address) bool {
		return a.ID == prAddrId && a.Active
	})).Return(nil)

	result, err := service.Update(ctx, user, cmd)

	assert.NoError(t, err)
	assert.True(t, result.Active)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_Update_Deactivate(t *testing.T) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	prAddrId := entities.NewId()
	existingPrAddr := entities.Address{
		ID:     prAddrId,
		Type:   entities.ProtectedAddress,
		Email:  "protected@example.com",
		Owner:  user,
		Active: true,
	}

	deactivate := false
	cmd := PrAddrUpdateCmd{
		PrAddrId: prAddrId,
		Active:   &deactivate,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)

	// deactivateAliasesForPrAddr: bulk-deactivate active aliases forwarding to this praddr
	addressRepo.On("BatchUpdate", ctx, mock.AnythingOfType("entities.AddressFilter"), mock.AnythingOfType("entities.AddressBulkUpdateFields")).Return(nil)

	// Update praddr to inactive
	addressRepo.On("Update", ctx, mock.MatchedBy(func(a entities.Address) bool {
		return a.ID == prAddrId && !a.Active
	})).Return(nil)

	result, err := service.Update(ctx, user, cmd)

	assert.NoError(t, err)
	assert.False(t, result.Active)
	addressRepo.AssertExpectations(t)
}

func TestProtectedAddrService_Update_SetActive_NotAuthorized(t *testing.T) {
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
		ID:     prAddrId,
		Type:   entities.ProtectedAddress,
		Email:  "protected@example.com",
		Owner:  owner,
		Active: true,
	}

	deactivate := false
	cmd := PrAddrUpdateCmd{
		PrAddrId: prAddrId,
		Active:   &deactivate,
	}

	addressRepo.On("GetById", ctx, prAddrId).Return(existingPrAddr, nil)

	_, err := service.Update(ctx, nonOwner, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	addressRepo.AssertNotCalled(t, "Update")
}

func sendVerifyFixture(t *testing.T, mutate func(*entities.Address)) (*ProtectedAddrService, entities.User, entities.Address) {
	service, addressRepo, _ := setupProtectedAddrService(t)
	owner := entities.User{ID: entities.NewId(), Type: entities.RegularUser, Login: "owner@test.com"}
	addr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: owner,
	}
	if mutate != nil {
		mutate(&addr)
	}
	addressRepo.On("GetById", mock.Anything, addr.ID).Return(addr, nil)
	return service, owner, addr
}

func TestProtectedAddrService_SendVerifyEmail_NotOwner(t *testing.T) {
	service, _, addr := sendVerifyFixture(t, nil)
	other := entities.User{ID: entities.NewId(), Type: entities.RegularUser, Login: "other@test.com"}

	err := service.SendVerifyEmail(context.Background(), other, addr.ID)

	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Empty(t, service.verifySentAt)
}

func TestProtectedAddrService_SendVerifyEmail_NotProtectedAddress(t *testing.T) {
	service, owner, addr := sendVerifyFixture(t, func(a *entities.Address) { a.Type = entities.AliasAddress })

	err := service.SendVerifyEmail(context.Background(), owner, addr.ID)

	assert.ErrorIs(t, err, entities.ErrValidation)
}

func TestProtectedAddrService_SendVerifyEmail_AlreadyVerified(t *testing.T) {
	service, owner, addr := sendVerifyFixture(t, func(a *entities.Address) { a.Verified = true })

	err := service.SendVerifyEmail(context.Background(), owner, addr.ID)

	assert.ErrorIs(t, err, entities.ErrValidation)
}

func TestProtectedAddrService_SendVerifyEmail_Throttled(t *testing.T) {
	service, owner, addr := sendVerifyFixture(t, nil)
	require.True(t, service.allowVerifySend(addr.ID, time.Now()))

	err := service.SendVerifyEmail(context.Background(), owner, addr.ID)

	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Contains(t, err.Error(), "try again later")
}

func TestProtectedAddrService_AllowVerifySend(t *testing.T) {
	service, _, _ := setupProtectedAddrService(t)
	id1, id2 := entities.NewId(), entities.NewId()
	now := time.Now()

	assert.True(t, service.allowVerifySend(id1, now))
	assert.False(t, service.allowVerifySend(id1, now.Add(verifyResendInterval-time.Second)))
	assert.True(t, service.allowVerifySend(id2, now), "other addresses are not affected")
	assert.True(t, service.allowVerifySend(id1, now.Add(verifyResendInterval)))
	assert.NotContains(t, service.verifySentAt, id2, "expired entries are dropped")
}

type validateTokenEnv struct {
	service   *ProtectedAddrService
	addrRepo  *MockAddressRepo
	tokenRepo *MockAddressVerifyRepo
}

func setupValidateTokenEnv(t *testing.T) validateTokenEnv {
	t.Helper()
	addrRepo := new(MockAddressRepo)
	tokenRepo := new(MockAddressVerifyRepo)
	repof := &factory.RepoFactory{Address: addrRepo, Chain: new(MockChainRepo), AddressVerify: tokenRepo}
	service, err := NewProtectedAddrService(repof, "", testSMTPCfg(), nil)
	require.NoError(t, err)
	return validateTokenEnv{service, addrRepo, tokenRepo}
}

// issueToken creates a verification token for addrId and registers it in the token repo mock.
func (e validateTokenEnv) issueToken(t *testing.T, addrId entities.Id) string {
	t.Helper()
	token, userToken, err := entities.NewAddressVerifyToken(time.Now().Add(time.Hour), addrId)
	require.NoError(t, err)
	e.tokenRepo.On("GetById", mock.Anything, token.ID).Return(token, nil)
	e.tokenRepo.On("Delete", mock.Anything, mock.Anything, token.ID).Return(nil).Maybe()
	return userToken
}

func testPrAddr(owner entities.User, verified bool) entities.Address {
	return entities.Address{
		ID:       entities.NewId(),
		Type:     entities.ProtectedAddress,
		Email:    "protected@example.com",
		Owner:    owner,
		Verified: verified,
	}
}

func TestProtectedAddrService_ValidateVerifyToken_Success(t *testing.T) {
	e := setupValidateTokenEnv(t)
	owner := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	addr := testPrAddr(owner, false)
	tokenStr := e.issueToken(t, addr.ID)
	e.addrRepo.On("GetById", mock.Anything, addr.ID).Return(addr, nil)
	e.addrRepo.On("Update", mock.Anything, mock.MatchedBy(func(a entities.Address) bool { return a.Verified })).Return(nil)

	got, err := e.service.ValidateVerifyToken(context.Background(), owner, addr.ID, tokenStr)

	require.NoError(t, err)
	assert.True(t, got.Verified)
	e.addrRepo.AssertExpectations(t)
}

func TestProtectedAddrService_ValidateVerifyToken_VerifiedAddressOwner(t *testing.T) {
	e := setupValidateTokenEnv(t)
	owner := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	addr := testPrAddr(owner, true)
	tokenStr := e.issueToken(t, addr.ID)
	e.addrRepo.On("GetById", mock.Anything, addr.ID).Return(addr, nil)

	got, err := e.service.ValidateVerifyToken(context.Background(), owner, addr.ID, tokenStr)

	require.NoError(t, err)
	assert.Equal(t, addr.ID, got.ID)
	e.addrRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// F19/F27: a valid token for the attacker's own address must not unlock a victim's verified address
func TestProtectedAddrService_ValidateVerifyToken_TokenForOtherAddress(t *testing.T) {
	e := setupValidateTokenEnv(t)
	attacker := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	victim := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	own := testPrAddr(attacker, false)
	victimAddr := testPrAddr(victim, true)
	tokenStr := e.issueToken(t, own.ID)

	got, err := e.service.ValidateVerifyToken(context.Background(), attacker, victimAddr.ID, tokenStr)

	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.Address{}, got)
	e.addrRepo.AssertNotCalled(t, "GetById", mock.Anything, mock.Anything)
	e.addrRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// a user who holds a valid token for the address but does not own it gets nothing and verifies nothing
func TestProtectedAddrService_ValidateVerifyToken_NotOwner(t *testing.T) {
	for _, verified := range []bool{false, true} {
		e := setupValidateTokenEnv(t)
		owner := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
		other := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
		addr := testPrAddr(owner, verified)
		tokenStr := e.issueToken(t, addr.ID)
		e.addrRepo.On("GetById", mock.Anything, addr.ID).Return(addr, nil)

		got, err := e.service.ValidateVerifyToken(context.Background(), other, addr.ID, tokenStr)

		assert.ErrorIs(t, err, entities.ErrNotAuthorized, "verified=%v", verified)
		assert.Equal(t, entities.Address{}, got)
		e.addrRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	}
}

func TestProtectedAddrService_ValidateVerifyToken_Admin(t *testing.T) {
	e := setupValidateTokenEnv(t)
	owner := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	admin := entities.User{ID: entities.NewId(), Type: entities.AdminUser}
	addr := testPrAddr(owner, true)
	tokenStr := e.issueToken(t, addr.ID)
	e.addrRepo.On("GetById", mock.Anything, addr.ID).Return(addr, nil)

	_, err := e.service.ValidateVerifyToken(context.Background(), admin, addr.ID, tokenStr)

	assert.NoError(t, err)
}

func TestSMTPTLSSettings(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		mode    string
		policy  gomail.TLSPolicy
		ssl     bool
		wantErr bool
	}{
		{"remote host defaults to mandatory", "smtp.example.com", "", gomail.TLSMandatory, false, false},
		{"localhost defaults to opportunistic", "localhost", "", gomail.TLSOpportunistic, false, false},
		{"loopback IPv4 defaults to opportunistic", "127.0.0.1", "", gomail.TLSOpportunistic, false, false},
		{"loopback IPv6 defaults to opportunistic", "::1", "", gomail.TLSOpportunistic, false, false},
		{"mandatory", "smtp.example.com", "mandatory", gomail.TLSMandatory, false, false},
		{"mode is case-insensitive", "smtp.example.com", " Mandatory ", gomail.TLSMandatory, false, false},
		{"ssl", "smtp.example.com", "ssl", gomail.TLSMandatory, true, false},
		{"explicit opportunistic on remote host", "smtp.example.com", "opportunistic", gomail.TLSOpportunistic, false, false},
		{"none", "smtp.example.com", "none", gomail.NoTLS, false, false},
		{"explicit mode wins over loopback default", "localhost", "mandatory", gomail.TLSMandatory, false, false},
		{"unknown mode", "smtp.example.com", "starttls", 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, ssl, err := smtpTLSSettings(config.MailNotificationConfig{SMTPHost: tt.host, TLSMode: tt.mode})
			if tt.wantErr {
				assert.ErrorIs(t, err, entities.ErrConfiguration)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.policy, policy)
			assert.Equal(t, tt.ssl, ssl)
		})
	}
}

func TestNewProtectedAddrService_InvalidTLSMode(t *testing.T) {
	cfg := testSMTPCfg()
	cfg.TLSMode = "bogus"

	service, err := NewProtectedAddrService(&factory.RepoFactory{}, "", cfg, nil)

	assert.ErrorIs(t, err, entities.ErrConfiguration)
	assert.Nil(t, service)
}

func TestSetupSMTPClient_WithCredentials(t *testing.T) {
	cfg := testSMTPCfg()
	cfg.SMTPHost = "smtp.example.com"
	cfg.SMTPUsername = "user"
	cfg.SMTPPassword = "pass"

	client, err := setupSMTPClient(cfg)

	require.NoError(t, err)
	assert.NotNil(t, client)
}
