package fileshare

import "errors"

var ErrAuthMalformed = errors.New("malformed authentication token")
var ErrAuthInvalid = errors.New("invalid authentication token")

type AuthProvider interface {
	GetUser(jwt string) (*User, error)
}
