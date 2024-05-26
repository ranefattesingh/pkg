package json

import (
	"context"
	"encoding/json"
	"net/http"
)

type Validator interface {
	Valid(ctx context.Context) error
}

func DecodeAndValidateJSON[T Validator](r *http.Request) (T, error) {
	var v T

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return v, err
	}

	return v, v.Valid(r.Context())
}
