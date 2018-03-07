package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	//"strings"
)

const (
	DELIM_TRIPLE   = "├"
	DELIM_END      = "└"
	DELIM_HORIZON  = "───"
	DELIM_VERTICAL = "│"
)

type FileList []os.FileInfo

func (fl FileList) Len() int {
	return len(fl)
}

func (fl FileList) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl FileList) Less(i, j int) bool {
	return fl[i].Name() < fl[j].Name()
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

func dirTree(out io.Writer, path string, printFiles bool) error {
	return outFiles(out, path, printFiles, "")
}

func outFiles(out io.Writer, path string, printFiles bool, indent string) error {
	files, _ := ioutil.ReadDir(path)
	sort.Sort(FileList(files))
	for ix, item := range files {
		if ix == len(files)-1 {
			fmt.Fprintf(out, "%s%s%s%s\n", indent, DELIM_END, DELIM_HORIZON, item.Name())
			if item.IsDir() {
				outFiles(out, filepath.Join(path, item.Name()), printFiles, indent+" ")
			}
		} else {
			fmt.Fprintf(out, "%s%s%s%s\n", indent, DELIM_TRIPLE, DELIM_HORIZON, item.Name())
			if item.IsDir() {
				outFiles(out, filepath.Join(path, item.Name()), printFiles, indent+DELIM_VERTICAL+" ")
			}
		}
	}
	return nil
}
