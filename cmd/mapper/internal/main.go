package internal

import (
	"fmt"

	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
)

type Resolver interface {
	VarLhs() *jen.Statement
	VarRhs() *jen.Statement

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
	return fmt.Sprintf("%s%d", name, index)
}

func GenTypeName(T *mapper.Type) *jen.Statement {
	if T.PkgPath != "" {
		return jen.Qual(T.PkgPath, T.Type)
	}
	return jen.Id(T.Type)
}

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
