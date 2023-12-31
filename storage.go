package fileshare

import (
	"errors"
	"io"
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
	CreateFile(name string) (io.WriteCloser, error)
	OpenFile(name string) (io.ReadCloser, fs.FileInfo, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

type AuthenticatedStorageProvider interface {
	CreateFile(name string, user *User) (io.WriteCloser, error)
	OpenFile(name string, user *User) (io.ReadCloser, fs.FileInfo, error)
	ReadDir(name string, user *User) ([]fs.DirEntry, error)
	CanRead(name string, user *User) bool
	CanWrite(name string, user *User) bool
}
