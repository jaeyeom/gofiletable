package table

import (
	"fmt"
	"os"
	"testing"

	"github.com/jaeyeom/gofiletable/filesystem"
)

func TestPutAndGet(t *testing.T) {
	const (
		getOp = iota
		putOp
	)
	type Operation struct {
		op    int
		key   string
		value string
		err   error
	}
	examples := []struct {
		tableOption TableOption
		operations  []Operation
	}{{
		TableOption{"/test-table-0001", filesystem.NewMemoryFileSystem(), false},
		[]Operation{},
	}, {
		TableOption{"/test-table-0002", filesystem.NewMemoryFileSystem(), true},
		[]Operation{},
	}, {
		TableOption{"/test-table-0003", filesystem.NewMemoryFileSystem(), false},
		[]Operation{
			{putOp, "hello", "world", nil},
			{putOp, "hello", "world2", nil},
			{getOp, "world", "", os.ErrNotExist},
			{getOp, "hello", "world2", nil},
		},
	}, {
		TableOption{"/test-table-0004", filesystem.NewMemoryFileSystem(), true},
		[]Operation{
			{putOp, "hello", "world", nil},
			{putOp, "hello", "world2", nil},
			{getOp, "world", "", os.ErrNotExist},
			{getOp, "hello", "world2", nil},
		},
	}, {
		TableOption{"/test-table-0005", filesystem.NewMemoryFileSystem(), true},
		[]Operation{
			{putOp, "hello", "world", nil},
			{putOp, "hello", "world1", nil},
			{putOp, "hello", "world123", nil},
			{putOp, "hello", "world45678", nil},
			{putOp, "hello", "world9", nil},
			{putOp, "hello", "world100", nil},
			{putOp, "hello", "world23", nil},
			{getOp, "world", "", os.ErrNotExist},
			{getOp, "hello", "world23", nil},
		},
	}}
	for i, testCase := range examples {
		tbl, err := Create(testCase.tableOption)
		if err != nil {
			t.Error(err)
		}
		for j, e := range testCase.operations {
			if e.op == getOp {
				value, err := tbl.Get([]byte(e.key))
				if string(value) != e.value {
					t.Errorf("%d.%d. %s expected but %s found", i, j, e.value, value)
				}
				if err != e.err {
					t.Errorf("%d.%d. %v", i, j, err)
				}
			} else if e.op == putOp {
				if err := tbl.Put([]byte(e.key), []byte(e.value)); err != e.err {
					t.Errorf("%d.%d. %v", i, j, err)
				}
			}
		}
		if err := tbl.Drop(); err != nil {
			t.Error(err)
		}
	}
}

func ExampleGetSnapshots() {
	tbl, err := Create(TableOption{"/test-table-0000", filesystem.NewMemoryFileSystem(), true})
	if err != nil {
		fmt.Println(err)
	}
	tbl.Put([]byte("key"), []byte("history1"))
	tbl.Put([]byte("key"), []byte("history2"))
	tbl.Put([]byte("key2"), []byte("history3"))
	tbl.Put([]byte("key"), []byte("history4"))
	c, cerr := tbl.GetSnapshots([]byte("key"))
	for snapshot := range c {
		fmt.Println(string(snapshot.Value))
	}
	fmt.Println(<-cerr)
	c, cerr = tbl.GetSnapshots([]byte("key2"))
	for snapshot := range c {
		fmt.Println(string(snapshot.Value))
	}
	fmt.Println(<-cerr)
	c, cerr = tbl.GetSnapshots([]byte("key3"))
	for snapshot := range c {
		fmt.Println(string(snapshot.Value))
	}
	fmt.Println(<-cerr)
	fmt.Println("KEYS:")
	for key := range tbl.Keys() {
		fmt.Println(string(key))
	}
	if err := tbl.Drop(); err != nil {
		fmt.Println(err)
	}
	// Output:
	// history1
	// history2
	// history4
	// <nil>
	// history3
	// <nil>
	// file does not exist
	// KEYS:
	// key
	// key2
}

func ExamplePutSnapshots() {
	tbl, err := Create(TableOption{"/test-table-0000", filesystem.NewMemoryFileSystem(), true})
	if err != nil {
		fmt.Println(err)
	}
	tbl.PutSnapshots([]byte("key"), []Snapshot{{
		Info:  SnapshotInfo{100, 7},
		Value: []byte("history"),
	}, {
		Info:  SnapshotInfo{200, 8},
		Value: []byte("history2"),
	}, {
		Info:  SnapshotInfo{300, 5},
		Value: []byte("test0"),
	}})
	c, cerr := tbl.GetSnapshots([]byte("key"))
	for snapshot := range c {
		fmt.Println(string(snapshot.Value))
	}
	fmt.Println(<-cerr)
	// Output:
	// history
	// history2
	// test0
	// <nil>
}
