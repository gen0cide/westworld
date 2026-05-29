// Command parsecheck parses + statically validates every generated
// .routine under examples/scenarios, reporting any that fail. It's the
// cheap offline gate before a live sweep — catches DSL/validator errors
// without needing the server.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/gen0cide/westworld/runtime"
)

func main() {
	root := "examples/scenarios"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	var files []string
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && filepath.Ext(p) == ".routine" {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)

	var fails int
	for _, f := range files {
		if _, err := runtime.ParseRoutineFile(f); err != nil {
			fmt.Printf("FAIL %s\n     %v\n", f, err)
			fails++
		}
	}
	fmt.Printf("\nparsed %d files, %d failed\n", len(files), fails)
	if fails > 0 {
		os.Exit(1)
	}
}
