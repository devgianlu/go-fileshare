package storage

import (
	"github.com/devgianlu/go-fileshare"
	"io/fs"
	"os"
)

type localStorageProvider struct {
	fs fs.ReadDirFS
}

func NewLocalStorageProvider(base string) fileshare.StorageProvider {
	return &localStorageProvider{os.DirFS(base).(fs.ReadDirFS)}
}

func (p *localStorageProvider) OpenFile(name string) (fs.File, error) {
	return p.fs.Open(name)
}

func (p *localStorageProvider) ReadDir(name string) ([]fs.DirEntry, error) {
	return p.fs.ReadDir(name)
}
