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
		{
			name:  "active true",
			input: map[string][]string{"active": {"true"}},
			want: AddressFilter{
				Filter: Filter{Page: DefaulPageNumber, PageSize: DefaultPageSize},
				Active: func() *bool { v := true; return &v }(),
			},
		},
		{
			name:  "active false",
			input: map[string][]string{"active": {"false"}},
			want: AddressFilter{
				Filter: Filter{Page: DefaulPageNumber, PageSize: DefaultPageSize},
				Active: func() *bool { v := false; return &v }(),
			},
		},
		{
			name:    "invalid active value",
			input:   map[string][]string{"active": {"notabool"}},
			wantErr: ErrValidation,
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
				if tt.want.Active != nil {
					if got.Active == nil {
						t.Errorf("Active = nil, want %v", *tt.want.Active)
					} else if *got.Active != *tt.want.Active {
						t.Errorf("Active = %v, want %v", *got.Active, *tt.want.Active)
					}
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
		{
			name:  "active true",
			input: map[string][]string{"active": {"true"}},
			want: UserFilter{
				Filter: Filter{Page: DefaulPageNumber, PageSize: DefaultPageSize},
				Active: func() *bool { v := true; return &v }(),
			},
		},
		{
			name:  "active false",
			input: map[string][]string{"active": {"false"}},
			want: UserFilter{
				Filter: Filter{Page: DefaulPageNumber, PageSize: DefaultPageSize},
				Active: func() *bool { v := false; return &v }(),
			},
		},
		{
			name:    "invalid active value",
			input:   map[string][]string{"active": {"notabool"}},
			wantErr: ErrValidation,
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
				if tt.want.Active != nil {
					if got.Active == nil {
						t.Errorf("Active = nil, want %v", *tt.want.Active)
					} else if *got.Active != *tt.want.Active {
						t.Errorf("Active = %v, want %v", *got.Active, *tt.want.Active)
					}
				}
			}
		})
	}
}

func TestNewApiTokensFilter(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name    string
		input   map[string][]string
		want    ApiTokenFilter
		wantErr error
	}{
		{
			name: "valid user_ids filter",
			input: map[string][]string{
				"user_ids":  {"uid1", "uid2"},
				"page":      {"2"},
				"page_size": {"5"},
			},
			want: ApiTokenFilter{
				Filter:  Filter{Page: 2, PageSize: 5},
				UserIds: []Id{"uid1", "uid2"},
			},
		},
		{
			name:  "active true",
			input: map[string][]string{"user_ids": {"uid1"}, "active": {"true"}},
			want: ApiTokenFilter{
				Filter:  Filter{Page: DefaulPageNumber, PageSize: DefaultPageSize},
				UserIds: []Id{"uid1"},
				Active:  &trueVal,
			},
		},
		{
			name:  "active false",
			input: map[string][]string{"user_ids": {"uid1"}, "active": {"false"}},
			want: ApiTokenFilter{
				Filter:  Filter{Page: DefaulPageNumber, PageSize: DefaultPageSize},
				UserIds: []Id{"uid1"},
				Active:  &falseVal,
			},
		},
		{
			name:    "invalid active value",
			input:   map[string][]string{"user_ids": {"uid1"}, "active": {"notabool"}},
			wantErr: ErrValidation,
		},
		{
			name: "default page values",
			input: map[string][]string{
				"user_ids": {"uid1"},
			},
			want: ApiTokenFilter{
				Filter:  Filter{Page: DefaulPageNumber, PageSize: DefaultPageSize},
				UserIds: []Id{"uid1"},
			},
		},
		{
			name:    "invalid page",
			input:   map[string][]string{"page": {"invalid"}},
			wantErr: ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewApiTokensFilter(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewApiTokensFilter() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Page != tt.want.Page {
				t.Errorf("Page = %v, want %v", got.Page, tt.want.Page)
			}
			if got.PageSize != tt.want.PageSize {
				t.Errorf("PageSize = %v, want %v", got.PageSize, tt.want.PageSize)
			}
			if len(got.UserIds) != len(tt.want.UserIds) {
				t.Errorf("UserIds length = %v, want %v", len(got.UserIds), len(tt.want.UserIds))
			}
			if tt.want.Active != nil {
				if got.Active == nil {
					t.Errorf("Active = nil, want %v", *tt.want.Active)
				} else if *got.Active != *tt.want.Active {
					t.Errorf("Active = %v, want %v", *got.Active, *tt.want.Active)
				}
			} else if got.Active != nil {
				t.Errorf("Active = %v, want nil", *got.Active)
			}
		})
	}
}
