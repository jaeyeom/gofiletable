package filesystem

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// MemoryFileSystem is a fake file system.
type MemoryFileSystem struct {
	files map[string][]byte
}

// NewMemoryFileSystem creates an in-memory file system.
func NewMemoryFileSystem() *MemoryFileSystem {
	return &MemoryFileSystem{
		files: map[string][]byte{
			string(filepath.Separator): nil,
		},
	}
}

type fileWriter interface {
	// Writes file to the filesystem.
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// fileCloser implements Close function in addition to bytes.Buffer.
type fileCloser struct {
	bytes.Buffer
	fw   fileWriter
	path string
	perm os.FileMode
}

// Close commits the buffer to the file system.
func (f *fileCloser) Close() error {
	return f.fw.WriteFile(f.path, f.Bytes(), f.perm)
}

type MemoryFile struct {
	path    string
	content []byte
	mode    os.FileMode
	modTime time.Time
}

func (mf MemoryFile) Name() string {
	return filepath.Base(mf.path)
}

func (mf MemoryFile) Size() int64 {
	return int64(len(mf.content))
}

func (mf MemoryFile) Mode() os.FileMode {
	return mf.mode
}

func (mf MemoryFile) ModTime() time.Time {
	return mf.modTime
}

func (mf MemoryFile) IsDir() bool {
	return mf.mode.IsDir()
}

func (mf MemoryFile) Sys() interface{} {
	return nil
}

func (mf MemoryFile) Close() error {
	return nil
}

// ensureDirName returns a clean path with the directory suffix.
func ensureDirName(path string) string {
	cleaned := filepath.Clean(path)
	if cleaned == string(filepath.Separator) {
		return cleaned
	}
	return cleaned + string(filepath.Separator)
}

// parentDir returns a parent directory of the given path.
func parentDir(path string) string {
	return ensureDirName(filepath.Dir(filepath.Clean(path)))
}

// Mkdir creates a new directory with the specified name and permission bits
// (before umask). If there is an error, it will be of type *os.PathError.
func (mfs MemoryFileSystem) Mkdir(name string, perm os.FileMode) error {
	current := filepath.Clean(name)
	if _, exists := mfs.files[parentDir(current)]; !exists {
		return &os.PathError{
			Op:   "mkdir",
			Path: name,
			Err:  errors.New("No such file or directory"),
		}
	}
	mfs.files[ensureDirName(current)] = nil
	return nil
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil. The parameter perm is just for
// compatibility and does nothing. If path is already a directory,
// MkdirAll does nothing and returns nil.
func (mfs MemoryFileSystem) MkdirAll(path string, perm os.FileMode) error {
	current := filepath.Clean(path)
	if current == "." || current == string(filepath.Separator) {
		return nil
	}
	if _, exists := mfs.files[ensureDirName(current)]; exists {
		return nil
	}
	mfs.MkdirAll(parentDir(current), perm)
	return mfs.Mkdir(ensureDirName(current), perm)
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
	mfs.files[cleaned] = nil
	f := fileCloser{
		Buffer: *bytes.NewBuffer(mfs.files[cleaned]),
		fw:     mfs,
		path:   name,
		perm:   0666,
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

// Walk implements filepath.Walk function for memory file system.
func (mfs MemoryFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	cleaned := filepath.Clean(root)
	paths := []string{}
	for path, _ := range mfs.files {
		if strings.HasPrefix(path, cleaned) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		// TODO: Provide the right information.
		mode := os.FileMode(0777)
		if strings.HasSuffix(path, "/") {
			mode |= os.ModeDir
		}
		mf := MemoryFile{
			path:    path,
			content: []byte{},
			mode:    mode,
		}
		err := walkFn(filepath.Clean(path), mf, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteFile writes data to a file named by filename. If the file does not
// exist, WriteFile creates it with permissions perm; otherwise WriteFile
// truncates it before writing.
func (mfs MemoryFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	cleaned := filepath.Clean(filename)
	mfs.files[cleaned] = data
	_ = perm // TODO: Implement perm.
	return nil
}
