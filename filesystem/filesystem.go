// Package filesystem has an interface for file operations.
package filesystem

import (
	"io"
	"os"
	"path/filepath"
)

// FileSystem is an interface for a filesystem. It's possible to
// implement in-memory file system, for example.
type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(path string) error
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.ReadWriteCloser, error)
	Remove(name string) error
	Walk(root string, walkFn filepath.WalkFunc) error
}

// osFileSystem is a FileSystem implementation that just simply calls
// functions in the go os package library.
type osFileSystem struct {
}

var OSFileSystem osFileSystem = osFileSystem{}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission
// bits perm are used for all directories that MkdirAll creates. If
// path is already a directory, MkdirAll does nothing and returns nil.
func (osFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// RemoveAll removes path and any children it contains. It removes
// everything it can but returns the first error it encounters. If the
// path does not exist, RemoveAll returns nil (no error).
func (osFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading.
func (osFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Create creates the named file mode 0666 (before umask), truncating
// it if it already exists. If successful, returns a writer to the
// file. If there is an error, it will be of type *PathError.
func (osFileSystem) Create(name string) (io.ReadWriteCloser, error) {
	return os.Create(name)
}

// Remove removes the named file or directory. If there is an error,
// it will be of type *PathError.
func (osFileSystem) Remove(name string) error {
	return os.Remove(name)
}

// Walk calls the default Walk function.
func (osFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}
