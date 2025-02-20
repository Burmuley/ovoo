package rest

import "github.com/Burmuley/ovoo/internal/entities"

// UserTypeFStr converts a string representation of user type to its corresponding integer value.
// It returns -1 if the provided user type is not recognized.
func UserTypeFStr(utype string) int {
	switch utype {
	case "regular":
		return int(entities.RegularUser)
	case "admin":
		return int(entities.AdminUser)
	case "milter":
		return int(entities.MilterUser)
	default:
		return -1
	}
}
