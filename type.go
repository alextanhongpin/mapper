package mapper

import (
	"go/types"
)

type Visitor interface {
	Visit(T types.Type) bool
}

func Walk(visitor Visitor, T types.Type) bool {
	if next := visitor.Visit(T); !next {
		return next
	}

	switch u := T.(type) {
	case *types.Named:
		return Walk(visitor, u.Underlying())
	case *types.Pointer:
		return Walk(visitor, u.Elem())
	case *types.Array:
		return Walk(visitor, u.Elem())
	case *types.Slice:
		return Walk(visitor, u.Elem())
	case *types.Map:
		return Walk(visitor, u.Elem())
	default:
		return types.IdenticalIgnoreTags(T, u)
	}
}
func IsPointer(T types.Type) bool {
	_, ok := T.(*types.Pointer)
	return ok
}

func IsStruct(T types.Type) bool {
	_, ok := T.Underlying().(*types.Struct)
	return ok
}

func IsSlice(T types.Type) bool {
	_, ok := T.(*types.Slice)
	return ok
}

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
	_ = Walk(v, T)
	return v.u
}

func IsUnderlyingError(T types.Type) bool {
	U := NewUnderlyingType(T)
	return U.String() == "error"
}

func IsUnderlyingIdentical(lhs, rhs types.Type) bool {
	return UnderlyingSignature(lhs) == UnderlyingSignature(rhs)
}

func UnderlyingSignature(T types.Type) string {
	return types.TypeString(NewUnderlyingType(T), nil)
}

func IsIdentical(lhs, rhs types.Type) bool {
	return types.TypeString(lhs, nil) == types.TypeString(rhs, nil)
}
