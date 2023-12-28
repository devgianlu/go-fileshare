package fileshare

import "errors"

var ErrAuthMalformed = errors.New("malformed authentication token")
var ErrAuthInvalid = errors.New("invalid authentication token")

type TokenProvider interface {
	GetUser(token string) (string, error)
	GetToken(nickname string) (string, error)
}

type AuthPasswordUser struct {
	Nickname string
	Passwd   string
}

type AuthPassword struct {
	Users []AuthPasswordUser
}

type AuthGithub struct {
	CallbackBaseURL string `yaml:"callback_base_url"`
	ClientID        string `yaml:"client_id"`
	ClientSecret    string `yaml:"client_secret"`
}

type OAuth2ProviderPayload struct {
	Code  string
	State string
}

type OAuth2AuthProvider interface {
	AuthProvider
	Callback() (string, error)
}

type AuthProvider interface {
	Authenticate(payload any) (string, error)
}
