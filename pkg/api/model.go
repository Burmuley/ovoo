package api

import "encoding/json"

// Error type implements Error interface
func (m Error) Error() string {
	errByte, err := json.Marshal(m)
	if err != nil {
		return m.ErrorDescription
	}

	return string(errByte)
}
