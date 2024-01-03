package storage

import (
	"errors"
	"github.com/devgianlu/go-fileshare"
	"io"
	"io/fs"
	"testing"
)

type mockDirEntry struct {
	name string
	dir  bool
}

func (m *mockDirEntry) Name() string {
	return m.name
}

func (m *mockDirEntry) IsDir() bool {
	return m.dir
}

func (m *mockDirEntry) Type() fs.FileMode {
	panic("mock")
}

func (m *mockDirEntry) Info() (fs.FileInfo, error) {
	panic("mock")
}

type mockStorageProvider struct {
	dirEntries []fs.DirEntry
}

func (p *mockStorageProvider) CreateFile(string) (io.WriteCloser, error) {
	return nil, nil
}

func (p *mockStorageProvider) OpenFile(string) (io.ReadCloser, fs.FileInfo, error) {
	return nil, nil, nil
}

func (p *mockStorageProvider) ReadDir(string) ([]fs.DirEntry, error) {
	return p.dirEntries, nil
}

func TestAclStorageProvider_CanRead(t *testing.T) {
	user := &fileshare.User{
		Nickname: "test",
		Admin:    false,
		ACL: []fileshare.PathACL{
			{
				Path:  "/test/foo/bar",
				Read:  true,
				Write: true,
			},
		},
	}

	storage := NewACLStorageProvider(&mockStorageProvider{}, nil)

	truePayloads := []string{
		"/test/foo/bar",
		"/test/foo/bar/",
		"/test/foo",
		"/test/foo/",
		"/test/foo/../foo/bar",
		"/test/foo/../../test/foo/bar",
		"/test/baz/../foo/bar",
		"/test/baz/baz/../../foo/bar",
	}
	for _, payload := range truePayloads {
		if !storage.CanRead(payload, user) {
			t.Fatalf("%s: expected true, got false", payload)
		}
	}

	falsePayloads := []string{
		"/test",
		"/test/",
		"/test/baz",
		"/test/foo/..",
		"/test/foo/../..",
		"/test/foo/../baz",
		"/test/foo/../../baz",
		"/",
		"/..",
		"/../..",
	}
	for _, payload := range falsePayloads {
		if storage.CanRead(payload, user) {
			t.Fatalf("%s: expected false, got true", payload)
		}
	}
}

func TestAclStorageProvider_CanWrite(t *testing.T) {
	user := &fileshare.User{
		Nickname: "test",
		Admin:    false,
		ACL: []fileshare.PathACL{
			{
				Path:  "/test/foo/bar",
				Read:  true,
				Write: true,
			},
		},
	}

	storage := NewACLStorageProvider(&mockStorageProvider{}, nil)

	truePayloads := []string{
		"/test/foo/bar",
		"/test/foo/bar/",
		"/test/foo/../foo/bar",
		"/test/foo/../../test/foo/bar",
		"/test/baz/../foo/bar",
		"/test/baz/baz/../../foo/bar",
	}
	for _, payload := range truePayloads {
		if !storage.CanWrite(payload, user) {
			t.Fatalf("%s: expected true, got false", payload)
		}
	}

	falsePayloads := []string{
		"/test",
		"/test/",
		"/test/foo",
		"/test/foo/",
		"/test/baz",
		"/test/foo/..",
		"/test/foo/../..",
		"/test/foo/../baz",
		"/test/foo/../../baz",
		"/",
		"/..",
		"/../..",
	}
	for _, payload := range falsePayloads {
		if storage.CanWrite(payload, user) {
			t.Fatalf("%s: expected false, got true", payload)
		}
	}
}

func TestAclStorageProvider_ReadDir1(t *testing.T) {
	user := &fileshare.User{
		Nickname: "test",
		Admin:    false,
		ACL: []fileshare.PathACL{
			{
				Path:  "/test",
				Read:  true,
				Write: true,
			},
		},
	}

	storage := NewACLStorageProvider(&mockStorageProvider{
		dirEntries: []fs.DirEntry{
			&mockDirEntry{"bar.txt", false},
			&mockDirEntry{"foo", true},
			&mockDirEntry{"test", true},
		},
	}, nil)

	payloads := []string{
		"/",
		"/..",
		"/../..",
		"/test/..",
		"/test/../..",
		"/test/foo/../..",
		"/test/foo/bar/../../..",
	}
	for _, payload := range payloads {
		if entries, _ := storage.ReadDir(payload, user); len(entries) != 1 || entries[0].Name() != "test" {
			t.Fatalf("%s: expected \"test\" entry, got %v", payload, entries)
		}
	}
}

func TestAclStorageProvider_ReadDir2(t *testing.T) {
	user := &fileshare.User{
		Nickname: "test",
		Admin:    false,
		ACL:      []fileshare.PathACL{},
	}

	storage := NewACLStorageProvider(&mockStorageProvider{
		dirEntries: []fs.DirEntry{
			&mockDirEntry{"bar.txt", false},
			&mockDirEntry{"baz.txt", false},
		},
	}, nil)

	payloads := []string{
		"/test",
		"/test/foo",
		"/test/foo/bar",
	}
	for _, payload := range payloads {
		if _, err := storage.ReadDir(payload, user); !errors.Is(err, fileshare.ErrStorageReadForbidden) {
			t.Fatalf("%s: expected read forbidden error, got %v", payload, err)
		}
	}
}
