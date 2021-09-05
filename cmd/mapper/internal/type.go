package internal

import (
	"go/types"

	"github.com/dave/jennifer/jen"
)

// GenTypeName generates the element type.
func GenTypeName(T types.Type) *jen.Statement {
	U := NewUnderlyingType(T)
	switch u := U.(type) {
	case *types.Named:
		return jen.Qual(u.Obj().Pkg().Path(), u.Obj().Name())
	default:
		return jen.Id(u.String())
	}
}

// GenType generates the basic type.
func GenType(T types.Type) *jen.Statement {
	return GenerateType(T)
}
