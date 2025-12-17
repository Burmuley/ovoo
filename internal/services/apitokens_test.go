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

func setupApiTokensService(t *testing.T) (*ApiTokensService, *MockApiTokensRepo) {
	tokensRepo := new(MockApiTokensRepo)

	repof := &factory.RepoFactory{
		ApiTokens: tokensRepo,
	}

	service, err := NewApiTokensService(repof)
	require.NoError(t, err)

	return service, tokensRepo
}

func TestNewApiTokensService(t *testing.T) {
	repof := &factory.RepoFactory{}
	service, err := NewApiTokensService(repof)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewApiTokensService_NilRepoFactory(t *testing.T) {
	service, err := NewApiTokensService(nil)

	assert.Error(t, err)
	assert.Nil(t, service)
	assert.ErrorIs(t, err, entities.ErrConfiguration)
}

func TestApiTokensService_GetById_Success(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	tokenId := entities.NewId()
	expectedToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Test Token",
		Description: "Test description",
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(expectedToken, nil)

	token, err := service.GetById(ctx, tokenId)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_GetById_NotFound(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	tokenId := entities.NewId()

	tokensRepo.On("GetById", ctx, tokenId).Return(entities.ApiToken{}, entities.ErrNotFound)

	token, err := service.GetById(ctx, tokenId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_GetByIdCurUser_Success_Owner(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	userId := entities.NewId()
	owner := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	tokenId := entities.NewId()
	expectedToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Test Token",
		Description: "Test description",
		Owner:       owner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(expectedToken, nil)

	token, err := service.GetByIdCurUser(ctx, owner, tokenId)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_GetByIdCurUser_Success_Admin(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	tokenOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	tokenId := entities.NewId()
	expectedToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Test Token",
		Description: "Test description",
		Owner:       tokenOwner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(expectedToken, nil)

	token, err := service.GetByIdCurUser(ctx, admin, tokenId)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_GetByIdCurUser_NotAuthorized(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	nonOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "nonowner@test.com",
	}

	tokenOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	tokenId := entities.NewId()
	existingToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Test Token",
		Description: "Test description",
		Owner:       tokenOwner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(existingToken, nil)

	token, err := service.GetByIdCurUser(ctx, nonOwner, tokenId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_GetAll_Success(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	userId := entities.NewId()
	user := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	expectedTokens := []entities.ApiToken{
		{ID: entities.NewId(), Name: "Token 1", Owner: user},
		{ID: entities.NewId(), Name: "Token 2", Owner: user},
	}

	tokensRepo.On("GetAllForUser", ctx, userId, mock.AnythingOfType("entities.ApiTokenFilter")).Return(expectedTokens, nil)

	tokens, err := service.GetAll(ctx, user)

	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, tokens)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_Create_Success(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	cmd := ApiTokenCreateCmd{
		Name:        "Test Token",
		Description: "Test description",
		ExpireIn:    30, // 30 days
	}

	tokensRepo.On("Create", ctx, mock.AnythingOfType("entities.ApiToken")).Return(nil)

	token, err := service.Create(ctx, user, cmd)

	assert.NoError(t, err)
	assert.Equal(t, cmd.Name, token.Name)
	assert.Equal(t, cmd.Description, token.Description)
	assert.True(t, token.Active)
	assert.NotEmpty(t, token.ID)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_Create_EmptyName(t *testing.T) {
	service, _ := setupApiTokensService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	cmd := ApiTokenCreateCmd{
		Name:        "",
		Description: "Test description",
		ExpireIn:    30,
	}

	token, err := service.Create(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Contains(t, err.Error(), "name field cannot be empty")
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_Create_WhitespaceName(t *testing.T) {
	service, _ := setupApiTokensService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	cmd := ApiTokenCreateCmd{
		Name:        "   ",
		Description: "Test description",
		ExpireIn:    30,
	}

	token, err := service.Create(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Contains(t, err.Error(), "name field cannot be empty")
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_Create_InvalidExpireIn(t *testing.T) {
	service, _ := setupApiTokensService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	cmd := ApiTokenCreateCmd{
		Name:        "Test Token",
		Description: "Test description",
		ExpireIn:    0, // Invalid
	}

	token, err := service.Create(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Contains(t, err.Error(), "expire_in value cannot be less than 1")
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_Create_NegativeExpireIn(t *testing.T) {
	service, _ := setupApiTokensService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	cmd := ApiTokenCreateCmd{
		Name:        "Test Token",
		Description: "Test description",
		ExpireIn:    -5, // Invalid
	}

	token, err := service.Create(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Contains(t, err.Error(), "expire_in value cannot be less than 1")
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_Update_Success(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	userId := entities.NewId()
	owner := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	tokenId := entities.NewId()
	existingToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Old Name",
		Description: "Old description",
		TokenHash:   "somehashvalue1234567890abcdef1234567890abcdef1234567890abcdef12",
		Salt:        "somesaltvalue",
		Owner:       owner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	newName := "New Name"
	newDescription := "New description"
	cmd := ApiTokenUpdateCmd{
		TokenId:     tokenId,
		Name:        &newName,
		Description: &newDescription,
	}

	updatedToken := existingToken
	updatedToken.Name = newName
	updatedToken.Description = newDescription

	tokensRepo.On("GetById", ctx, tokenId).Return(existingToken, nil)
	tokensRepo.On("Update", ctx, mock.AnythingOfType("entities.ApiToken")).Return(updatedToken, nil)

	token, err := service.Update(ctx, owner, cmd)

	assert.NoError(t, err)
	assert.Equal(t, newName, token.Name)
	assert.Equal(t, newDescription, token.Description)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_Update_NotAuthorized(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	nonOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "nonowner@test.com",
	}

	tokenOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	tokenId := entities.NewId()
	existingToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Test Token",
		Description: "Test description",
		Owner:       tokenOwner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	newName := "New Name"
	cmd := ApiTokenUpdateCmd{
		TokenId: tokenId,
		Name:    &newName,
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(existingToken, nil)

	token, err := service.Update(ctx, nonOwner, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_Update_TokenNotFound(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	tokenId := entities.NewId()
	newName := "New Name"
	cmd := ApiTokenUpdateCmd{
		TokenId: tokenId,
		Name:    &newName,
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(entities.ApiToken{}, entities.ErrNotFound)

	token, err := service.Update(ctx, user, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_Update_CannotActivateExpiredToken(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	userId := entities.NewId()
	owner := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	tokenId := entities.NewId()
	existingToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Expired Token",
		Description: "This token is expired",
		Owner:       owner,
		Active:      false,
		Expiration:  time.Now().Add(-24 * time.Hour), // Expired
	}

	activate := true
	cmd := ApiTokenUpdateCmd{
		TokenId: tokenId,
		Active:  &activate,
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(existingToken, nil)

	token, err := service.Update(ctx, owner, cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrValidation)
	assert.Contains(t, err.Error(), "can not activate expired token")
	assert.Equal(t, entities.ApiToken{}, token)
}

func TestApiTokensService_Update_DeactivateToken(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	userId := entities.NewId()
	owner := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	tokenId := entities.NewId()
	existingToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Active Token",
		Description: "This token is active",
		TokenHash:   "somehashvalue1234567890abcdef1234567890abcdef1234567890abcdef12",
		Salt:        "somesaltvalue",
		Owner:       owner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	deactivate := false
	cmd := ApiTokenUpdateCmd{
		TokenId: tokenId,
		Active:  &deactivate,
	}

	updatedToken := existingToken
	updatedToken.Active = false

	tokensRepo.On("GetById", ctx, tokenId).Return(existingToken, nil)
	tokensRepo.On("Update", ctx, mock.AnythingOfType("entities.ApiToken")).Return(updatedToken, nil)

	token, err := service.Update(ctx, owner, cmd)

	assert.NoError(t, err)
	assert.False(t, token.Active)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_Delete_Success_Owner(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	userId := entities.NewId()
	owner := entities.User{
		ID:    userId,
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	tokenId := entities.NewId()
	existingToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Test Token",
		Description: "Test description",
		Owner:       owner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(existingToken, nil)
	tokensRepo.On("Delete", ctx, tokenId).Return(nil)

	_, err := service.Delete(ctx, owner, tokenId)

	assert.NoError(t, err)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_Delete_Success_Admin(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	admin := entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}

	tokenOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	tokenId := entities.NewId()
	existingToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Test Token",
		Description: "Test description",
		Owner:       tokenOwner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(existingToken, nil)
	tokensRepo.On("Delete", ctx, tokenId).Return(nil)

	_, err := service.Delete(ctx, admin, tokenId)

	assert.NoError(t, err)
	tokensRepo.AssertExpectations(t)
}

func TestApiTokensService_Delete_NotAuthorized(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	nonOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "nonowner@test.com",
	}

	tokenOwner := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "owner@test.com",
	}

	tokenId := entities.NewId()
	existingToken := entities.ApiToken{
		ID:          tokenId,
		Name:        "Test Token",
		Description: "Test description",
		Owner:       tokenOwner,
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	tokensRepo.On("GetById", ctx, tokenId).Return(existingToken, nil)

	_, err := service.Delete(ctx, nonOwner, tokenId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotAuthorized)
}

func TestApiTokensService_Delete_TokenNotFound(t *testing.T) {
	service, tokensRepo := setupApiTokensService(t)
	ctx := context.Background()

	user := entities.User{
		ID:    entities.NewId(),
		Type:  entities.RegularUser,
		Login: "user@test.com",
	}

	tokenId := entities.NewId()

	tokensRepo.On("GetById", ctx, tokenId).Return(entities.ApiToken{}, entities.ErrNotFound)

	_, err := service.Delete(ctx, user, tokenId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrNotFound)
}
