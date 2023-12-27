package fileshare

import "errors"

var ErrAuthMalformed = errors.New("malformed authentication token")
var ErrAuthInvalid = errors.New("invalid authentication token")

type TokenProvider interface {
	GetUser(token string) (string, error)
	GetToken(nickname string) (string, error)
}
