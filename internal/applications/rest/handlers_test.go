package rest

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/applications/rest/middleware"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
	"github.com/Burmuley/ovoo/internal/services"
)

// --- mock repository implementations ---

type mockAddressRepo struct{ mock.Mock }

func (m *mockAddressRepo) GetById(ctx context.Context, id entities.Id) (entities.Address, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entities.Address), args.Error(1)
}
func (m *mockAddressRepo) GetByEmail(ctx context.Context, email entities.Email) ([]entities.Address, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Address), args.Error(1)
}
func (m *mockAddressRepo) GetAll(ctx context.Context, filter entities.AddressFilter) ([]entities.Address, entities.PaginationMetadata, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Address), args.Get(1).(entities.PaginationMetadata), args.Error(2)
}
func (m *mockAddressRepo) Create(ctx context.Context, address entities.Address) error {
	return m.Called(ctx, address).Error(0)
}
func (m *mockAddressRepo) BatchCreate(ctx context.Context, addresses []entities.Address) error {
	return m.Called(ctx, addresses).Error(0)
}
func (m *mockAddressRepo) Update(ctx context.Context, address entities.Address) error {
	return m.Called(ctx, address).Error(0)
}
func (m *mockAddressRepo) DeleteById(ctx context.Context, id entities.Id) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockAddressRepo) BatchDeleteById(ctx context.Context, ids []entities.Id) error {
	return m.Called(ctx, ids).Error(0)
}

type mockChainRepo struct{ mock.Mock }

func (m *mockChainRepo) GetByHash(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(entities.Chain), args.Error(1)
}
func (m *mockChainRepo) GetByFilters(ctx context.Context, filter entities.ChainFilter) ([]entities.Chain, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Chain), args.Error(1)
}
func (m *mockChainRepo) Create(ctx context.Context, chain entities.Chain) error {
	return m.Called(ctx, chain).Error(0)
}
func (m *mockChainRepo) BatchCreate(ctx context.Context, chains []entities.Chain) error {
	return m.Called(ctx, chains).Error(0)
}
func (m *mockChainRepo) Delete(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(entities.Chain), args.Error(1)
}
func (m *mockChainRepo) BatchDelete(ctx context.Context, hashes []entities.Hash) error {
	return m.Called(ctx, hashes).Error(0)
}

type mockUsersRepo struct{ mock.Mock }

func (m *mockUsersRepo) GetAll(ctx context.Context, filter entities.UserFilter) ([]entities.User, entities.PaginationMetadata, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.User), args.Get(1).(entities.PaginationMetadata), args.Error(2)
}
func (m *mockUsersRepo) GetById(ctx context.Context, id entities.Id) (entities.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entities.User), args.Error(1)
}
func (m *mockUsersRepo) GetByLogin(ctx context.Context, login string) (entities.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(entities.User), args.Error(1)
}
func (m *mockUsersRepo) Create(ctx context.Context, user entities.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUsersRepo) BatchCreate(ctx context.Context, users []entities.User) error {
	return m.Called(ctx, users).Error(0)
}
func (m *mockUsersRepo) Update(ctx context.Context, user entities.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUsersRepo) Delete(ctx context.Context, id entities.Id) error {
	return m.Called(ctx, id).Error(0)
}

type mockTokensRepo struct{ mock.Mock }

func (m *mockTokensRepo) GetById(ctx context.Context, tokenId entities.Id) (entities.ApiToken, error) {
	args := m.Called(ctx, tokenId)
	return args.Get(0).(entities.ApiToken), args.Error(1)
}
func (m *mockTokensRepo) GetAll(ctx context.Context, filter entities.ApiTokenFilter) ([]entities.ApiToken, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.ApiToken), args.Error(1)
}
func (m *mockTokensRepo) Create(ctx context.Context, token entities.ApiToken) error {
	return m.Called(ctx, token).Error(0)
}
func (m *mockTokensRepo) Update(ctx context.Context, token entities.ApiToken) (entities.ApiToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(entities.ApiToken), args.Error(1)
}
func (m *mockTokensRepo) BatchCreate(ctx context.Context, tokens []entities.ApiToken) error {
	return m.Called(ctx, tokens).Error(0)
}
func (m *mockTokensRepo) Delete(ctx context.Context, tokenId entities.Id) error {
	return m.Called(ctx, tokenId).Error(0)
}
func (m *mockTokensRepo) BatchDeleteById(ctx context.Context, ids []entities.Id) error {
	return m.Called(ctx, ids).Error(0)
}
func (m *mockTokensRepo) BatchDeleteForUser(ctx context.Context, id entities.Id) error {
	return m.Called(ctx, id).Error(0)
}

// --- test helpers ---

func buildTestApplication(t *testing.T, addrRepo *mockAddressRepo) *Application {
	t.Helper()
	repof := &factory.RepoFactory{
		Address:   addrRepo,
		Chain:     new(mockChainRepo),
		Users:     new(mockUsersRepo),
		ApiTokens: new(mockTokensRepo),
	}

	aliasesSvc, err := services.NewAliasesService("test.com", []string{"alpha", "bravo", "charlie"}, repof)
	require.NoError(t, err)
	prAddrsSvc, err := services.NewProtectedAddrService(repof)
	require.NoError(t, err)
	usersSvc, err := services.NewUsersService(repof)
	require.NoError(t, err)
	chainsSvc, err := services.NewChainsService("test.com", repof)
	require.NoError(t, err)
	tokensSvc, err := services.NewApiTokensService(repof)
	require.NoError(t, err)

	gw := &services.ServiceGateway{
		Aliases: aliasesSvc,
		Users:   usersSvc,
		PrAddrs: prAddrsSvc,
		Chains:  chainsSvc,
		Tokens:  tokensSvc,
	}

	return &Application{
		svcGw:  gw,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func testUser() entities.User {
	return entities.User{
		ID:   entities.NewId(),
		Type: entities.AdminUser,
	}
}

func withUser(r *http.Request, u entities.User) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, u)
	return r.WithContext(ctx)
}

func testAlias(ownerID entities.Id) entities.Address {
	prAddrID := entities.NewId()
	prAddr := entities.Address{
		ID:    prAddrID,
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: entities.User{ID: ownerID, Type: entities.AdminUser},
	}
	return entities.Address{
		ID:             entities.NewId(),
		Type:           entities.AliasAddress,
		Email:          "alias@test.com",
		ForwardAddress: &prAddr,
		Owner:          entities.User{ID: ownerID, Type: entities.AdminUser},
		Active:         true,
	}
}

func testProtectedAddr(ownerID entities.Id) entities.Address {
	return entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: entities.User{ID: ownerID, Type: entities.AdminUser},
		Active: true,
	}
}

// --- UpdateAlias tests ---

func TestUpdateAlias_NilMetadata(t *testing.T) {
	addrRepo := new(mockAddressRepo)
	app := buildTestApplication(t, addrRepo)

	user := testUser()
	alias := testAlias(user.ID)

	addrRepo.On("GetById", mock.Anything, alias.ID).Return(alias, nil)
	addrRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"active": true}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/aliases/"+alias.ID.String(), body)
	req.SetPathValue("id", alias.ID.String())
	req = withUser(req, user)

	w := httptest.NewRecorder()
	require.NotPanics(t, func() { app.UpdateAlias(w, req) })

	assert.Equal(t, http.StatusOK, w.Code)
	addrRepo.AssertExpectations(t)
}

func TestUpdateAlias_WithMetadata(t *testing.T) {
	addrRepo := new(mockAddressRepo)
	app := buildTestApplication(t, addrRepo)

	user := testUser()
	alias := testAlias(user.ID)

	addrRepo.On("GetById", mock.Anything, alias.ID).Return(alias, nil)
	addrRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"metadata": {"comment": "test comment", "service_name": "my-service"}}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/aliases/"+alias.ID.String(), body)
	req.SetPathValue("id", alias.ID.String())
	req = withUser(req, user)

	w := httptest.NewRecorder()
	app.UpdateAlias(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	addrRepo.AssertExpectations(t)
}

func TestUpdateAlias_EmptyBody(t *testing.T) {
	addrRepo := new(mockAddressRepo)
	app := buildTestApplication(t, addrRepo)

	user := testUser()
	alias := testAlias(user.ID)

	addrRepo.On("GetById", mock.Anything, alias.ID).Return(alias, nil)
	addrRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/aliases/"+alias.ID.String(), body)
	req.SetPathValue("id", alias.ID.String())
	req = withUser(req, user)

	w := httptest.NewRecorder()
	require.NotPanics(t, func() { app.UpdateAlias(w, req) })

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- UpdatePrAddr tests ---

func TestUpdatePrAddr_NilMetadata(t *testing.T) {
	addrRepo := new(mockAddressRepo)
	app := buildTestApplication(t, addrRepo)

	user := testUser()
	prAddr := testProtectedAddr(user.ID)

	addrRepo.On("GetById", mock.Anything, prAddr.ID).Return(prAddr, nil)
	addrRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"active": true}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/praddrs/"+prAddr.ID.String(), body)
	req.SetPathValue("id", prAddr.ID.String())
	req = withUser(req, user)

	w := httptest.NewRecorder()
	require.NotPanics(t, func() { app.UpdatePrAddr(w, req) })

	assert.Equal(t, http.StatusOK, w.Code)
	addrRepo.AssertExpectations(t)
}

func TestUpdatePrAddr_WithMetadata(t *testing.T) {
	addrRepo := new(mockAddressRepo)
	app := buildTestApplication(t, addrRepo)

	user := testUser()
	prAddr := testProtectedAddr(user.ID)

	addrRepo.On("GetById", mock.Anything, prAddr.ID).Return(prAddr, nil)
	addrRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"metadata": {"comment": "my notes", "service_name": "acme"}}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/praddrs/"+prAddr.ID.String(), body)
	req.SetPathValue("id", prAddr.ID.String())
	req = withUser(req, user)

	w := httptest.NewRecorder()
	app.UpdatePrAddr(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	addrRepo.AssertExpectations(t)
}

func TestUpdatePrAddr_EmptyBody(t *testing.T) {
	addrRepo := new(mockAddressRepo)
	app := buildTestApplication(t, addrRepo)

	user := testUser()
	prAddr := testProtectedAddr(user.ID)

	addrRepo.On("GetById", mock.Anything, prAddr.ID).Return(prAddr, nil)
	addrRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/praddrs/"+prAddr.ID.String(), body)
	req.SetPathValue("id", prAddr.ID.String())
	req = withUser(req, user)

	w := httptest.NewRecorder()
	require.NotPanics(t, func() { app.UpdatePrAddr(w, req) })

	assert.Equal(t, http.StatusOK, w.Code)
}
