package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
	. "github.com/dave/jennifer/jen"
)

type OutputTypeVisitor struct {
	code *Statement
}

func NewOutputTypeVisitor() *OutputTypeVisitor {
	return &OutputTypeVisitor{
		code: Null(),
	}
}

func (v *OutputTypeVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case
		*types.Pointer,
		*types.Slice,
		*types.Array:
		v.code = v.code.Nil()
		return false
	case *types.Named:
		o := u.Obj()
		p := o.Pkg()
		v.code = v.code.Qual(p.Path(), o.Name()).Values()
		return false
	default:
		v.code = v.code.Id(u.String())
	}
	return true
}

func GenerateOutputType(T types.Type, hasError bool) *Statement {
	v := NewOutputTypeVisitor()
	_ = mapper.Walk(v, T)
	if hasError {
		return List(v.code, Err())
	}
	return v.code
}
