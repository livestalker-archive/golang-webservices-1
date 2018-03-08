package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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

func (fl FileList) OnlyDirs() FileList {
	var count int
	for _, item := range fl {
		if item.IsDir() {
			count++
		}
	}
	newFileList := make(FileList, count)
	var i int
	for _, item := range fl {
		if item.IsDir() {
			newFileList[i] = item
			i++
		}
	}
	return newFileList
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
	if !printFiles {
		files = FileList(files).OnlyDirs()
	}
	sort.Sort(FileList(files))
	for ix, item := range files {
		var prefix, newIdent, strSize string
		if ix == len(files)-1 {
			prefix = fmt.Sprintf("%s%s%s", indent, DELIM_END, DELIM_HORIZON)
			newIdent = indent + "\t"
		} else {
			prefix = fmt.Sprintf("%s%s%s", indent, DELIM_TRIPLE, DELIM_HORIZON)
			newIdent = indent + DELIM_VERTICAL + "\t"
		}
		if item.IsDir() {
			fmt.Fprintf(out, "%s%s\n", prefix, item.Name())
			outFiles(out, filepath.Join(path, item.Name()), printFiles, newIdent)
		} else {
			if item.Size() == 0 {
				strSize = "empty"
			} else {
				strSize = fmt.Sprintf("%db", item.Size())
			}
			fmt.Fprintf(out, "%s%s (%s)\n", prefix, item.Name(), strSize)
		}
	}
	return nil
}
