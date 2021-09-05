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
