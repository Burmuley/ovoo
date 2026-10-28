package entities

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordHash_Success(t *testing.T) {
	hash, err := NewPasswordHash("s3cr3t-p@ssw0rd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hash) == 0 {
		t.Fatal("hash is empty")
	}

	if hash == "s3cr3t-p@ssw0rd" {
		t.Fatal("password was stored in plain text")
	}

	if !ValidPassword("s3cr3t-p@ssw0rd", hash) {
		t.Error("generated hash does not validate against its own password")
	}
}

func TestNewPasswordHash_EmptyPassword(t *testing.T) {
	hash, err := NewPasswordHash("")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want %v", err, ErrValidation)
	}

	if hash != "" {
		t.Errorf("hash = %q, want empty", hash)
	}
}

func TestNewPasswordHash_DifferentSaltPerCall(t *testing.T) {
	first, err := NewPasswordHash("same-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := NewPasswordHash("same-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first == second {
		t.Error("two hashes of the same password are identical, salting is broken")
	}
}

func TestValidPassword(t *testing.T) {
	hash, err := NewPasswordHash("correct-horse")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{name: "matching password", password: "correct-horse", hash: hash, want: true},
		{name: "wrong password", password: "battery-staple", hash: hash, want: false},
		{name: "empty password against a real hash", password: "", hash: hash, want: false},
		{name: "empty hash", password: "correct-horse", hash: "", want: false},
		{name: "empty password and empty hash", password: "", hash: "", want: false},
		{name: "non-bcrypt hash", password: "correct-horse", hash: "correct-horse", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidPassword(tt.password, tt.hash); got != tt.want {
				t.Errorf("ValidPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidPassword_LegacyEmptyStringHash guards the auth bypass that shipped before
// UsersService.Create was fixed: it stored a bcrypt hash of the empty string, so any such
// account could be reached with empty credentials. Those rows still exist in databases
// created by older releases and must never authenticate.
func TestValidPassword_LegacyEmptyStringHash(t *testing.T) {
	// built with bcrypt directly - NewPasswordHash refuses to produce this any more
	legacy, err := bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// sanity check: bcrypt itself does consider the empty password a match
	if bcrypt.CompareHashAndPassword(legacy, []byte("")) != nil {
		t.Fatal("test fixture is not a hash of the empty string")
	}

	if ValidPassword("", string(legacy)) {
		t.Error("empty password authenticated against a legacy hash of the empty string")
	}

	if ValidPassword("anything", string(legacy)) {
		t.Error("arbitrary password authenticated against a legacy hash of the empty string")
	}
}
