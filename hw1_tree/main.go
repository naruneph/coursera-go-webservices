package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var prefix []bool

func getLastFile(path string, printFiles bool) (lastFile string, err error) {
	dir := filepath.Dir(path)
	files, err := ioutil.ReadDir(dir)

	if !printFiles {
		for i := len(files) - 1; i >= 0; i-- {
			if files[i].IsDir() {
				lastFile = files[i].Name()
				break
			}
		}

	} else {
		lastFile = files[len(files)-1].Name()
	}
	return

}

func makeLine(path string, root string, size int64, lastFile string) string {
	root = strings.TrimRight(root, string(os.PathSeparator))
	root = strings.TrimLeft(root, string(os.PathSeparator))
	rootItems := strings.Split(root, string(os.PathSeparator))
	items := strings.Split(path, string(os.PathSeparator))

	for i, v := range items {
		if v == rootItems[len(rootItems)-1] {
			items = items[i+1:]
			break
		}
	}

	res := ""
	depth := len(items)

	switch {
	case len(prefix) < depth && lastFile == items[len(items)-1]:
		prefix = append(prefix, false)
	case len(prefix) < depth && lastFile != items[len(items)-1]:
		prefix = append(prefix, true)
	case len(prefix) >= depth && lastFile == items[len(items)-1]:
		prefix[depth-1] = false
	case len(prefix) >= depth && lastFile != items[len(items)-1]:
		prefix[depth-1] = true
	}

	for i := 0; i < depth-1; i++ {
		switch prefix[i] {
		case false:
			res += "\t"
		case true:
			res += "│\t"
		}
	}

	if lastFile == items[len(items)-1] {
		res += "└───"
	} else {
		res += "├───"
	}

	res += items[len(items)-1]

	if size > 0 {
		res += fmt.Sprintf("%s%d%s", " (", size, "b)")
	} else if size == 0 {
		res += " (empty)"
	}

	return res
}

func dirTree(out io.Writer, root string, printFiles bool) error {

	var resList []string
	var res string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !info.IsDir() && !printFiles {
			return nil
		}

		if path != root {

			lastFile, err := getLastFile(path, printFiles)

			if err != nil {
				return err
			}

			if !info.IsDir() {
				res = makeLine(path, root, info.Size(), lastFile)
			} else {
				res = makeLine(path, root, -1, lastFile)
			}
			if res != "" {
				resList = append(resList, res)
			}

		}

		return nil
	})

	for _, v := range resList {
		fmt.Fprintf(out, "%s\n", v)
	}

	return err

}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
