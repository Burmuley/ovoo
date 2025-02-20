package entities

import (
	"testing"
	"time"
)

func TestApiToken_Validate(t *testing.T) {
	type fields struct {
		ID          Id
		Token       string
		Description string
		Owner       User
		Expiration  time.Time
		Active      bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "valid token",
			wantErr: false,
			fields: fields{
				ID:          NewId(),
				Token:       generateTokenString(),
				Description: "test token",
				Owner: User{
					ID:    NewId(),
					Login: "test_user",
					Type:  MilterUser,
				},
				Expiration: time.Now().Add(time.Hour * 2),
				Active:     true,
			},
		},
		{
			name:    "no owner",
			wantErr: true,
			fields: fields{
				ID:          NewId(),
				Token:       generateTokenString(),
				Description: "",
				Expiration:  time.Now().Add(time.Hour * 2),
				Active:      true,
			},
		},
		{
			name:    "no description",
			wantErr: false,
			fields: fields{
				ID:          NewId(),
				Token:       generateTokenString(),
				Description: "",
				Owner: User{
					ID:    NewId(),
					Login: "test_user",
					Type:  MilterUser,
				},
				Expiration: time.Now().Add(time.Hour * 2),
				Active:     true,
			},
		},
		{
			name:    "empty token",
			wantErr: true,
			fields: fields{
				ID:          NewId(),
				Description: "",
				Owner: User{
					ID:    NewId(),
					Login: "test_user",
					Type:  MilterUser,
				},
				Expiration: time.Now().Add(time.Hour * 2),
				Active:     true,
			},
		},
		{
			name:    "empty id",
			wantErr: true,
			fields: fields{
				Token:       generateTokenString(),
				Description: "test token",
				Owner: User{
					ID:    NewId(),
					Login: "test_user",
					Type:  MilterUser,
				},
				Expiration: time.Now().Add(time.Hour * 2),
				Active:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("blah blah")
			tr := &ApiToken{
				ID:          tt.fields.ID,
				Token:       tt.fields.Token,
				Description: tt.fields.Description,
				Owner:       tt.fields.Owner,
				Expiration:  tt.fields.Expiration,
				Active:      tt.fields.Active,
			}
			if err := tr.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ApiToken.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApiToken_Expired(t *testing.T) {
	type fields struct {
		ID          Id
		Token       string
		Description string
		Owner       User
		Expiration  time.Time
		Active      bool
	}
	tests := []struct {
		name  string
		token *ApiToken
		want  bool
	}{
		{
			name:  "expiration time higher than now",
			want:  false,
			token: NewToken(time.Now().Add(time.Hour*2), "test tolen", User{ID: NewId(), Login: "test_owner", Type: MilterUser}),
		},
		{
			name:  "expiration time lower than now",
			want:  true,
			token: NewToken(time.Now().Add(time.Hour*-2), "test tolen", User{ID: NewId(), Login: "test_owner", Type: MilterUser}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.Expired(); got != tt.want {
				t.Errorf("ApiToken.Expired() = %v, want %v", got, tt.want)
			}
		})
	}
}
