package mapper

import "go/types"

func NewTypeName(T types.Type) *types.TypeName {
	named, ok := T.(*types.Named)
	if !ok {
		return nil
	}
	return named.Obj()
}
