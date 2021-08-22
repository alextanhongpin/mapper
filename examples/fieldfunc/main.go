package main

import (
	"fmt"

	"github.com/google/uuid"
)

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	// Supports error as second return parameter.
	AtoB(A) (B, error)
	SliceAtoB([]A) ([]B, error)
	VariadicAtoB(...A) ([]B, error)
}

type A struct {
	// Defining local function to perform field conversion.
	ID int `map:",IntToString"`

	// Automatically maps if the input and output are both collection.
	IDs []string `map:",github.com/google/uuid/Parse"`

	// Defining external function to perform field conversion.
	ExternalID int `map:",github.com/alextanhongpin/mapper/examples/IntToString"`

	// Use standard packages.
	//Num  int      `map:",fmt/Sprint"` // NOTE: This does not work because fmt.Sprint input is interface, however, we check if the input type matches string.
	Nums []string `map:",strconv/Atoi"`

	// Another example of external function, which returns error as second return
	// parameter.
	UUID string `map:",github.com/google/uuid/Parse"`
}

type B struct {
	ID         string
	IDs        []uuid.UUID
	ExternalID string
	Nums       []int
	UUID       uuid.UUID
}

// IntToString that resides locally.
func IntToString(i int) string {
	return fmt.Sprint(i)
}
