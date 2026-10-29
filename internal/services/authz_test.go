package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Burmuley/ovoo/internal/entities"
)

// Helper function to create test users
func createTestUser(userType entities.UserType) entities.User {
	return entities.User{
		ID:   entities.NewId(),
		Type: userType,
	}
}

// Helper function to create test address with owner
func createTestAddress(owner entities.User) entities.Address {
	return entities.Address{
		ID:    entities.NewId(),
		Owner: owner,
		Email: "test@example.com",
	}
}

// Helper function to create test API token with owner
func createTestApiToken(owner entities.User) entities.ApiToken {
	return entities.ApiToken{
		ID:    entities.NewId(),
		Owner: owner,
	}
}

// Tests for canGetAliases
func TestCanGetAliases_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	assert.True(t, canGetAliases(user))
}

func TestCanGetAliases_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.True(t, canGetAliases(user))
}

func TestCanGetAliases_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	assert.False(t, canGetAliases(user))
}

// Tests for canGetAlias
func TestCanGetAlias_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canGetAlias(admin, addr))
}

func TestCanGetAlias_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canGetAlias(owner, addr))
}

func TestCanGetAlias_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canGetAlias(nonOwner, addr))
}

func TestCanGetAlias_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canGetAlias(milter, addr))
}

// Tests for canCreateAlias
func TestCanCreateAlias_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	assert.True(t, canCreateAlias(user))
}

func TestCanCreateAlias_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.True(t, canCreateAlias(user))
}

func TestCanCreateAlias_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	assert.False(t, canCreateAlias(user))
}

// Tests for canDeleteAlias
func TestCanDeleteAlias_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canDeleteAlias(admin, addr))
}

func TestCanDeleteAlias_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canDeleteAlias(owner, addr))
}

func TestCanDeleteAlias_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canDeleteAlias(nonOwner, addr))
}

func TestCanDeleteAlias_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canDeleteAlias(milter, addr))
}

// Tests for canUpdateAlias
func TestCanUpdateAlias_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canUpdateAlias(admin, addr))
}

func TestCanUpdateAlias_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canUpdateAlias(owner, addr))
}

func TestCanUpdateAlias_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canUpdateAlias(nonOwner, addr))
}

func TestCanUpdateAlias_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canUpdateAlias(milter, addr))
}

// Tests for canGetPrAddr
func TestCanGetPrAddr_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canGetPrAddr(admin, addr))
}

func TestCanGetPrAddr_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canGetPrAddr(owner, addr))
}

func TestCanGetPrAddr_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canGetPrAddr(nonOwner, addr))
}

func TestCanGetPrAddr_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canGetPrAddr(milter, addr))
}

// Tests for canCreatePrAddr
func TestCanCreatePrAddr_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	assert.True(t, canCreatePrAddr(user))
}

func TestCanCreatePrAddr_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.True(t, canCreatePrAddr(user))
}

func TestCanCreatePrAddr_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	assert.False(t, canCreatePrAddr(user))
}

// Tests for canUpdatePrAddr
func TestCanUpdatePrAddr_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canUpdatePrAddr(admin, addr))
}

func TestCanUpdatePrAddr_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canUpdatePrAddr(owner, addr))
}

func TestCanUpdatePrAddr_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canUpdatePrAddr(nonOwner, addr))
}

func TestCanUpdatePrAddr_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canUpdatePrAddr(milter, addr))
}

// Tests for canDeletePrAddr
func TestCanDeletePrAddr_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canDeletePrAddr(admin, addr))
}

func TestCanDeletePrAddr_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canDeletePrAddr(owner, addr))
}

func TestCanDeletePrAddr_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canDeletePrAddr(nonOwner, addr))
}

func TestCanDeletePrAddr_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canDeletePrAddr(milter, addr))
}

// createTestForwardChain builds a forward chain (external -> alias -> protected address) as the
// milter creates it: the alias and the protected address belong to owner, while the generated
// reply alias and the external sender address are owned by the milter service user.
func createTestForwardChain(owner, milter entities.User) entities.Chain {
	return entities.Chain{
		Hash:            entities.NewHash("external@example.com", "alias@test.com"),
		OrigFromAddress: createTestTypedAddress(entities.ExternalAddress, milter),
		OrigToAddress:   createTestTypedAddress(entities.AliasAddress, owner),
		FromAddress:     createTestTypedAddress(entities.ReplyAliasAddress, milter),
		ToAddress:       createTestTypedAddress(entities.ProtectedAddress, owner),
	}
}

// createTestReplyChain builds the reply chain paired with the forward chain above. Note that its
// OrigToAddress is the milter-owned reply alias, so ownership can only be established from the
// alias and the protected address.
func createTestReplyChain(owner, milter entities.User) entities.Chain {
	return entities.Chain{
		Hash:            entities.NewHash("protected@example.com", "reply@test.com"),
		OrigFromAddress: createTestTypedAddress(entities.ProtectedAddress, owner),
		OrigToAddress:   createTestTypedAddress(entities.ReplyAliasAddress, milter),
		FromAddress:     createTestTypedAddress(entities.AliasAddress, owner),
		ToAddress:       createTestTypedAddress(entities.ExternalAddress, milter),
	}
}

func createTestTypedAddress(atype entities.AddressType, owner entities.User) entities.Address {
	addr := createTestAddress(owner)
	addr.Type = atype
	return addr
}

// Tests for canGetChain
func TestCanGetChain_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	milter := createTestUser(entities.MilterUser)
	chain := createTestForwardChain(createTestUser(entities.RegularUser), milter)

	assert.True(t, canGetChain(user, chain))
}

func TestCanGetChain_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	chain := createTestForwardChain(createTestUser(entities.RegularUser), milter)

	assert.True(t, canGetChain(milter, chain))
}

func TestCanGetChain_RegularUser_OwnForwardChain(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	chain := createTestForwardChain(owner, createTestUser(entities.MilterUser))

	assert.True(t, canGetChain(owner, chain))
}

// A milter-created reply chain carries the milter user on its OrigToAddress, so an
// OrigToAddress-only ownership check would wrongly reject the real owner here.
func TestCanGetChain_RegularUser_OwnReplyChain(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	chain := createTestReplyChain(owner, createTestUser(entities.MilterUser))

	assert.True(t, canGetChain(owner, chain))
}

func TestCanGetChain_RegularUser_OtherUsersChain(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	chain := createTestForwardChain(createTestUser(entities.RegularUser), milter)
	other := createTestUser(entities.RegularUser)

	assert.False(t, canGetChain(other, chain))
}

// The milter owning the reply alias and the external address must not be enough on its own: a
// regular user who owns neither the alias nor the protected address gets nothing.
func TestCanGetChain_RegularUser_OwnsOnlyUntrustedAddresses(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	chain := createTestForwardChain(createTestUser(entities.RegularUser), user)

	assert.False(t, canGetChain(user, chain))
}

// A malformed chain whose alias and protected address belong to different users is readable by
// neither of them.
func TestCanGetChain_RegularUser_SplitOwnership(t *testing.T) {
	aliasOwner := createTestUser(entities.RegularUser)
	praddrOwner := createTestUser(entities.RegularUser)
	milter := createTestUser(entities.MilterUser)

	chain := createTestForwardChain(aliasOwner, milter)
	chain.ToAddress = createTestTypedAddress(entities.ProtectedAddress, praddrOwner)

	assert.False(t, canGetChain(aliasOwner, chain))
	assert.False(t, canGetChain(praddrOwner, chain))
}

// A chain holding no alias and no protected address carries no trustworthy owner at all.
func TestCanGetChain_RegularUser_NoOwnedAddressTypes(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	chain := entities.Chain{
		Hash:            entities.NewHash("external@example.com", "reply@test.com"),
		OrigFromAddress: createTestTypedAddress(entities.ExternalAddress, user),
		OrigToAddress:   createTestTypedAddress(entities.ReplyAliasAddress, user),
		FromAddress:     createTestTypedAddress(entities.ReplyAliasAddress, user),
		ToAddress:       createTestTypedAddress(entities.ExternalAddress, user),
	}

	assert.False(t, canGetChain(user, chain))
}

// Tests for canCreateChain
func TestCanCreateChain_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	addr := createTestAddress(createTestUser(entities.RegularUser))
	assert.True(t, canCreateChain(user, addr))
}

func TestCanCreateChain_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	addr := createTestAddress(createTestUser(entities.RegularUser))
	assert.True(t, canCreateChain(user, addr))
}

func TestCanCreateChain_RegularUser_NotOwner(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	addr := createTestAddress(createTestUser(entities.RegularUser))
	assert.False(t, canCreateChain(user, addr))
}

func TestCanCreateChain_RegularUser_Owner(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)
	assert.True(t, canCreateChain(owner, addr))
}

func TestCanCreateChain_MilterUser_OwnAddress(t *testing.T) {
	// a milter user owning the alias still passes through the service-user branch
	milter := createTestUser(entities.MilterUser)
	addr := createTestAddress(milter)
	assert.True(t, canCreateChain(milter, addr))
}

// Tests for canDeleteChain
func TestCanDeleteChain_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	assert.True(t, canDeleteChain(user))
}

func TestCanDeleteChain_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	assert.True(t, canDeleteChain(user))
}

func TestCanDeleteChain_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.False(t, canDeleteChain(user))
}

// Tests for canGetUser
func TestCanGetUser_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	targetUserId := entities.NewId()

	assert.True(t, canGetUser(admin, targetUserId))
}

func TestCanGetUser_RegularUserSelf(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.True(t, canGetUser(user, user.ID))
}

func TestCanGetUser_RegularUserOther(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	otherUserId := entities.NewId()

	assert.False(t, canGetUser(user, otherUserId))
}

func TestCanGetUser_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	targetUserId := entities.NewId()

	assert.False(t, canGetUser(milter, targetUserId))
}

// Tests for canCreateUser
func TestCanCreateUser_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	assert.True(t, canCreateUser(user))
}

func TestCanCreateUser_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.False(t, canCreateUser(user))
}

func TestCanCreateUser_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	assert.False(t, canCreateUser(user))
}

// Tests for canUpdateUser
func TestCanUpdateUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	regular := createTestUser(entities.RegularUser)
	milter := createTestUser(entities.MilterUser)

	assert.True(t, canUpdateUser(admin, entities.NewId()))
	assert.True(t, canUpdateUser(regular, regular.ID))
	assert.False(t, canUpdateUser(regular, entities.NewId()))
	assert.False(t, canUpdateUser(milter, entities.NewId()))
	assert.False(t, canUpdateUser(milter, milter.ID))
}

// Tests for canSetUserType
func TestCanSetUserType(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	regular := createTestUser(entities.RegularUser)
	milter := createTestUser(entities.MilterUser)

	assert.True(t, canSetUserType(admin, regular))
	assert.True(t, canSetUserType(admin, createTestUser(entities.AdminUser)))
	assert.False(t, canSetUserType(admin, admin), "admin can not change own type")
	assert.False(t, canSetUserType(regular, regular), "regular user can not change own type")
	assert.False(t, canSetUserType(regular, createTestUser(entities.RegularUser)))
	assert.False(t, canSetUserType(milter, milter))
}

// Tests for canDeleteUser
func TestCanDeleteUser_AdminUserDeletingOther(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	targetUser := createTestUser(entities.RegularUser)

	assert.True(t, canDeleteUser(admin, targetUser))
}

func TestCanDeleteUser_AdminUserDeletingSelf(t *testing.T) {
	admin := createTestUser(entities.AdminUser)

	assert.False(t, canDeleteUser(admin, admin))
}

func TestCanDeleteUser_RegularUserDeletingOther(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	targetUser := createTestUser(entities.RegularUser)

	assert.False(t, canDeleteUser(user, targetUser))
}

func TestCanDeleteUser_RegularUserDeletingSelf(t *testing.T) {
	user := createTestUser(entities.RegularUser)

	assert.False(t, canDeleteUser(user, user))
}

func TestCanDeleteUser_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	targetUser := createTestUser(entities.RegularUser)

	assert.False(t, canDeleteUser(milter, targetUser))
}

// Tests for canCreateApiToken
func TestCanCreateApiToken_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	assert.True(t, canCreateApiToken(user))
}

func TestCanCreateApiToken_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.True(t, canCreateApiToken(user))
}

func TestCanCreateApiToken_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	assert.True(t, canCreateApiToken(user))
}

// Tests for canGetApiToken
func TestCanGetApiToken_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.True(t, canGetApiToken(admin, token))
}

func TestCanGetApiToken_OwnerUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.True(t, canGetApiToken(owner, token))
}

func TestCanGetApiToken_NonOwnerUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.False(t, canGetApiToken(nonOwner, token))
}

// Tests for canUpdateApiToken
func TestCanUpdateApiToken_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.True(t, canUpdateApiToken(admin, token))
}

func TestCanUpdateApiToken_OwnerUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.True(t, canUpdateApiToken(owner, token))
}

func TestCanUpdateApiToken_NonOwnerUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.False(t, canUpdateApiToken(nonOwner, token))
}

// Tests for canDeleteApiToken
func TestCanDeleteApiToken_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.True(t, canDeleteApiToken(admin, token))
}

func TestCanDeleteApiToken_OwnerUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.True(t, canDeleteApiToken(owner, token))
}

func TestCanDeleteApiToken_NonOwnerUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.False(t, canDeleteApiToken(nonOwner, token))
}

// Helper to create a test domain with the given owner
func createTestDomain(owner entities.User) entities.CustomDomain {
	return entities.CustomDomain{
		ID:    entities.NewId(),
		Name:  "example.com",
		Owner: owner,
	}
}

// Tests for canVerifyDomain
func TestCanVerifyDomain_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	domain := createTestDomain(owner)
	assert.True(t, canVerifyDomain(admin, domain))
}

func TestCanVerifyDomain_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	domain := createTestDomain(owner)
	assert.True(t, canVerifyDomain(owner, domain))
}

func TestCanVerifyDomain_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	domain := createTestDomain(owner)
	assert.False(t, canVerifyDomain(nonOwner, domain))
}

func TestCanVerifyDomain_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	owner := createTestUser(entities.RegularUser)
	domain := createTestDomain(owner)
	assert.False(t, canVerifyDomain(milter, domain))
}

// Test for canGetUsers panic
func TestCanGetUsers_Panics(t *testing.T) {
	assert.Panics(t, func() {
		canGetUsers()
	})
}

// Tests for canSetActiveAlias
func TestCanSetActiveAlias_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canSetActiveAlias(addr, admin))
}

func TestCanSetActiveAlias_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canSetActiveAlias(addr, owner))
}

func TestCanSetActiveAlias_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canSetActiveAlias(addr, nonOwner))
}

func TestCanSetActiveAlias_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canSetActiveAlias(addr, milter))
}

// Tests for canSetActivePrAddr
func TestCanSetActivePrAddr_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canSetActivePrAddr(addr, admin))
}

func TestCanSetActivePrAddr_OwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.True(t, canSetActivePrAddr(addr, owner))
}

func TestCanSetActivePrAddr_NonOwnerRegularUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	addr := createTestAddress(owner)

	assert.False(t, canSetActivePrAddr(addr, nonOwner))
}

// Tests for canSetActiveApiToken
func TestCanSetActiveApiToken_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	owner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.True(t, canSetActiveApiToken(token, admin))
}

func TestCanSetActiveApiToken_OwnerUser(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.True(t, canSetActiveApiToken(token, owner))
}

func TestCanSetActiveApiToken_NonOwner(t *testing.T) {
	owner := createTestUser(entities.RegularUser)
	nonOwner := createTestUser(entities.RegularUser)
	token := createTestApiToken(owner)

	assert.False(t, canSetActiveApiToken(token, nonOwner))
}

// Tests for canSetActiveUser
func TestCanSetActiveUser_AdminUser_OtherUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	targetUser := createTestUser(entities.RegularUser)

	assert.True(t, canSetActiveUser(targetUser, admin))
}

func TestCanSetActiveUser_AdminUser_Self(t *testing.T) {
	admin := createTestUser(entities.AdminUser)

	assert.False(t, canSetActiveUser(admin, admin))
}

func TestCanSetActiveUser_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	targetUser := createTestUser(entities.RegularUser)

	assert.False(t, canSetActiveUser(targetUser, user))
}
