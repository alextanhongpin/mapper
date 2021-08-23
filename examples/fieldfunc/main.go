package main

import (
	"database/sql"
	"errors"
	"fmt"

	examples "github.com/alextanhongpin/mapper/examples"
	"github.com/google/uuid"
)

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type Mapper
type Mapper interface {
	// Supports error as second return parameter.
	AtoB(A) (B, error)
	SliceAtoB([]A) ([]B, error)
	VariadicAtoB(...A) ([]B, error)

	// No errors.
	CtoD(C) D
	SliceCtoD([]C) []D
	VariadicCtoD(...C) []D

	ConvertImportedFunc(examples.CustomField) (CustomField, error)
	ConvertImportedFuncPointer(examples.CustomField) (*CustomField, error)
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

	Remarks      sql.NullString `map:",NullStringToPointer"`
	RemarksError sql.NullString `map:",NullStringToPointerError"`
	PtrString    *string        `map:",PointerStringToNullString"`
}

type B struct {
	ID           string
	IDs          []uuid.UUID
	ExternalID   string
	Nums         []int
	UUID         uuid.UUID
	Remarks      *string
	RemarksError *string
}

type C struct {
	ID int `map:",IntToString"`
}

type D struct {
	ID string
}

type CustomField struct {
	Num int
}

// IntToString that resides locally.
func IntToString(i int) string {
	return fmt.Sprint(i)
}

func NullStringToPointer(str sql.NullString) *string {
	if str.Valid {
		return &str.String
	}
	return nil
}

func NullStringToPointerError(str sql.NullString) (*string, error) {
	if str.Valid {
		return &str.String, nil
	}
	return nil, errors.New("not found")
}

func PointerStringToNullString(in *string) sql.NullString {
	if in == nil || *in == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		Valid:  true,
		String: *in,
	}
}
