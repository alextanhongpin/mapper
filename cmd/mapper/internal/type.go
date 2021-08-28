package internal

import (
	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
)

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
