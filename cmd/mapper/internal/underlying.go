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
	case *types.Named, *types.Map, *types.Basic:
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

func IsUnderlyingError(T types.Type) bool {
	U := NewUnderlyingType(T)
	return U.String() == "error"
}
