package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewAddressVerifyToken(t *testing.T) {
	t.Run("valid creation", func(t *testing.T) {
		addrId := NewId()
		expiry := time.Now().UTC().Add(time.Hour)

		token, userToken, err := NewAddressVerifyToken(expiry, addrId)
		if err != nil {
			t.Fatalf("NewAddressVerifyToken() unexpected error = %v", err)
		}

		if token.ID == "" {
			t.Error("NewAddressVerifyToken() token.ID is empty")
		}
		if token.Hash == "" {
			t.Error("NewAddressVerifyToken() token.Hash is empty")
		}
		if token.Salt == "" {
			t.Error("NewAddressVerifyToken() token.Salt is empty")
		}
		if token.AddrId != addrId {
			t.Errorf("NewAddressVerifyToken() token.AddrId = %v, want %v", token.AddrId, addrId)
		}
		if !token.ExpiryAt.Equal(expiry) {
			t.Errorf("NewAddressVerifyToken() token.ExpiryAt = %v, want %v", token.ExpiryAt, expiry)
		}
		if userToken == "" {
			t.Error("NewAddressVerifyToken() userToken is empty")
		}
	})

	t.Run("round trip through decode and validate", func(t *testing.T) {
		addrId := NewId()
		expiry := time.Now().UTC().Add(time.Hour)

		token, userToken, err := NewAddressVerifyToken(expiry, addrId)
		if err != nil {
			t.Fatalf("NewAddressVerifyToken() unexpected error = %v", err)
		}

		decodedId, rawToken, err := DecodeAddressVerifyToken(userToken)
		if err != nil {
			t.Fatalf("DecodeAddressVerifyToken() unexpected error = %v", err)
		}
		if decodedId != token.ID {
			t.Errorf("DecodeAddressVerifyToken() id = %v, want %v", decodedId, token.ID)
		}

		if err := token.Validate(rawToken); err != nil {
			t.Errorf("Validate() unexpected error = %v", err)
		}
	})

	t.Run("zero-value AddrId is accepted", func(t *testing.T) {
		expiry := time.Now().UTC().Add(time.Hour)

		token, userToken, err := NewAddressVerifyToken(expiry, Id(""))
		if err != nil {
			t.Fatalf("NewAddressVerifyToken() unexpected error = %v", err)
		}
		if token.AddrId != Id("") {
			t.Errorf("NewAddressVerifyToken() token.AddrId = %v, want empty", token.AddrId)
		}
		if userToken == "" {
			t.Error("NewAddressVerifyToken() userToken is empty")
		}
	})
}

func TestDecodeAddressVerifyToken(t *testing.T) {
	validToken, validUserToken, err := NewAddressVerifyToken(time.Now().UTC().Add(time.Hour), NewId())
	if err != nil {
		t.Fatalf("failed to set up fixture token: %v", err)
	}

	tests := []struct {
		name         string
		userToken    string
		wantErr      bool
		wantSentinel error
	}{
		{
			name:      "valid token",
			userToken: validUserToken,
			wantErr:   false,
		},
		{
			name:      "empty input string",
			userToken: "",
			wantErr:   true,
		},
		{
			name:      "non-base62 characters",
			userToken: "not-base62!!!",
			wantErr:   true,
		},
		{
			name:         "wrong prefix",
			userToken:    Base62Encode([]byte(strings.Join([]string{"nota", "id", "tok"}, "_"))),
			wantErr:      true,
			wantSentinel: ErrValidation,
		},
		{
			name:         "too few chunks",
			userToken:    Base62Encode([]byte(strings.Join([]string{PrAddrVerifyTokenPrefix, "onlyid"}, "_"))),
			wantErr:      true,
			wantSentinel: ErrValidation,
		},
		{
			name:         "too many chunks",
			userToken:    Base62Encode([]byte(strings.Join([]string{PrAddrVerifyTokenPrefix, "id", "tok", "extra"}, "_"))),
			wantErr:      true,
			wantSentinel: ErrValidation,
		},
		{
			name:         "decodes to string '0'",
			userToken:    "0",
			wantErr:      true,
			wantSentinel: ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, rawToken, err := DecodeAddressVerifyToken(tt.userToken)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DecodeAddressVerifyToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantSentinel != nil && !errors.Is(err, tt.wantSentinel) {
				t.Errorf("DecodeAddressVerifyToken() error = %v, want sentinel %v", err, tt.wantSentinel)
			}
			if !tt.wantErr {
				if id != validToken.ID {
					t.Errorf("DecodeAddressVerifyToken() id = %v, want %v", id, validToken.ID)
				}
				if rawToken == "" {
					t.Error("DecodeAddressVerifyToken() rawToken is empty")
				}
			}
		})
	}
}

func TestAddressVerifyToken_Validate(t *testing.T) {
	expiry := time.Now().UTC().Add(time.Hour)
	token, userToken, err := NewAddressVerifyToken(expiry, NewId())
	if err != nil {
		t.Fatalf("failed to set up fixture token: %v", err)
	}
	_, rawToken, err := DecodeAddressVerifyToken(userToken)
	if err != nil {
		t.Fatalf("failed to decode fixture token: %v", err)
	}

	t.Run("correct token validates", func(t *testing.T) {
		if err := token.Validate(rawToken); err != nil {
			t.Errorf("Validate() unexpected error = %v", err)
		}
	})

	t.Run("wrong token value", func(t *testing.T) {
		other, err := RandString(32)
		if err != nil {
			t.Fatalf("failed to generate random string: %v", err)
		}
		if err := token.Validate(other); !errors.Is(err, ErrValidation) {
			t.Errorf("Validate() error = %v, want sentinel %v", err, ErrValidation)
		}
	})

	t.Run("empty token string", func(t *testing.T) {
		if err := token.Validate(""); !errors.Is(err, ErrValidation) {
			t.Errorf("Validate() error = %v, want sentinel %v", err, ErrValidation)
		}
	})

	t.Run("mismatched hash lengths do not panic", func(t *testing.T) {
		shortSaltToken := AddressVerifyToken{
			ID:       NewId(),
			AddrId:   NewId(),
			Hash:     "abcd",
			Salt:     "shortsalt",
			ExpiryAt: expiry,
		}
		if err := shortSaltToken.Validate(rawToken); !errors.Is(err, ErrValidation) {
			t.Errorf("Validate() error = %v, want sentinel %v", err, ErrValidation)
		}
	})
}

func TestAddressVerifyToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		expiryAt time.Time
		want     bool
	}{
		{
			name:     "future expiry",
			expiryAt: time.Now().UTC().Add(time.Hour),
			want:     false,
		},
		{
			name:     "past expiry",
			expiryAt: time.Now().UTC().Add(-time.Hour),
			want:     true,
		},
		{
			name:     "just past boundary",
			expiryAt: time.Now().UTC().Add(-time.Millisecond),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := AddressVerifyToken{ExpiryAt: tt.expiryAt}
			if got := token.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
