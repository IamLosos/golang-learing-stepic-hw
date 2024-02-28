package main

import (
	"fmt"
	"io" // see https://habrahabr.ru/post/306914/
	"io/fs"
	"os" // see https://pkg.go.dev/os#Open
	"sort"
)

const (
	tPref = "├───"
	gPref = "└───"
	lPref = "│\t"
	ePref = "\t"
)

func getCurrentPrefix(isLast bool) string {
	if isLast {
		return gPref
	}

	return tPref
}

func getFollowingPrefix(isLast bool) string {
	if isLast {
		return ePref
	}

	return lPref
}

func SortFileNameAscend(files []fs.DirEntry) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
}

func filterNonDirectory(files []fs.DirEntry) []fs.DirEntry {
	i := 0 // output index
	for _, x := range files {
		if x.IsDir() {
			// copy and increment index
			files[i] = x
			i++
		}
	}

	return files[0:i]
}

func getFileFormatedString(f fs.DirEntry) string {
	if f.IsDir() {
		return f.Name()
	}

	var sizeString string
	fi, err := f.Info()
	if err != nil {
		panic(err.Error())
	}

	s := fi.Size()
	if s == 0 {
		sizeString = "empty"
	} else {
		sizeString = fmt.Sprintf("%db", s)
	}

	return fmt.Sprintf("%s (%s)", f.Name(), sizeString)
}

func dirTreeRecursion(out io.Writer, path string, printFiles bool, prevPrefix string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		panic(err.Error())
	}

	if !printFiles {
		files = filterNonDirectory(files)
	}

	if len(files) == 0 {
		return nil
	}

	SortFileNameAscend(files)

	for i, file := range files {
		isLast := len(files)-1 == i
		// fmt.Printf("idx %d of %d , isLast = %t\n", i, len(files), isLast)
		// fmt.Printf("prev: %s ; cur: %s ; fol: %s\n", prevPrefix, getCurrentPrefix(isLast), getFollowingPrefix(isLast))

		//fmt.Printf("%s%s%s\n", prevPrefix, getCurrentPrefix(isLast), file.Name())
		fmt.Fprintf(out, "%s%s%s\n", prevPrefix, getCurrentPrefix(isLast), getFileFormatedString(file))

		if file.IsDir() {
			dirTreeRecursion(out, fmt.Sprintf("%s%s%s",
				path, string(os.PathSeparator),
				file.Name()),
				printFiles,
				fmt.Sprintf("%s%s", prevPrefix, getFollowingPrefix(isLast)))
		}
	}

	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return dirTreeRecursion(out, path, printFiles, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}

	// fmt.Println(os.Args[0])
	// fmt.Println(os.Args[1])
	// fmt.Println(os.Args[2])

	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
	//`go run main.go . -f`
}
