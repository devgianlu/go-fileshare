package http

import (
	"archive/tar"
	"compress/gzip"
	"github.com/devgianlu/go-fileshare"
	"io"
	"path/filepath"
)

func compressFolderToArchive(storage fileshare.AuthenticatedStorageProvider, user *fileshare.User, path string, w io.Writer) error {
	gw := gzip.NewWriter(w)
	aw := tar.NewWriter(gw)

	var addFolderToArchive func(dir string) error
	addFolderToArchive = func(dir string) error {
		entries, err := storage.ReadDir(dir, user)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				// add sub-folders recursively
				if err := addFolderToArchive(filepath.Join(dir, entry.Name())); err != nil {
					return err
				}

				continue
			}

			fileInfo, err := entry.Info()
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
			if err != nil {
				return err
			}

			// ensure we use the full path
			header.Name = filepath.Join(dir, entry.Name())
			if err := aw.WriteHeader(header); err != nil {
				return err
			}

			file, _, err := storage.OpenFile(header.Name, user)
			if err != nil {
				return err
			}

			if _, err = io.Copy(aw, file); err != nil {
				_ = file.Close()
				return err
			}

			_ = file.Close()
		}

		return nil
	}

	if err := addFolderToArchive(path); err != nil {
		return err
	}

	if err := aw.Close(); err != nil {
		return err
	}

	return gw.Close()
}
