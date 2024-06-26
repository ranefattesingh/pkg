package json

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Response struct {
	Success bool   `json:"success"`
	Payload any    `json:"payload,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

func EncodeResponseJSON(rw http.ResponseWriter, httpStatusCode int, content any) error {
	rw.WriteHeader(httpStatusCode)

	if content != nil {
		return encode(rw, content, nil)
	}

	return nil
}

func EncodeErrorJSON(rw http.ResponseWriter, err error) error {
	if err != nil {
		return encode(rw, nil, err)
	}

	return nil
}

func encode(w http.ResponseWriter, content any, err error) error {
	success := true

	var jErr *Error
	if err != nil {
		if !errors.As(err, &jErr) {
			jErr = &Error{
				HTTPStatusCode: http.StatusInternalServerError,
				Code:           http.StatusInternalServerError,
				Message:        err.Error(),
			}
		}

		w.WriteHeader(jErr.HTTPStatusCode)

		success = false
	}

	response := Response{
		Success: success,
		Payload: content,
		Error:   jErr,
	}

	encodingErr := json.NewEncoder(w).Encode(response)
	if encodingErr != nil {
		return encodingErr
	}

	return nil
}
