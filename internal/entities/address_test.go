package entities

import "testing"

func TestAddress_Validate(t *testing.T) {
	type fields struct {
		Type           AddressType
		ID             Id
		Email          Email
		ForwardAddress *Address
		Owner          User
		Metadata       AddressMetadata
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid alias address",
			fields: fields{
				Type: AliasAddress,
				ID:   NewId(),
				Email: func() Email {
					email, _ := GenAliasEmail("domain.local", testDict)
					return email
				}(),
				ForwardAddress: &Address{ID: NewId(), Email: Email("some@protected.email"), Type: ProtectedAddress, Owner: User{ID: NewId()}},
				Owner: User{
					ID: NewId(),
				},
			},
			wantErr: false,
		},
		{
			name: "empty owner",
			fields: fields{
				Type: AliasAddress,
				ID:   NewId(),
				Email: func() Email {
					email, _ := GenAliasEmail("domain.local", testDict)
					return email
				}(),
				ForwardAddress: &Address{ID: NewId(), Email: Email("some@protected.email"), Type: ProtectedAddress},
			},
			wantErr: true,
		},
		{
			name: "valid protected address",
			fields: fields{
				Type: ProtectedAddress,
				ID:   NewId(),
				Email: func() Email {
					email, _ := GenAliasEmail("domain.local", testDict)
					return email
				}(),
				Owner: User{
					ID: NewId(),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid protected address (fwd email set)",
			fields: fields{
				Type: ProtectedAddress,
				ID:   NewId(),
				Email: func() Email {
					email, _ := GenAliasEmail("domain.local", testDict)
					return email
				}(),
				ForwardAddress: &Address{ID: NewId(), Email: Email("some@protected.email"), Type: ProtectedAddress},
				Owner: User{
					ID: NewId(),
				},
			},
			wantErr: true,
		},
		{
			name: "valid external address",
			fields: fields{
				Type:  ExternalAddress,
				ID:    NewId(),
				Email: "external@company.com",
				Owner: User{
					ID: NewId(),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid external address (fwd email set)",
			fields: fields{
				Type:           ExternalAddress,
				ID:             NewId(),
				Email:          "external@company.com",
				ForwardAddress: &Address{ID: NewId(), Email: Email("some@protected.email"), Type: ExternalAddress},
				Owner: User{
					ID: NewId(),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid id",
			fields: fields{
				Type: AliasAddress,
				ID:   "some invalid ID",
				Email: func() Email {
					email, _ := GenAliasEmail("domain.local", testDict)
					return email
				}(),
				ForwardAddress: &Address{ID: NewId(), Email: Email("some@protected.email"), Type: ProtectedAddress},
				Owner: User{
					ID: NewId(),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			fields: fields{
				Type:           AliasAddress,
				ID:             NewId(),
				Email:          Email("some#weird.email"),
				ForwardAddress: &Address{ID: NewId(), Email: Email("some@protected.email"), Type: ProtectedAddress},
				Owner: User{
					ID: NewId(),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid fwd email",
			fields: fields{
				Type: ReplyAliasAddress,
				ID:   NewId(),
				Email: func() Email {
					email, _ := GenAliasEmail("domain.local", testDict)
					return email
				}(),
				ForwardAddress: &Address{ID: NewId(), Email: Email("some#invalid_protected.email"), Type: ExternalAddress},
				Owner: User{
					ID: NewId(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Address{
				Type:           tt.fields.Type,
				ID:             tt.fields.ID,
				Email:          tt.fields.Email,
				ForwardAddress: tt.fields.ForwardAddress,
				Owner:          tt.fields.Owner,
				Metadata:       tt.fields.Metadata,
			}
			if err := a.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Address.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
