package mapper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alextanhongpin/pkg/stringcase"
)

// FooBar becomes foo_bar_gen.go
func fileNameFromType(typeName string) string {
	// FooBar to foo_bar
	fileName := stringcase.SnakeCase(typeName)

	// foo_bar to foo_bar_gen.go
	fileName = fmt.Sprintf("%s_gen.go", fileName)

	return fileName
}

// fullPath returns the full path to the package, relative to the caller.
func fullPath(rel string) string {
	path, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("mapper: error getting working directory: %w", err))
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

func hasExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".go"
}
