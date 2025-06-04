package entities

import (
	"testing"
	"time"
)

func TestChain_Validate(t *testing.T) {
	type fields struct {
		Hash            Hash
		FromAddress     Address
		ToAddress       Address
		OrigFromAddress Address
		OrigToAddress   Address
		CreatedAt       time.Time
		UpdatedAt       time.Time
		UpdatedBy       User
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "valid chain",
			wantErr: false,
			fields: fields{
				Hash: NewHash("from@address.com", "to@address.com"),
				FromAddress: Address{
					ID:    NewId(),
					Email: "from@address.com",
					Type:  ExternalAddress,
					Owner: User{ID: NewId(), Login: "test_owner", Type: MilterUser},
				},
				ToAddress: Address{
					ID:             NewId(),
					Email:          "to@address.com",
					ForwardAddress: &Address{ID: NewId(), Email: "from@address.com", Type: ProtectedAddress, Owner: User{ID: NewId()}},
					Type:           ReplyAliasAddress,
					Owner:          User{ID: NewId(), Login: "test_owner", Type: MilterUser},
				},
				OrigFromAddress: Address{
					ID:    NewId(),
					Email: "from@address.com",
					Type:  ExternalAddress,
				},
				OrigToAddress: Address{
					ID:    NewId(),
					Email: "to@address.com",
					Type:  ReplyAliasAddress,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				UpdatedBy: User{ID: NewId(), Login: "test_updater", Type: MilterUser},
			},
		},
		{
			name:    "invalid chain - no from address",
			wantErr: true,
			fields: fields{
				Hash: NewHash("from@address.com", "to@address.com"),
				ToAddress: Address{
					ID:             NewId(),
					Email:          "to@address.com",
					ForwardAddress: &Address{ID: NewId(), Email: "from@address.com", Type: ProtectedAddress, Owner: User{ID: NewId()}},
					Type:           ReplyAliasAddress,
					Owner:          User{ID: NewId(), Login: "test_owner", Type: MilterUser},
				},
				OrigFromAddress: Address{
					ID:    NewId(),
					Email: "from@address.com",
					Type:  ExternalAddress,
				},
				OrigToAddress: Address{
					ID:    NewId(),
					Email: "to@address.com",
					Type:  ReplyAliasAddress,
				},
				CreatedAt: time.Now(),
			},
		},
		{
			name:    "invalid chain - no to address",
			wantErr: true,
			fields: fields{
				Hash: NewHash("from@address.com", "to@address.com"),
				FromAddress: Address{
					ID:    NewId(),
					Email: "from@address.com",
					Type:  ExternalAddress,
					Owner: User{ID: NewId(), Login: "test_owner", Type: MilterUser},
				},
				OrigFromAddress: Address{
					ID:    NewId(),
					Email: "from@address.com",
					Type:  ExternalAddress,
				},
				OrigToAddress: Address{
					ID:    NewId(),
					Email: "to@address.com",
					Type:  ReplyAliasAddress,
				},
				CreatedAt: time.Now(),
			},
		},
		{
			name:    "invalid chain - wrong hash",
			wantErr: true,
			fields: fields{
				Hash: "some random string lol",
				FromAddress: Address{
					ID:    NewId(),
					Email: "other@address.com",
					Type:  ExternalAddress,
					Owner: User{ID: NewId(), Login: "test_owner", Type: MilterUser},
				},
				ToAddress: Address{
					ID:             NewId(),
					Email:          "to@address.com",
					ForwardAddress: &Address{ID: NewId(), Email: "from@address.com", Type: ProtectedAddress, Owner: User{ID: NewId()}},
					Type:           ReplyAliasAddress,
					Owner:          User{ID: NewId(), Login: "test_owner", Type: MilterUser},
				},
				OrigFromAddress: Address{
					ID:    NewId(),
					Email: "other@address.com",
					Type:  ExternalAddress,
				},
				OrigToAddress: Address{
					ID:    NewId(),
					Email: "to@address.com",
					Type:  ReplyAliasAddress,
				},
				CreatedAt: time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Chain{
				Hash:            tt.fields.Hash,
				FromAddress:     tt.fields.FromAddress,
				ToAddress:       tt.fields.ToAddress,
				OrigFromAddress: tt.fields.OrigFromAddress,
				OrigToAddress:   tt.fields.OrigToAddress,
				CreatedAt:       tt.fields.CreatedAt,
				UpdatedAt:       tt.fields.UpdatedAt,
				UpdatedBy:       tt.fields.UpdatedBy,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Chain.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
