package fileshare

type User struct {
	Nickname string
	Admin    bool
	ACL      []PathACL
}

type UsersProvider interface {
	GetUser(nickname string) (*User, error)
}
