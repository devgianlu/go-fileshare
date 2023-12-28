package fileshare

const UserNicknameAnonymous = "anonymous"

type User struct {
	Nickname string
	Admin    bool
	ACL      []PathACL
}

func (u User) Anonymous() bool {
	return u.Nickname == UserNicknameAnonymous
}

type UsersProvider interface {
	GetUser(nickname string) (*User, error)
}
