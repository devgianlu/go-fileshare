package auth

import "github.com/devgianlu/go-fileshare"

type configUsersProvider struct {
	users []fileshare.User
}

func NewConfigUsersProvider(users []fileshare.User) fileshare.UsersProvider {
	return &configUsersProvider{users}
}

func (p *configUsersProvider) GetUser(nickname string) (*fileshare.User, error) {
	for _, user := range p.users {
		if user.Nickname == nickname {
			return &user, nil
		}
	}

	return nil, nil
}
