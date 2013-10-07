package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jaeyeom/gofiletable/table"
)

var (
	addr      = flag.String("addr", ":9001", "address of server")
	tablePath = flag.String("table_path", "", "path to the backend table")
)

var tbl *table.Table

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

// indexHandler is index page handler.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<html><body>")
	defer fmt.Fprint(w, "</body></html>")
	if r.URL.Path == "/" {
		fmt.Fprint(w, "<ul>")
		for key := range tbl.Keys() {
			encoded := encodeKey(key)
			fmt.Fprintf(w, "<li><a href=\"/%s\">%s</a></li>", string(encoded), string(key))
		}
		fmt.Fprint(w, "</ul>")
		return
	}
	splitted := strings.Split(r.URL.Path, "/")
	if len(splitted) == 2 && splitted[1] != "" {
		key, err := decodeKey([]byte(splitted[1]))
		if err != nil {
			log.Println(err)
			return
		}
		cs, cerr := tbl.GetSnapshots(key)
		for s := range cs {
			fmt.Fprintf(w, "<h2>%s</h2>", time.Unix(0, int64(s.Info.Timestamp)))
			fmt.Fprintf(w, "<p>\n%s\n</p>", string(s.Value))
		}
		err = <-cerr
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	flag.Parse()
	var err error
	tbl, err = table.Create(table.TableOption{
		BaseDirectory: *tablePath,
		KeepSnapshots: true,
	})
	if err != nil {
		log.Println(err)
		return
	}
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(*addr, nil)
}
