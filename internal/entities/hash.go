package entities

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// Hash represents a SHA-512/256 hash as a string.
type Hash string

// NewHash creates a new Hash from two input strings.
// It concatenates the inputs with a '+' separator and generates a SHA-512/256 hash.
func NewHash(s1, s2 string) Hash {
	return SimpleHash(strings.Join([]string{s1, s2}, "+"))
}

func SimpleHash(s string) Hash {
	hash := sha512.Sum512_256([]byte(s))
	return Hash(fmt.Sprintf("%x", hash))
}

// Validate checks if the Hash is valid.
// It returns an error if the hash length is not 64 characters or if it contains invalid characters.
func (h Hash) Validate() error {
	if len(string(h)) != 64 {
		return fmt.Errorf("wrong hash length")
	}

	reg := regexp.MustCompile("^[a-fA-F0-9]{64}$")
	if !reg.MatchString(string(h)) {
		return fmt.Errorf("wrong hash pattern")
	}

	return nil
}

// String returns the string representation of the Hash.
func (h Hash) String() string {
	return string(h)
}

// HashSaltToken creates a SHA-256 hash of a token by prepending the salt.
// The salt adds randomness to prevent rainbow table attacks.
// Returns the hex-encoded hash string.
func HashSaltToken(salt, token string) string {
	sum := sha256.Sum256(append([]byte(salt), []byte(token)...))
	return hex.EncodeToString(sum[:])
}
