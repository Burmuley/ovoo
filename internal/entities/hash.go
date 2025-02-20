package entities

import (
	"crypto/sha512"
	"fmt"
	"regexp"
	"strings"
)

// Hash represents a SHA-512/256 hash as a string.
type Hash string

// NewHash creates a new Hash from two input strings.
// It concatenates the inputs with a '+' separator and generates a SHA-512/256 hash.
func NewHash(s1, s2 string) Hash {
	hash := sha512.Sum512_256([]byte(strings.Join([]string{s1, s2}, "+")))
	return Hash(fmt.Sprintf("%x", hash))
}

// Validate checks if the Hash is valid.
// It returns an error if the hash length is not 64 characters or if it contains invalid characters.
func (h Hash) Validate() error {
	if len(string(h)) != 64 {
		return fmt.Errorf("%w: validating hash: wrong hash length", ErrValidation)
	}

	reg := regexp.MustCompile("^[a-fA-F0-9]{64}$")
	if !reg.MatchString(string(h)) {
		return fmt.Errorf("%w: validating hash: wrong hash pattern", ErrValidation)
	}

	return nil
}

// String returns the string representation of the Hash.
func (h Hash) String() string {
	return string(h)
}
