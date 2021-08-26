package internal

import (
	"fmt"

	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
)

// Resolver ...
type Resolver interface {
	LhsVar() *jen.Statement
	RhsVar() *jen.Statement

	Lhs() interface{}
	Rhs() mapper.StructField

	LhsType() *jen.Statement
	RhsType() *jen.Statement

	Tag() *mapper.Tag
	//Mapper() interface{}

	// Increases the assignment count
	Assign()

	IsField() bool
	IsMethod() bool
}

func argsWithIndex(name string, index int) string {
	if index < 0 {
		index = 0
	}
	return fmt.Sprintf("%s%d", name, index)
}

// GenTypeName generates the element type.
func GenTypeName(T *mapper.Type) *jen.Statement {
	if T.PkgPath != "" {
		return jen.Qual(T.PkgPath, T.Type)
	}
	return jen.Id(T.Type)
}

// GenType generates the basic type.
func GenType(T *mapper.Type) *jen.Statement {
	return jen.Do(func(s *jen.Statement) {
		if T.IsSlice {
			s.Add(jen.Index())
		}
		if T.IsPointer {
			s.Add(jen.Op("*"))
		}
	}).Add(GenTypeName(T))
}
