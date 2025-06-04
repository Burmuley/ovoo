package entities

import (
	"fmt"
	"strconv"
)

const (
	DefaultPageSize  int = 5
	DefaulPageNumber     = 1
)

type Filter struct {
	Page     int
	PageSize int
	Ids      []Id
}

func NewFilter(input map[string][]string) (Filter, error) {
	filter := Filter{}
	for key, vals := range input {
		switch key {
		case "id":
			ids := make([]Id, 0, len(vals))
			for _, val := range vals {
				ids = append(ids, Id(val))
			}
			filter.Ids = ids
		case "page":
			page, err := strconv.Atoi(vals[0]) // only count first value
			if err != nil {
				return Filter{}, fmt.Errorf("%w: invalid page value='%s'", ErrValidation, vals[0])
			}
			filter.Page = page
		case "page_size":
			pageSize, err := strconv.Atoi(vals[0]) // only count first value
			if err != nil {
				return Filter{}, fmt.Errorf("%w: invalid page_size value='%s'", ErrValidation, vals[0])
			}
			filter.PageSize = pageSize
		}
	}

	if filter.PageSize == 0 {
		filter.PageSize = DefaultPageSize
	}

	if filter.Page == 0 {
		filter.Page = DefaulPageNumber
	}

	return filter, nil
}

type AddressFilter struct {
	Filter
	Types        []AddressType
	Emails       []Email
	Owners       []Id
	ServiceNames []string
}

func NewAddressFilter(input map[string][]string) (AddressFilter, error) {
	af := AddressFilter{}
	filter, err := NewFilter(input)
	if err != nil {
		return AddressFilter{}, err
	}

	af.Filter = filter
	for key, vals := range input {
		switch key {
		case "type":
			types := make([]AddressType, 0, len(vals))
			for _, val := range vals {
				atype, err := strconv.Atoi(val)
				if err != nil || AddressType(atype) > ExternalAddress {
					return AddressFilter{}, fmt.Errorf("%w: unsupported address type '%s'", ErrValidation, val)
				}
				types = append(types, AddressType(atype))
			}
			af.Types = types
		case "owner":
			owners := make([]Id, 0, len(vals))
			for _, val := range vals {
				owners = append(owners, Id(val))
			}
			af.Owners = owners
		case "email":
			emails := make([]Email, 0, len(vals))
			for _, val := range vals {
				emails = append(emails, Email(val))
			}
			af.Emails = emails
		case "service_name":
			snames := make([]string, 0, len(vals))
			for _, val := range vals {
				snames = append(snames, val)
			}
			af.ServiceNames = snames

		}
	}

	return af, nil
}

type UserFilter struct {
	Filter
	Types  []UserType
	Logins []string
}

func NewUserFilter(input map[string][]string) (UserFilter, error) {
	uf := UserFilter{}
	filter, err := NewFilter(input)
	if err != nil {
		return UserFilter{}, err
	}

	uf.Filter = filter
	for filter, vals := range input {
		switch filter {
		case "type":
			types := make([]UserType, 0, len(vals))
			for _, val := range vals {
				utype, err := strconv.Atoi(val)
				if err != nil || utype > int(MilterUser) {
					return UserFilter{}, fmt.Errorf("%w: unsupported user type '%s'", ErrValidation, val)
				}
				types = append(types, UserType(utype))
			}
			uf.Types = types
		case "login":
			logins := make([]string, 0, len(vals))
			for _, val := range vals {
				logins = append(logins, val)
			}
			uf.Logins = logins
		}
	}

	return uf, nil
}

type ApiTokenFilter struct {
	Filter
}

func NewApiTokensFilter(input map[string][]string) (ApiTokenFilter, error) {
	af := ApiTokenFilter{}
	filter, err := NewFilter(input)
	if err != nil {
		return ApiTokenFilter{}, err
	}
	af.Filter = filter
	return af, nil
}
