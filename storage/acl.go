package storage

import (
	"fmt"
	"github.com/devgianlu/go-fileshare"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

type aclStorageProvider struct {
	underlying fileshare.StorageProvider
	defaultACL []fileshare.PathACL
}

func NewACLStorageProvider(storage fileshare.StorageProvider, defaultACL []fileshare.PathACL) (fileshare.AuthenticatedStorageProvider, error) {
	// errors
	// - read false, write true not allowed
	// - read false, write false redundant
	// - path does not exist

	// TODO: make sure all paths in ACL exist
	return &aclStorageProvider{storage, defaultACL}, nil
}

func (p *aclStorageProvider) evalACL(path string, user *fileshare.User) (read bool, write bool) {
	if user.Admin {
		panic("cannot evaluate ACL for admin user")
	}

	path = filepath.Join("/", path)

	var acls []fileshare.PathACL
	filterAcls := func(list []fileshare.PathACL) error {
		for _, acl := range list {
			rel, err := filepath.Rel(acl.Path, path)
			if err != nil {
				return err
			}

			// ACL does not apply to this file
			if strings.HasPrefix(rel, "../") {
				continue
			}

			acls = append(acls, acl)
		}

		return nil
	}

	if err := filterAcls(user.ACL); err != nil {
		log.WithError(err).WithField("module", "storage").
			Errorf("failed evaluating user ACL for %s, bailing out", path)
		return false, false
	}

	// no user ACL defined for path, check default
	if len(acls) == 0 {
		if err := filterAcls(p.defaultACL); err != nil {
			log.WithError(err).WithField("module", "storage").
				Errorf("failed evaluating default ACL for %s, bailing out", path)
			return false, false
		}
	}

	// no ACL defined for path, default deny
	if len(acls) == 0 {
		return false, false
	} else if len(acls) == 1 {
		return acls[0].Read, acls[0].Write
	}

	panic("unsupported") // FIXME: support multiple rules matching the path
}

func (p *aclStorageProvider) CreateFile(name string, user *fileshare.User) (io.WriteCloser, error) {
	if user.Admin {
		return p.underlying.CreateFile(name)
	}

	_, write := p.evalACL(name, user)
	if !write {
		return nil, fileshare.NewError("cannot write file", fileshare.ErrStorageWriteForbidden, fmt.Errorf("user %s is not allowed to write to %s", user.Nickname, name))
	}

	return p.underlying.CreateFile(name)
}

func (p *aclStorageProvider) OpenFile(name string, user *fileshare.User) (io.ReadCloser, fs.FileInfo, error) {
	if user.Admin {
		return p.underlying.OpenFile(name)
	}

	read, _ := p.evalACL(name, user)
	if !read {
		return nil, nil, fileshare.NewError("cannot read file", fileshare.ErrStorageReadForbidden, fmt.Errorf("user %s is not allowed to read from %s", user.Nickname, name))
	}

	return p.underlying.OpenFile(name)
}

func (p *aclStorageProvider) ReadDir(name string, user *fileshare.User) ([]fs.DirEntry, error) {
	if user.Admin {
		return p.underlying.ReadDir(name)
	}

	entries, err := p.underlying.ReadDir(name)
	if err != nil {
		return nil, err
	}

	var allowedEntries []fs.DirEntry
	for _, entry := range entries {
		read, _ := p.evalACL(filepath.Join(name, entry.Name()), user)
		if !read {
			continue
		}

		allowedEntries = append(allowedEntries, entry)
	}

	return allowedEntries, nil
}

func (p *aclStorageProvider) CanWrite(name string, user *fileshare.User) bool {
	if user.Admin {
		return true
	}

	_, write := p.evalACL(name, user)
	return write
}
