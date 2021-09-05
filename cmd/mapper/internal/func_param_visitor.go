package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type FuncParamVisitor struct {
	fields       mapper.StructFields
	methods      map[string]*mapper.Func
	isCollection bool
	isPointer    bool
	obj          *types.TypeName
}

func NewFuncParamVisitor() *FuncParamVisitor {
	return &FuncParamVisitor{}
}

func (v *FuncParamVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Pointer:
		v.isPointer = true
	case *types.Slice, *types.Array:
		v.isCollection = true
	case *types.Named:
		v.methods = mapper.ExtractNamedMethods(u)
		v.obj = u.Obj()
	case *types.Struct:
		v.fields = mapper.ExtractStructFields(u)
		return false
	default:
		panic("not handled")
	}
	return true
}

func (v FuncParamVisitor) Fields() mapper.StructFields {
	return v.fields
}

func (v FuncParamVisitor) Methods() map[string]*mapper.Func {
	return v.methods
}

func (v FuncParamVisitor) IsCollection() bool {
	return v.isCollection
}

func (v FuncParamVisitor) IsPointer() bool {
	return v.isPointer
}
