package entities

import (
	"fmt"
	"strings"
	"time"
)

type UserType int8

const (
	RegularUser UserType = iota
	AdminUser
	MilterUser
)

func UserTypeAtoi(uStr string) int {
	m := map[string]int{
		"regular": int(RegularUser),
		"admin":   int(AdminUser),
		"milter":  int(MilterUser),
	}

	uInt, ok := m[uStr]
	if !ok {
		return 99
	}

	return uInt
}

func UserTypeItoa(uInt int) string {
	m := map[int]string{
		int(RegularUser): "regular",
		int(AdminUser):   "admin",
		int(MilterUser):  "milter",
	}

	uStr, ok := m[uInt]
	if !ok {
		return "unknown"
	}

	return uStr
}

// User represents a user in the system with various attributes.
type User struct {
	ID             Id
	Type           UserType
	Login          string
	FirstName      string
	LastName       string
	PasswordHash   string
	FailedAttempts int
	LockoutUntil   time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UpdatedBy      *User
	CreatedBy      *User
}

// Validate checks if the User object is valid and returns an error if not.
func (u User) Validate() error {
	if err := u.ID.Validate(); err != nil {
		return err
	}

	if len(u.Login) == 0 {
		return fmt.Errorf("login can not be empty")
	}

	if u.Type > 2 {
		return fmt.Errorf("invalid user type")
	}

	return nil
}

// String returns a string representation of the User, combining FirstName and LastName.
func (u User) String() string {
	return strings.TrimSpace(strings.Join([]string{u.FirstName, u.LastName}, " "))
}
