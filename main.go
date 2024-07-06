package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SortOrder int

type Duplicate struct {
	Size  int64
	Hash  string
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
	data := getData(os.Args[1], format)

	sortSizes(data, direction)
	showSizeInfo(data)

	if checkDuplicates() {
		showDuplicates(data)
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

func getData(root, format string) *[]Duplicate {
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

	data := make([]Duplicate, 0, len(duplicates))
	for size, files := range duplicates {
		if len(files) > 1 {
			data = append(data, Duplicate{Size: size, Files: files})
		}
	}

	return &data
}

func isValidFormat(path string, format string) bool {
	return format == "" || strings.TrimPrefix(filepath.Ext(path), ".") == format
}

func sortSizes(data *[]Duplicate, direction SortOrder) {
	sort.Slice(*data, func(i, j int) bool {
		if direction == sortDesc {
			return (*data)[i].Size > (*data)[j].Size
		}
		return (*data)[i].Size < (*data)[j].Size
	})
}

func showSizeInfo(data *[]Duplicate) {
	for _, info := range *data {
		fmt.Println()
		fmt.Printf("%d bytes\n", info.Size)
		for _, filename := range info.Files {
			fmt.Println(filename)
		}
	}
}

func checkDuplicates() bool {
	var answer string

	for {
		fmt.Println()
		fmt.Println("Check for duplicates?")

		fmt.Scanln(&answer)

		switch answer {
		case "yes":
			return true
		case "no":
			return false
		default:
			fmt.Println()
			fmt.Println("Wrong option")
		}
	}
}

func showDuplicates(data *[]Duplicate) {
	var i = 0

	for _, info := range *data {
		duplicates := findDuplicates(info.Files)
		if len(duplicates) == 0 {
			continue
		}

		fmt.Println()
		fmt.Printf("%d bytes\n", info.Size)
		for _, duplicate := range duplicates {
			fmt.Printf("Hash: %s\n", duplicate.Hash)
			for _, filename := range duplicate.Files {
				i++
				fmt.Printf("%d. %s\n", i, filename)
			}
		}
	}
}

func findDuplicates(files []string) (duplicates []Duplicate) {
	hashedFiles := make(map[string][]string)
	for _, filename := range files {
		hash := getFileHash(filename)
		hashedFiles[hash] = append(hashedFiles[hash], filename)
	}

	for hash, hf := range hashedFiles {
		if len(hf) > 1 {
			duplicates = append(duplicates, Duplicate{Hash: hash, Files: hf})
		}
	}

	return
}

func getFileHash(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}
