package main

import (
	"fmt"

	"github.com/google/uuid"
)

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	// Supports error as second return parameter.
	AtoB(A) (B, error)
}

type A struct {
	// Defining local function to perform field conversion.
	ID int `map:",IntToString"`

	// Defining external function to perform field conversion.
	ExternalID int `map:",github.com/alextanhongpin/mapper/examples/IntToString"`

	// Another example of external function, which returns error as second return
	// parameter.
	UUID string `map:",github.com/google/uuid/Parse"`
}

type B struct {
	ID         string
	UUID       uuid.UUID
	ExternalID string
}

// IntToString that resides locally.
func IntToString(i int) string {
	return fmt.Sprint(i)
}
