package main

import (
	"strconv"
	"time"

	examples "github.com/alextanhongpin/mapper/examples"
)

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	AtoB(A) B
	ExternalAtoB(examples.A) examples.B
	CtoD(C) (D, error)
}

type Status string

// Automatically maps methods. However, tags does not work if you use methods.
type A struct {
	id     int
	str    string
	bool   bool
	slice  []string
	m      map[string]int
	t      time.Time
	status Status `map:"Status(),StatusToString"`
}

func (a *A) ID() int            { return a.id }
func (a A) Str() string         { return a.str }
func (a A) Bool() bool          { return a.bool }
func (a A) Slice() []string     { return a.slice }
func (a A) Map() map[string]int { return a.m }
func (a A) Time() time.Time     { return a.t }
func (a A) Status() Status      { return a.status }

type B struct {
	ID     int
	Str    string
	Bool   bool
	Slice  []string
	Map    map[string]int
	Time   time.Time
	Status string
}

type C struct {
	id string
}

func (c C) ID() (int, error) {
	return strconv.Atoi(c.id)
}

type D struct {
	ID int
}

func StatusToString(s Status) string {
	return string(s)
}
