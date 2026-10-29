package entities

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase62Decode(t *testing.T) {
	type args struct {
		domain    string
		wordsDict []string
	}

	tests := []struct {
		name     string
		original string
		encoded  string
		want     string
		wantErr  bool
	}{
		{
			name:     "Test #1",
			original: "qwe123",
			encoded:  Base62Encode([]byte("qwe123")),
			want:     "qwe123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Base62Decode(tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("Base62Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Base62Decode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBase62Decode_TooLong(t *testing.T) {
	_, err := Base62Decode(strings.Repeat("a", MaxBase62DecodeLen+1))
	assert.Error(t, err)

	_, err = Base62Decode(strings.Repeat("a", MaxBase62DecodeLen))
	assert.NoError(t, err)
}

func TestDecodeAddressVerifyToken_TooLong(t *testing.T) {
	_, _, err := DecodeAddressVerifyToken(strings.Repeat("a", MaxBase62DecodeLen+1))
	assert.Error(t, err)
}
