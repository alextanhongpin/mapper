package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type FuncParamVisitor struct {
	fields       mapper.StructFields
	methods      map[string]*mapper.Func
	isCollection bool
}

func NewFuncParamVisitor() *FuncParamVisitor {
	return &FuncParamVisitor{}
}

func (v *FuncParamVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Slice, *types.Array:
		v.isCollection = true
	case *types.Named:
		v.methods = mapper.ExtractNamedMethods(u)
		return true
	case *types.Struct:
		v.fields = mapper.ExtractStructFields(u)
		return false
	default:
		panic("not implemented")
	}
	return false
}
