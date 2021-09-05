package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
	. "github.com/dave/jennifer/jen"
)

type TypeVisitor struct {
	code *Statement
}

func NewTypeVisitor() *TypeVisitor {
	return &TypeVisitor{
		code: Null(),
	}
}

func (v *TypeVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Pointer:
		v.code = v.code.Op("*")
	case *types.Slice:
		v.code = v.code.Index()
	case *types.Array:
		v.code = v.code.Index(Lit(u.Len()))
	case *types.Named:
		o := u.Obj()
		p := o.Pkg()
		v.code = v.code.Qual(p.Path(), o.Name())
		return false
	default:
		v.code = v.code.Id(u.String())
	}
	return true
}

func GenerateType(T types.Type) *Statement {
	v := NewTypeVisitor()
	_ = mapper.Walk(v, T)
	return v.code
}
