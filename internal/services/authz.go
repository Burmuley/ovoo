package services

import (
	"github.com/Burmuley/ovoo/internal/entities"
)

// canGetAliases determines if the given user can retrieve the list of aliases.
// Returns true if the user is not of type MilterUser.
func canGetAliases(cuser entities.User) bool {
	return cuser.Type != entities.MilterUser
}

// canGetAlias determines if the given user can retrieve the specific alias (address).
// Returns true if the user is an Admin, or if the address is owned by a RegularUser whose id matches the user's id.
func canGetAlias(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canCreateAlias determines if the given user can create a new alias.
// Returns true if the user is an Admin or RegularUser.
func canCreateAlias(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canDeleteAlias determines if the user can delete the given alias (address).
// Returns true if the user is an Admin, or if the address is owned by a RegularUser with the same id.
func canDeleteAlias(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canUpdateAlias determines if the given user can update the specific alias (address).
// Returns true if the user is an Admin, or if the address is owned by a RegularUser with the same id.
func canUpdateAlias(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canGetPrAddr determines if the given user can retrieve the specified primary address.
// Returns true if the user is an Admin, or if the address is owned by a RegularUser with the same id.
func canGetPrAddr(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canCreatePrAddr determines if the given user can create a new primary address.
// Returns true if the user is an Admin or RegularUser.
func canCreatePrAddr(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canUpdatePrAddr determines if the given user can update the specified primary address.
// Returns true if the user is an Admin, or if the address is owned by a RegularUser with the same id.
func canUpdatePrAddr(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canDeletePrAddr determines if the given user can delete the specified primary address.
// Returns true if the user is an Admin, or if the address is owned by a RegularUser with the same id.
func canDeletePrAddr(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canGetChain determines if the user can retrieve chain-related data.
// Returns true if the user is an Admin or a MilterUser.
func canGetChain(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.MilterUser {
		return true
	}

	return false
}

// canCreateChain determines if the user can create a new chain entry.
// Returns true if the user is an Admin or a MilterUser.
func canCreateChain(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.MilterUser {
		return true
	}

	return false
}

// canDeleteChain determines if the user can delete a chain entry.
// Returns true if the user is an Admin or a MilterUser.
func canDeleteChain(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.MilterUser {
		return true
	}

	return false
}

// canGetUsers is a stub for determining if fetching all users is permitted.
// Currently not implemented and will panic if called.
func canGetUsers() bool {
	panic("implement me!")
}

// canGetUser determines if cuser can get access to a user with id user_id.
// Returns true if the user is an Admin, or if cuser is a RegularUser accessing their own user_id.
func canGetUser(cuser entities.User, user_id entities.Id) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == user_id && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canCreateUser determines if the given user can create a new user account.
// Returns true if the user is an Admin.
func canCreateUser(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	return false
}

// canUpdateUser determines if cuser can update a user with id user_id.
// Returns true if the user is an Admin, or if cuser is a RegularUser updating their own account.
func canUpdateUser(cuser entities.User, user_id entities.Id) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == user_id && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

// canDeleteUser determines if cuser can delete the given target user.
// Returns true if cuser is an Admin and the target user is not themselves.
func canDeleteUser(cuser, targetU entities.User) bool {
	if cuser.Type == entities.AdminUser && targetU.ID != cuser.ID {
		return true
	}

	return false
}

// canCreateApiToken determines if the given user can create a new API token.
// Always returns true.
func canCreateApiToken(cuser entities.User) bool {
	return true
}

// canGetApiToken determines if the user can access the specific API token.
// Returns true if the user is an Admin or if they are the owner of the token.
func canGetApiToken(cuser entities.User, token entities.ApiToken) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == token.Owner.ID {
		return true
	}

	return false
}

// canUpdateApiToken determines if the user can update the given API token.
// Returns true if the user is an Admin or if they are the owner of the token.
func canUpdateApiToken(cuser entities.User, token entities.ApiToken) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == token.Owner.ID {
		return true
	}

	return false
}

// canDeleteApiToken determines if the user can delete the given API token.
// Returns true if the user is an Admin or if they are the owner of the token.
func canDeleteApiToken(cuser entities.User, token entities.ApiToken) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == token.Owner.ID {
		return true
	}

	return false
}
