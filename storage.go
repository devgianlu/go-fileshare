package fileshare

import (
	"errors"
	"io/fs"
)

var ErrStorageReadForbidden = errors.New("user is not allowed to read from this location")
var ErrStorageWriteForbidden = errors.New("user is not allowed to write to this location")

type PathACL struct {
	Path  string
	Read  bool
	Write bool
}

type StorageProvider interface {
	OpenFile(filename string) (fs.File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

type AuthenticatedStorageProvider interface {
	OpenFile(name string, user *User) (fs.File, error)
	ReadDir(name string, user *User) ([]fs.DirEntry, error)
}
