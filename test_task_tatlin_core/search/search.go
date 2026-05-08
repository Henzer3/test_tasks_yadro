package search

import (
	"bufio"
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
)

type names struct {
	name  string
	count int
}

type fileStatNames struct {
	fileName string
	names    []names
}

func New(fileName string) (*fileStatNames, error) {
	res := &fileStatNames{
		fileName: fileName,
	}

	if err := res.getStat(); err != nil {
		return nil, err
	}

	return res, nil
}

func (f *fileStatNames) DoAscendingOrder() {
	slices.SortFunc(f.names, func(a, b names) int {
		return cmp.Compare(a.count, b.count)
	})
}

func (f *fileStatNames) DoDescendingOrder() {
	slices.SortFunc(f.names, func(a, b names) int {
		return cmp.Compare(b.count, a.count)
	})
}

func (f *fileStatNames) ShowStats() {
	fmt.Printf("Names in file: %s\n\n", f.fileName)
	for _, v := range f.names {
		fmt.Printf("%s %d\n", v.name, v.count)
	}
}

func (f *fileStatNames) getStat() (statErr error) {
	file, err := os.Open(f.fileName)
	if err != nil {
		return fmt.Errorf("cant open file: %s err: %w", f.fileName, err)
	}

	defer func() {
		if err := file.Close(); err != nil && statErr == nil {
			statErr = fmt.Errorf("cant close file: %s, err: %w", f.fileName, err)
		}
	}()

	namesMap := make(map[string]int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())

		if name == "" {
			continue
		}

		namesMap[name] += 1

	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan file error, file: %s err: %w", f.fileName, err)
	}

	nameSlice := make([]names, 0, len(namesMap))

	for name, count := range namesMap {
		nameSlice = append(nameSlice, names{name: name, count: count})
	}

	f.names = nameSlice

	return nil
}
