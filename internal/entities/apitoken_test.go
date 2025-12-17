package entities

import (
	"testing"
	"time"
)

func TestApiToken_Validate(t *testing.T) {
	type fields struct {
		ID          Id
		Name        string
		TokenHash   string
		Salt        string
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
			fields: func() fields {
				salt, _ := RandString(16)
				rawToken, _ := RandString(32)
				return fields{
					ID:          NewId(),
					Name:        "Test token",
					TokenHash:   HashApiToken(salt, rawToken),
					Salt:        salt,
					Description: "test token",
					Owner: User{
						ID:    NewId(),
						Login: "test_user",
						Type:  MilterUser,
					},
					Expiration: time.Now().Add(time.Hour * 2),
					Active:     true,
				}
			}(),
		},
		{
			name:    "no owner",
			wantErr: true,
			fields: func() fields {
				salt, _ := RandString(16)
				rawToken, _ := RandString(32)
				return fields{
					ID:          NewId(),
					Name:        "Test token",
					TokenHash:   HashApiToken(salt, rawToken),
					Salt:        salt,
					Description: "",
					Expiration:  time.Now().Add(time.Hour * 2),
					Active:      true,
				}
			}(),
		},
		{
			name:    "no description",
			wantErr: false,
			fields: func() fields {
				salt, _ := RandString(16)
				rawToken, _ := RandString(32)
				return fields{
					ID:          NewId(),
					Name:        "Test token",
					TokenHash:   HashApiToken(salt, rawToken),
					Salt:        salt,
					Description: "",
					Owner: User{
						ID:    NewId(),
						Login: "test_user",
						Type:  MilterUser,
					},
					Expiration: time.Now().Add(time.Hour * 2),
					Active:     true,
				}
			}(),
		},
		{
			name:    "empty token",
			wantErr: true,
			fields: func() fields {
				salt, _ := RandString(16)
				return fields{
					ID:          NewId(),
					Name:        "Test token",
					Salt:        salt,
					Description: "",
					Owner: User{
						ID:    NewId(),
						Login: "test_user",
						Type:  MilterUser,
					},
					Expiration: time.Now().Add(time.Hour * 2),
					Active:     true,
				}
			}(),
		},
		{
			name:    "empty id",
			wantErr: true,
			fields: func() fields {
				salt, _ := RandString(16)
				rawToken, _ := RandString(32)
				return fields{
					Name:        "Test token",
					TokenHash:   HashApiToken(salt, rawToken),
					Salt:        salt,
					Description: "test token",
					Owner: User{
						ID:    NewId(),
						Login: "test_user",
						Type:  MilterUser,
					},
					Expiration: time.Now().Add(time.Hour * 2),
					Active:     true,
				}
			}(),
		},
		{
			name:    "empty name",
			wantErr: true,
			fields: func() fields {
				salt, _ := RandString(16)
				rawToken, _ := RandString(32)
				return fields{
					ID:          NewId(),
					TokenHash:   HashApiToken(salt, rawToken),
					Salt:        salt,
					Description: "test token",
					Owner: User{
						ID:    NewId(),
						Login: "test_user",
						Type:  MilterUser,
					},
					Expiration: time.Now().Add(time.Hour * 2),
					Active:     true,
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &ApiToken{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				TokenHash:   tt.fields.TokenHash,
				Salt:        tt.fields.Salt,
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
	type test struct {
		name  string
		token *ApiToken
		want  bool
	}
	tests := []test{
		func() test {
			token, _ := NewToken(time.Now().Add(time.Hour*2), "test token", "test token description", User{ID: NewId(), Login: "test_owner", Type: MilterUser})
			return test{
				name:  "expiration time higher than now",
				want:  false,
				token: token,
			}
		}(),
		func() test {
			token, _ := NewToken(time.Now().Add(time.Hour*-2), "test token", "test token description", User{ID: NewId(), Login: "test_owner", Type: MilterUser})
			return test{
				name:  "expiration time lower than now",
				want:  true,
				token: token,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.Expired(); got != tt.want {
				t.Errorf("ApiToken.Expired() = %v, want %v", got, tt.want)
			}
		})
	}
}
