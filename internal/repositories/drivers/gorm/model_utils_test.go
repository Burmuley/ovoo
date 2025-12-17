package gorm

import (
	"testing"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/stretchr/testify/assert"
)

func createTestUser() entities.User {
	return entities.User{
		ID:           entities.NewId(),
		FirstName:    "John",
		LastName:     "Doe",
		Login:        "john.doe@example.com",
		Type:         entities.RegularUser,
		PasswordHash: "hashedpassword123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func createTestAddress(owner entities.User) entities.Address {
	return entities.Address{
		ID:    entities.NewId(),
		Type:  entities.AliasAddress,
		Email: entities.Email("test@example.com"),
		Owner: owner,
		Metadata: entities.AddressMetadata{
			Comment:     "Test comment",
			ServiceName: "TestService",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UpdatedBy: owner,
	}
}

func TestUserFromEntity(t *testing.T) {
	user := createTestUser()

	gormUser := userFromEntity(user)

	assert.Equal(t, string(user.ID), gormUser.ID)
	assert.Equal(t, user.FirstName, gormUser.FirstName)
	assert.Equal(t, user.LastName, gormUser.LastName)
	assert.Equal(t, user.Login, gormUser.Login)
	assert.Equal(t, int(user.Type), gormUser.Type)
	assert.Equal(t, user.PasswordHash, gormUser.PwdHash)
	assert.Equal(t, user.FailedAttempts, gormUser.FailedAttempts)
}

func TestUserFromEntity_WithUpdatedBy(t *testing.T) {
	updatedByUser := createTestUser()
	updatedByUser.ID = entities.NewId()
	updatedByUser.Login = "updater@example.com"

	user := createTestUser()
	user.UpdatedBy = &updatedByUser

	gormUser := userFromEntity(user)

	assert.NotNil(t, gormUser.UpdatedBy)
	assert.Equal(t, string(updatedByUser.ID), gormUser.UpdatedBy.ID)
	assert.Equal(t, updatedByUser.Login, gormUser.UpdatedBy.Login)
}

func TestUserToEntity(t *testing.T) {
	gormUser := User{
		Model: Model{
			ID:        string(entities.NewId()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FirstName:      "Jane",
		LastName:       "Smith",
		Login:          "jane.smith@example.com",
		Type:           int(entities.AdminUser),
		PwdHash:        "hashedpassword456",
		FailedAttempts: 2,
		LockoutUntil:   time.Now().Add(1 * time.Hour),
	}

	user := userToEntity(gormUser)

	assert.Equal(t, entities.Id(gormUser.ID), user.ID)
	assert.Equal(t, gormUser.FirstName, user.FirstName)
	assert.Equal(t, gormUser.LastName, user.LastName)
	assert.Equal(t, gormUser.Login, user.Login)
	assert.Equal(t, entities.UserType(gormUser.Type), user.Type)
	assert.Equal(t, gormUser.PwdHash, user.PasswordHash)
	assert.Equal(t, gormUser.FailedAttempts, user.FailedAttempts)
}

func TestUserFromEntityList(t *testing.T) {
	users := []entities.User{
		createTestUser(),
		createTestUser(),
	}
	users[1].ID = entities.NewId()
	users[1].Login = "user2@example.com"

	gormUsers := userFromEntityList(users)

	assert.Len(t, gormUsers, 2)
	assert.Equal(t, string(users[0].ID), gormUsers[0].ID)
	assert.Equal(t, string(users[1].ID), gormUsers[1].ID)
}

func TestAddressFromEntity(t *testing.T) {
	owner := createTestUser()
	address := createTestAddress(owner)

	gormAddr := addressFromEntity(address)

	assert.Equal(t, string(address.ID), gormAddr.ID)
	assert.Equal(t, string(address.Email), gormAddr.Email)
	assert.Equal(t, int(address.Type), gormAddr.Type)
	assert.Equal(t, address.Metadata.Comment, gormAddr.Metadata.Comment)
	assert.Equal(t, address.Metadata.ServiceName, gormAddr.Metadata.ServiceName)
	assert.Equal(t, string(owner.ID), gormAddr.Owner.ID)
}

func TestAddressFromEntity_WithForwardAddress(t *testing.T) {
	owner := createTestUser()
	forwardAddress := createTestAddress(owner)
	forwardAddress.ID = entities.NewId()
	forwardAddress.Email = entities.Email("forward@example.com")

	address := createTestAddress(owner)
	address.ForwardAddress = &forwardAddress

	gormAddr := addressFromEntity(address)

	assert.NotNil(t, gormAddr.ForwardAddress)
	assert.Equal(t, string(forwardAddress.ID), gormAddr.ForwardAddress.ID)
	assert.Equal(t, string(forwardAddress.Email), gormAddr.ForwardAddress.Email)
}

func TestAddressToEntity(t *testing.T) {
	owner := User{
		Model: Model{ID: string(entities.NewId())},
		Login: "owner@example.com",
	}

	gormAddr := Address{
		Model: Model{
			ID:        string(entities.NewId()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Email: "test@example.com",
		Type:  int(entities.AliasAddress),
		Owner: owner,
		Metadata: AddressMetadata{
			Comment:     "Test comment",
			ServiceName: "TestService",
		},
		UpdatedBy: owner,
	}

	address := addressToEntity(gormAddr)

	assert.Equal(t, entities.Id(gormAddr.ID), address.ID)
	assert.Equal(t, entities.Email(gormAddr.Email), address.Email)
	assert.Equal(t, entities.AddressType(gormAddr.Type), address.Type)
	assert.Equal(t, gormAddr.Metadata.Comment, address.Metadata.Comment)
	assert.Equal(t, gormAddr.Metadata.ServiceName, address.Metadata.ServiceName)
}

func TestAddressFromEntityList(t *testing.T) {
	owner := createTestUser()
	addresses := []entities.Address{
		createTestAddress(owner),
		createTestAddress(owner),
	}
	addresses[1].ID = entities.NewId()
	addresses[1].Email = entities.Email("test2@example.com")

	gormAddrs := addressFromEntityList(addresses)

	assert.Len(t, gormAddrs, 2)
	assert.Equal(t, string(addresses[0].ID), gormAddrs[0].ID)
	assert.Equal(t, string(addresses[1].ID), gormAddrs[1].ID)
}

func TestAddressToEntityList(t *testing.T) {
	owner := User{
		Model: Model{ID: string(entities.NewId())},
		Login: "owner@example.com",
	}

	gormAddrs := []Address{
		{
			Model: Model{ID: string(entities.NewId())},
			Email: "test1@example.com",
			Type:  int(entities.AliasAddress),
			Owner: owner,
		},
		{
			Model: Model{ID: string(entities.NewId())},
			Email: "test2@example.com",
			Type:  int(entities.ProtectedAddress),
			Owner: owner,
		},
	}

	addresses := addressToEntityList(gormAddrs)

	assert.Len(t, addresses, 2)
	assert.Equal(t, entities.Id(gormAddrs[0].ID), addresses[0].ID)
	assert.Equal(t, entities.Id(gormAddrs[1].ID), addresses[1].ID)
}

func TestChainFromEntity(t *testing.T) {
	owner := createTestUser()
	fromAddr := createTestAddress(owner)
	toAddr := createTestAddress(owner)
	toAddr.ID = entities.NewId()
	toAddr.Email = entities.Email("to@example.com")

	origFromAddr := createTestAddress(owner)
	origFromAddr.ID = entities.NewId()
	origFromAddr.Email = entities.Email("origfrom@example.com")

	origToAddr := createTestAddress(owner)
	origToAddr.ID = entities.NewId()
	origToAddr.Email = entities.Email("origto@example.com")

	chain := entities.Chain{
		Hash:            entities.NewHash(string(origFromAddr.Email), string(origToAddr.Email)),
		FromAddress:     fromAddr,
		ToAddress:       toAddr,
		OrigFromAddress: origFromAddr,
		OrigToAddress:   origToAddr,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UpdatedBy:       owner,
	}

	gormChain := chainFromEntity(chain)

	assert.Equal(t, string(chain.Hash), gormChain.Hash)
	assert.Equal(t, string(fromAddr.ID), gormChain.FromAddress.ID)
	assert.Equal(t, string(toAddr.ID), gormChain.ToAddress.ID)
	assert.Equal(t, string(origFromAddr.ID), gormChain.OrigFromAddress.ID)
	assert.Equal(t, string(origToAddr.ID), gormChain.OrigToAddress.ID)
}

func TestChainToEntity(t *testing.T) {
	owner := User{
		Model: Model{ID: string(entities.NewId())},
		Login: "owner@example.com",
	}

	fromAddr := Address{
		Model: Model{ID: string(entities.NewId())},
		Email: "from@example.com",
		Owner: owner,
	}

	toAddr := Address{
		Model: Model{ID: string(entities.NewId())},
		Email: "to@example.com",
		Owner: owner,
	}

	origFromAddr := Address{
		Model: Model{ID: string(entities.NewId())},
		Email: "origfrom@example.com",
		Owner: owner,
	}

	origToAddr := Address{
		Model: Model{ID: string(entities.NewId())},
		Email: "origto@example.com",
		Owner: owner,
	}

	gormChain := Chain{
		Hash:            "testhash123",
		FromAddress:     fromAddr,
		ToAddress:       toAddr,
		OrigFromAddress: origFromAddr,
		OrigToAddress:   origToAddr,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UpdatedBy:       owner,
	}

	chain := chainToEntity(gormChain)

	assert.Equal(t, entities.Hash(gormChain.Hash), chain.Hash)
	assert.Equal(t, entities.Id(fromAddr.ID), chain.FromAddress.ID)
	assert.Equal(t, entities.Id(toAddr.ID), chain.ToAddress.ID)
	assert.Equal(t, entities.Id(origFromAddr.ID), chain.OrigFromAddress.ID)
	assert.Equal(t, entities.Id(origToAddr.ID), chain.OrigToAddress.ID)
}

func TestChainFromEntityList(t *testing.T) {
	owner := createTestUser()
	fromAddr := createTestAddress(owner)
	toAddr := createTestAddress(owner)
	toAddr.ID = entities.NewId()

	origFromAddr := createTestAddress(owner)
	origFromAddr.ID = entities.NewId()
	origFromAddr.Email = entities.Email("origfrom@example.com")

	origToAddr := createTestAddress(owner)
	origToAddr.ID = entities.NewId()
	origToAddr.Email = entities.Email("origto@example.com")

	chains := []entities.Chain{
		{
			Hash:            entities.NewHash(string(origFromAddr.Email), string(origToAddr.Email)),
			FromAddress:     fromAddr,
			ToAddress:       toAddr,
			OrigFromAddress: origFromAddr,
			OrigToAddress:   origToAddr,
			UpdatedBy:       owner,
		},
	}

	gormChains := chainFromEntityList(chains)

	assert.Len(t, gormChains, 1)
	assert.Equal(t, string(chains[0].Hash), gormChains[0].Hash)
}

func TestChainToEntityList(t *testing.T) {
	owner := User{
		Model: Model{ID: string(entities.NewId())},
		Login: "owner@example.com",
	}

	fromAddr := Address{
		Model: Model{ID: string(entities.NewId())},
		Email: "from@example.com",
		Owner: owner,
	}

	toAddr := Address{
		Model: Model{ID: string(entities.NewId())},
		Email: "to@example.com",
		Owner: owner,
	}

	origFromAddr := Address{
		Model: Model{ID: string(entities.NewId())},
		Email: "origfrom@example.com",
		Owner: owner,
	}

	origToAddr := Address{
		Model: Model{ID: string(entities.NewId())},
		Email: "origto@example.com",
		Owner: owner,
	}

	gormChains := []Chain{
		{
			Hash:            "hash1",
			FromAddress:     fromAddr,
			ToAddress:       toAddr,
			OrigFromAddress: origFromAddr,
			OrigToAddress:   origToAddr,
			UpdatedBy:       owner,
		},
	}

	chains := chainToEntityList(gormChains)

	assert.Len(t, chains, 1)
	assert.Equal(t, entities.Hash(gormChains[0].Hash), chains[0].Hash)
}

func TestApiTokenFromEntity(t *testing.T) {
	owner := createTestUser()
	token := entities.ApiToken{
		ID:          entities.NewId(),
		Name:        "Test Token",
		TokenHash:   "hashvalue123",
		Salt:        "saltvalue",
		Description: "Test description",
		Owner:       owner,
		Expiration:  time.Now().Add(24 * time.Hour),
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   owner,
	}

	gormToken := apiTokenFromEntity(token)

	assert.Equal(t, token.ID.String(), gormToken.ID)
	assert.Equal(t, token.Name, gormToken.Name)
	assert.Equal(t, token.TokenHash, gormToken.TokenHash)
	assert.Equal(t, token.Salt, gormToken.Salt)
	assert.Equal(t, token.Description, gormToken.Description)
	assert.Equal(t, token.Active, gormToken.Active)
	assert.Equal(t, string(owner.ID), gormToken.Owner.ID)
}

func TestApiTokenToEntity(t *testing.T) {
	owner := User{
		Model: Model{ID: string(entities.NewId())},
		Login: "owner@example.com",
	}

	gormToken := ApiToken{
		Model: Model{
			ID:        string(entities.NewId()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Token",
		TokenHash:   "hashvalue123",
		Salt:        "saltvalue",
		Description: "Test description",
		Owner:       owner,
		Expiration:  time.Now().Add(24 * time.Hour),
		Active:      true,
		UpdatedBy:   owner,
	}

	token := apiTokenToEntity(gormToken)

	assert.Equal(t, entities.Id(gormToken.ID), token.ID)
	assert.Equal(t, gormToken.Name, token.Name)
	assert.Equal(t, gormToken.TokenHash, token.TokenHash)
	assert.Equal(t, gormToken.Salt, token.Salt)
	assert.Equal(t, gormToken.Description, token.Description)
	assert.Equal(t, gormToken.Active, token.Active)
}

func TestApiTokenFromEntityList(t *testing.T) {
	owner := createTestUser()
	tokens := []entities.ApiToken{
		{
			ID:        entities.NewId(),
			Name:      "Token 1",
			TokenHash: "hash1",
			Owner:     owner,
			UpdatedBy: owner,
		},
		{
			ID:        entities.NewId(),
			Name:      "Token 2",
			TokenHash: "hash2",
			Owner:     owner,
			UpdatedBy: owner,
		},
	}

	gormTokens := apiTokenFromEntityList(tokens)

	assert.Len(t, gormTokens, 2)
	assert.Equal(t, tokens[0].ID.String(), gormTokens[0].ID)
	assert.Equal(t, tokens[1].ID.String(), gormTokens[1].ID)
}
