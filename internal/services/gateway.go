package services

import (
	"fmt"
	"reflect"

	"github.com/Burmuley/ovoo/internal/entities"
)

type ServiceGateway struct {
	Aliases *AliasesService
	Users   *UsersService
	PrAddrs *ProtectedAddrService
	Chains  *ChainsService
	Tokens  *ApiTokensService
}

// New creates a new ServiceGateway instance with the provided service implementations.
// It accepts variable number of services and assigns them to the appropriate fields
// in the ServiceGateway struct based on their type.
//
// Parameters:
//   - services: Variable number of service implementations to be included in the gateway
//
// Returns:
//   - *ServiceGateway: A fully initialized service gateway if successful
//   - error: An error if an unknown service type is provided or if any required service is missing
func New(services ...any) (*ServiceGateway, error) {
	f := &ServiceGateway{}
	for _, uc := range services {
		switch t := uc.(type) {
		case *AliasesService:
			f.Aliases = t
		case *UsersService:
			f.Users = t
		case *ProtectedAddrService:
			f.PrAddrs = t
		case *ChainsService:
			f.Chains = t
		case *ApiTokensService:
			f.Tokens = t
		default:
			return nil, fmt.Errorf("%w: unknown service type %T", entities.ErrConfiguration, t)
		}
	}

	if err := checkNilServices(f); err != nil {
		return nil, fmt.Errorf("validating services: %w", err)
	}

	return f, nil
}

// checkNilServices verifies that all fields in the ServiceGateway are initialized
// and not nil. It uses reflection to examine each field and returns an error if
// any field is nil.
//
// Parameters:
//   - gw: Pointer to a ServiceGateway instance to check
//
// Returns:
//   - error: nil if all fields are initialized, otherwise an error describing
//     which field is nil
func checkNilServices(gw *ServiceGateway) error {
	val := reflect.ValueOf(gw).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		if field.IsNil() {
			return fmt.Errorf("field %s can not be nil", fieldName)
		}
	}

	return nil
}
