package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
	. "github.com/dave/jennifer/jen"
)

type InputTypeVisitor struct {
	code     *Statement
	variadic bool
}

func NewInputTypeVisitor(variadic bool) *InputTypeVisitor {
	return &InputTypeVisitor{
		code:     Null(),
		variadic: variadic,
	}
}

func (v *InputTypeVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Pointer:
		v.code = v.code.Op("*")
	case *types.Slice:
		if v.variadic {
			v.code = v.code.Op("...")
		} else {
			v.code = v.code.Index()
		}
	case *types.Array:
		if v.variadic {
			v.code = v.code.Op("...")
		} else {
			v.code = v.code.Index(Lit(u.Len()))
		}
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

func GenerateInputType(T types.Type, variadic bool) *Statement {
	v := NewInputTypeVisitor(variadic)
	_ = mapper.Walk(v, T)
	return v.code
}
