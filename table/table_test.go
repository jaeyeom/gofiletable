package table

import (
	"testing"

	"github.com/jaeyeom/gofiletable/filesystem"
)

func TestCreateAndDrop(t *testing.T) {
	tbl, err := Create(TableOption{"/test-table-0002", filesystem.NewMemoryFileSystem()})
	if err != nil {
		t.Error("failed to create a table")
	}
	if err := tbl.Drop(); err != nil {
		t.Error("failed to drop a table")
	}
}

func TestPutAndGet(t *testing.T) {
	tbl, err := Create(TableOption{"/test-table-0002", filesystem.NewMemoryFileSystem()})
	if err != nil {
		t.Error(err)
	}
	if err := tbl.Put([]byte("hello"), []byte("world")); err != nil {
		t.Error(err)
	}
	value, _ := tbl.Get([]byte("world"))
	if value != nil {
		t.Errorf("nil expected but %s found", value)
	}
	value, _ = tbl.Get([]byte("hello"))
	if string(value) != "world" {
		t.Errorf("world expected but %s found", value)
	}
	tbl.Remove([]byte("hello"))
	value, _ = tbl.Get([]byte("hello"))
	if value != nil {
		t.Errorf("nil expected but %s found", value)
	}
	if err := tbl.Drop(); err != nil {
		t.Error(err)
	}
}
