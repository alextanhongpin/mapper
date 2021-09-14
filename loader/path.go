package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alextanhongpin/pkg/stringcase"
)

// FullPath returns the full path to the package, relative to the caller.
func FullPath(rel string) string {
	path, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get package directory: %v", err))
	}
	path = filepath.Join(path, rel)
	return path
}

// addSuffixToFileName adds a suffix to the filename, before the extension, to allow main.go -> main_gen.go
func addSuffixToFileName(path, suffix string) string {
	ext := filepath.Ext(path)
	path = path[:len(path)-len(ext)]
	return path + suffix + ext
}

// safeAddSuffixToFileName only adds the suffix if the user generated name does not already contains the suffix.
func safeAddSuffixToFileName(path, suffix string) string {
	if strings.Contains(path, suffix) {
		return path
	}
	return addSuffixToFileName(path, suffix)
}

func isFile(path string) bool {
	return filepath.Ext(path) != ""
}

func safeAddFileName(path, fileName string) string {
	if isFile(path) {
		return path
	}
	return filepath.Join(path, fileName)
}

func FileNameFromTypeName(input, output, typename string) string {
	if isFile(output) {
		return output
	}

	dir := FullPath(output)
	if output == "" {
		dir = filepath.Dir(FullPath(input))
	}

	// Foo becomes foo
	fileName := fmt.Sprintf("%s.go", stringcase.SnakeCase(typename))

	// foo becomes foo_gen
	fileName = safeAddSuffixToFileName(fileName, "_gen")

	// path/to/main.go becomes path/to/foo_gen.go
	return safeAddFileName(dir, fileName)
}

func FileName(p string) string {
	file := filepath.Base(p)
	ext := filepath.Ext(file)
	if ext == "" {
		return file
	}
	return file[:len(file)-len(ext)]
}
