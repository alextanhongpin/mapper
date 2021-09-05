package mapper

import "go/types"

// StructField for the example below.
//type Foo struct {
//  Name sql.NullString `json:"name"
//}
type StructField struct {
	Name string `example:"Name"`
	// Useful when the output directory doesn't match the existing ones.
	Pkg      string // e.g. yourpkg
	PkgPath  string // e.g. github.com/your-org/yourpkg
	Exported bool   // e.g. true
	Tag      *Tag   // e.g. `map:"RenameField,CustomFunction"`
	Ordinal  int    // The original position of the struct field.
	Type     types.Type
	//*Type
	*types.Var
}

func NewStructFields(T types.Type) StructFields {
	v := NewStructVisitor()
	_ = Walk(v, T)
	return v.fields
}

type StructVisitor struct {
	fields StructFields
}

func NewStructVisitor() *StructVisitor {
	return &StructVisitor{}
}

func (v *StructVisitor) Visit(T types.Type) bool {
	switch u := T.Underlying().(type) {
	case *types.Struct:
		v.fields = ExtractStructFields(u)
	}
	return true
}
