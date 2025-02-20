package entities

import (
	"strings"
	"testing"
)

func TestHash_Validate(t *testing.T) {
	tests := []struct {
		name    string
		h       Hash
		wantErr bool
	}{
		{
			name:    "simple test #1",
			h:       NewHash("test_str1", "test_str2"),
			wantErr: false,
		},
		{
			name:    "plain string",
			h:       "some string which is not a hash",
			wantErr: true,
		},
		{
			name: "wrong letters",
			h: func() Hash {
				hash := NewHash("test string 123", "test string 234")
				return Hash(strings.Replace(string(hash), "e", "X", 3))
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.h.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Hash.Validate() error = %v, wantErr %v, value: %s", err, tt.wantErr, tt.h)
			}
		})
	}
}
