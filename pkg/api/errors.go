package api

const (
	ERR_CODE_NOT_FOUND = iota + 10
	ERR_CODE_DUPLICATE_ENTRY
	ERR_CODE_BACKEND_ERROR
	ERR_CODE_UNAUTHORIZED
	ERR_CODE_UNAUTHENTICATED
	ERR_CODE_BAD_REQUEST
)

var (
	ErrUnauthenticated = Error{"unauthenticated request", ERR_CODE_UNAUTHENTICATED}
	ErrUnauthorized    = Error{"user is not authorized to perform requested action", ERR_CODE_UNAUTHORIZED}
	ErrNotFound        = Error{"the requested entity was not found", ERR_CODE_NOT_FOUND}
	ErrBadRequest      = Error{"malformed request", ERR_CODE_BAD_REQUEST}
	ErrDuplicateEntry  = Error{"duplicate entry", ERR_CODE_DUPLICATE_ENTRY}
	ErrBackendError    = Error{"backend error occurred", ERR_CODE_BACKEND_ERROR}
)
