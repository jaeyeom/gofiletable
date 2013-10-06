package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ExampleMkdirAll() {
	if filepath.Separator != '/' {
		// Testing only on Unix like file system.
		// TODO: Test other platforms like Windows.
		return
	}
	ls := func(path string, f os.FileInfo, err error) error {
		fmt.Println(path, f.IsDir())
		return nil
	}
	mfs := NewMemoryFileSystem()
	mfs.Walk("/", ls)
	fmt.Println()

	mfs.MkdirAll("/path/to/hello/world", 0700)
	mfs.Walk("/", ls)
	fmt.Println()

	mfs.Create("/path/toto")
	mfs.RemoveAll("/path/to")
	mfs.Walk("/", ls)

	// Output:
	// / true
	//
	// / true
	// /path true
	// /path/to true
	// /path/to/hello true
	// /path/to/hello/world true
	//
	// / true
	// /path true
	// /path/toto false
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
