package postgresql

import (
	"net/url"
	"strings"
)

func EncodeConnectionString(connectionString string) string {
	urlBeginIndex := strings.Index(connectionString, "//")
	right := strings.LastIndex(connectionString, "@")
	left := urlBeginIndex + strings.Index(connectionString[urlBeginIndex:], ":") + 1
	password := connectionString[left:right]
	encodedPassword := url.QueryEscape(password)

	return strings.Replace(connectionString, password, encodedPassword, 1)
}
