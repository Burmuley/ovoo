package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

func TestNew_AllServicesProvided(t *testing.T) {
	repof := &factory.RepoFactory{}

	aliasesService := &AliasesService{repof: repof, domain: "test.com", wordsDictionary: []string{"word"}}
	usersService := &UsersService{repof: repof}
	prAddrsService := &ProtectedAddrService{repof: repof}
	chainsService := &ChainsService{repof: repof, domain: "test.com"}
	tokensService := &ApiTokensService{repof: repof}

	gateway, err := New(aliasesService, usersService, prAddrsService, chainsService, tokensService)

	require.NoError(t, err)
	assert.NotNil(t, gateway)
	assert.Equal(t, aliasesService, gateway.Aliases)
	assert.Equal(t, usersService, gateway.Users)
	assert.Equal(t, prAddrsService, gateway.PrAddrs)
	assert.Equal(t, chainsService, gateway.Chains)
	assert.Equal(t, tokensService, gateway.Tokens)
}

func TestNew_MissingService(t *testing.T) {
	repof := &factory.RepoFactory{}

	aliasesService := &AliasesService{repof: repof, domain: "test.com", wordsDictionary: []string{"word"}}
	usersService := &UsersService{repof: repof}
	// Missing prAddrsService, chainsService, tokensService

	gateway, err := New(aliasesService, usersService)

	assert.Error(t, err)
	assert.Nil(t, gateway)
	assert.Contains(t, err.Error(), "can not be nil")
}

func TestNew_UnknownServiceType(t *testing.T) {
	unknownService := struct{ name string }{name: "unknown"}

	gateway, err := New(unknownService)

	assert.Error(t, err)
	assert.Nil(t, gateway)
	assert.Contains(t, err.Error(), "unknown service type")
}

func TestNew_EmptyServices(t *testing.T) {
	gateway, err := New()

	assert.Error(t, err)
	assert.Nil(t, gateway)
	assert.Contains(t, err.Error(), "can not be nil")
}

func TestNew_DuplicateServices(t *testing.T) {
	repof := &factory.RepoFactory{}

	aliasesService1 := &AliasesService{repof: repof, domain: "test1.com", wordsDictionary: []string{"word"}}
	aliasesService2 := &AliasesService{repof: repof, domain: "test2.com", wordsDictionary: []string{"word"}}
	usersService := &UsersService{repof: repof}
	prAddrsService := &ProtectedAddrService{repof: repof}
	chainsService := &ChainsService{repof: repof, domain: "test.com"}
	tokensService := &ApiTokensService{repof: repof}

	// Second aliases service should override the first one
	gateway, err := New(aliasesService1, aliasesService2, usersService, prAddrsService, chainsService, tokensService)

	require.NoError(t, err)
	assert.NotNil(t, gateway)
	// The second aliasesService should be set
	assert.Equal(t, aliasesService2, gateway.Aliases)
}

func TestCheckNilServices_AllFieldsSet(t *testing.T) {
	repof := &factory.RepoFactory{}

	gw := &ServiceGateway{
		Aliases: &AliasesService{repof: repof, domain: "test.com", wordsDictionary: []string{"word"}},
		Users:   &UsersService{repof: repof},
		PrAddrs: &ProtectedAddrService{repof: repof},
		Chains:  &ChainsService{repof: repof, domain: "test.com"},
		Tokens:  &ApiTokensService{repof: repof},
	}

	err := checkNilServices(gw)

	assert.NoError(t, err)
}

func TestCheckNilServices_AliasesNil(t *testing.T) {
	repof := &factory.RepoFactory{}

	gw := &ServiceGateway{
		Aliases: nil,
		Users:   &UsersService{repof: repof},
		PrAddrs: &ProtectedAddrService{repof: repof},
		Chains:  &ChainsService{repof: repof, domain: "test.com"},
		Tokens:  &ApiTokensService{repof: repof},
	}

	err := checkNilServices(gw)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Aliases")
}

func TestCheckNilServices_UsersNil(t *testing.T) {
	repof := &factory.RepoFactory{}

	gw := &ServiceGateway{
		Aliases: &AliasesService{repof: repof, domain: "test.com", wordsDictionary: []string{"word"}},
		Users:   nil,
		PrAddrs: &ProtectedAddrService{repof: repof},
		Chains:  &ChainsService{repof: repof, domain: "test.com"},
		Tokens:  &ApiTokensService{repof: repof},
	}

	err := checkNilServices(gw)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Users")
}

func TestCheckNilServices_PrAddrsNil(t *testing.T) {
	repof := &factory.RepoFactory{}

	gw := &ServiceGateway{
		Aliases: &AliasesService{repof: repof, domain: "test.com", wordsDictionary: []string{"word"}},
		Users:   &UsersService{repof: repof},
		PrAddrs: nil,
		Chains:  &ChainsService{repof: repof, domain: "test.com"},
		Tokens:  &ApiTokensService{repof: repof},
	}

	err := checkNilServices(gw)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PrAddrs")
}

func TestCheckNilServices_ChainsNil(t *testing.T) {
	repof := &factory.RepoFactory{}

	gw := &ServiceGateway{
		Aliases: &AliasesService{repof: repof, domain: "test.com", wordsDictionary: []string{"word"}},
		Users:   &UsersService{repof: repof},
		PrAddrs: &ProtectedAddrService{repof: repof},
		Chains:  nil,
		Tokens:  &ApiTokensService{repof: repof},
	}

	err := checkNilServices(gw)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Chains")
}

func TestCheckNilServices_TokensNil(t *testing.T) {
	repof := &factory.RepoFactory{}

	gw := &ServiceGateway{
		Aliases: &AliasesService{repof: repof, domain: "test.com", wordsDictionary: []string{"word"}},
		Users:   &UsersService{repof: repof},
		PrAddrs: &ProtectedAddrService{repof: repof},
		Chains:  &ChainsService{repof: repof, domain: "test.com"},
		Tokens:  nil,
	}

	err := checkNilServices(gw)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Tokens")
}
