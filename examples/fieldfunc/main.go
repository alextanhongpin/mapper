package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

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
	ID           int
	IDs          []string
	ExternalID   int
	Nums         []string
	UUID         string
	Remarks      sql.NullString
	RemarksError sql.NullString
	PtrString    *string
}

type B struct {
	// Defining local function to perform field conversion.
	ID string `map:",IntToString"`

	// Another field, referring to the same source field.
	AliasID string `map:"ID,IntToString"`

	// Automatically maps if the input and output are both collection.
	IDs []uuid.UUID `map:",github.com/google/uuid/Parse"`

	// Defining external function to perform field conversion.
	ExternalID string `map:",github.com/alextanhongpin/mapper/examples/IntToString"`

	// Use standard packages.
	//Num  int      `map:",fmt/Sprint"` // NOTE: This does not work because fmt.Sprint input is interface, however, we check if the input type matches string.
	Nums []int `map:",strconv/Atoi"`

	// Another example of external function, which returns error as second return
	// parameter.
	UUID         uuid.UUID      `map:",github.com/google/uuid/Parse"`
	Remarks      *string        `map:",NullStringToPointer"`
	RemarksError *string        `map:",NullStringToPointerError"`
	PtrString    sql.NullString `map:",PointerStringToNullString"`
}

type C struct {
	ID int
}

type D struct {
	ID string `map:",IntToString"`
}

type CustomField struct {
	Num int `map:",StringToInt"`
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

func StringToInt(str string) (int, error) {
	return strconv.Atoi(str)
}
