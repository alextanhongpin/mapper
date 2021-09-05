package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type UnderlyingVisitor struct {
	u types.Type
}

func (v *UnderlyingVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Named, *types.Map, *types.Basic, *types.Struct:
		v.u = u
		return false
	default:
		return true
	}
}

// NewUnderlyingType extracts the underlying type of a Type.
func NewUnderlyingType(T types.Type) types.Type {
	v := &UnderlyingVisitor{}
	_ = mapper.Walk(v, T)
	return v.u
}
