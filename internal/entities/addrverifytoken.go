package entities

import (
	"crypto/subtle"
	"fmt"
	"strings"
	"time"
)

const PrAddrVerifyTokenPrefix = "pratk"

type AddressVerifyToken struct {
	ID        Id
	AddrId    Id
	Hash      string
	Salt      string
	ExpiryAt  time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	UpdatedBy User
	Owner     User
}

func NewAddressVerifyToken(expiration time.Time, addrId Id) (AddressVerifyToken, string, error) {
	// create verification token
	rawToken, err := RandString(32)
	if err != nil {
		return AddressVerifyToken{}, "", err
	}

	// hand the salt
	salt, err := RandString(16)
	if err != nil {
		return AddressVerifyToken{}, "", err
	}

	hash := HashSaltToken(salt, rawToken)
	token := AddressVerifyToken{
		ID:       NewId(),
		AddrId:   addrId,
		Hash:     hash,
		Salt:     salt,
		ExpiryAt: expiration,
	}

	userToken := Base62Encode([]byte(strings.Join(
		[]string{
			PrAddrVerifyTokenPrefix,
			token.ID.String(),
			rawToken,
		}, "_",
	)))

	return token, userToken, nil
}

func DecodeAddressVerifyToken(token string) (Id, string, error) {
	rawToken, err := Base62Decode(token)
	if err != nil {
		return "", "", err
	}

	if !strings.HasPrefix(rawToken, PrAddrVerifyTokenPrefix) {
		return "", "", fmt.Errorf("%w: not an address verify token", ErrValidation)
	}

	chunks := strings.Split(rawToken, "_")
	if len(chunks) != 3 {
		return "", "", fmt.Errorf("%w: invalid address verify token format", ErrValidation)
	}

	return Id(chunks[1]), chunks[2], nil
}

func (v AddressVerifyToken) Validate(token string) error {
	nh := HashSaltToken(v.Salt, token)
	if subtle.ConstantTimeCompare([]byte(v.Hash), []byte(nh)) != 1 {
		return fmt.Errorf("%w: invalid token", ErrValidation)
	}

	return nil
}

func (v AddressVerifyToken) IsExpired() bool {
	if time.Now().UTC().After(v.ExpiryAt) {
		return true
	}

	return false
}
