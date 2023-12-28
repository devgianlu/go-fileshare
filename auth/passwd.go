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
	for i, user := range auth.Users {
		// check config is correct
		if len(user.Nickname) == 0 || len(user.Passwd) == 0 {
			return nil, fmt.Errorf("invalid config for %s", user.Nickname)
		}

		// check password is bcrypt
		if _, err := bcrypt.Cost([]byte(user.Passwd)); err != nil {
			return nil, fmt.Errorf("invalid bcrypt hash for %s: %w", user.Nickname, err)
		}

		// check no duplicates
		for j, user_ := range auth.Users {
			if i == j {
				continue
			} else if user.Nickname == user_.Nickname {
				return nil, fmt.Errorf("duplicate user %s", user.Nickname)
			}
		}
	}

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
