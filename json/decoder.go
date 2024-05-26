package json

import (
	"encoding/json"
	"net/http"
)

func DecodeJSON[T any](r *http.Request) (T, error) {
	var v T

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return v, err
	}

	return v, err
}
