package main

import examples "github.com/alextanhongpin/mapper/examples"

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	AtoB(A) *B
	ExternalAtoB(examples.A) *examples.B
	CtoD(C) D
}

type A struct {
	ID    int
	Str   string
	Bool  bool
	Slice []string
	Map   map[string]int
	Ptr   *C
}

type B struct {
	ID    int
	Str   string
	Bool  bool
	Slice []string
	Map   map[string]int
	Ptr   *D
}

type C struct {
	Name string
}

type D struct {
	Name string
}
