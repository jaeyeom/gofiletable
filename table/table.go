// Package table implements operations on the table.
package table

import (
	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	"github.com/jaeyeom/gofiletable/filesystem"
)

type Table struct {
	baseDirectory string
	fileSystem    filesystem.FileSystem
}

// encodeKey encodes key to base64 URL encoder to avoid illegal
// characters in the filename.
func encodeKey(key []byte) []byte {
	size := (len(key) + 2) / 3 * 4
	encoded := make([]byte, size)
	base64.URLEncoding.Encode(encoded, key)
	return encoded
}

// Create creates a table. Actually it just creates an empty
// directory.
func Create(baseDirectory string) (*Table, error) {
	// TODO: Produce error if the table already exists.
	tbl := Table{baseDirectory, filesystem.OSFileSystem{}}
	if err := tbl.Recover(); err != nil {
		return nil, err
	}
	return &tbl, nil
}

// Open opens a table in the baseDirectory.
func Open(baseDirectory string) (*Table, error) {
	return Create(baseDirectory)
}

// Drop drops the table tbl. It removes all the data in the table and
// the directory.
func (tbl Table) Drop() error {
	return tbl.fileSystem.RemoveAll(tbl.baseDirectory)
}

// Recover creates the table directory.
func (tbl Table) Recover() error {
	return tbl.fileSystem.MkdirAll(tbl.baseDirectory, 0700)
}

// Get gets the value of the key in the table.
func (tbl Table) Get(key []byte) ([]byte, error) {
	filename := string(encodeKey(key))
	path := filepath.Join(tbl.baseDirectory, filename)
	f, err := tbl.fileSystem.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// Put writes the data into the table.
func (tbl Table) Put(key []byte, value []byte) error {
	filename := string(encodeKey(key))
	path := filepath.Join(tbl.baseDirectory, filename)
	f, err := tbl.fileSystem.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(value)
	return err
}

// Remove removes an item in the table.
func (tbl Table) Remove(key []byte) error {
	filename := string(encodeKey(key))
	path := filepath.Join(tbl.baseDirectory, filename)
	return tbl.fileSystem.Remove(path)
}
