package storage

import (
	"github.com/devgianlu/go-fileshare"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type localStorageProvider struct {
	base string
}

func NewLocalStorageProvider(base string) fileshare.StorageProvider {
	return &localStorageProvider{base}
}

func (p *localStorageProvider) CreateFile(name string) (io.WriteCloser, error) {
	path := filepath.Join(p.base, filepath.Clean("/"+name))
	return os.Create(path)
}

func (p *localStorageProvider) OpenFile(name string) (fs.File, error) {
	path := filepath.Join(p.base, filepath.Clean("/"+name))
	return os.Open(path)
}

func (p *localStorageProvider) ReadDir(name string) ([]fs.DirEntry, error) {
	path := filepath.Join(p.base, filepath.Clean("/"+name))
	return os.ReadDir(path)
}
