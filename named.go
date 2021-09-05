package mapper

import (
	"fmt"
	"go/types"
)

func NewTypeName(T types.Type) *types.TypeName {
	named, ok := T.(*types.Named)
	if !ok {
		return nil
	}
	return named.Obj()
}

type NamedVisitor struct {
	methods map[string]*Func
	obj     *types.TypeName
	pkg     *types.Package
}

func NewNamedVisitor(T types.Type) *NamedVisitor {
	v := &NamedVisitor{}
	_ = Walk(v, T)
	return v
}

func (v *NamedVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Named:
		v.methods = newNamedMethods(u)
		v.obj = u.Obj()
		if v.obj != nil {
			v.pkg = v.obj.Pkg()
		}
		return false
	}
	return true
}

func (v NamedVisitor) Methods() map[string]*Func {
	return v.methods
}

func (v NamedVisitor) Obj() *types.TypeName {
	return v.obj
}

func (v NamedVisitor) Pkg() *types.Package {
	return v.pkg
}

func newNamedMethods(T types.Type) map[string]*Func {
	t, ok := T.(*types.Named)
	if !ok {
		panic(fmt.Sprintf("mapper: %s is not a named typed", T))
	}

	result := make(map[string]*Func)
	for i := 0; i < t.NumMethods(); i++ {
		method := t.Method(i)
		if !method.Exported() {
			continue
		}
		fn := NewFunc(method, t.Obj())
		result[fn.Name] = fn
	}

	return result
}
