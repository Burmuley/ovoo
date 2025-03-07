package entities

import "golang.org/x/crypto/bcrypt"

func NewPasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func ValidPassword(password, hashedPasword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPasword), []byte(password))
	return err == nil
}
