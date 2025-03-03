package rest

import (
	"encoding/json"
	"io"
)

func readBody(body io.ReadCloser, data any) error {
	rawBody, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rawBody, data); err != nil {
		return err
	}

	return nil
}
