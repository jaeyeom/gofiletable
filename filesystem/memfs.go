package filesystem

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// MemoryFileSystem is a fake file system.
type MemoryFileSystem struct {
	files map[string][]byte
}

// NewMemoryFileSystem creates an in-memory file system.
func NewMemoryFileSystem() *MemoryFileSystem {
	return &MemoryFileSystem{
		files: map[string][]byte{string(filepath.Separator): nil},
	}
}

// fileCloser is to implement Close function.
type fileCloser struct {
	bytes.Buffer
	files *map[string][]byte
	path string
}

// Close commits the buffer to the file system.
func (w *fileCloser) Close() error {
	buf, err := ioutil.ReadAll(w)
	if err != nil {
		return err
	}
	(*w.files)[w.path] = buf
	return nil
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil. The parameter perm is just for
// compatibility and does nothing. If path is already a directory,
// MkdirAll does nothing and returns nil.
func (mfs MemoryFileSystem) MkdirAll(path string, perm os.FileMode) error {
	current := filepath.Clean(path)
	for current != "." && current != string(filepath.Separator) {
		mfs.files[current+string(filepath.Separator)] = nil
		current = filepath.Dir(current)
	}
	return nil
}

// RemoveAll removes path and any children it contains. It removes
// everything it can but returns the first error it encounters. If the
// path does not exist, RemoveAll returns nil (no error).
func (mfs MemoryFileSystem) RemoveAll(path string) error {
	cleaned := filepath.Clean(path) + string(filepath.Separator)
	for k, _ := range mfs.files {
		if strings.HasPrefix(k, cleaned) {
			delete(mfs.files, k)
		}
	}
	return nil
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading.
func (mfs MemoryFileSystem) Open(name string) (io.ReadCloser, error) {
	cleaned := filepath.Clean(name)
	content, ok := mfs.files[cleaned]
	if !ok {
		return nil, os.ErrNotExist
	}
	return ioutil.NopCloser(bytes.NewReader(content)), nil
}

// Create creates the named file, truncating it if it already
// exists, and returns a writer to the file.
func (mfs MemoryFileSystem) Create(name string) (io.ReadWriteCloser, error) {
	cleaned := filepath.Clean(name)
	mfs.files[cleaned] = make([]byte, 0, 10)
	f := fileCloser{
		*bytes.NewBuffer(mfs.files[cleaned]),
		&mfs.files,
		name,
	}
	return &f, nil
}

// Remove removes the named file or directory. If there is an error,
// it will be of type *PathError.
func (mfs MemoryFileSystem) Remove(name string) error {
	cleaned := filepath.Clean(name)
	delete(mfs.files, cleaned)
	return nil
}
