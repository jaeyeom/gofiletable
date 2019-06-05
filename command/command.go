// Binary command is a command line tool for tables.
package main // import "github.com/jaeyeom/gofiletable/command"

import (
	"flag"
	"fmt"
	"log"

	"github.com/jaeyeom/gofiletable/table"
)

// ls prints the list of keys of each path.
func ls(tablePaths []string) {
	for _, tablePath := range tablePaths {
		tbl, err := table.Create(table.TableOption{
			BaseDirectory: tablePath,
			KeepSnapshots: true,
		})
		if err != nil {
			log.Println("Error on path", tablePath, ":", err)
			return
		}
		for key := range tbl.Keys() {
			fmt.Println(string(key))
		}
	}
}

// cat prints the value of the key.
func cat(tablePath string, key string) {
	tbl, err := table.Create(table.TableOption{
		BaseDirectory: tablePath,
		KeepSnapshots: true,
	})
	if err != nil {
		log.Println(err)
		return
	}
	value, err := tbl.Get([]byte(key))
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(value))
}

// help prints help message. If cmd is empty, prints the list of commands.
func help(cmd string) {
	helpDetails := map[string]string{
		"ls":  "ls path [path...] - prints list of keys from each path",
		"cat": "cat path key - prints the value",
	}
	if cmd == "" {
		fmt.Println("Available commands are:")
		for cmd, details := range helpDetails {
			fmt.Println(cmd, ":", details)
		}
	} else {
		fmt.Println(helpDetails[cmd])
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || (len(args) == 1 && args[0] == "help") {
		help("")
		return
	}
	if len(args) == 2 && args[0] == "help" {
		help(args[1])
		return
	}
	cmd := args[0]
	if cmd == "ls" {
		if len(args) < 2 {
			help("ls")
			return
		}
		ls(args[1:])
		return
	}
	if cmd == "cat" {
		if len(args) != 3 {
			help("cat")
			return
		}
		cat(args[1], args[2])
	}
}
