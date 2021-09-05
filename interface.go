package mapper

import (
	"go/types"
)

type InterfaceVisitor struct {
	methods map[string]*Func
}

func (v *InterfaceVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Interface:
		v.methods = newInterfaceMethods(u)
		return false
	}
	return true
}

func NewInterfaceMethods(T types.Type) map[string]*Func {
	v := &InterfaceVisitor{}
	_ = Walk(v, T)
	return v.methods
}

func newInterfaceMethods(in *types.Interface) map[string]*Func {
	result := make(map[string]*Func)
	for i := 0; i < in.NumMethods(); i++ {
		fn := NewFunc(in.Method(i))
		result[fn.Name] = fn
	}
	return result
}
