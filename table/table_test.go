package table

import (
	"testing"
)

func TestCreateAndDrop(t *testing.T) {
	// TODO: Avoid creating a real directory and file.
	tbl, err := Create("/tmp/test-table-0000")
	if err != nil {
		t.Error("failed to create a table")
	}
	if err := tbl.Drop(); err != nil {
		t.Error("failed to drop a table")
	}
}

func TestPutAndGet(t *testing.T) {
	// TODO: Avoid creating a real directory and file.
	tbl, err := Create("/tmp/test-table-0001")
	if err != nil {
		t.Error(err)
	}
	if err := tbl.Put([]byte("hello"), []byte("world")); err != nil {
		t.Error(err)
	}
	value, err := tbl.Get([]byte("hello"))
	if err != nil {
		t.Error(err)
	}
	if string(value) != "world" {
		t.Errorf("world expected but %s found", value)
	}
	if err := tbl.Drop(); err != nil {
		t.Error(err)
	}
}
