package main

import examples "github.com/alextanhongpin/mapper/examples"

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	AtoB(A) B
	ExternalAtoB(examples.A) examples.B
}

type A struct {
	ID    int
	Str   string
	Bool  bool
	Slice []string
	Map   map[string]int
}

type B struct {
	// Map ID field only.
	ID    int
	Str   string         `json:"str" map:"-"`
	Bool  bool           `map:"-"`
	Slice []string       `map:"-"`
	Map   map[string]int `map:"-"`
}
