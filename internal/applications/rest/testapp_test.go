package rest

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
	"github.com/Burmuley/ovoo/internal/services"
)

type testApp struct {
	app        *Application
	addrRepo   *mockAddressRepo
	chainRepo  *mockChainRepo
	usersRepo  *mockUsersRepo
	tokensRepo *mockTokensRepo
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	ta := &testApp{
		addrRepo:   new(mockAddressRepo),
		chainRepo:  new(mockChainRepo),
		usersRepo:  new(mockUsersRepo),
		tokensRepo: new(mockTokensRepo),
	}
	repof := &factory.RepoFactory{
		Address:   ta.addrRepo,
		Chain:     ta.chainRepo,
		Users:     ta.usersRepo,
		ApiTokens: ta.tokensRepo,
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
	ta.app = &Application{
		svcGw:  gw,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	return ta
}

// testUserFull returns an admin user with all fields needed to pass entity validation.
func testUserFull() entities.User {
	return entities.User{
		ID:    entities.NewId(),
		Type:  entities.AdminUser,
		Login: "admin@test.com",
	}
}

func testToken(ownerID entities.Id) entities.ApiToken {
	return entities.ApiToken{
		ID:          entities.NewId(),
		Name:        "test-token",
		Description: "a test token",
		Active:      true,
		// Login required to pass entities.ApiToken.Validate() → Owner.Validate()
		Owner:      entities.User{ID: ownerID, Type: entities.AdminUser, Login: "admin@test.com"},
		Expiration: time.Now().Add(24 * time.Hour),
		TokenHash:  "somehash",
	}
}

func testChain() entities.Chain {
	owner := entities.User{ID: entities.NewId(), Type: entities.AdminUser, Active: true}
	origFrom := entities.Address{
		ID:    entities.NewId(),
		Email: "external@example.com",
		Type:  entities.ExternalAddress,
		Owner: owner,
	}
	prAddr := &entities.Address{
		ID:     entities.NewId(),
		Email:  "protected@example.com",
		Type:   entities.ProtectedAddress,
		Owner:  owner,
		Active: true,
	}
	alias := entities.Address{
		ID:             entities.NewId(),
		Email:          "alias@test.com",
		Type:           entities.AliasAddress,
		Owner:          owner,
		ForwardAddress: prAddr,
		Active:         true,
	}
	hash := entities.NewHash(string(origFrom.Email), string(alias.Email))
	return entities.Chain{
		Hash:            hash,
		FromAddress:     origFrom,
		ToAddress:       alias,
		OrigFromAddress: origFrom,
		OrigToAddress:   alias,
	}
}
