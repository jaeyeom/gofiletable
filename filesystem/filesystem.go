package filesystem

import (
	"io"
	"os"
)

type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(path string) error
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	Remove(name string) error
}

type OSFileSystem struct {
}

func (OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (OSFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (OSFileSystem) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

func (OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}
