package main

import examples "github.com/alextanhongpin/mapper/examples"

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	AtoB(A) B
	ExternalAtoB(examples.A) examples.B
}

// Automatically maps methods. However, tags does not work if you use methods.
type A struct {
	id    int
	str   string
	bool  bool
	slice []string
	m     map[string]int
}

func (a *A) ID() int            { return a.id }
func (a A) Str() string         { return a.str }
func (a A) Bool() bool          { return a.bool }
func (a A) Slice() []string     { return a.slice }
func (a A) Map() map[string]int { return a.m }

type B struct {
	ID    int
	Str   string
	Bool  bool
	Slice []string
	Map   map[string]int
}
