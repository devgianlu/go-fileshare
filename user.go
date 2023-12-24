package fileshare

type User struct {
	Nickname string
	Admin    bool
}

type UsersProvider interface {
	GetUser(nickname string) (*User, error)
}
