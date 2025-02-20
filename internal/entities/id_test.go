package entities

import "testing"

func TestId_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      Id
		wantErr bool
	}{
		{
			name:    "valid id",
			id:      NewId(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.id.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Id.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
