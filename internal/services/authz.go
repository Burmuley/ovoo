package services

import (
	"github.com/Burmuley/ovoo/internal/entities"
)

func canGetAliases(cuser entities.User) bool {
	return cuser.Type != entities.MilterUser
}

func canGetAlias(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canCreateAlias(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canDeleteAlias(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canUpdateAlias(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canGetPrAddr(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canCreatePrAddr(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canUpdatePrAddr(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canDeletePrAddr(cuser entities.User, addr entities.Address) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if addr.Owner.ID == cuser.ID && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canGetChain(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.MilterUser {
		return true
	}

	return false
}

func canCreateChain(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.MilterUser {
		return true
	}

	return false
}

func canDeleteChain(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser || cuser.Type == entities.MilterUser {
		return true
	}

	return false
}

func canGetUsers() bool {
	panic("implement me!")
}

func canGetUser(cuser entities.User, user_id entities.Id) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == user_id && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canCreateUser(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	return false
}

func canUpdateUser(cuser entities.User, user_id entities.Id) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == user_id && cuser.Type == entities.RegularUser {
		return true
	}

	return false
}

func canDeleteUser(cuser entities.User) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	return false
}

func canCreateApiToken(cuser entities.User) bool {
	return true
}

func canGetApiToken(cuser entities.User, token entities.ApiToken) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == token.Owner.ID {
		return true
	}

	return false
}

func canUpdateApiToken(cuser entities.User, token entities.ApiToken) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == token.Owner.ID {
		return true
	}

	return false
}

func canDeleteApiToken(cuser entities.User, token entities.ApiToken) bool {
	if cuser.Type == entities.AdminUser {
		return true
	}

	if cuser.ID == token.Owner.ID {
		return true
	}

	return false
}
