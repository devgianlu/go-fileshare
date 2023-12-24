package fileshare

import (
	"io/fs"
)

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
	OpenFile(filename string, user *User) (fs.File, error)
	ReadDir(name string, user *User) ([]fs.DirEntry, error)
}
