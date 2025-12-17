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

// Tests for canGetChain
func TestCanGetChain_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	assert.True(t, canGetChain(user))
}

func TestCanGetChain_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	assert.True(t, canGetChain(user))
}

func TestCanGetChain_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.False(t, canGetChain(user))
}

// Tests for canCreateChain
func TestCanCreateChain_AdminUser(t *testing.T) {
	user := createTestUser(entities.AdminUser)
	assert.True(t, canCreateChain(user))
}

func TestCanCreateChain_MilterUser(t *testing.T) {
	user := createTestUser(entities.MilterUser)
	assert.True(t, canCreateChain(user))
}

func TestCanCreateChain_RegularUser(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.False(t, canCreateChain(user))
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
func TestCanUpdateUser_AdminUser(t *testing.T) {
	admin := createTestUser(entities.AdminUser)
	targetUserId := entities.NewId()

	assert.True(t, canUpdateUser(admin, targetUserId))
}

func TestCanUpdateUser_RegularUserSelf(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	assert.True(t, canUpdateUser(user, user.ID))
}

func TestCanUpdateUser_RegularUserOther(t *testing.T) {
	user := createTestUser(entities.RegularUser)
	otherUserId := entities.NewId()

	assert.False(t, canUpdateUser(user, otherUserId))
}

func TestCanUpdateUser_MilterUser(t *testing.T) {
	milter := createTestUser(entities.MilterUser)
	targetUserId := entities.NewId()

	assert.False(t, canUpdateUser(milter, targetUserId))
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

// Test for canGetUsers panic
func TestCanGetUsers_Panics(t *testing.T) {
	assert.Panics(t, func() {
		canGetUsers()
	})
}
