package entities

import "testing"

func TestUser_Validate(t *testing.T) {
	type fields struct {
		Type         UserType
		ID           Id
		Login        string
		FirstName    string
		LastName     string
		PasswordHash string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid user",
			fields: fields{
				Type:      AdminUser,
				ID:        NewId(),
				Login:     "superadmin",
				FirstName: "Super",
				LastName:  "Admin",
			},
			wantErr: false,
		},
		{
			name: "invalid user - id",
			fields: fields{
				Type:      AdminUser,
				ID:        "some string",
				Login:     "superadmin",
				FirstName: "Super",
				LastName:  "Admin",
			},
			wantErr: true,
		},
		{
			name: "invalid user - type",
			fields: fields{
				Type:      44,
				ID:        NewId(),
				Login:     "superadmin",
				FirstName: "Super",
				LastName:  "Admin",
			},
			wantErr: true,
		},
		{
			name: "invalid user - login",
			fields: fields{
				Type:      MilterUser,
				ID:        NewId(),
				Login:     "",
				FirstName: "Super",
				LastName:  "Admin",
			},
			wantErr: true,
		},
		{
			name: "valid user - no names",
			fields: fields{
				Type:  RegularUser,
				ID:    NewId(),
				Login: "superadmin",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := User{
				Type:         tt.fields.Type,
				ID:           tt.fields.ID,
				Login:        tt.fields.Login,
				FirstName:    tt.fields.FirstName,
				LastName:     tt.fields.LastName,
				PasswordHash: tt.fields.PasswordHash,
			}
			if err := u.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
