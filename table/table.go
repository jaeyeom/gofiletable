// Package table implements operations on the table.
package table

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/jaeyeom/gofiletable/filesystem"
)

// TableOption stores options for opening a table.
type TableOption struct {
	BaseDirectory string
	FileSystem    filesystem.FileSystem
	KeepSnapshots bool
}

// Table stores state of the table. The actual data isn't stored in the struct.
type Table struct {
	baseDirectory string
	fileSystem    filesystem.FileSystem
	keepSnapshots bool
}

var (
	// ErrHeaderSizeMismatch is returned when the header size does
	// not match.
	ErrHeaderSizeMismatch = errors.New("gofiletable: header size mismatch")

	// ErrNoSnapshots is returned when there is no snapshots with
	// a given key.
	ErrNoSnapshots = errors.New("gofiletable: no snapshots")
)

// Header is a struct for the header of the table. It has the size of
// header and also indexing of the snapshots.
type Header struct {
	ByteSize  uint64 // Size of the header binary representation
	Snapshots []SnapshotInfo
}

// SnapshotInfo has the timestamp when the snapshot was written and
// the bytesize of the snapshot. By adding the byte size, it's
// possible to know the offset of the snapshot.
type SnapshotInfo struct {
	Timestamp uint64
	ByteSize  uint64
}

// Snapshot has the SnapshotInfo and the actual value.
type Snapshot struct {
	Info  SnapshotInfo
	Value []byte
}

// ByteReadCounter implements a counter that counts the number of
// bytes read.
type ByteReadCounter struct {
	Reader *bufio.Reader
	Count  uint64
}

// Read counts underlying reader while counting the number of bytes.
func (brc *ByteReadCounter) Read(p []byte) (n int, err error) {
	n, err = brc.Reader.Read(p)
	brc.Count += uint64(n)
	return
}

// ReadByte reads 1 byte from the reader.
func (brc *ByteReadCounter) ReadByte() (c byte, err error) {
	brc.Count += 1
	return brc.Reader.ReadByte()
}

// encodeKey encodes key to base64 URL encoder to avoid illegal
// characters in the filename.
func encodeKey(key []byte) []byte {
	size := base64.URLEncoding.EncodedLen(len(key))
	encoded := make([]byte, size)
	base64.URLEncoding.Encode(encoded, key)
	return encoded
}

// decodeKey decodes base64 URL to the key.
func decodeKey(encoded []byte) ([]byte, error) {
	size := base64.URLEncoding.DecodedLen(len(encoded))
	key := make([]byte, size)
	n, err := base64.URLEncoding.Decode(key, encoded)
	return key[0:n], err
}

// Create creates a table. Actually it just creates an empty
// directory.
func Create(option TableOption) (*Table, error) {
	// TODO: Produce error if the table already exists.
	tbl := Table{
		baseDirectory: option.BaseDirectory,
		fileSystem:    option.FileSystem,
		keepSnapshots: option.KeepSnapshots,
	}
	if tbl.fileSystem == nil {
		tbl.fileSystem = filesystem.OSFileSystem
	}
	if err := tbl.Recover(); err != nil {
		return nil, err
	}
	return &tbl, nil
}

// Open opens a table in the baseDirectory.
func Open(option TableOption) (*Table, error) {
	return Create(option)
}

// readHeader reads header from the reader r and returns the header
// struct. ErrHeaderSizeMismatch is returned when the header size does
// not match.
func readHeader(r *bufio.Reader) (header *Header, err error) {
	brc := &ByteReadCounter{
		Reader: r,
		Count:  0,
	}
	header = &Header{}
	header.ByteSize, err = binary.ReadUvarint(brc)
	if err != nil {
		return
	}
	snapshotSize, err := binary.ReadUvarint(brc)
	if err != nil {
		return
	}
	snapshots := make([]SnapshotInfo, snapshotSize)
	for i := uint64(0); i < snapshotSize; i++ {
		err = binary.Read(brc, binary.BigEndian, &snapshots[i].Timestamp)
		if err != nil {
			return
		}
		snapshots[i].ByteSize, err = binary.ReadUvarint(brc)
		if err != nil {
			return
		}
	}
	if brc.Count > header.ByteSize {
		err = ErrHeaderSizeMismatch
		return
	}
	// TODO: Remove this when Seek() function is implemented.
	for brc.Count < header.ByteSize {
		if _, err = brc.ReadByte(); err != nil {
			return
		}

	}
	header.Snapshots = snapshots
	return
}

// WriteTo writes the header to w and returns the number of bytes
// actually written to w. Header size can be 16 and it will be
// recalculated if the size is bigger.
func (header *Header) WriteTo(w io.Writer) (n int64, err error) {
	buf := bytes.NewBuffer(nil)
	bin := make([]byte, binary.MaxVarintLen64)
	// Write uint64 binary encoding of the snapshot size to buf.
	buf.Write(bin[0:binary.PutUvarint(bin, uint64(len(header.Snapshots)))])
	// Write each snapshot to buf.
	for _, snapshot := range header.Snapshots {
		binary.Write(buf, binary.BigEndian, snapshot.Timestamp)
		buf.Write(bin[0:binary.PutUvarint(bin, snapshot.ByteSize)])
	}
	// Find the variable size of header size.
	headerSizeSize := uint64(binary.PutUvarint(bin, header.ByteSize))
	// Recalculate header byte size until it gets right.
	for header.ByteSize < headerSizeSize+uint64(buf.Len()) {
		header.ByteSize = headerSizeSize + uint64(buf.Len())
		headerSizeSize = uint64(binary.PutUvarint(bin, header.ByteSize))
	}
	n1, err := w.Write(bin[0:headerSizeSize])
	n += int64(n1)
	if err != nil {
		return
	}
	n1, err = w.Write(buf.Bytes())
	n += int64(n1)
	for uint64(n) < header.ByteSize {
		n1, err = w.Write([]byte{0})
		n += int64(n1)
		if err != nil {
			return
		}
	}
	return
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
	if !tbl.keepSnapshots {
		filename := string(encodeKey(key))
		path := filepath.Join(tbl.baseDirectory, filename)
		f, err := tbl.fileSystem.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return ioutil.ReadAll(f)
	}
	var value []byte
	c, cerr := tbl.GetSnapshots(key)
	for snapshot := range c {
		value = snapshot.Value
	}
	return value, <-cerr
}

// GetSnapshots returns a channel of snapshot.
func (tbl Table) GetSnapshots(key []byte) (<-chan *Snapshot, <-chan error) {
	c := make(chan *Snapshot)
	cerr := make(chan error, 1)
	go func() {
		defer close(c)
		defer close(cerr)
		filename := string(encodeKey(key))
		path := filepath.Join(tbl.baseDirectory, filename)
		f, err := tbl.fileSystem.Open(path)
		if err != nil {
			cerr <- err
			return
		}
		defer f.Close()
		r := bufio.NewReader(f)
		h, err := readHeader(r)
		if err != nil {
			cerr <- err
			return
		}
		if len(h.Snapshots) == 0 {
			cerr <- ErrNoSnapshots
			return
		}
		for _, snapshot := range h.Snapshots {
			value := make([]byte, snapshot.ByteSize)
			_, err = io.ReadFull(r, value)
			if err != nil {
				cerr <- err
				break
			}
			c <- &Snapshot{snapshot, value}
		}
		return
	}()
	return c, cerr
}

// PutSnapshots rewrites the whole snapshots of the key with the given snapshots.
func (tbl Table) PutSnapshots(key []byte, snapshots []Snapshot) error {
	filename := string(encodeKey(key))
	path := filepath.Join(tbl.baseDirectory, filename)
	f, err := tbl.fileSystem.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	header := &Header{
		ByteSize:  16,
		Snapshots: []SnapshotInfo{},
	}
	for _, snapshot := range snapshots {
		header.Snapshots = append(header.Snapshots, snapshot.Info)
	}
	header.WriteTo(f)
	for _, snapshot := range snapshots {
		_, err = f.Write(snapshot.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Put writes the data into the table.
func (tbl Table) Put(key []byte, value []byte) error {
	filename := string(encodeKey(key))
	path := filepath.Join(tbl.baseDirectory, filename)
	var header *Header
	var valueArea []byte
	if tbl.keepSnapshots {
		f, err := tbl.fileSystem.Open(path)
		if err == nil {
			defer f.Close()
			r := bufio.NewReader(f)
			header, err = readHeader(r)
			if err != nil {
				return err
			}
			valueArea, err = ioutil.ReadAll(r)
			if err != nil {
				return err
			}
		} else {
			header = &Header{
				ByteSize:  16,
				Snapshots: nil,
			}
		}
	}
	f, err := tbl.fileSystem.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if header != nil {
		header.Snapshots = append(header.Snapshots, SnapshotInfo{uint64(time.Now().UnixNano()), uint64(len(value))})
		header.WriteTo(f)
	}

	if valueArea != nil {
		_, err = f.Write(valueArea)
		if err != nil {
			return err
		}
	}
	_, err = f.Write(value)
	if err != nil {
		return err
	}
	return err
}

// Remove removes an item in the table.
func (tbl Table) Remove(key []byte) error {
	filename := string(encodeKey(key))
	path := filepath.Join(tbl.baseDirectory, filename)
	return tbl.fileSystem.Remove(path)
}

// Keys returns a channel of keys.
func (tbl Table) Keys() (c chan []byte) {
	c = make(chan []byte)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		key, err := decodeKey([]byte(name))
		if err != nil {
			return err
		}
		c <- key
		return nil
	}
	go func() {
		defer close(c)
		tbl.fileSystem.Walk(tbl.baseDirectory, walkFunc)
	}()
	return c
}
