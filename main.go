package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type SortOrder int

type FileInfo struct {
	Size int64
	Path string
}

type GroupInfo struct {
	Size  int64
	Hash  string
	Files []string
	Index int
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
	files := getFiles(os.Args[1], format)

	filesBySize := groupBySize(files)
	sortFiles(filesBySize)
	showSizes(filesBySize)

	if !confirm("Check for duplicates?") {
		return
	}

	duplicates := getDuplicates(filesBySize)
	if len(*duplicates) == 0 {
		return
	}

	showDuplicates(duplicates)

	if !confirm("Delete files?") {
		return
	}

	deleteFiles(duplicates)
}

func getFileFormat() string {
	fmt.Println()
	fmt.Println("Enter file format:")
	return getString()
}

func getFiles(root, format string) *[]FileInfo {
	var files []FileInfo

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isValidFormat(path, format) {
			files = append(files, FileInfo{Size: info.Size(), Path: path})
		}

		return nil
	})

	return &files
}

func isValidFormat(path string, format string) bool {
	return format == "" || strings.TrimPrefix(filepath.Ext(path), ".") == format
}

func sortFiles(files *[]GroupInfo) {
	direction := getSortDirection()

	sort.Slice(*files, func(i, j int) bool {
		if direction == sortDesc {
			return (*files)[i].Size > (*files)[j].Size
		}
		return (*files)[i].Size < (*files)[j].Size
	})
}

func groupBySize(data *[]FileInfo) *[]GroupInfo {
	var groups map[int64][]string
	var duplicates []GroupInfo

	groups = make(map[int64][]string)

	for _, info := range *data {
		groups[info.Size] = append(groups[info.Size], info.Path)
	}

	for size, files := range groups {
		if len(files) > 1 {
			duplicates = append(duplicates, GroupInfo{Size: size, Files: files})
		}
	}

	return &duplicates
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

func showSizes(files *[]GroupInfo) {
	for _, info := range *files {
		fmt.Println()
		fmt.Printf("%d bytes\n", info.Size)
		for _, filename := range info.Files {
			fmt.Println(filename)
		}
	}
}

func confirm(prompt string) bool {
	var answer string

	for {
		fmt.Println()
		fmt.Println(prompt)

		answer = getString()

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

func getDuplicates(filesBySize *[]GroupInfo) *[]GroupInfo {
	var hashes []string
	var groupsByHash = make(map[string]GroupInfo)
	var duplicates []GroupInfo

	for _, info := range *filesBySize {
		for _, filename := range info.Files {
			hash := getFileHash(filename)

			if !slices.Contains(hashes, hash) {
				hashes = append(hashes, hash)
			}

			group, ok := groupsByHash[hash]
			if !ok {
				group = GroupInfo{Size: info.Size, Hash: hash}
			}

			group.Files = append(group.Files, filename)
			groupsByHash[hash] = group
		}
	}

	for _, hash := range hashes {
		group := groupsByHash[hash]
		if len(group.Files) > 1 {
			duplicates = append(duplicates, group)
		}
	}

	return &duplicates
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

func showDuplicates(data *[]GroupInfo) {
	var size int64 = -1
	var i int = 1

	for gi, group := range *data {
		if size != group.Size {
			fmt.Println()
			fmt.Printf("%d bytes\n", group.Size)
		}

		// save index for later
		(*data)[gi].Index = i

		fmt.Printf("Hash: %s\n", group.Hash)
		for _, filename := range group.Files {
			fmt.Printf("%d. %s\n", i, filename)
			i++
		}
	}
}

func deleteFiles(data *[]GroupInfo) {
	var numFiles = getNumFiles(data)
	var filesToDelete = getFileNumbers(numFiles)
	var size int64

	for _, pos := range filesToDelete {
		info := getFileInfo(data, pos)
		os.Remove(info.Path)
		size += info.Size
	}

	fmt.Println()
	fmt.Printf("Total freed up space: %d bytes\n", size)
}

func getNumFiles(data *[]GroupInfo) (total int) {
	for _, group := range *data {
		total += len(group.Files)
	}
	return
}

func getFileNumbers(total int) []int {
	var input string
	var numbers []int

	for {
		fmt.Println()
		fmt.Println("Enter file numbers to delete:")

		input = getString()
		numbers = numbers[:0] // reset length

		scanner := bufio.NewScanner(strings.NewReader(input))
		scanner.Split(bufio.ScanWords)

		ok := true
		for scanner.Scan() {
			num, err := strconv.Atoi(scanner.Text())
			if err != nil || num > total {
				ok = false
				break
			}

			numbers = append(numbers, num)
		}

		if ok {
			return numbers
		}

		fmt.Println()
		fmt.Println("Wrong format")
	}
}

func getFileInfo(data *[]GroupInfo, pos int) FileInfo {
	for _, group := range *data {
		if pos-group.Index >= len(group.Files) {
			continue
		}

		return FileInfo{
			Size: group.Size,
			Path: group.Files[pos-group.Index],
		}
	}

	panic("Something went wrong")
}

func getString() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}
