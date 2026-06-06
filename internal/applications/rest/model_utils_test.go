package rest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Burmuley/ovoo/internal/entities"
)

func TestUserTypeFStr(t *testing.T) {
	tests := []struct {
		input    string
		expected entities.UserType
	}{
		{"regular", entities.RegularUser},
		{"admin", entities.AdminUser},
		{"milter", entities.MilterUser},
		{"unknown", 99},
		{"", 99},
		{"ADMIN", 99}, // case-sensitive
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, userTypeFStr(tt.input))
		})
	}
}

func TestUserTypeTStr(t *testing.T) {
	tests := []struct {
		input    entities.UserType
		expected string
	}{
		{entities.RegularUser, "regular"},
		{entities.AdminUser, "admin"},
		{entities.MilterUser, "milter"},
		{99, "unknown"},
		{42, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, userTypeTStr(tt.input))
		})
	}
}

func TestAddrTypeTStr(t *testing.T) {
	tests := []struct {
		input    entities.AddressType
		expected string
	}{
		{entities.AliasAddress, "alias"},
		{entities.ExternalAddress, "external"},
		{entities.ProtectedAddress, "protected_address"},
		{entities.ReplyAliasAddress, "reply_alias"},
		{99, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, addrTypeTStr(tt.input))
		})
	}
}

func TestPgmTMetadata(t *testing.T) {
	pgm := entities.PaginationMetadata{
		CurrentPage:  2,
		FirstPage:    1,
		LastPage:     5,
		PageSize:     10,
		TotalRecords: 42,
	}
	result := pgmTMetadata(pgm)
	assert.Equal(t, 2, result.CurrentPage)
	assert.Equal(t, 1, result.FirstPage)
	assert.Equal(t, 5, result.LastPage)
	assert.Equal(t, 10, result.PageSize)
	assert.Equal(t, 42, result.TotalRecords)
}

func TestPgmTMetadata_Zero(t *testing.T) {
	result := pgmTMetadata(entities.PaginationMetadata{})
	assert.Equal(t, 0, result.CurrentPage)
	assert.Equal(t, 0, result.TotalRecords)
}

func TestUserTResponse(t *testing.T) {
	active := true
	u := entities.User{
		ID:        entities.NewId(),
		Login:     "admin@example.com",
		FirstName: "Alice",
		LastName:  "Smith",
		Type:      entities.AdminUser,
		Active:    active,
	}
	result := userTResponse(u)
	assert.Equal(t, string(u.ID), result.Id)
	assert.Equal(t, u.Login, result.Login)
	assert.Equal(t, u.FirstName, result.FirstName)
	assert.Equal(t, u.LastName, result.LastName)
	assert.Equal(t, "admin", result.Type)
	assert.NotNil(t, result.Active)
	assert.Equal(t, active, *result.Active)
}

func TestUserTResponse_RegularUser(t *testing.T) {
	u := entities.User{ID: entities.NewId(), Type: entities.RegularUser}
	result := userTResponse(u)
	assert.Equal(t, "regular", result.Type)
}

func TestUserTResponse_UnknownType(t *testing.T) {
	u := entities.User{ID: entities.NewId(), Type: 99}
	result := userTResponse(u)
	assert.Equal(t, "unknown", result.Type)
}

func TestAddressTAliasData(t *testing.T) {
	ownerID := entities.NewId()
	prAddr := &entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
	}
	alias := entities.Address{
		ID:             entities.NewId(),
		Type:           entities.AliasAddress,
		Email:          "alias@test.com",
		ForwardAddress: prAddr,
		Owner:          entities.User{ID: ownerID, Type: entities.AdminUser},
		Metadata: entities.AddressMetadata{
			Comment:     "test comment",
			ServiceName: "my-service",
		},
		Active: true,
	}
	result := addressTAliasData(alias)
	assert.Equal(t, "alias@test.com", string(result.Email))
	assert.Equal(t, "protected@example.com", string(result.ForwardEmail))
	assert.Equal(t, alias.ID.String(), result.Id)
	assert.NotNil(t, result.Metadata.Comment)
	assert.Equal(t, "test comment", *result.Metadata.Comment)
	assert.NotNil(t, result.Metadata.ServiceName)
	assert.Equal(t, "my-service", *result.Metadata.ServiceName)
	assert.NotNil(t, result.Active)
	assert.True(t, *result.Active)
}

func TestAddressTPrAddrData(t *testing.T) {
	ownerID := entities.NewId()
	prAddr := entities.Address{
		ID:    entities.NewId(),
		Type:  entities.ProtectedAddress,
		Email: "protected@example.com",
		Owner: entities.User{ID: ownerID, Type: entities.AdminUser},
		Metadata: entities.AddressMetadata{
			Comment:     "note",
			ServiceName: "acme",
		},
		Active: true,
	}
	result := addressTPrAddrData(prAddr)
	assert.Equal(t, "protected@example.com", string(result.Email))
	assert.Equal(t, prAddr.ID.String(), result.Id)
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, "note", *result.Metadata.Comment)
	assert.Equal(t, "acme", *result.Metadata.ServiceName)
	assert.NotNil(t, result.Active)
	assert.True(t, *result.Active)
}

func TestChainTChainData(t *testing.T) {
	owner := entities.User{ID: entities.NewId(), Type: entities.AdminUser}
	origFrom := entities.Address{
		ID: entities.NewId(), Email: "external@example.com",
		Type:  entities.ExternalAddress,
		Owner: owner,
	}
	alias := entities.Address{
		ID: entities.NewId(), Email: "alias@test.com",
		Type:  entities.AliasAddress,
		Owner: owner,
	}
	hash := entities.NewHash(string(origFrom.Email), string(alias.Email))
	chain := entities.Chain{
		Hash:            hash,
		FromAddress:     origFrom,
		ToAddress:       alias,
		OrigFromAddress: origFrom,
		OrigToAddress:   alias,
	}
	result := chainTChainData(chain)
	assert.Equal(t, hash.String(), result.Hash)
	assert.Equal(t, "external@example.com", result.FromEmail)
	assert.Equal(t, "alias@test.com", result.ToEmail)
	assert.Equal(t, "external@example.com", result.OrigFromAddress.Email)
	assert.Equal(t, "external", result.OrigFromAddress.Type)
	assert.Equal(t, "alias@test.com", result.OrigToAddress.Email)
	assert.Equal(t, "alias", result.OrigToAddress.Type)
}

func TestTokenTApiTokenData(t *testing.T) {
	expiration := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	token := entities.ApiToken{
		ID:          entities.NewId(),
		Name:        "mytoken",
		Description: "a test token",
		Active:      true,
		Expiration:  expiration,
	}
	result := tokenTApiTokenData(token)
	assert.Equal(t, string(token.ID), *result.Id)
	assert.True(t, result.Active)
	assert.NotNil(t, result.Description)
	assert.Equal(t, "a test token", *result.Description)
	assert.Equal(t, "mytoken", result.Name)
}

func TestTokenTApiTokenDataOnCreate(t *testing.T) {
	token := entities.ApiToken{
		ID:          entities.NewId(),
		Name:        "mytoken",
		Description: "a test token",
		Active:      true,
		Expiration:  time.Now().Add(24 * time.Hour),
		Token:       "ovtk_abc123_xyz789",
	}
	result := tokenTApiTokenDataOnCreate(token)
	assert.Equal(t, string(token.ID), *result.Id)
	assert.True(t, result.Active)
	assert.Equal(t, "ovtk_abc123_xyz789", result.ApiToken)
	assert.Equal(t, "a test token", *result.Description)
}

func TestTokenTApiTokenDataOnCreate_EmptyToken(t *testing.T) {
	token := entities.ApiToken{
		ID:   entities.NewId(),
		Name: "tok",
	}
	result := tokenTApiTokenDataOnCreate(token)
	assert.Equal(t, "", result.ApiToken)
}

func TestCustomDomainTDomainData(t *testing.T) {
	owner := entities.User{ID: entities.NewId(), Type: entities.RegularUser, Login: "owner@example.com"}
	domain := entities.CustomDomain{
		ID:    entities.NewId(),
		Name:  "example.com",
		Owner: owner,
		VerificationData: entities.DomainVerificationData{
			RecordType:             entities.TXTRecord,
			Name:                   "_ovoo_check_abc",
			Value:                  "OVOO_ID=xyz",
			LastVerificationResult: "",
		},
	}
	result := customDomainTDomainData(domain)
	assert.Equal(t, domain.ID.String(), result.Id)
	assert.Equal(t, domain.Name, result.Name)
	assert.NotNil(t, result.VerificationData)
	assert.Equal(t, DomainVerificationDNSRecordType(entities.TXTRecord), result.VerificationData.RecordType)
	assert.Equal(t, "_ovoo_check_abc", result.VerificationData.Name)
	assert.Equal(t, "OVOO_ID=xyz", result.VerificationData.Value)
	assert.Nil(t, result.VerificationData.LastVerificationResult)
}

func TestCustomDomainTDomainData_WithLastVerificationResult(t *testing.T) {
	owner := entities.User{ID: entities.NewId(), Type: entities.RegularUser, Login: "owner@example.com"}
	domain := entities.CustomDomain{
		ID:    entities.NewId(),
		Name:  "example.com",
		Owner: owner,
		VerificationData: entities.DomainVerificationData{
			RecordType:             entities.CNAMERecord,
			Name:                   "_ovoo_check_def",
			Value:                  "abc.ovoocheck.local.",
			LastVerificationResult: "validation error: record not found",
		},
	}
	result := customDomainTDomainData(domain)
	assert.NotNil(t, result.VerificationData.LastVerificationResult)
	assert.Equal(t, "validation error: record not found", *result.VerificationData.LastVerificationResult)
	assert.Equal(t, DomainVerificationDNSRecordType(entities.CNAMERecord), result.VerificationData.RecordType)
}
