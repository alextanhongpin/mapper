package mapper

import "go/types"

func UnderlyingString(T types.Type, q types.Qualifier) string {
	u := NewUnderlyingType(T)
	return types.TypeString(u, q)
}
