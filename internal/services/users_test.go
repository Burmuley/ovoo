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

func setupUsersService(t *testing.T) (*UsersService, *MockUsersRepo, *MockAddressRepo, *MockApiTokensRepo, *MockChainRepo) {
	usersRepo := new(MockUsersRepo)
	addressRepo := new(MockAddressRepo)
	tokensRepo := new(MockApiTokensRepo)
	chainRepo := new(MockChainRepo)

	repof := &factory.RepoFactory{
		Users:     usersRepo,
		Address:   addressRepo,
		ApiTokens: tokensRepo,
		Chain:     chainRepo,
	}

	service, err := NewUsersService(repof)
	require.NoError(t, err)

	return service, usersRepo, addressRepo, tokensRepo, chainRepo
}

func TestNewUsersService(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewUsersService(repof)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewUsersService_NilRepoFactory(t *testing.T) {
	service, err := NewUsersService(nil)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

func TestUsersService_Create_Success(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	cmd := UserCreateCmd{
		FirstName: "John",
		LastName:  "Doe",
		Login:     "john@test.com",
		Type:      entities.RegularUser,
	}

	usersRepo.On("Create", ctx, mock.AnythingOfType("entities.User")).Return(nil)

	user, err := service.Create(ctx, adminUser, cmd)

	assert.NoError(t, err)
	assert.Equal(t, cmd.FirstName, user.FirstName)
	assert.Equal(t, cmd.LastName, user.LastName)
	assert.Equal(t, cmd.Login, user.Login)
	assert.Equal(t, cmd.Type, user.Type)
	assert.NotEmpty(t, user.ID)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_Create_NotAuthorized(t *testing.T) {
	service, _, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	regularUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	cmd := UserCreateCmd{
		FirstName: "John",
		LastName:  "Doe",
		Login:     "john@test.com",
		Type:      entities.RegularUser,
	}

	user, err := service.Create(ctx, regularUser, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_Create_ValidationError(t *testing.T) {
	service, _, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	// Empty login should fail validation
	cmd := UserCreateCmd{
		FirstName: "John",
		LastName:  "Doe",
		Login:     "",
		Type:      entities.RegularUser,
	}

	user, err := service.Create(ctx, adminUser, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_CreatePriv_Success(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	user := entities.User{
		FirstName:    "John",
		LastName:     "Doe",
		Login:        "john@test.com",
		Type:         entities.RegularUser,
		PasswordHash: "password123",
	}

	usersRepo.On("Create", ctx, mock.AnythingOfType("entities.User")).Return(nil)

	createdUser, err := service.CreatePriv(ctx, user)

	assert.NoError(t, err)
	assert.Equal(t, user.FirstName, createdUser.FirstName)
	assert.Equal(t, user.LastName, createdUser.LastName)
	assert.Equal(t, user.Login, createdUser.Login)
	assert.NotEmpty(t, createdUser.ID)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_Update_Success(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	existingUser := entities.User{
		ID:        userId,
		FirstName: "John",
		LastName:  "Doe",
		Login:     "john@test.com",
		Type:      entities.RegularUser,
	}

	newFirstName := "Jane"
	newLastName := "Smith"
	cmd := UserUpdateCmd{
		UserID:    userId,
		FirstName: &newFirstName,
		LastName:  &newLastName,
	}

	usersRepo.On("GetById", ctx, userId).Return(existingUser, nil)
	usersRepo.On("Update", ctx, mock.AnythingOfType("entities.User")).Return(nil)

	updatedUser, err := service.Update(ctx, adminUser, cmd)

	assert.NoError(t, err)
	assert.Equal(t, newFirstName, updatedUser.FirstName)
	assert.Equal(t, newLastName, updatedUser.LastName)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_Update_NotAuthorized(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	regularUser := entities.User{
		ID:    entities.NewId(), // Different ID
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	existingUser := entities.User{
		ID:        userId,
		FirstName: "John",
		LastName:  "Doe",
		Login:     "john@test.com",
		Type:      entities.RegularUser,
	}

	newFirstName := "Jane"
	cmd := UserUpdateCmd{
		UserID:    userId,
		FirstName: &newFirstName,
	}

	usersRepo.On("GetById", ctx, userId).Return(existingUser, nil)

	user, err := service.Update(ctx, regularUser, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_Update_UserNotFound(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	newFirstName := "Jane"
	cmd := UserUpdateCmd{
		UserID:    userId,
		FirstName: &newFirstName,
	}

	usersRepo.On("GetById", ctx, userId).Return(entities.User{}, entities.ErrNotFound)

	user, err := service.Update(ctx, adminUser, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_Update_RegularUserSelf(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	regularUser := entities.User{
		ID:    userId, // Same ID
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	existingUser := entities.User{
		ID:        userId,
		FirstName: "John",
		LastName:  "Doe",
		Login:     "user@test.com",
		Type:      entities.RegularUser,
	}

	newFirstName := "Jane"
	cmd := UserUpdateCmd{
		UserID:    userId,
		FirstName: &newFirstName,
	}

	usersRepo.On("GetById", ctx, userId).Return(existingUser, nil)
	usersRepo.On("Update", ctx, mock.AnythingOfType("entities.User")).Return(nil)

	updatedUser, err := service.Update(ctx, regularUser, cmd)

	assert.NoError(t, err)
	assert.Equal(t, newFirstName, updatedUser.FirstName)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_Delete_Success(t *testing.T) {
	service, usersRepo, addressRepo, tokensRepo, chainRepo := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	targetUser := entities.User{
		ID:        userId,
		FirstName: "John",
		LastName:  "Doe",
		Login:     "john@test.com",
		Type:      entities.RegularUser,
	}

	usersRepo.On("GetById", ctx, userId).Return(targetUser, nil)

	// Mock for deletePrAddrsForUser
	addressRepo.On("GetAll", ctx, mock.AnythingOfType("entities.AddressFilter")).Return(
		[]entities.Address{},
		entities.PaginationMetadata{},
		nil,
	)
	addressRepo.On("BatchDeleteById", ctx, mock.AnythingOfType("[]entities.Id")).Return(nil)

	// Mock for deleting API tokens
	tokensRepo.On("BatchDeleteForUser", ctx, userId).Return(nil)

	// Mock for deleting user
	usersRepo.On("Delete", ctx, userId).Return(nil)

	// Prevent unused mock warning
	_ = chainRepo

	deletedUser, err := service.Delete(ctx, adminUser, userId)

	assert.NoError(t, err)
	assert.Equal(t, targetUser, deletedUser)
	usersRepo.AssertExpectations(t)
	tokensRepo.AssertExpectations(t)
}

func TestUsersService_Delete_NotAuthorized_RegularUser(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	regularUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	targetUser := entities.User{
		ID:        userId,
		FirstName: "John",
		LastName:  "Doe",
		Login:     "john@test.com",
		Type:      entities.RegularUser,
	}

	usersRepo.On("GetById", ctx, userId).Return(targetUser, nil)

	user, err := service.Delete(ctx, regularUser, userId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_Delete_AdminCannotDeleteSelf(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	adminId := entities.NewId()
	adminUser := entities.User{
		ID:    adminId,
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	usersRepo.On("GetById", ctx, adminId).Return(adminUser, nil)

	user, err := service.Delete(ctx, adminUser, adminId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_Delete_InvalidId(t *testing.T) {
	service, _, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	invalidId := entities.Id("invalid")

	user, err := service.Delete(ctx, adminUser, invalidId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_GetById_Success(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	expectedUser := entities.User{
		ID:        userId,
		FirstName: "John",
		LastName:  "Doe",
		Login:     "john@test.com",
		Type:      entities.RegularUser,
	}

	usersRepo.On("GetById", ctx, userId).Return(expectedUser, nil)

	user, err := service.GetById(ctx, adminUser, userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_GetById_NotAuthorized(t *testing.T) {
	service, _, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	regularUser := entities.User{
		ID:    entities.NewId(), // Different ID
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	user, err := service.GetById(ctx, regularUser, userId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_GetById_RegularUserSelf(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	regularUser := entities.User{
		ID:    userId, // Same ID
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	expectedUser := entities.User{
		ID:        userId,
		FirstName: "John",
		LastName:  "Doe",
		Login:     "user@test.com",
		Type:      entities.RegularUser,
	}

	usersRepo.On("GetById", ctx, userId).Return(expectedUser, nil)

	user, err := service.GetById(ctx, regularUser, userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_GetById_InvalidId(t *testing.T) {
	service, _, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	invalidId := entities.Id("")

	user, err := service.GetById(ctx, adminUser, invalidId)

	assert.Error(t, err)
	// Empty ID fails validation after authorization check passes for admin
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_GetByIdPriv_Success(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	expectedUser := entities.User{
		ID:        userId,
		FirstName: "John",
		LastName:  "Doe",
		Login:     "john@test.com",
		Type:      entities.RegularUser,
	}

	usersRepo.On("GetById", ctx, userId).Return(expectedUser, nil)

	user, err := service.GetByIdPriv(ctx, adminUser, userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_GetByLogin_Success(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	login := "john@test.com"
	expectedUser := entities.User{
		ID:        entities.NewId(),
		FirstName: "John",
		LastName:  "Doe",
		Login:     login,
		Type:      entities.RegularUser,
	}

	usersRepo.On("GetByLogin", ctx, login).Return(expectedUser, nil)

	user, err := service.GetByLogin(ctx, login)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_GetByLogin_NotFound(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	login := "nonexistent@test.com"

	usersRepo.On("GetByLogin", ctx, login).Return(entities.User{}, entities.ErrNotFound)

	user, err := service.GetByLogin(ctx, login)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	assert.Equal(t, entities.User{}, user)
}

func TestUsersService_GetAll_AdminUser(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	adminUser := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	filters := map[string][]string{
		"page":      {"1"},
		"page_size": {"10"},
	}

	expectedUsers := []entities.User{
		{ID: entities.NewId(), Login: "user1@test.com"},
		{ID: entities.NewId(), Login: "user2@test.com"},
	}
	expectedMetadata := entities.PaginationMetadata{
		CurrentPage:  1,
		PageSize:     10,
		TotalRecords: 2,
	}

	usersRepo.On("GetAll", ctx, mock.AnythingOfType("entities.UserFilter")).Return(expectedUsers, expectedMetadata, nil)

	users, metadata, err := service.GetAll(ctx, adminUser, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
	assert.Equal(t, expectedMetadata, metadata)
	usersRepo.AssertExpectations(t)
}

func TestUsersService_GetAll_RegularUser(t *testing.T) {
	service, usersRepo, _, _, _ := setupUsersService(t)
	ctx := context.Background()

	userId := entities.NewId()
	regularUser := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	filters := map[string][]string{
		"page":      {"1"},
		"page_size": {"10"},
	}

	expectedUsers := []entities.User{
		{ID: userId, Login: "user@test.com"},
	}
	expectedMetadata := entities.PaginationMetadata{
		CurrentPage:  1,
		PageSize:     10,
		TotalRecords: 1,
	}

	// For regular user, only their own record should be returned
	usersRepo.On("GetAll", ctx, mock.AnythingOfType("entities.UserFilter")).Return(expectedUsers, expectedMetadata, nil)

	users, metadata, err := service.GetAll(ctx, regularUser, filters)

	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
	assert.Equal(t, expectedMetadata, metadata)
	usersRepo.AssertExpectations(t)
}
