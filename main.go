package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type SortOrder int

type SizeInfo struct {
	Size  int64
	Files []string
}

const (
	sortDesc SortOrder = iota + 1
	sortAsc
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Directory is not specified")
		return
	}

	format := getFileFormat()
	direction := getSortDirection()
	data := getDuplicates(os.Args[1], format)

	sortSizes(&data, direction)

	for _, info := range data {
		fmt.Println()
		fmt.Printf("%d bytes\n", info.Size)
		for _, filename := range info.Files {
			fmt.Println(filename)
		}
	}
}

func getFileFormat() (ext string) {
	fmt.Println()
	fmt.Println("Enter file format:")
	fmt.Scanln(&ext)
	return
}

func getSortDirection() (sort SortOrder) {
	fmt.Println()
	fmt.Println("Size sorting options:")
	fmt.Println("1. Descending")
	fmt.Println("2. Ascending")

	for {
		fmt.Scan(&sort)
		if sort == sortDesc || sort == sortAsc {
			return
		}

		fmt.Println()
		fmt.Println("Wrong option")
		fmt.Println()
		fmt.Println("Enter a sorting option:")
	}
}

func getDuplicates(root, format string) []SizeInfo {
	duplicates := make(map[int64][]string)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isValidFormat(path, format) {
			size := info.Size()
			duplicates[size] = append(duplicates[size], path)
		}

		return nil
	})

	data := make([]SizeInfo, 0, len(duplicates))
	for size, files := range duplicates {
		if len(files) > 1 {
			data = append(data, SizeInfo{Size: size, Files: files})
		}
	}

	return data
}

func isValidFormat(path string, format string) bool {
	return format == "" || filepath.Ext(path)[1:] == format
}

func sortSizes(data *[]SizeInfo, direction SortOrder) {
	sort.Slice(*data, func(i, j int) bool {
		if direction == sortDesc {
			return (*data)[i].Size > (*data)[j].Size
		}
		return (*data)[i].Size < (*data)[j].Size
	})
}
