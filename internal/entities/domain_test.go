package entities

import "testing"

func TestCustomDomain_Validate(t *testing.T) {
	validOwner := User{
		ID:    NewId(),
		Login: "owner@example.com",
		Type:  RegularUser,
	}

	tests := []struct {
		name    string
		domain  CustomDomain
		wantErr bool
	}{
		{
			name: "valid domain",
			domain: CustomDomain{
				ID:    NewId(),
				Name:  "example.com",
				Owner: validOwner,
			},
			wantErr: false,
		},
		{
			name: "invalid ID",
			domain: CustomDomain{
				ID:    Id("not-a-ulid"),
				Name:  "example.com",
				Owner: validOwner,
			},
			wantErr: true,
		},
		{
			name: "invalid owner ID",
			domain: CustomDomain{
				ID:   NewId(),
				Name: "example.com",
				Owner: User{
					ID:    Id("not-a-ulid"),
					Login: "owner@example.com",
					Type:  RegularUser,
				},
			},
			wantErr: true,
		},
		{
			name: "valid global domain",
			domain: CustomDomain{
				ID:     NewId(),
				Name:   "global.example.com",
				Global: true,
				Owner:  validOwner,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.domain.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("CustomDomain.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
