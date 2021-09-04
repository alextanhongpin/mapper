package internal

import (
	"go/types"
)

func IsStruct(T types.Type) bool {
	_, ok := T.Underlying().(*types.Struct)
	return ok
}
