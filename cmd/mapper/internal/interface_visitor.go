package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type InterfaceVisitor struct {
	methods map[string]*mapper.Func
}

func (v *InterfaceVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Interface:
		v.methods = mapper.ExtractInterfaceMethods(u)
	}
	return true
}

func GenerateInterfaceMethods(T types.Type) map[string]*mapper.Func {
	v := &InterfaceVisitor{}
	_ = mapper.Walk(v, T.Underlying())
	return v.methods
}
