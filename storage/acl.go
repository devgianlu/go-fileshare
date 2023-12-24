package storage

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	"io/fs"
)

type aclStorageProvider struct {
	underlying fileshare.StorageProvider
	defaultACL []fileshare.PathACL
}

func NewACLStorageProvider(storage fileshare.StorageProvider, defaultACL []fileshare.PathACL) (fileshare.AuthenticatedStorageProvider, error) {
	// TODO: make sure all paths in ACL exist
	return &aclStorageProvider{storage, defaultACL}, nil
}

func (p *aclStorageProvider) OpenFile(filename string, user *fileshare.User) (fs.File, error) {
	if user.Admin {
		return p.underlying.OpenFile(filename)
	}

	return nil, fmt.Errorf("unsupported") // TODO
}

func (p *aclStorageProvider) ReadDir(name string, user *fileshare.User) ([]fs.DirEntry, error) {
	if user.Admin {
		return p.underlying.ReadDir(name)
	}

	return nil, fmt.Errorf("unsupported") // TODO
}
