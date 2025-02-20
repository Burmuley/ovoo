package entities

import "testing"

func TestEmail_Validate(t *testing.T) {
	tests := []struct {
		name    string
		e       Email
		wantErr bool
	}{
		{
			name:    "valid simple email",
			e:       "burmuley@gmail.com",
			wantErr: false,
		},
		{
			name:    "valid long email",
			e:       "some-weird_corporate.email-nonsense@microhard.worst.software.omg_die.com",
			wantErr: false,
		},
		{
			name:    "invalid email #1",
			e:       "i_am_not_at_email.com",
			wantErr: true,
		},
		{
			name:    "invalid email #2",
			e:       "some@string that looks.like_email.com",
			wantErr: true,
		},
		{
			name:    "invalid email #3",
			e:       "iam#atypo.net",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Email.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
