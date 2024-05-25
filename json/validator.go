package json

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Validator interface {
	Valid(ctx context.Context) error
}

func DecodeAndValidateJSON[T Validator](r *http.Request) (T, error) {
	var v T

	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	return v, v.Valid(r.Context())
}
