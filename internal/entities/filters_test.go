package entities

import (
	"errors"
	"testing"
)

func TestNewAddressFilter(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string][]string
		want    AddressFilter
		wantErr error
	}{
		{
			name: "valid filters",
			input: map[string][]string{
				"type":         {"1"},
				"owner":        {"owner1", "owner2"},
				"id":           {"id1", "id2"},
				"email":        {"test@test.com"},
				"service_name": {"service1"},
				"page":         {"2"},
				"page_size":    {"10"},
			},
			want: AddressFilter{
				Filter: Filter{
					Page:     2,
					PageSize: 10,
					Ids:      []Id{"id1", "id2"},
				},
				Types:        []AddressType{1},
				Emails:       []Email{"test@test.com"},
				Owners:       []Id{"owner1", "owner2"},
				ServiceNames: []string{"service1"},
			},
		},
		{
			name: "default page values",
			input: map[string][]string{
				"type": {"1"},
			},
			want: AddressFilter{
				Filter: Filter{
					Page:     DefaulPageNumber,
					PageSize: DefaultPageSize,
				},
				Types: []AddressType{1},
			},
		},
		{
			name: "invalid type",
			input: map[string][]string{
				"type": {"invalid"},
			},
			wantErr: ErrValidation,
		},
		{
			name: "invalid page",
			input: map[string][]string{
				"page": {"invalid"},
			},
			wantErr: ErrValidation,
		},
		{
			name: "unsupported filter",
			input: map[string][]string{
				"unknown": {"value"},
			},
			want: AddressFilter{
				Filter: Filter{
					Page:     DefaulPageNumber,
					PageSize: DefaultPageSize,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAddressFilter(tt.input)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("NewAddressFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil {
				if got.Page != tt.want.Page {
					t.Errorf("Page = %v, want %v", got.Page, tt.want.Page)
				}
				if got.PageSize != tt.want.PageSize {
					t.Errorf("PageSize = %v, want %v", got.PageSize, tt.want.PageSize)
				}
				if len(got.Ids) != len(tt.want.Ids) {
					t.Errorf("Ids length = %v, want %v", len(got.Ids), len(tt.want.Ids))
				}
				if len(got.Types) != len(tt.want.Types) {
					t.Errorf("Types length = %v, want %v", len(got.Types), len(tt.want.Types))
				}
				if len(got.Emails) != len(tt.want.Emails) {
					t.Errorf("Emails length = %v, want %v", len(got.Emails), len(tt.want.Emails))
				}
				if len(got.Owners) != len(tt.want.Owners) {
					t.Errorf("Owners length = %v, want %v", len(got.Owners), len(tt.want.Owners))
				}
				if len(got.ServiceNames) != len(tt.want.ServiceNames) {
					t.Errorf("ServiceNames length = %v, want %v", len(got.ServiceNames), len(tt.want.ServiceNames))
				}
			}
		})
	}
}

func TestNewUserFilter(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string][]string
		want    UserFilter
		wantErr error
	}{
		{
			name: "valid filters",
			input: map[string][]string{
				"type":      {"1"},
				"id":        {"id1", "id2"},
				"login":     {"user1", "user2"},
				"page":      {"2"},
				"page_size": {"10"},
			},
			want: UserFilter{
				Filter: Filter{
					Page:     2,
					PageSize: 10,
					Ids:      []Id{"id1", "id2"},
				},
				Types:  []UserType{1},
				Logins: []string{"user1", "user2"},
			},
		},
		{
			name: "default page values",
			input: map[string][]string{
				"type": {"1"},
			},
			want: UserFilter{
				Filter: Filter{
					Page:     1,
					PageSize: DefaultPageSize,
				},
				Types: []UserType{1},
			},
		},
		{
			name: "invalid type",
			input: map[string][]string{
				"type": {"invalid"},
			},
			wantErr: ErrValidation,
		},
		{
			name: "invalid page",
			input: map[string][]string{
				"page": {"invalid"},
			},
			wantErr: ErrValidation,
		},
		{
			name: "unsupported filter",
			input: map[string][]string{
				"unknown": {"value"},
			},
			want: UserFilter{
				Filter: Filter{
					Page:     DefaulPageNumber,
					PageSize: DefaultPageSize,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUserFilter(tt.input)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("NewUserFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil {
				if got.Page != tt.want.Page {
					t.Errorf("Page = %v, want %v", got.Page, tt.want.Page)
				}
				if got.PageSize != tt.want.PageSize {
					t.Errorf("PageSize = %v, want %v", got.PageSize, tt.want.PageSize)
				}
				if len(got.Ids) != len(tt.want.Ids) {
					t.Errorf("Ids length = %v, want %v", len(got.Ids), len(tt.want.Ids))
				}
				if len(got.Types) != len(tt.want.Types) {
					t.Errorf("Types length = %v, want %v", len(got.Types), len(tt.want.Types))
				}
				if len(got.Logins) != len(tt.want.Logins) {
					t.Errorf("Logins length = %v, want %v", len(got.Logins), len(tt.want.Logins))
				}
			}
		})
	}
}
