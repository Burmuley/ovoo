package entities

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

const (
	aliasPrefixMinLength = 10
)

var customPrefixReg = regexp.MustCompile(fmt.Sprintf(`^[a-z0-9\-]{%d,}$`, aliasPrefixMinLength))

// GenAliasEmail generates an alias email address using the provided domain and dictionary.
// Parameters:
//   - domain: the domain part of the email address (e.g., "example.com")
//   - dict: slice of strings used as a word dictionary to generate email aliases
//
// Returns:
//   - Email: the generated alias email address
//   - error: error if validation fails or random number generation fails
func GenAliasEmail(domain string, dict []string, prefix *string) (Email, error) {
	if len(domain) == 0 {
		return "", fmt.Errorf("%w: domain can not be empty", ErrValidation)
	}
	if len(dict) == 0 {
		return "", fmt.Errorf("%w: dictionary can not be empty", ErrValidation)
	}

	var email string

	if prefix == nil {
		words := make([]string, 2)
		for i := range words {
			r, err := rand.Int(rand.Reader, big.NewInt(int64(len(dict))))
			if err != nil {
				return "", fmt.Errorf("%w: making new random number: %w", ErrValidation, err)
			}
			words[i] = dict[r.Uint64()]
		}

		email = strings.Join(words, "-")
	} else {
		if !customPrefixReg.MatchString(*prefix) {
			return "", fmt.Errorf(
				"%w: custom prefix must be at least %d symbols from the following range: a-z, 0-9, dash", ErrValidation, aliasPrefixMinLength,
			)
		}
		email = *prefix
	}

	hash := SimpleHash(email)
	email = fmt.Sprintf("%s-%s", email, hash.String()[0:3])
	email = fmt.Sprintf("%s@%s", email, domain)
	return Email(email), nil
}

// GenReplyAliasEmail generates a reply alias email address based on the sender's email,
// the original alias email, and the domain. It returns the generated Email, a Hash, and
// an error if any occurred during the process.
func GenReplyAliasEmail(sourceEmail, destEmail Email, domain string) (Email, Hash, error) {
	if err := sourceEmail.Validate(); err != nil {
		return "", "", fmt.Errorf("%w: invalid sender email: %w", ErrValidation, err)
	}

	if err := destEmail.Validate(); err != nil {
		return "", "", fmt.Errorf("%w: invalid alias email: %w", ErrValidation, err)
	}

	sem := maskEmail(sourceEmail.String()) // sender email masked
	aem := maskEmail(destEmail.String())   // alias email masked
	hash := NewHash(sem, aem)
	reply_alias := strings.Join([]string{sem, hash.String()[0:10], hash.String()[56:64]}, "_")
	return Email(strings.Join([]string{reply_alias, domain}, "@")), hash, nil
}

// maskEmail takes an email address as input and returns a masked version of it.
// It replaces '@' with '_at_' and joins all alphanumeric parts with underscores.
func maskEmail(email string) string {
	r := regexp.MustCompile(`([a-zA-Z0-9]+)`)
	email = strings.ReplaceAll(email, "@", "_at_")
	email = strings.Join(r.FindAllString(email, -1), "_")
	return email
}
