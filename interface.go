package mapper

import (
	"go/types"
)

type InterfaceVisitor struct {
	methods map[string]*Func
	obj     *types.TypeName
}

func (v *InterfaceVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Named:
		v.obj = u.Obj()
	case *types.Interface:
		v.methods = newInterfaceMethods(u, v.obj)
		return false
	}
	return true
}

func NewInterfaceMethods(T types.Type) map[string]*Func {
	v := &InterfaceVisitor{}
	_ = Walk(v, T)
	return v.methods
}

func newInterfaceMethods(in *types.Interface, obj *types.TypeName) map[string]*Func {
	result := make(map[string]*Func)
	for i := 0; i < in.NumMethods(); i++ {
		fn := NewFunc(in.Method(i), obj)
		result[fn.Name] = fn
	}
	return result
}
