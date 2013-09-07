package filesystem

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func assertMapEqual(t *testing.T, m1, m2 map[string][]byte) {
	if len(m1) != len(m2) {
		t.Errorf("length %d != %d", len(m1), len(m2))
		return
	}
	for k, m1v := range m1 {
		m2v, ok := m2[k]
		if !ok {
			t.Errorf("key %s does not exist on the right", k)
		} else if string(m1v) != string(m2v) {
			t.Errorf("value %v != %v for key %s", m1v, m2v, k)
		}
	}
	for k, _ := range m2 {
		_, ok := m1[k]
		if !ok {
			t.Errorf("key %s does not exist on the left", k)
		}
	}
}

func TestMkdirAll(t *testing.T) {
	if filepath.Separator != '/' {
		// Testing only on Unix like file system.
		// TODO: Test other platforms like Windows.
		return
	}
	mfs := NewMemoryFileSystem()
	assertMapEqual(t, map[string][]byte{
		"/": nil,
	}, mfs.files)
		
	mfs.MkdirAll("/path/to/hello/world", 0700)
	assertMapEqual(t, map[string][]byte{
		"/": nil,
		"/path/": nil,
		"/path/to/": nil,
		"/path/to/hello/": nil,
		"/path/to/hello/world/": nil,
	}, mfs.files)

	mfs.Create("/path/toto")

	mfs.RemoveAll("/path/to")
	assertMapEqual(t, map[string][]byte{
		"/": nil,
		"/path/": nil,
		"/path/toto": []byte{},
	}, mfs.files)
}

func ExampleWriteFile() {
	if filepath.Separator != '/' {
		// Testing only on Unix like file system.
		// TODO: Test other platforms like Windows.
		return
	}
	mfs := NewMemoryFileSystem()
	mfs.MkdirAll("/path/to/hello/world", 0700)
	w, _ := mfs.Create("/path/to/myfile.txt")
	w.Write([]byte("content"))
	w.Close()
	r, _ := mfs.Open("/path/to/myfile.txt")
	buf, _ := ioutil.ReadAll(r)
	fmt.Println(string(buf))
	// Output: content
}
