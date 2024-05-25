package json

type Error struct {
	HTTPStatusCode int    `json:"-"`
	Code           int    `json:"code"`
	Message        string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}
