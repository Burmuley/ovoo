package entities

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
)

// Base62Encode converts a byte slice to a base62 encoded string.
// Base62 encoding uses the characters 0-9, a-z, and A-Z to represent
// binary data as a string.
//
// Parameters:
//   - b: the byte slice to encode
//
// Returns:
//   - string: the base62 encoded representation of the input
func Base62Encode(b []byte) string {
	var i big.Int
	i.SetBytes(b)
	return i.Text(62)
}

// Base62Decode converts a base62 encoded string back to a string.
// Base62 decoding interprets the characters 0-9, a-z, and A-Z to
// convert the string back to its original binary data.
//
// Parameters:
//   - s: the base62 encoded string to decode
//
// Returns:
//   - string: the decoded string
//   - error: an error if the input is not a valid base62 string
func Base62Decode(s string) (string, error) {
	var i big.Int
	_, ok := i.SetString(s, 62)
	if !ok {
		return "", errors.New("can not parse base62 string")
	}

	return string(i.Bytes()), nil
}

// RandString generates a cryptographically secure random string of specified byte length.
// It uses crypto/rand to generate random bytes which are then encoded using base64 URL-safe encoding.
//
// Parameters:
//   - nByte: the number of random bytes to generate.
//
// Returns:
//   - string: the base64 URL-safe encoded random string.
//   - error: an error if random byte generation fails.
func RandString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	// return base64.RawURLEncoding.EncodeToString(b), nil
	return Base62Encode(b), nil
}
