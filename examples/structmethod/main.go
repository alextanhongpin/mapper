package main

import (
	"path/filepath"
)

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	// Supports error as second return parameter.
	AtoB(A) (B, error)
}

type A struct {
	URL string

	ExternalURL string
}

type B struct {
	// Defining local function to perform field conversion.
	URL string `map:",URLBuilder.Build"`

	// Defining external function to perform field conversion.
	ExternalURL string `map:",github.com/alextanhongpin/mapper/examples/URLBuilder.Build"`
}

type URLBuilder struct {
	Domain string
}

func (u URLBuilder) Build(path string) (string, error) {
	return filepath.Join(u.Domain, path), nil
}
