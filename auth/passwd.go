package auth

import (
	"errors"
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"golang.org/x/crypto/bcrypt"
)

const AuthProviderTypePassword = "passwd"

type PasswordAuthProviderPayload struct {
	Nickname string
	Password string
}

type passwordAuthProvider struct {
	users []fileshare.AuthPasswordUser
}

func NewPasswordAuthProvider(auth fileshare.AuthPassword) (fileshare.AuthProvider, error) {
	for _, user := range auth.Users {
		// check config is correct
		if len(user.Nickname) == 0 || len(user.Passwd) == 0 {
			return nil, fmt.Errorf("invalid config for %s", user.Nickname)
		}
	}

	// TODO: check duplicate users

	return &passwordAuthProvider{auth.Users}, nil
}

func (p *passwordAuthProvider) Authenticate(payload_ any) (string, error) {
	payload, ok := payload_.(PasswordAuthProviderPayload)
	if !ok {
		panic("invalid payload type")
	}

	for _, user := range p.users {
		if user.Nickname != payload.Nickname {
			continue
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Passwd), []byte(payload.Password)); errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", fmt.Errorf("wrong password for %s", user.Nickname)
		} else if err != nil {
			return "", err
		} else {
			return user.Nickname, nil
		}
	}

	return "", fmt.Errorf("user %s not found", payload.Nickname)
}
