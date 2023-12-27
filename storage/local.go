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

func (p *localStorageProvider) OpenFile(name string) (io.ReadCloser, fs.FileInfo, error) {
	path := filepath.Join(p.base, filepath.Clean("/"+name))

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	if fileInfo.IsDir() {
		_ = file.Close()
		return nil, fileInfo, nil
	} else {
		return file, fileInfo, nil
	}
}

func (p *localStorageProvider) ReadDir(name string) ([]fs.DirEntry, error) {
	path := filepath.Join(p.base, filepath.Clean("/"+name))
	return os.ReadDir(path)
}
