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
