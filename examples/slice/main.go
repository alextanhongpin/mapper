package main

import (
	"strings"

	examples "github.com/alextanhongpin/mapper/examples"
)

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	AtoB(A) B
	SliceAtoB([]A) []B
	ExternalAtoB([]examples.A) []examples.B
	VariadicError(...A) ([]B, error)
	Variadic(...A) []B
	CtoD(C) (D, error)
}

type A struct {
	ID    int
	Str   string
	Bool  bool
	Slice []string
	Ints  []int
	Map   map[string]int
}

type B struct {
	ID    int
	Str   string
	Bool  bool
	Slice []string `map:",Upper"`
	Ints  []int    `map:",AddOne"`
	Map   map[string]int
}

type C struct {
	Ints  []string
	Items []string
}

type D struct {
	Ints  []int  `map:",strconv/Atoi"`
	Items string `map:",Join"`
}

func Upper(in []string) []string {
	result := make([]string, len(in))
	for i, s := range in {
		result[i] = strings.ToUpper(s)
	}
	return result
}

func AddOne(n int) int {
	return n + 1
}

func Join(in []string) (string, error) {
	return strings.Join(in, ","), nil
}
