// Package table implements operations on the table.
package table

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Table struct {
	baseDirectory string
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
	tbl := Table{baseDirectory}
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
	return os.RemoveAll(tbl.baseDirectory)
}

// Recover creates the table directory.
func (tbl Table) Recover() error {
	return os.MkdirAll(tbl.baseDirectory, 0700)
}

// Get gets the value of the key in the table.
func (tbl Table) Get(key []byte) ([]byte, error) {
	filename := string(encodeKey(key))
	path := filepath.Join(tbl.baseDirectory, filename)
	f, err := os.Open(path)
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
	f, err := os.Create(path)
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
	return os.Remove(path)
}
