package entities

import (
	"fmt"
	"strconv"
)

const (
	DefaultPageSize  int = 5
	DefaulPageNumber int = 1
)

type Filter struct {
	Page     int
	PageSize int
	Ids      []Id
}

// NewFilter parses and returns a Filter from the given input map.
// The input map is expected to contain string slices for filter keys like "id", "page", and "page_size".
// "id": a list of string ids (will be converted to Id type)
// "page": page number (only the first element is used if slice contains more items)
// "page_size": size of the page (only the first element is used)
//
// Defaults are assigned for Page and PageSize if not present or set to zero.
// Returns an error if "page" or "page_size" values are invalid.
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
	Types             []AddressType
	Emails            []Email
	Owners            []Id
	ServiceNames      []string
	ForwardAddressIds []Id
	Active            *bool
	Search            string
}

// NewAddressFilter parses and returns an AddressFilter from the given input map.
// Populates AddressFilter fields such as Types, Emails, Owners, and ServiceNames
// based on the corresponding filter keys provided in input. Returns an error
// if any validation fails (e.g., unsupported address type).
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
		case "active":
			active, err := strconv.ParseBool(vals[len(vals)-1]) // include last value only
			if err != nil {
				return AddressFilter{}, fmt.Errorf("%w: value for 'active' field must be boolean", ErrValidation)
			}
			af.Active = &active
		case "q":
			if len(vals) > 0 && vals[0] != "" {
				af.Search = vals[0]
			}
		}
	}

	return af, nil
}

type UserFilter struct {
	Filter
	Types  []UserType
	Logins []string
	Active *bool
}

// NewUserFilter parses and returns a UserFilter from the given input map.
// Populates UserFilter fields like Types and Logins based on filter keys.
// Returns an error if the user type value is not supported or invalid.
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
		case "active":
			active, err := strconv.ParseBool(vals[len(vals)-1]) // include last value only
			if err != nil {
				return UserFilter{}, fmt.Errorf("%w: value for 'active' field must be boolean", ErrValidation)
			}
			uf.Active = &active
		}
	}

	return uf, nil
}

type ApiTokenFilter struct {
	Filter
	UserIds []Id
	Active  *bool
}

// NewApiTokensFilter constructs an ApiTokenFilter using the provided input map.
// Only the generic Filter fields are considered; no additional fields are set.
// Returns an error if any value in the input filter causes validation to fail.
func NewApiTokensFilter(input map[string][]string) (ApiTokenFilter, error) {
	af := ApiTokenFilter{}
	filter, err := NewFilter(input)
	if err != nil {
		return ApiTokenFilter{}, err
	}

	af.Filter = filter
	for filter, vals := range input {
		switch filter {
		case "user_ids":
			ids := make([]Id, 0, len(vals))
			for _, val := range vals {
				ids = append(ids, Id(val))
			}
			af.UserIds = ids
		case "active":
			active, err := strconv.ParseBool(vals[len(vals)-1]) // include last value only
			if err != nil {
				return ApiTokenFilter{}, fmt.Errorf("%w: value for 'active' field must be boolean", ErrValidation)
			}
			af.Active = &active
		}
	}
	return af, nil
}

type ChainFilter struct {
	Filter
	OrigFromAddrIds []Id
	OrigToAddrIds   []Id
	FromAddrsIds    []Id
	ToAddrIds       []Id
}

type CustomDomainFilter struct {
	Filter
	Active        *bool
	Verified      *bool
	Owners        []Id
	IncludeGlobal bool
}

func NewCustomDomainFilter(input map[string][]string) (CustomDomainFilter, error) {
	cdf := CustomDomainFilter{}
	filter, err := NewFilter(input)
	if err != nil {
		return CustomDomainFilter{}, err
	}

	cdf.Filter = filter
	for filter, vals := range input {
		switch filter {
		case "owner":
			ids := make([]Id, 0, len(vals))
			for _, val := range vals {
				ids = append(ids, Id(val))
			}
			cdf.Owners = ids
		case "active":
			if len(vals) > 0 {
				active, err := strconv.ParseBool(vals[len(vals)-1]) // include last value only
				if err != nil {
					return CustomDomainFilter{}, fmt.Errorf("%w: value for 'active' field must be boolean", ErrValidation)
				}
				cdf.Active = &active
			}
		case "global":
			if len(vals) > 0 {
				global, err := strconv.ParseBool(vals[len(vals)-1]) // include last value only
				if err != nil {
					return CustomDomainFilter{}, fmt.Errorf("%w: value for 'global' field must be boolean", ErrValidation)
				}
				cdf.IncludeGlobal = global
			}
		case "verified":
			if len(vals) > 0 {
				verified, err := strconv.ParseBool(vals[len(vals)-1]) // include last value only
				if err != nil {
					return CustomDomainFilter{}, fmt.Errorf("%w: value for 'verified' field must be boolean", ErrValidation)
				}
				cdf.Verified = &verified
			}
		}
	}

	return cdf, nil
}
