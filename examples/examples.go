package examples

import (
	"fmt"
	"path/filepath"
)

func IntToString(i int) string {
	return fmt.Sprint(i)
}

type URLBuilder struct {
	Domain string
}

func (u URLBuilder) Build(path string) string {
	return filepath.Join(u.Domain, path)
}

type URLer interface {
	Build(path string) string
}

type A struct {
	ID    int
	Str   string
	Bool  bool
	Slice []string
	Map   map[string]int
}

type B struct {
	ID    int
	Str   string
	Bool  bool
	Slice []string
	Map   map[string]int
}
