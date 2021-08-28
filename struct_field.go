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
	*Type
	*types.Var
}
