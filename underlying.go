package mapper

import "go/types"

type UnderlyingVisitor struct {
	t         types.Type
	u         types.Type
	isPointer bool
	isSlice   bool
	isArray   bool
	len       int64
	obj       *types.TypeName
	pkg       *types.Package
}

func NewUnderlyingVisitor(T types.Type) *UnderlyingVisitor {
	v := &UnderlyingVisitor{t: T}
	_ = Walk(v, T)
	return v
}

func (v *UnderlyingVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Slice:
		v.isSlice = true
	case *types.Array:
		v.isArray = true
		v.len = u.Len()
	case *types.Pointer:
		v.isPointer = true
	case *types.Named:
		v.u = u
		v.obj = u.Obj()
		v.pkg = u.Obj().Pkg()
		return false
	case *types.Map, *types.Basic, *types.Struct:
		v.u = u
		return false
	}
	return true
}

func (u UnderlyingVisitor) Type() types.Type {
	return u.t
}

func (u UnderlyingVisitor) Underlying() types.Type {
	return u.u
}

func (u UnderlyingVisitor) IsPointer() bool {
	return u.isPointer
}

func (u UnderlyingVisitor) IsSlice() bool {
	return u.isSlice
}

func (u UnderlyingVisitor) IsArray() (bool, int64) {
	return u.isArray, u.len
}

func (u UnderlyingVisitor) Obj() *types.TypeName {
	return u.obj
}

func (u UnderlyingVisitor) Pkg() *types.Package {
	return u.pkg
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

func UnderlyingSignature(T types.Type) string {
	return types.TypeString(NewUnderlyingType(T), nil)
}

func IsUnderlyingIdentical(lhs, rhs types.Type) bool {
	return UnderlyingSignature(lhs) == UnderlyingSignature(rhs)
}

func IsIdentical(lhs, rhs types.Type) bool {
	return types.TypeString(lhs, nil) == types.TypeString(rhs, nil)
}
